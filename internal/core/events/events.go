package events

import "time"

type Event struct {
	Type      string
	Payload   any
	Timestamp time.Time
}

type Bus struct {
	subscribers map[string][]func(Event)
}

func NewBus() *Bus {
	return &Bus{subscribers: make(map[string][]func(Event))}
}

func (b *Bus) Subscribe(eventType string, fn func(Event)) {
	b.subscribers[eventType] = append(b.subscribers[eventType], fn)
}

func (b *Bus) Publish(eventType string, payload any) {
	evt := Event{Type: eventType, Payload: payload, Timestamp: time.Now()}
	for _, fn := range b.subscribers[eventType] {
		fn(evt)
	}
}
