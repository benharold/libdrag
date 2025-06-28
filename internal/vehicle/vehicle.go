package vehicle

import (
	"context"
	"fmt"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// VehicleInterface defines vehicle monitoring
type VehicleInterface interface {
	component.Component
	GetLane() int
	IsStaged() bool
	GetPosition() float64
}

// SimpleVehicle implements a basic vehicle for testing
type SimpleVehicle struct {
	id       events.ComponentID
	lane     int
	staged   bool
	position float64
	status   component.ComponentStatus
}

func NewSimpleVehicle(lane int) *SimpleVehicle {
	return &SimpleVehicle{
		id:   events.ComponentID(fmt.Sprintf("vehicle_%d", lane)),
		lane: lane,
		status: component.ComponentStatus{
			ID:       events.ComponentID(fmt.Sprintf("vehicle_%d", lane)),
			Status:   "ready",
			Metadata: make(map[string]interface{}),
		},
	}
}

func (sv *SimpleVehicle) GetID() events.ComponentID            { return sv.id }
func (sv *SimpleVehicle) GetLane() int                         { return sv.lane }
func (sv *SimpleVehicle) IsStaged() bool                       { return sv.staged }
func (sv *SimpleVehicle) GetPosition() float64                 { return sv.position }
func (sv *SimpleVehicle) GetStatus() component.ComponentStatus { return sv.status }

func (sv *SimpleVehicle) Initialize(ctx context.Context, bus events.EventBus, config config.Config) error {
	return nil
}

func (sv *SimpleVehicle) Start(ctx context.Context) error {
	sv.status.Status = "running"
	return nil
}

func (sv *SimpleVehicle) Stop() error {
	sv.status.Status = "stopped"
	return nil
}

func (sv *SimpleVehicle) HandleEvent(ctx context.Context, event events.Event) error {
	return nil
}
