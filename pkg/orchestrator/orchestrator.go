package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/internal/vehicle"
	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
	"github.com/benharold/libdrag/pkg/timing"
	"github.com/benharold/libdrag/pkg/tree"
)

// RaceState defines race progression states
type RaceState string

const (
	RaceStateIdle      RaceState = "idle"
	RaceStatePreparing RaceState = "preparing"
	RaceStateStaging   RaceState = "staging"
	RaceStateArmed     RaceState = "armed"
	RaceStateRunning   RaceState = "running"
	RaceStateComplete  RaceState = "complete"
	RaceStateAborted   RaceState = "aborted"
	RaceStateError     RaceState = "error"
)

// RaceStatus represents overall race state
type RaceStatus struct {
	State       RaceState                                        `json:"state"`
	StartTime   time.Time                                        `json:"start_time,omitempty"`
	Components  map[events.ComponentID]component.ComponentStatus `json:"components"`
	ActiveLanes []int                                            `json:"active_lanes"`
	LastError   error                                            `json:"last_error,omitempty"`
}

// RaceOrchestrator coordinates all race components
type RaceOrchestrator struct {
	mu            sync.RWMutex
	bus           events.EventBus
	config        config.Config
	components    map[events.ComponentID]component.Component
	status        RaceStatus
	eventLog      []events.Event
	timingSystem  *timing.TimingSystem
	christmasTree *tree.ChristmasTree
}

func NewRaceOrchestrator() *RaceOrchestrator {
	return &RaceOrchestrator{
		components: make(map[events.ComponentID]component.Component),
		status: RaceStatus{
			State:       RaceStateIdle,
			Components:  make(map[events.ComponentID]component.ComponentStatus),
			ActiveLanes: make([]int, 0),
		},
		eventLog: make([]events.Event, 0),
	}
}

func (ro *RaceOrchestrator) Initialize(ctx context.Context, components []component.Component, cfg config.Config) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.config = cfg
	ro.bus = events.NewSimpleEventBus()

	// Start event bus
	if err := ro.bus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %v", err)
	}

	// Subscribe to all events for logging
	ro.bus.SubscribeAll(ro.logEvent)

	// Initialize all components
	for _, comp := range components {
		ro.components[comp.GetID()] = comp

		// Store references to key components
		switch c := comp.(type) {
		case *timing.TimingSystem:
			ro.timingSystem = c
		case *tree.ChristmasTree:
			ro.christmasTree = c
		}

		if err := comp.Initialize(ctx, ro.bus, cfg); err != nil {
			return fmt.Errorf("failed to initialize component %s: %v", comp.GetID(), err)
		}

		if err := comp.Start(ctx); err != nil {
			return fmt.Errorf("failed to start component %s: %v", comp.GetID(), err)
		}

		ro.status.Components[comp.GetID()] = comp.GetStatus()
	}

	// Subscribe to race completion events
	ro.bus.Subscribe(events.EventRunComplete, ro.handleRunComplete)

	ro.status.State = RaceStateIdle

	// Publish system ready event
	event := &events.BaseEvent{
		Type:      events.EventSystemReady,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{"components_count": len(components)},
	}
	ro.bus.Publish(ctx, event)

	return nil
}

func (ro *RaceOrchestrator) StartRace(leftVehicle, rightVehicle vehicle.VehicleInterface) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	if ro.status.State != RaceStateIdle {
		return fmt.Errorf("cannot start race: current state is %s", ro.status.State)
	}

	ro.status.State = RaceStatePreparing
	ro.status.StartTime = time.Now()
	ro.status.ActiveLanes = []int{1, 2}

	fmt.Println("🏁 libdrag Race Orchestrator: Starting new race")

	// Notify components about race preparation
	event := &events.BaseEvent{
		Type:      events.EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data: map[string]interface{}{
			"lanes": ro.status.ActiveLanes,
		},
	}

	return ro.bus.Publish(context.Background(), event)
}

func (ro *RaceOrchestrator) GetRaceStatus() RaceStatus {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	// Update component statuses
	status := ro.status
	status.Components = make(map[events.ComponentID]component.ComponentStatus)
	for id, comp := range ro.components {
		status.Components[id] = comp.GetStatus()
	}

	return status
}

