package events

import (
	"sync"
	"time"
)

// EventType defines the type of events in the system
type EventType string

// Event types
const (
	// EventTreePreStage Tree events
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

	// EventTimingBeamTrigger Timing events
	EventTimingBeamTrigger EventType = "timing.beam_trigger"
	EventTimingReaction    EventType = "timing.reaction"
	EventTiming60Foot      EventType = "timing.60_foot"
	EventTiming330Foot     EventType = "timing.330_foot"
	EventTimingEighthMile  EventType = "timing.eighth_mile"
	EventTimingQuarterMile EventType = "timing.quarter_mile"
	EventTimingTrapSpeed   EventType = "timing.trap_speed"

	// EventAutoStartActivated Auto-start events
	EventAutoStartActivated    EventType = "autostart.activated"
	EventStagingTimeoutFoul    EventType = "autostart.staging_timeout_foul"
	EventTreeSequenceTriggered EventType = "autostart.tree_sequence_triggered"
	EventAutoStartFault        EventType = "autostart.fault"
	EventAutoStartReset        EventType = "autostart.reset"

	// EventRaceStart Race events
	EventRaceStart    EventType = "race.start"
	EventRaceComplete EventType = "race.complete"
	EventRaceAbort    EventType = "race.abort"
	EventRaceFoul     EventType = "race.foul"

	// EventBeamBroken Beam events
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

// subscription holds a handler and its ID for removal
type subscription struct {
	id      int
	handler EventHandler
}

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu          sync.RWMutex
	handlers    map[EventType][]subscription
	allHandlers []subscription // Handlers that receive all events
	asyncMode   bool
	eventQueue  chan Event
	done        chan struct{}
	wg          sync.WaitGroup
	nextID      int
}

// NewEventBus creates a new event bus
func NewEventBus(asyncMode bool) *EventBus {
	eb := &EventBus{
		handlers:    make(map[EventType][]subscription),
		allHandlers: make([]subscription, 0),
		asyncMode:   asyncMode,
		done:        make(chan struct{}),
		nextID:      1,
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

	// Create subscription with unique ID
	sub := subscription{
		id:      eb.nextID,
		handler: handler,
	}
	eb.nextID++

	eb.handlers[eventType] = append(eb.handlers[eventType], sub)

	// Return unsubscribe function that uses the ID
	return func() {
		eb.unsubscribeByID(eventType, sub.id, false)
	}
}

// SubscribeAll adds a handler that receives all events
func (eb *EventBus) SubscribeAll(handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Create subscription with unique ID
	sub := subscription{
		id:      eb.nextID,
		handler: handler,
	}
	eb.nextID++

	eb.allHandlers = append(eb.allHandlers, sub)

	// Return unsubscribe function that uses the ID
	return func() {
		eb.unsubscribeByID("", sub.id, true)
	}
}

// unsubscribeByID removes a subscription by ID
func (eb *EventBus) unsubscribeByID(eventType EventType, id int, allEvents bool) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if allEvents {
		// Remove from all-event handlers
		for i, sub := range eb.allHandlers {
			if sub.id == id {
				eb.allHandlers = append(eb.allHandlers[:i], eb.allHandlers[i+1:]...)
				break
			}
		}
	} else {
		// Remove from specific event handlers
		handlers := eb.handlers[eventType]
		for i, sub := range handlers {
			if sub.id == id {
				eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
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
	handlers := make([]subscription, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	allHandlers := make([]subscription, len(eb.allHandlers))
	copy(allHandlers, eb.allHandlers)
	eb.mu.RUnlock()

	// Deliver to specific handlers
	for _, sub := range handlers {
		sub.handler(event)
	}

	// Deliver to all-event handlers
	for _, sub := range allHandlers {
		sub.handler(event)
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

	eb.handlers = make(map[EventType][]subscription)
	eb.allHandlers = make([]subscription, 0)
	eb.nextID = 1
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
