package douyin

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"noctua/internal/signer"
	"noctua/pkg/engine"
	"noctua/pkg/httpx"
	"noctua/pkg/logger"
	"noctua/types"
	"strconv"
	"time"
)

const MaxRetries = 5

// DouYinApiClient 负责抖音 API 请求
type DouYinApiClient struct {
	userAgent      string
	verifyParams   VerifyParams
	signClient     *signer.SignServerClient
	currentSession *types.Session
	missingSession func()
	discardSession func(*types.Session)
	acquireSession func(*types.Session) (*types.Session, error)
	refreshSession func(*types.Session) (*types.Session, error)
}

// NewDouYinApiClient 创建 DouYinApiClient
func NewDouYinApiClient(signClient *signer.SignServerClient) *DouYinApiClient {
	return &DouYinApiClient{
		signClient: signClient,
		userAgent:  DOUYIN_FIXED_USER_AGENT,
	}
}

func (c *DouYinApiClient) OnAcquireSession(fn func(session *types.Session) (*types.Session, error)) {
	c.acquireSession = fn
}

func (c *DouYinApiClient) OnRefreshSession(fn func(session *types.Session) (*types.Session, error)) {
	c.refreshSession = fn
}

func (c *DouYinApiClient) OnDiscardSession(fn func(currentSession *types.Session)) {
	c.discardSession = fn
}

func (c *DouYinApiClient) OnMissingSession(fn func()) {
	c.missingSession = fn
}

func (c *DouYinApiClient) withSession() error {
	newSession, err := c.acquireSession(c.currentSession)
	if err != nil {
		return err
	}
	if c.currentSession == nil {
		c.BuildVerifyParams(newSession.Account.UserAgent)
	} else {
		if c.currentSession.Account.UID != newSession.Account.UID {
			c.BuildVerifyParams(newSession.Account.UserAgent)
		}
	}

	c.currentSession = newSession
	return nil
}

func (c *DouYinApiClient) CurrentSession() *types.Session {
	return c.currentSession
}

// getHeaders 返回 HTTP 头部信息
func (c *DouYinApiClient) getHeaders() (map[string]string, error) {
	err := c.withSession()
	if err != nil {
		return nil, err
	}
	// 预处理session
	if c.currentSession == nil {
		// 发送事件
		c.missingSession()
		// 返回响应
		return nil, fmt.Errorf("Build fetch header failed, session is nil")
	}
	cookieString, err := JsonToCookieString(c.currentSession.Account.Cookie)
	if err != nil {
		return nil, err
	}
	// 返回请求头
	return map[string]string{
		"Content-Type":    "application/json;charset=UTF-8",
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Origin":          "https://www.douyin.com",
		"Referer":         "https://www.douyin.com/",
		"User-Agent":      c.currentSession.Account.UserAgent,
		"Cookie":          cookieString,
	}, nil
}

func (c *DouYinApiClient) BuildVerifyParams(currentUserAgent string) {
	userAgent := c.userAgent
	if len(currentUserAgent) > 0 {
		userAgent = currentUserAgent
	}
	// 处理verifyParams
	verifyParams, err := BuildVerifyParams(userAgent)
	if err != nil {
		logger.Log.Errorf("build verify params err: %v", err)
	}
	c.verifyParams = verifyParams
}

// getCommonParams 获取通用参数
func (c *DouYinApiClient) BuildCommonParams() map[string]string {
	deviceInfo := &engine.DeviceInfo{
		CookieEnabled:   true,
		BrowserOnline:   true,
		ScreenHeight:    900,
		ScreenWidth:     1600,
		CpuCoreNum:      8,
		DeviceMemory:    8,
		Downlink:        4.45,
		RoundTripTime:   100,
		BrowserLanguage: "zh-CN",
		BrowserPlatform: "MacIntel",
		BrowserName:     "Chrome",
		BrowserVersion:  "134.0.0.0",
		EngineName:      "Blink",
		EngineVersion:   "134.0.0.0",
		OsName:          "Mac+OS",
		OsVersion:       "10.15.7",
		EffectiveType:   "4g",
	}
	if c.currentSession != nil {
		err := json.Unmarshal([]byte(c.currentSession.Account.DeviceInfo), deviceInfo)
		if err != nil {
			logger.Log.Warnf("get device info failed, use default info")
		}
	}
	return map[string]string{
		"device_platform":             "webapp",
		"aid":                         "6383",
		"channel":                     "channel_pc_web",
		"pc_client_type":              "1",
		"publish_video_strategy_type": "2",
		"version_code":                "170400",
		"version_name":                "17.4.0",
		"platform":                    "PC",
		"browser_language":            deviceInfo.BrowserLanguage,
		"browser_platform":            deviceInfo.BrowserPlatform,
		"browser_name":                deviceInfo.BrowserName,
		"browser_version":             deviceInfo.BrowserVersion,
		"browser_online":              strconv.FormatBool(deviceInfo.BrowserOnline),
		"engine_name":                 deviceInfo.EngineName,
		"engine_version":              deviceInfo.EngineVersion,
		"os_name":                     deviceInfo.OsName,
		"os_version":                  deviceInfo.OsVersion,
		"effective_type":              deviceInfo.EffectiveType,
		"webid":                       c.verifyParams.WebID,
		"msToken":                     c.verifyParams.MsToken,
		"cookie_enabled":              strconv.FormatBool(deviceInfo.CookieEnabled),
		"screen_width":                strconv.Itoa(deviceInfo.ScreenWidth),
		"screen_height":               strconv.Itoa(deviceInfo.ScreenHeight),
		"cpu_core_num":                strconv.Itoa(deviceInfo.CpuCoreNum),
		"device_memory":               strconv.FormatFloat(deviceInfo.DeviceMemory, 'f', -1, 64),
		"downlink":                    strconv.FormatFloat(deviceInfo.Downlink, 'f', -1, 64),
		"round_trip_time":             strconv.Itoa(deviceInfo.RoundTripTime),
	}
}

