package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SimpleEventBus implements EventBus with in-memory message passing
type SimpleEventBus struct {
	mu           sync.RWMutex
	subscribers  map[EventType][]subscription
	allSubs      []subscription
	running      bool
	eventChannel chan Event
	ctx          context.Context
	cancel       context.CancelFunc
}

type subscription struct {
	id      string
	handler EventHandler
	active  bool
}

func (s *subscription) Unsubscribe() error {
	s.active = false
	return nil
}

func (s *subscription) IsActive() bool {
	return s.active
}

// NewSimpleEventBus creates a new event bus
func NewSimpleEventBus() *SimpleEventBus {
	return &SimpleEventBus{
		subscribers:  make(map[EventType][]subscription),
		allSubs:      make([]subscription, 0),
		eventChannel: make(chan Event, 1000), // Buffered channel
	}
}

func (bus *SimpleEventBus) Publish(ctx context.Context, event Event) error {
	select {
	case bus.eventChannel <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event bus full, dropping event: %s", event.GetType())
	}
}

func (bus *SimpleEventBus) Subscribe(eventType EventType, handler EventHandler) Subscription {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	sub := subscription{
		id:      fmt.Sprintf("%s_%d", eventType, time.Now().UnixNano()),
		handler: handler,
		active:  true,
	}

	bus.subscribers[eventType] = append(bus.subscribers[eventType], sub)
	return &sub
}

func (bus *SimpleEventBus) SubscribeAll(handler EventHandler) Subscription {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	sub := subscription{
		id:      fmt.Sprintf("all_%d", time.Now().UnixNano()),
		handler: handler,
		active:  true,
	}

	bus.allSubs = append(bus.allSubs, sub)
	return &sub
}

func (bus *SimpleEventBus) Start(ctx context.Context) error {
	bus.ctx, bus.cancel = context.WithCancel(ctx)
	bus.running = true

	go bus.eventLoop()
	return nil
}

func (bus *SimpleEventBus) Stop() error {
	if bus.cancel != nil {
		bus.cancel()
	}
	bus.running = false
	return nil
}

func (bus *SimpleEventBus) eventLoop() {
	for {
		select {
		case event := <-bus.eventChannel:
			bus.deliverEvent(event)
		case <-bus.ctx.Done():
			return
		}
	}
}

func (bus *SimpleEventBus) deliverEvent(event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	// Deliver to specific event type subscribers
	if subs, exists := bus.subscribers[event.GetType()]; exists {
		for _, sub := range subs {
			if sub.active {
				go func(s subscription) {
					if err := s.handler(bus.ctx, event); err != nil {
						fmt.Printf("⚠️  Error handling event %s: %v\n", event.GetType(), err)
					}
				}(sub)
			}
		}
	}

	// Deliver to "all events" subscribers
	for _, sub := range bus.allSubs {
		if sub.active {
			go func(s subscription) {
				if err := s.handler(bus.ctx, event); err != nil {
					fmt.Printf("⚠️  Error handling event %s: %v\n", event.GetType(), err)
				}
			}(sub)
		}
	}
}
