package proxy

import (
	"context"
	"fmt"
	"noctua/pkg/logger"
	"noctua/pkg/utils/math"
	"noctua/types"
	"strings"
	"sync"
	"time"
)

const ProxyTimeout = 10 * time.Second
const ProxyChannelCap = 100
const CheckWorkerInterval = 10 * time.Second

// ProxyPoolConfig 配置
type ProxyPoolConfig struct {
	MinDynamic     int
	MinStatic      int
	DynamicEnabled bool
	StaticEnabled  bool
}

// channel 代理通道
type channel struct {
	mu    sync.Mutex
	queue chan *types.ProxyInfo
}

// ProxyPoolStatus 定义代理池的状态信息
type ProxyPoolStatus struct {
	DynamicEnabled    bool                      `json:"dynamicEnabled"`    // 是否启用动态代理
	StaticEnabled     bool                      `json:"staticEnabled"`     // 是否启用静态代理
	TotalProxies      int                       `json:"totalProxies"`      // 所有代理总数
	InUseProxies      int                       `json:"inUseProxies"`      // 使用中的代理数
	DynamicProxies    map[string]*ChannelStatus `json:"dynamicProxies"`    // 动态代理通道状态
	StaticProxies     map[string]*ChannelStatus `json:"staticProxies"`     // 静态代理通道状态
	EventQueueLength  int                       `json:"eventQueueLength"`  // 事件队列长度
	EnsureQueueLength int                       `json:"ensureQueueLength"` // 补充队列长度
}

// ChannelStatus 定义通道状态
type ChannelStatus struct {
	Length   int       `json:"length"`   // 当前通道中的代理数
	Capacity int       `json:"capacity"` // 通道容量
	LastUsed time.Time `json:"lastUsed"` // 最后使用时间（可选）
}

// ProxyPool 代理池
type ProxyPool struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	config     ProxyPoolConfig
	proxies    *sync.Map               // 所有代理
	inUse      *sync.Map               // 使用中的代理
	dynamic    map[string]*channel     // 动态代理通道
	static     map[string]*channel     // 静态代理通道
	ensureChan chan types.ProxyRequest // 代理补充请求
}

// NewProxyPool 初始化代理池
func NewProxyPool(parentCtx context.Context, config ProxyPoolConfig) *ProxyPool {
	if config.MinStatic == 0 {
		config.MinStatic = 1
	}
	if config.MinDynamic == 0 {
		config.MinDynamic = 1
	}
	ctx, cancel := context.WithCancel(parentCtx)
	pool := &ProxyPool{
		ctx:        ctx,
		cancel:     cancel,
		config:     config,
		proxies:    &sync.Map{},
		inUse:      &sync.Map{},
		dynamic:    make(map[string]*channel),
		static:     make(map[string]*channel),
		ensureChan: make(chan types.ProxyRequest, 20),
	}
	pool.wg.Add(2)

	go pool.ensureWorker()
	go pool.checkWorker()
	return pool
}

// ensureWorker 代理补充协程
func (p *ProxyPool) ensureWorker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			logger.Log.Info("Ensure worker stopped")
			return
		case _ = <-p.ensureChan:
		}
	}
}

// addToChannel 将代理添加到对应通道
func (p *ProxyPool) addToChannel(proxy *types.ProxyInfo) {
	region := p.getRegionFromKey(proxy.ProxyKey)
	if region == "" {
		logger.Log.Warnf("Invalid region for proxy %s", proxy.ProxyKey)
		return
	}

	ch := p.getChannel(proxy.ProxyType, region)
	if ch == nil {
		return
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()
	select {
	case ch.queue <- proxy:
	default:
		logger.Log.Warnf("Channel full for %s in %s, dropping %s", proxy.ProxyType, region, proxy.ProxyKey)
	}
}

// getChannel 获取或创建通道
func (p *ProxyPool) getChannel(proxyType, region string) *channel {
	var m map[string]*channel
	if proxyType == "dynamic" && p.config.DynamicEnabled {
		m = p.dynamic
	} else if proxyType == "static" && p.config.StaticEnabled {
		m = p.static
	} else {
		return nil
	}

	ch, exists := m[region]
	if !exists {
		ch = &channel{
			queue: make(chan *types.ProxyInfo, ProxyChannelCap),
		}
		m[region] = ch
	}
	return ch
}

// checkWorker 定期检查代理状态
func (p *ProxyPool) checkWorker() {
	defer p.wg.Done()
	ticker := time.NewTicker(CheckWorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			logger.Log.Info("Check worker stopped")
			return
		case <-ticker.C:
			p.checkProxies()
		}
	}
}

// checkProxies 检查并移除过期或无效代理
func (p *ProxyPool) checkProxies() {
	p.proxies.Range(func(key, value interface{}) bool {
		proxy := value.(*types.ProxyInfo)
		if !proxy.GetExpireTime().After(time.Now()) || !proxy.Useable {
			p.RemoveProxy(proxy)
		}
		return true
	})
}