// processQueryParams 预处理 URL 参数，获取 `a_bogus` 签名
func (c *DouYinApiClient) processQueryParams(uri string, params map[string]string, needSign bool) (map[string]string, error) {
	finalParams := c.BuildCommonParams()
	if params != nil {
		for k, v := range params {
			finalParams[k] = v
		}
	}
	if needSign {
		signParams := httpx.EncodeURLParams(finalParams)
		// 判断agent
		userAgent := c.userAgent
		if c.currentSession != nil && c.currentSession.Account.UserAgent != "" {
			userAgent = c.currentSession.Account.UserAgent
		}
		signResp, err := c.signClient.DouyinSign(&signer.DouyinSignRequest{
			URI:         uri,
			QueryParams: signParams,
			UserAgent:   userAgent,
		})
		if err != nil {
			return nil, err
		}
		finalParams["a_bogus"] = signResp.Data.ABogus
	}
	return finalParams, nil
}

// sendRequest 统一发送请求
func (c *DouYinApiClient) fetch(uri string, params map[string]string, requestParams *CallRequestParams, retryCount ...int) ([]byte, error) {
	currentRetry := 0
	if len(retryCount) > 0 {
		currentRetry = retryCount[0]
	}

	if currentRetry >= MaxRetries {
		return nil, fmt.Errorf("Max retry limit (%d) reached for uri: %s", MaxRetries, uri)
	}
	// 处理header
	var headers map[string]string
	// header overwrite
	if requestParams.Headers != nil {
		headers = requestParams.Headers
	} else {
		defaultHeaders, err := c.getHeaders()
		if err != nil {
			return nil, err
		}
		headers = defaultHeaders
	}
	// 处理请求参数
	processedParams, err := c.processQueryParams(uri, params, requestParams.NeedSign)
	if err != nil {
		return nil, err
	}
	client := resty.New().
		//SetDebug(true).
		SetBaseURL(DOUYIN_API_URL).
		SetAuthScheme("").
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).   // 初次重试等待 2s
		SetRetryMaxWaitTime(5 * time.Second) // 最大等待时间
	// 判断session是否正常
	if c.currentSession == nil {
		// 发送事件
		c.missingSession()
		// 返回响应
		return nil, fmt.Errorf("fetch failed, session is nil")
	} else {
		//初次使用代理
		if c.currentSession.ProxyInfo.Useable {
			client.SetProxy(c.currentSession.ProxyInfo.BuildProtocol())
		}
	}
	// 增加重试条件
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if r.StatusCode() == 200 {
			return false
		}
		newSession, err := c.refreshSession(c.currentSession)
		if err != nil {
			return false
		}
		c.currentSession = newSession
		if c.currentSession.ProxyInfo.Useable {
			client.SetProxy(c.currentSession.ProxyInfo.BuildProtocol())
		}
		return true
	})
	excutor := client.R()
	for headerKey, headerVal := range headers {
		excutor.SetHeader(headerKey, headerVal)
	}
	redoFetch := false
	resp, err := excutor.SetQueryParams(processedParams).Get(uri)
	if err != nil {
		// 自动重试失败后改为手动重试
		if resp != nil && resp.Request.Attempt > 1 {
			logger.Log.Errorf("Failed to retry for %d times, err：%s", resp.Request.Attempt-1, err.Error())
			redoFetch = true
		} else {
			return nil, err
		}
	}
	respBody := resp.Body()
	// 如果相应为空需要更换账号
	if !redoFetch && len(string(respBody)) == 0 || string(respBody) == "blocked" {
		logger.Log.Errorf("Account may blocked. try again later to confirm: %s", c.currentSession.Account.UserID)
		redoFetch = true
	}
	// redoFetch重新更换账号
	if redoFetch {
		if c.discardSession != nil && currentRetry > 2 {
			c.discardSession(c.currentSession)
			c.currentSession = nil
		}
		return c.fetch(uri, params, requestParams, currentRetry+1)
	}
	return respBody, err
}

// Pong 获取用户信息
func (c *DouYinApiClient) Pong() (*PongResp, error) {
	selfInfo := &PongResp{}
	client := resty.New().
		SetBaseURL(DOUYIN_INDEX_URL).
		SetAuthScheme("")
	if c.currentSession.ProxyInfo.Useable == true {
		client.SetProxy(c.currentSession.ProxyInfo.BuildProtocol())
	}
	uri := "/aweme/v1/web/query/user/"
	queryParams, err := c.processQueryParams(uri, nil, false)
	if err != nil {
		return nil, err
	}
	cookieString, err := JsonToCookieString(c.currentSession.Account.Cookie)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":    "application/json;charset=UTF-8",
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Origin":          "https://www.douyin.com",
		"Referer":         "https://www.douyin.com/",
		"User-Agent":      c.currentSession.Account.UserAgent,
		"Cookie":          cookieString,
	}
	_, err = client.R().SetHeaders(headers).SetResult(selfInfo).SetQueryParams(queryParams).Get(uri)
	if err != nil {
		return nil, err
	}
	return selfInfo, err
}
