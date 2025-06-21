package bus

import (
	"noctua/pkg/logger"
	"reflect"
	"sync"
)

type EventBus struct {
	events chan interface{}
	subs   map[reflect.Type][]chan interface{} // 按事件类型存储订阅通道
	mu     sync.RWMutex
	closed bool
}

func NewEventBus(bufferSize int) *EventBus {
	eb := &EventBus{
		events: make(chan interface{}, bufferSize),
		subs:   make(map[reflect.Type][]chan interface{}),
	}
	go eb.run()
	return eb
}

func (eb *EventBus) Publish(event interface{}) {
	eb.mu.RLock()
	if eb.closed {
		eb.mu.RUnlock()
		return
	}
	select {
	case eb.events <- event:
	default:
		logger.Log.Errorf("EventBus full, dropping event: %T", event)
	}
	eb.mu.RUnlock()
}

// SubscribeToType 订阅特定类型的事件
func (eb *EventBus) SubscribeToType(eventType interface{}, eventSize int) <-chan interface{} {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.closed {
		return nil
	}
	ch := make(chan interface{}, eventSize)
	typ := reflect.TypeOf(eventType)
	eb.subs[typ] = append(eb.subs[typ], ch)
	return ch
}

func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if !eb.closed {
		close(eb.events)
		for _, chans := range eb.subs {
			for _, ch := range chans {
				close(ch)
			}
		}
		eb.closed = true
	}
}

func (eb *EventBus) run() {
	for event := range eb.events {
		eb.mu.RLock()
		typ := reflect.TypeOf(event)
		for subType, chans := range eb.subs {
			if subType == typ { // 只分发匹配类型的事件
				for _, ch := range chans {
					select {
					case ch <- event:
					default:
						logger.Log.Warnf("Subscriber channel full, dropping event: %T", event)
					}
				}
			}
		}
		eb.mu.RUnlock()
	}
}
