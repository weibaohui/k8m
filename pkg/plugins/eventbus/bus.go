package eventbus

import "sync"

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

	for _, ch := range b.subscribers[e.Type] {
		select {
		case ch <- e:
		default:
			// 丢弃慢消费者
		}
	}
}
