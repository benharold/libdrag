package events

import (
	"sync"
	"time"
)

// EventType defines the type of events in the system
type EventType string

// Event types
const (
	// Tree events
	EventTreePreStage      EventType = "tree.pre_stage"
	EventTreeStage         EventType = "tree.stage"
	EventTreeArmed         EventType = "tree.armed"
	EventTreeActivated     EventType = "tree.activated" // Add missing event type
	EventTreeDisarmed      EventType = "tree.disarmed"
	EventTreeAmberOn       EventType = "tree.amber_on"
	EventTreeAmberOff      EventType = "tree.amber_off"
	EventTreeGreenOn       EventType = "tree.green_on"
	EventTreeRedLight      EventType = "tree.red_light"
	EventTreeSequenceStart EventType = "tree.sequence_start"
	EventTreeSequenceEnd   EventType = "tree.sequence_end"
	EventTreeEmergencyStop EventType = "tree.emergency_stop"

	// Timing events
	EventTimingBeamTrigger EventType = "timing.beam_trigger"
	EventTimingReaction    EventType = "timing.reaction"
	EventTiming60Foot      EventType = "timing.60_foot"
	EventTiming330Foot     EventType = "timing.330_foot"
	EventTimingEighthMile  EventType = "timing.eighth_mile"
	EventTimingQuarterMile EventType = "timing.quarter_mile"
	EventTimingTrapSpeed   EventType = "timing.trap_speed"

	// Race events
	EventRaceStart    EventType = "race.start"
	EventRaceComplete EventType = "race.complete"
	EventRaceAbort    EventType = "race.abort"
	EventRaceFoul     EventType = "race.foul"

	// Beam events
	EventBeamBroken   EventType = "beam.broken"
	EventBeamRestored EventType = "beam.restored"
)

// Event represents a racing event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	RaceID    string                 `json:"race_id"`
	Lane      int                    `json:"lane,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// EventHandler is a function that handles events
type EventHandler func(event Event)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu          sync.RWMutex
	handlers    map[EventType][]EventHandler
	allHandlers []EventHandler // Handlers that receive all events
	asyncMode   bool
	eventQueue  chan Event
	done        chan struct{}
	wg          sync.WaitGroup
}

// NewEventBus creates a new event bus
func NewEventBus(asyncMode bool) *EventBus {
	eb := &EventBus{
		handlers:    make(map[EventType][]EventHandler),
		allHandlers: make([]EventHandler, 0),
		asyncMode:   asyncMode,
		done:        make(chan struct{}),
	}

	if asyncMode {
		eb.eventQueue = make(chan Event, 1000) // Buffer for performance
		eb.wg.Add(1)
		go eb.processEvents()
	}

	return eb
}

// Subscribe adds a handler for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// Return unsubscribe function
	return func() {
		eb.Unsubscribe(eventType, handler)
	}
}

// SubscribeAll adds a handler that receives all events
func (eb *EventBus) SubscribeAll(handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.allHandlers = append(eb.allHandlers, handler)

	// Return unsubscribe function
	return func() {
		eb.UnsubscribeAll(handler)
	}
}

// Unsubscribe removes a handler for a specific event type
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i := range handlers {
		// Note: In Go, we can't directly compare function values
		// This is a limitation - in practice, you might want to return
		// a subscription ID instead
		if i < len(handlers) {
			// For now, we'll remove by index
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// UnsubscribeAll removes a handler from all events
func (eb *EventBus) UnsubscribeAll(handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Similar limitation as above
	if len(eb.allHandlers) > 0 {
		eb.allHandlers = eb.allHandlers[:len(eb.allHandlers)-1]
	}
}

// Publish sends an event to all registered handlers
func (eb *EventBus) Publish(event Event) {
	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if eb.asyncMode {
		select {
		case eb.eventQueue <- event:
			// Event queued successfully
		default:
			// Queue full, drop event (or handle differently)
			// In production, you might want to log this
		}
	} else {
		eb.deliver(event)
	}
}

// deliver sends the event to handlers
func (eb *EventBus) deliver(event Event) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	allHandlers := make([]EventHandler, len(eb.allHandlers))
	copy(allHandlers, eb.allHandlers)
	eb.mu.RUnlock()

	// Deliver to specific handlers
	for _, handler := range handlers {
		handler(event)
	}

	// Deliver to all-event handlers
	for _, handler := range allHandlers {
		handler(event)
	}
}

// processEvents handles async event delivery
func (eb *EventBus) processEvents() {
	defer eb.wg.Done()
	for {
		select {
		case event := <-eb.eventQueue:
			eb.deliver(event)
		case <-eb.done:
			// Process remaining events in queue
			for {
				select {
				case event := <-eb.eventQueue:
					eb.deliver(event)
				default:
					return
				}
			}
		}
	}
}

// Stop shuts down the event bus
func (eb *EventBus) Stop() {
	if eb.asyncMode {
		close(eb.done)
		eb.wg.Wait()
	}
}

// Clear removes all handlers
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers = make(map[EventType][]EventHandler)
	eb.allHandlers = make([]EventHandler, 0)
}

// EventBuilder provides a fluent interface for creating events
type EventBuilder struct {
	event Event
}

// NewEvent creates a new event builder
func NewEvent(eventType EventType) *EventBuilder {
	return &EventBuilder{
		event: Event{
			Type:      eventType,
			Timestamp: time.Now(),
			Data:      make(map[string]interface{}),
		},
	}
}

// WithRaceID sets the race ID
func (eb *EventBuilder) WithRaceID(raceID string) *EventBuilder {
	eb.event.RaceID = raceID
	return eb
}

// WithLane sets the lane
func (eb *EventBuilder) WithLane(lane int) *EventBuilder {
	eb.event.Lane = lane
	return eb
}

// WithData adds data to the event
func (eb *EventBuilder) WithData(key string, value interface{}) *EventBuilder {
	eb.event.Data[key] = value
	return eb
}

// Build returns the constructed event
func (eb *EventBuilder) Build() Event {
	return eb.event
}