func (ro *RaceOrchestrator) GetResults() map[int]*timing.TimingResults {
	if ro.timingSystem == nil {
		return nil
	}

	results := make(map[int]*timing.TimingResults)
	results[1] = ro.timingSystem.GetResults(1)
	results[2] = ro.timingSystem.GetResults(2)

	return results
}

// GetTimingSystem returns the timing system component
func (ro *RaceOrchestrator) GetTimingSystem() *timing.TimingSystem {
	ro.mu.RLock()
	defer ro.mu.RUnlock()
	return ro.timingSystem
}

func (ro *RaceOrchestrator) GetTreeStatus() tree.TreeStatus {
	if ro.christmasTree == nil {
		return tree.TreeStatus{}
	}
	return ro.christmasTree.GetTreeStatus()
}

func (ro *RaceOrchestrator) Stop() {
	for _, comp := range ro.components {
		comp.Stop()
	}
	if ro.bus != nil {
		ro.bus.Stop()
	}
}

func (ro *RaceOrchestrator) logEvent(ctx context.Context, event events.Event) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.eventLog = append(ro.eventLog, event)
	return nil
}

func (ro *RaceOrchestrator) handleRunComplete(ctx context.Context, event events.Event) error {
	data := event.GetData()
	result, ok := data["results"].(*timing.TimingResults)
	if !ok {
		return fmt.Errorf("invalid results data")
	}

	fmt.Printf("🏁 libdrag: Lane %d completed: %.3fs", result.Lane, *result.QuarterMileTime)
	if result.TrapSpeed != nil {
		fmt.Printf(" @ %.1f mph", *result.TrapSpeed)
	}
	fmt.Println()

	// Check if all lanes are complete
	ro.mu.Lock()
	defer ro.mu.Unlock()

	allComplete := true
	for _, lane := range ro.status.ActiveLanes {
		if result := ro.timingSystem.GetResults(lane); result == nil || !result.IsComplete {
			allComplete = false
			break
		}
	}

	if allComplete {
		ro.status.State = RaceStateComplete
		fmt.Println("🏆 libdrag: Race completed!")

		// Determine winner
		ro.determineWinner()
	}

	return nil
}

func (ro *RaceOrchestrator) determineWinner() {
	results := ro.GetResults()
	var bestTime *float64
	var winner int

	for lane, result := range results {
		if result != nil && result.QuarterMileTime != nil && !result.IsFoul {
			if bestTime == nil || *result.QuarterMileTime < *bestTime {
				bestTime = result.QuarterMileTime
				winner = lane
			}
		}
	}

	if winner > 0 {
		fmt.Printf("🥇 libdrag: Winner: Lane %d with %.3fs\n", winner, *bestTime)

		// Calculate margin
		for lane, result := range results {
			if lane != winner && result != nil && result.QuarterMileTime != nil {
				margin := *result.QuarterMileTime - *bestTime
				fmt.Printf("📊 libdrag: Lane %d lost by %.3fs\n", lane, margin)
			}
		}
	}
}

func (ro *RaceOrchestrator) Reset() error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	// Only allow reset if race is complete, aborted, or in error state
	if ro.status.State != RaceStateComplete && ro.status.State != RaceStateAborted && ro.status.State != RaceStateError {
		return fmt.Errorf("cannot reset race: current state is %s", ro.status.State)
	}

	// Reset race status
	ro.status.State = RaceStateIdle
	ro.status.StartTime = time.Time{}
	ro.status.ActiveLanes = make([]int, 0)
	ro.status.LastError = nil

	// Clear event log
	ro.eventLog = make([]events.Event, 0)

	fmt.Println("🔄 libdrag Race Orchestrator: Race reset to idle state")

	// Publish reset event
	event := &events.BaseEvent{
		Type:      events.EventRaceReset,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{},
	}
	ro.bus.Publish(context.Background(), event)

	return nil
}

// IsRaceComplete returns true if the race is in a completed state
func (ro *RaceOrchestrator) IsRaceComplete() bool {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	return ro.status.State == RaceStateComplete || ro.status.State == RaceStateAborted || ro.status.State == RaceStateError
}
