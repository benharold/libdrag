package vehicle

import (
	"context"
	"fmt"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
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
	id       string
	lane     int
	position float64
	staged   bool
	status   component.ComponentStatus
}

func NewSimpleVehicle(lane int) *SimpleVehicle {
	return &SimpleVehicle{
		id:       fmt.Sprintf("vehicle_%d", lane),
		lane:     lane,
		position: 0.0,
		staged:   false,
		status: component.ComponentStatus{
			ID:       fmt.Sprintf("vehicle_%d", lane),
			Status:   "ready",
			Metadata: make(map[string]interface{}),
		},
	}
}

func (v *SimpleVehicle) GetID() string {
	return v.id
}

func (v *SimpleVehicle) GetLane() int {
	return v.lane
}

func (v *SimpleVehicle) IsStaged() bool {
	return v.staged
}

func (v *SimpleVehicle) GetPosition() float64 {
	return v.position
}

func (v *SimpleVehicle) GetStatus() component.ComponentStatus {
	return v.status
}

func (v *SimpleVehicle) Initialize(ctx context.Context, cfg config.Config) error {
	v.status.Status = "ready"
	return nil
}

func (v *SimpleVehicle) Arm(ctx context.Context) error {
	v.status.Status = "running"
	return nil
}

func (v *SimpleVehicle) EmergencyStop() error {
	v.status.Status = "stopped"
	return nil
}

// SetStaged sets the vehicle staging status
func (v *SimpleVehicle) SetStaged(staged bool) {
	v.staged = staged
}

// SetPosition sets the vehicle position on track
func (v *SimpleVehicle) SetPosition(position float64) {
	v.position = position
}
