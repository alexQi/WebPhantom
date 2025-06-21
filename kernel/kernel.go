package kernel

import (
	"context"
	"noctua/internal/proxy"
	"noctua/internal/scheduler"
	"noctua/kernel/bus"
	"noctua/kernel/session"
	"noctua/pkg/logger"
	"noctua/types"
	"reflect"
	systemRuntime "runtime"
	"sync"
)

type KernelConfig struct {
	ProxyConfig     proxy.ProxyPoolConfig
	SchedulerConfig scheduler.Config
	CrawlerConfig   CrawlerManagerConfig
}

type Kernel struct {
	Ctx             context.Context
	Version         string
	OS              string
	Scheduler       *scheduler.Scheduler
	EventBus        *bus.EventBus
	EventListener   *EventListener
	SessionManager  *session.Manager
	CrawlerManager  *CrawlerManager
	runtimeStarted  bool
	RuntimeChannel  chan types.RuntimeData
	runtimeHandlers struct {
		sync.RWMutex
		handlers []func(types.RuntimeData) // 回调函数列表
	}
}

func NewKernel(ctx context.Context, version string) *Kernel {
	// 迁移表
	MigrateModels()
	// 载入配置
	config := LoadConfig()
	// 处理OS
	var OS string
	if systemRuntime.GOOS == "darwin" {
		OS = "mac"
	} else if systemRuntime.GOOS == "windows" {
		OS = "win"
	}
	// 处理核心属性
	k := &Kernel{
		Ctx:            ctx,
		Version:        version,
		OS:             OS,
		RuntimeChannel: make(chan types.RuntimeData, 5000),
	}
	// 加载事件总线
	k.EventBus = bus.NewEventBus(2000)
	// 加载调度器
	k.Scheduler = scheduler.New(k.Ctx, config.SchedulerConfig)
	// 加载sessionManager
	k.SessionManager = session.NewManager(proxy.NewProxyPool(k.Ctx, config.ProxyConfig))
	// 创建爬虫管理器
	k.CrawlerManager = NewCrawlerManager(k.Scheduler.Context(), config.CrawlerConfig, k.SessionManager, k.EventBus, k.Scheduler, k.RuntimeChannel)
	// 加载Listener
	k.EventListener = NewEventListener(k.Ctx, k.EventBus, k.Scheduler, k.SessionManager, k.RuntimeChannel)
	// 启动listener
	k.EventListener.Start()
	return k
}

// AddRuntimeHandler 添加一个运行时数据处理回调
func (k *Kernel) AddRuntimeHandler(handler func(types.RuntimeData)) {
	k.runtimeHandlers.Lock()
	defer k.runtimeHandlers.Unlock()
	k.runtimeHandlers.handlers = append(k.runtimeHandlers.handlers, handler)
}

func (k *Kernel) RemoveRuntimeHandler(handler func(types.RuntimeData)) {
	k.runtimeHandlers.Lock()
	defer k.runtimeHandlers.Unlock()
	for i, h := range k.runtimeHandlers.handlers {
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
			k.runtimeHandlers.handlers = append(k.runtimeHandlers.handlers[:i], k.runtimeHandlers.handlers[i+1:]...)
			return
		}
	}
}

// ProcessRuntime 处理 RuntimeChannel 的数据并分发给所有注册的回调
func (k *Kernel) ProcessRuntime() {
	k.runtimeHandlers.Lock()
	if k.runtimeStarted {
		k.runtimeHandlers.Unlock()
		return
	}
	k.runtimeStarted = true
	k.runtimeHandlers.Unlock()
	go func() {
		for {
			select {
			case runtimeData := <-k.RuntimeChannel:
				k.runtimeHandlers.RLock()
				for _, handler := range k.runtimeHandlers.handlers {
					if handler != nil {
						func() {
							defer func() {
								if r := recover(); r != nil {
									logger.Log.Errorf("Runtime handler panicked: %v", r)
								}
							}()
							handler(runtimeData)
						}()
					}
				}
				k.runtimeHandlers.RUnlock()
			case <-k.Ctx.Done():
				logger.Log.Info("Runtime channel processing stopped due to context cancellation")
				return
			}
		}
	}()
}

func (k *Kernel) Status() {

}

func (k *Kernel) Stop() {

}
