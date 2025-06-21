package session

import (
	"fmt"
	"noctua/internal/model"
	"noctua/internal/proxy"
	"noctua/pkg/cache"
	"noctua/pkg/logger"
	"noctua/pkg/utils/str"
	"noctua/types"
	"strings"
	"sync"
	"time"
)

// SessionManagerStatus 定义会话管理的状态信息
type SessionManagerStatus struct {
	ActiveSessions int                    `json:"activeSessions"`
	TotalSessions  int                    `json:"totalSessions"`
	MediaDetails   map[string]MediaStatus `json:"mediaDetails"`
	ProxyStatus    *proxy.ProxyPoolStatus `json:"proxyStatus"`
	UserProxyPairs int                    `json:"userProxyPairs"`
}

// MediaStatus 定义每个媒体的状态
type MediaStatus struct {
	SessionCount int `json:"sessionCount"`
	ActiveCount  int `json:"activeCount"`
}

type SessionParams struct {
	MediaCode        string
	SessionType      string
	SessionRegion    string
	UserID           string
	KeepAlive        bool
	AllowNoneAccount bool
	AccountType      int
	ExcludeUserIdMap map[string]int
}

// Manager 管理账号的 Cookie 和代理
type Manager struct {
	mu           sync.Mutex
	mediaAccount *model.MediaAccount
	proxyPool    *proxy.ProxyPool
	sessionMap   map[string]*sync.Map // 修改为 map[string]*sync.Map
	userProxyMap sync.Map
}

// NewManager 创建 Manager
func NewManager(proxyPool *proxy.ProxyPool) *Manager {
	return &Manager{
		proxyPool:    proxyPool,
		mediaAccount: &model.MediaAccount{},
		sessionMap:   make(map[string]*sync.Map),
		userProxyMap: sync.Map{},
	}
}

func (sm *Manager) LoadMediaProxyCache() {
	cacheKeys := cache.CacheManager.Keys("media:proxy:*")
	for _, cacheKey := range cacheKeys {
		if cacheProxyString, ok := cache.CacheManager.Get(cacheKey); ok {
			uid, _ := strings.CutPrefix(cacheKey, "media:proxy:")
			sm.userProxyMap.Store(uid, cacheProxyString.(string))
		} else {
			logger.Log.Errorf("load media proxy cache key %s failed", cacheKey)
		}
	}
}

// SetMediaAccount 设置媒体账号
func (sm *Manager) SetMediaAccount(account *model.MediaAccount, isInUse bool, proxyKey string, expireTime time.Time) error {
	mediaCode := account.MediaCode
	if account.Status == 100 {
		if sessions, ok := sm.sessionMap[mediaCode]; ok {
			if session, ok := sessions.Load(account.UserID); ok {
				sm.proxyPool.ReleaseProxy(session.(*types.Session).ProxyInfo)
				sessions.Delete(account.UserID)
			}
		}
	} else {
		// 缓存代理关系
		if isInUse {
			err := cache.CacheManager.Set("media:proxy:"+account.UserID, proxyKey, expireTime.Sub(time.Now()))
			if err != nil {
				logger.Log.Errorf("set proxy cache failed: %v", err)
			}
		}
	}
	return nil
}

func (sm *Manager) RenewSession(params *SessionParams) (*types.Session, error) {
	// 未指定用户id
	var account *model.MediaAccount
	var err error
	if len(params.UserID) == 0 {
		account = &model.MediaAccount{
			UserID:    str.GenerateRandString(64),
			MediaCode: params.MediaCode,
			Type:      params.AccountType,
		}
	} else {
		// 查询数据库
		account, err = sm.mediaAccount.FindMediaAccount(&model.QueryMediaAccountParams{
			MediaCode: params.MediaCode,
			Type:      params.AccountType,
			UserID:    params.UserID,
		})
		if err != nil {
			account = &model.MediaAccount{
				UserID:    str.GenerateRandString(64),
				MediaCode: params.MediaCode,
				Type:      params.AccountType,
			}
		}
	}

	expireTime := time.Now().Add(24 * 7 * time.Hour)
	return &types.Session{
		Enabled:    true,
		Account:    account,
		ProxyInfo:  &types.ProxyInfo{},
		ExpireTime: expireTime,
	}, nil
}

