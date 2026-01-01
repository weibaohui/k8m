package eventbus

import (
	"sync"

	"k8s.io/klog/v2"
)

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
}

var (
	instance *EventBus
	once     sync.Once
)

func New() *EventBus {
	once.Do(func() {
		instance = &EventBus{
			subscribers: make(map[EventType][]chan Event),
		}
	})
	return instance
}

func (b *EventBus) Subscribe(t EventType) <-chan Event {
	ch := make(chan Event, 1) // 防阻塞
	b.mu.Lock()
	b.subscribers[t] = append(b.subscribers[t], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subscriberCount := len(b.subscribers[e.Type])
	klog.V(6).Infof("Publishing event type=%v to %d subscribers", e.Type, subscriberCount)

	droppedCount := 0
	for _, ch := range b.subscribers[e.Type] {
		select {
		case ch <- e:
		default:
			droppedCount++
			// 丢弃慢消费者
		}
	}

	if droppedCount > 0 {
		klog.V(6).Infof("Event type=%v dropped for %d slow consumers", e.Type, droppedCount)
	}
}