// GetAvailableProxy 获取可用代理
func (p *ProxyPool) GetAvailableProxy(req types.ProxyRequest) (*types.ProxyInfo, error) {
	if req.Region == "" {
		req.Region = "000000"
	}

	ch := p.getChannel(req.Type, req.Region)
	if ch == nil {
		return &types.ProxyInfo{}, nil
	}
	// 触发补充（异步）
	if req.Num > 0 || len(ch.queue) < p.getMinProxyCount(req.Type) {
		reqData := types.ProxyRequest{
			Num:    math.Max(req.Num, p.getMinProxyCount(req.Type)),
			Type:   req.Type,
			Region: req.Region,
		}
		select {
		case <-p.ctx.Done():
			return nil, p.ctx.Err()
		case p.ensureChan <- reqData:
		default:
			logger.Log.Warnf("Ensure channel full, dropping request for %d %s proxies in %s", reqData.Num, reqData.Type, reqData.Region)
		}
	}
	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case <-time.After(ProxyTimeout):
		return nil, fmt.Errorf("timeout waiting for %s proxy in %s after %v", req.Type, req.Region, ProxyTimeout)
	case proxy := <-ch.queue:
		if p.isProxyAvailable(proxy) {
			p.inUse.Store(proxy.ProxyKey, proxy)
			return proxy, nil
		}
		p.RemoveProxy(proxy)
		return p.GetAvailableProxy(req) // 递归获取下一个
	}
}

// isProxyAvailable 检查代理是否可用
func (p *ProxyPool) isProxyAvailable(proxy *types.ProxyInfo) bool {
	_, inUse := p.inUse.Load(proxy.ProxyKey)
	if inUse || !proxy.Useable || !proxy.GetExpireTime().After(time.Now()) {
		return false
	}
	valid, err := isValidProxy(proxy)
	return err == nil && valid
}

// ReleaseProxy 释放代理
func (p *ProxyPool) ReleaseProxy(proxy *types.ProxyInfo) {
	if !proxy.Useable || proxy.GetExpireTime().Before(time.Now()) {
		p.RemoveProxy(proxy)
		return
	}
	p.inUse.Delete(proxy.ProxyKey)
	p.addToChannel(proxy)
}

// RemoveProxy 移除代理
func (p *ProxyPool) RemoveProxy(proxy *types.ProxyInfo) {
	p.proxies.Delete(proxy.ProxyKey)
	p.inUse.Delete(proxy.ProxyKey)
}

// Stop 停止代理池
func (p *ProxyPool) Stop() {
	p.cancel()
	p.wg.Wait()
	for _, ch := range p.dynamic {
		close(ch.queue)
	}
	for _, ch := range p.static {
		close(ch.queue)
	}
	close(p.ensureChan)
}

// getMinProxyCount 获取最小代理数量
func (p *ProxyPool) getMinProxyCount(proxyType string) int {
	if proxyType == "dynamic" {
		return p.config.MinDynamic
	}
	return p.config.MinStatic
}

// getRegionFromKey 从 ProxyKey 中提取区域
func (p *ProxyPool) getRegionFromKey(proxyKey string) string {
	parts := strings.Split(proxyKey, "_")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "000000" // 默认区域
}

// Status 返回代理池状态
func (p *ProxyPool) Status() *ProxyPoolStatus {
	totalProxies := 0
	p.proxies.Range(func(_, _ interface{}) bool {
		totalProxies++
		return true
	})

	inUseProxies := 0
	p.inUse.Range(func(_, _ interface{}) bool {
		inUseProxies++
		return true
	})

	dynamicProxies := make(map[string]*ChannelStatus)
	for region, ch := range p.dynamic {
		ch.mu.Lock()
		dynamicProxies[region] = &ChannelStatus{
			Length:   len(ch.queue),
			Capacity: cap(ch.queue),
			// LastUsed 未实现，可选
		}
		ch.mu.Unlock()
	}

	staticProxies := make(map[string]*ChannelStatus)
	for region, ch := range p.static {
		ch.mu.Lock()
		staticProxies[region] = &ChannelStatus{
			Length:   len(ch.queue),
			Capacity: cap(ch.queue),
			// LastUsed 未实现，可选
		}
		ch.mu.Unlock()
	}

	return &ProxyPoolStatus{
		DynamicEnabled:    p.config.DynamicEnabled,
		StaticEnabled:     p.config.StaticEnabled,
		TotalProxies:      totalProxies,
		InUseProxies:      inUseProxies,
		DynamicProxies:    dynamicProxies,
		StaticProxies:     staticProxies,
		EnsureQueueLength: len(p.ensureChan),
	}
}
