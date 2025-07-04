package component

import (
	"context"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// ComponentStatus represents the current state of a component
type ComponentStatus struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"` // ready, running, error, stopped
	LastError error                  `json:"last_error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Component represents any system component
type Component interface {
	GetID() string
	Initialize(ctx context.Context, config config.Config) error
	Arm(ctx context.Context) error
	EmergencyStop() error
	GetStatus() ComponentStatus
}

// EventAwareComponent extends Component with event bus support
type EventAwareComponent interface {
	Component
	SetEventBus(eventBus *events.EventBus)
	SetRaceID(raceID string)
}