// GetSession 获取会话
func (sm *Manager) GetSession(params *SessionParams) (*types.Session, error) {
	// 初始化 mediaCode 对应的 sync.Map
	sm.mu.Lock()
	if _, ok := sm.sessionMap[params.MediaCode]; !ok {
		sm.sessionMap[params.MediaCode] = &sync.Map{}
	}
	sessions := sm.sessionMap[params.MediaCode]
	sm.mu.Unlock()

	// 检查现有 session
	excludeUserIds := make([]string, 0, len(params.ExcludeUserIdMap))
	for id := range params.ExcludeUserIdMap {
		excludeUserIds = append(excludeUserIds, id)
	}

	var foundSession *types.Session
	sessions.Range(func(key, value interface{}) bool {
		session := value.(*types.Session)
		if params.AccountType != session.Account.Type {
			return true
		}
		if _, ok := params.ExcludeUserIdMap[session.Account.UserID]; ok {
			return true
		}
		if session.InUsed {
			excludeUserIds = append(excludeUserIds, session.Account.UserID)
			return true
		}
		if time.Now().After(session.ExpireTime) {
			sessions.Delete(key)
			return true
		}
		if len(params.UserID) == 0 || params.UserID == session.Account.UserID {
			if params.KeepAlive {
				session.InUsed = true
				sessions.Store(key, session) // 在 Range 内更新
			}
			foundSession = session
			return false // 停止遍历
		}
		excludeUserIds = append(excludeUserIds, session.Account.UserID)
		return true
	})

	if foundSession != nil {
		return foundSession, nil
	}

	// 查询数据库
	account, err := sm.mediaAccount.FindMediaAccount(&model.QueryMediaAccountParams{
		MediaCode:      params.MediaCode,
		Type:           params.AccountType,
		UserID:         params.UserID,
		ExcludeUserIDs: excludeUserIds,
	})
	if err != nil {
		return nil, fmt.Errorf("query available account failed: %v", err)
	}
	if account.ID == 0 {
		if !params.AllowNoneAccount {
			return nil, fmt.Errorf("no available account for media: %s", params.MediaCode)
		}
		account = &model.MediaAccount{
			UserID:    str.GenerateRandString(64),
			MediaCode: params.MediaCode,
			Type:      params.AccountType,
		}
		logger.Log.Infof("Generated temp account %s for media %s", account.UserID, params.MediaCode)
	}

	// 获取代理
	proxyReq := types.ProxyRequest{
		Num:    1,
		Type:   "dynamic",
		Region: params.SessionRegion,
	}
	if params.AccountType > 1 {
		proxyReq.Type = "static"
		if proxyKey, ok := sm.userProxyMap.Load(account.UserID); ok {
			proxyReq.ProxyKey = proxyKey.(string)
		}
	}
	proxyInfo, err := sm.proxyPool.GetAvailableProxy(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("get proxy failed: %v", err)
	}

	expireTime := time.Now().Add(24 * 7 * time.Hour)
	if proxyInfo.Useable {
		expireTime = proxyInfo.GetExpireTime()
	}
	session := &types.Session{
		Enabled:    true,
		InUsed:     params.KeepAlive,
		Account:    account,
		ProxyInfo:  proxyInfo,
		ExpireTime: expireTime,
	}

	sessions.Store(account.UserID, session)

	if session.InUsed && account.IsReal > 0 && proxyInfo.Useable {
		err := cache.CacheManager.Set(
			"media:proxy:"+account.UserID,
			proxyInfo.ProxyKey,
			expireTime.Sub(time.Now()),
		)
		if err != nil {
			logger.Log.Errorf("set proxy cache failed: %v", err)
		}
	}
	return session, nil
}

