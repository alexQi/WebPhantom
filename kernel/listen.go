package kernel

import (
	"context"
	"fmt"
	"noctua/internal/scheduler"
	"noctua/kernel/bus"
	"noctua/kernel/session"
	"noctua/pkg/logger"
	"noctua/types"
)

type EventListener struct {
	eventBus       *bus.EventBus
	scheduler      *scheduler.Scheduler
	sm             *session.Manager
	mainCtx        context.Context
	ctx            context.Context    // 添加上下文用于控制关闭
	cancel         context.CancelFunc // 取消函数
	runtimeChannel chan types.RuntimeData
}

func NewEventListener(
	mainCtx context.Context,
	eventBus *bus.EventBus,
	scheduler *scheduler.Scheduler,
	sm *session.Manager,
	runtimeChannel chan types.RuntimeData,
) *EventListener {
	ctx, cancel := context.WithCancel(mainCtx) // 创建可取消的上下文
	return &EventListener{
		sm:             sm,
		ctx:            ctx,
		mainCtx:        mainCtx,
		cancel:         cancel,
		eventBus:       eventBus,
		scheduler:      scheduler,
		runtimeChannel: runtimeChannel,
	}
}

// Start
func (k *EventListener) Start() {
	// 检查是否已在运行
	if k.ctx.Err() != nil {
		k.ctx, k.cancel = context.WithCancel(k.mainCtx)
	}
	// 订阅采集开始事件
	k.ListenCrawlStart()
	// 订阅采集停止事件
	k.ListenCrawlEnd()
}

// Stop 停止监听并清理资源
func (k *EventListener) Stop() {
	k.cancel()         // 触发上下文取消
	k.eventBus.Close() // 关闭事件总线，停止事件分发
}

func (k *EventListener) listenEvent(eventType interface{}, capacity int, handler func(event interface{})) {
	sub := k.eventBus.SubscribeToType(eventType, capacity)
	go func() {
		eventName := fmt.Sprintf("%T", eventType)
		for {
			select {
			case event, ok := <-sub:
				if !ok {
					logger.Log.Infof("%s subscription closed", eventName)
					return
				}
				handler(event)
			case <-k.ctx.Done():
				logger.Log.Infof("%s listener stopped", eventName)
				return
			}
		}
	}()
}

// ListenCrawlStart
func (k *EventListener) ListenCrawlStart() {
	k.listenEvent(types.CrawlStartEvent{}, 100, func(event interface{}) {
		logger.Log.Infof("Crawl started ....")
	})
}

// ListenCrawlEnd
func (k *EventListener) ListenCrawlEnd() {
	k.listenEvent(types.CrawlEndEvent{}, 100, func(event interface{}) {
		k.scheduler.Shutdown()
	})
}
