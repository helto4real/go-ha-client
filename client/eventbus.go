package client

import "sync"

// eventBus is a simple event bus implementation to handle subscriptions and event publishing.
type eventBus[T any] struct {
	subscribers map[string][]chan T
	lock        sync.RWMutex
}

// NewEventBus creates a new EventBus instance.
func NewEventBus[T any]() *eventBus[T] {
	return &eventBus[T]{
		subscribers: make(map[string][]chan T, 10),
	}
}

// Subscribe adds a new subscriber to a topic.
func (eb *eventBus[T]) Subscribe(topic string) chan T {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	ch := make(chan T, 10) // Channel is created here and returned

	if prev, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(prev, ch)
	} else {
		eb.subscribers[topic] = []chan T{ch}
	}
    return ch
}

// Publish sends an event to all subscribers of a topic.
func (eb *eventBus[T]) Publish(topic string, event T) {
	eb.lock.RLock()
	defer eb.lock.RUnlock()

	channels, found := eb.subscribers[topic]
	if !found {
		return // No subscribers for the topic
	}

	for _, ch := range channels {
		ch <- event // Send event to subscriber
	}
}