// ReplaceSession 更换指定会话的代理
func (sm *Manager) ReplaceSession(mediaCode, userID, region string) (*types.Session, error) {
	// 加锁确保线程安全
	sm.mu.Lock()
	// 获取对应媒体的会话集合
	sessions, ok := sm.sessionMap[mediaCode]
	sm.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no sessions found for media: %s", mediaCode)
	}
	// 验证会话是否存在
	current, ok := sessions.Load(userID)
	if !ok {
		return nil, fmt.Errorf("session for user %s in media %s not found or mismatched", userID, mediaCode)
	}
	currentSession := current.(*types.Session)
	// 判断之前是否使用代理,未使用代理时直接返回保持不变
	if currentSession.ProxyInfo == nil || !currentSession.ProxyInfo.Useable {
		return currentSession, nil
	}
	// 移除旧的proxy
	sm.proxyPool.RemoveProxy(currentSession.ProxyInfo)
	// 构造代理请求
	proxyReq := types.ProxyRequest{
		Num:    1,
		Type:   "dynamic", // 默认动态代理
		Region: region,
	}
	// 获取新代理
	newProxyInfo, err := sm.proxyPool.GetAvailableProxy(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get new proxy for user %s: %v", userID, err)
	}
	// 更新会话过期时间
	if newProxyInfo.Useable {
		currentSession.ExpireTime = newProxyInfo.GetExpireTime() // 使用代理的过期时间
	}
	// 更新会话的代理信息
	currentSession.ProxyInfo = newProxyInfo
	sessions.Store(userID, currentSession) // 更新会话

	return currentSession, nil
}

// ReleaseSession 释放会话
func (sm *Manager) ReleaseSession(session *types.Session) {
	if session == nil {
		return
	}
	if sessions, ok := sm.sessionMap[session.Account.MediaCode]; ok {
		if s, ok := sessions.Load(session.Account.UserID); ok {
			s.(*types.Session).InUsed = false
			logger.Log.Infof("Released session for account %s", session.Account.UserID)
		}
	}
}

// InvalidateSession 标记账号失效
func (sm *Manager) InvalidateSession(session *types.Session) error {
	if session == nil || session.Account == nil {
		return nil
	}

	session.Account.Status = 100
	if _, err := session.Account.UpsertModel(); err != nil {
		logger.Log.Errorf("Failed to update account status: %v", err)
		return err
	}

	if sessions, ok := sm.sessionMap[session.Account.MediaCode]; ok {
		if s, ok := sessions.Load(session.Account.UserID); ok {
			sm.proxyPool.RemoveProxy(s.(*types.Session).ProxyInfo)
			sessions.Delete(session.Account.UserID)
		}
	}
	logger.Log.Infof("Invalidated account %s for media %s", session.Account.UserID, session.Account.MediaCode)
	return nil
}

// Status 返回会话管理状态
func (sm *Manager) Status() *SessionManagerStatus {
	totalSessions := 0
	activeSessions := 0
	mediaDetails := make(map[string]MediaStatus)

	for mediaCode, sessions := range sm.sessionMap {
		sessionCount := 0
		activeCount := 0
		sessions.Range(func(_, value interface{}) bool {
			session := value.(*types.Session)
			sessionCount++
			if session.InUsed {
				activeCount++
			}
			return true
		})
		totalSessions += sessionCount
		activeSessions += activeCount
		mediaDetails[mediaCode] = MediaStatus{
			SessionCount: sessionCount,
			ActiveCount:  activeCount,
		}
	}

	userProxyPairs := 0
	sm.userProxyMap.Range(func(_, _ interface{}) bool {
		userProxyPairs++
		return true
	})

	return &SessionManagerStatus{
		ActiveSessions: activeSessions,
		TotalSessions:  totalSessions,
		MediaDetails:   mediaDetails,
		UserProxyPairs: userProxyPairs,
		ProxyStatus:    sm.proxyPool.Status(),
	}
}
