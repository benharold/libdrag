package api

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/benharold/libdrag/internal/vehicle"
	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/orchestrator"
	"github.com/benharold/libdrag/pkg/timing"
	"github.com/benharold/libdrag/pkg/tree"
)

// LibDragAPI provides a mobile-friendly interface
type LibDragAPI struct {
	orchestrator *orchestrator.RaceOrchestrator
	mu           sync.RWMutex
}

func NewLibDragAPI() *LibDragAPI {
	return &LibDragAPI{
		orchestrator: orchestrator.NewRaceOrchestrator(),
	}
}

// Initialize the libdrag system
func (api *LibDragAPI) Initialize() error {
	// Create configuration
	cfg := config.NewDefaultConfig()

	// Create components
	timingSystem := timing.NewTimingSystem()
	christmasTree := tree.NewChristmasTree()

	components := []component.Component{
		timingSystem,
		christmasTree,
	}

	// Initialize system
	ctx := context.Background()
	return api.orchestrator.Initialize(ctx, components, cfg)
}

// StartRace starts a new drag race
func (api *LibDragAPI) StartRace() error {
	leftVehicle := vehicle.NewSimpleVehicle(1)
	rightVehicle := vehicle.NewSimpleVehicle(2)
	return api.orchestrator.StartRace(leftVehicle, rightVehicle)
}

// GetTreeStatusJSON returns christmas tree status as JSON
func (api *LibDragAPI) GetTreeStatusJSON() string {
	status := api.orchestrator.GetTreeStatus()
	jsonData, _ := json.Marshal(status)
	return string(jsonData)
}

// GetResultsJSON returns race results as JSON
func (api *LibDragAPI) GetResultsJSON() string {
	results := api.orchestrator.GetResults()
	jsonData, _ := json.Marshal(results)
	return string(jsonData)
}

// GetRaceStatusJSON returns race status as JSON
func (api *LibDragAPI) GetRaceStatusJSON() string {
	status := api.orchestrator.GetRaceStatus()
	jsonData, _ := json.Marshal(status)
	return string(jsonData)
}

// Stop shuts down the libdrag system
func (api *LibDragAPI) Stop() {
	api.orchestrator.Stop()
}

// IsRaceComplete checks if the race is finished
func (api *LibDragAPI) IsRaceComplete() bool {
	status := api.orchestrator.GetRaceStatus()
	return status.State == orchestrator.RaceStateComplete
}

// Version returns the libdrag version
func Version() string {
	return "libdrag v1.0.0 - Professional Drag Racing Library"
}
