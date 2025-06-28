package events

import (
	"context"
	"time"
)

// EventType defines the types of events in the system
type EventType string

const (
	// Timing Events
	EventBeamTriggered  EventType = "beam_triggered"
	EventBeamReleased   EventType = "beam_released"
	EventTimingComplete EventType = "timing_complete"

	// Christmas Tree Events
	EventPreStageOn        EventType = "pre_stage_on"
	EventStageOn           EventType = "stage_on"
	EventTreeSequenceStart EventType = "tree_sequence_start"
	EventAmberLight        EventType = "amber_light"
	EventGreenLight        EventType = "green_light"
	EventRedLight          EventType = "red_light"

	// Race Events
	EventRaceArmed   EventType = "race_armed"
	EventRaceStarted EventType = "race_started"
	EventRaceAborted EventType = "race_aborted"
	EventRunComplete EventType = "run_complete"
	EventSystemReady EventType = "system_ready"
	EventSystemError EventType = "system_error"

	// Vehicle Events
	EventVehicleEntered EventType = "vehicle_entered"
	EventVehicleStaged  EventType = "vehicle_staged"
)

// ComponentID identifies different system components
type ComponentID string

const (
	ComponentTimingSystem   ComponentID = "timing_system"
	ComponentChristmasTree  ComponentID = "christmas_tree"
	ComponentStarterControl ComponentID = "starter_control"
	ComponentScoreboard     ComponentID = "scoreboard"
	ComponentVehicleLeft    ComponentID = "vehicle_left"
	ComponentVehicleRight   ComponentID = "vehicle_right"
	ComponentOrchestrator   ComponentID = "orchestrator"
)

// Event represents any event in the drag race system
type Event interface {
	GetType() EventType
	GetTimestamp() time.Time
	GetSource() ComponentID
	GetData() map[string]interface{}
}

// BaseEvent implements the Event interface
type BaseEvent struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    ComponentID            `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

func (e *BaseEvent) GetType() EventType              { return e.Type }
func (e *BaseEvent) GetTimestamp() time.Time         { return e.Timestamp }
func (e *BaseEvent) GetSource() ComponentID          { return e.Source }
func (e *BaseEvent) GetData() map[string]interface{} { return e.Data }

// EventHandler processes events
type EventHandler func(ctx context.Context, event Event) error

// Subscription represents an event subscription
type Subscription interface {
	Unsubscribe() error
	IsActive() bool
}

// EventBus handles communication between components
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType EventType, handler EventHandler) Subscription
	SubscribeAll(handler EventHandler) Subscription
	Start(ctx context.Context) error
	Stop() error
}
