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
	State       RaceState                            `json:"state"`
	StartTime   time.Time                            `json:"start_time,omitempty"`
	Components  map[string]component.ComponentStatus `json:"components"`
	ActiveLanes []int                                `json:"active_lanes"`
	LastError   error                                `json:"last_error,omitempty"`
}

// RaceOrchestrator coordinates all race components using direct method calls
type RaceOrchestrator struct {
	mu            sync.RWMutex
	config        config.Config
	status        RaceStatus
	timingSystem  *timing.TimingSystem
	christmasTree *tree.ChristmasTree
	leftVehicle   *vehicle.SimpleVehicle
	rightVehicle  *vehicle.SimpleVehicle
	eventBus      *events.EventBus
	raceID        string
}

func NewRaceOrchestrator() *RaceOrchestrator {
	return &RaceOrchestrator{
		status: RaceStatus{
			State:       RaceStateIdle,
			Components:  make(map[string]component.ComponentStatus),
			ActiveLanes: []int{},
		},
	}
}

func (ro *RaceOrchestrator) Initialize(ctx context.Context, components []component.Component, cfg config.Config) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.config = cfg

	// Initialize components and identify their types
	for _, comp := range components {
		if err := comp.Initialize(ctx, cfg); err != nil {
			return fmt.Errorf("failed to initialize component %s: %v", comp.GetID(), err)
		}

		// Type-assert to get specific component references
		switch c := comp.(type) {
		case *timing.TimingSystem:
			ro.timingSystem = c
		case *tree.ChristmasTree:
			ro.christmasTree = c
		}

		// If component supports events, set event bus and race ID
		if eventAware, ok := comp.(component.EventAwareComponent); ok {
			if ro.eventBus != nil {
				eventAware.SetEventBus(ro.eventBus)
			}
			if ro.raceID != "" {
				eventAware.SetRaceID(ro.raceID)
			}
		}

		ro.status.Components[comp.GetID()] = comp.GetStatus()
	}

	// Verify we have required components
	if ro.timingSystem == nil {
		return fmt.Errorf("timing system component is required")
	}
	if ro.christmasTree == nil {
		return fmt.Errorf("christmas tree component is required")
	}

	// Arm components
	for _, comp := range components {
		if err := comp.Arm(ctx); err != nil {
			return fmt.Errorf("failed to start component %s: %v", comp.GetID(), err)
		}
	}

	ro.status.State = RaceStatePreparing
	return nil
}

func (ro *RaceOrchestrator) StartRace(leftVehicle, rightVehicle *vehicle.SimpleVehicle) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	fmt.Println("🏁 libdrag Race Orchestrator: Starting new race")

	ro.leftVehicle = leftVehicle
	ro.rightVehicle = rightVehicle
	ro.status.ActiveLanes = []int{1, 2}
	ro.status.StartTime = time.Now()
	ro.status.State = RaceStateStaging

	// Publish race start event
	if ro.eventBus != nil {
		ro.eventBus.Publish(
			events.NewEvent(events.EventRaceStart).
				WithRaceID(ro.raceID).
				Build(),
		)
	}

	// Reset and prepare timing system
	ro.timingSystem.StartRace()
	ro.timingSystem.AddVehicles([]int{1, 2})

	// Simulate staging process
	go ro.simulateRaceSequence()

	return nil
}

func (ro *RaceOrchestrator) simulateRaceSequence() {
	// Simulate vehicles entering pre-stage
	time.Sleep(500 * time.Millisecond)
	ro.christmasTree.SetPreStage(1)

	time.Sleep(200 * time.Millisecond)
	ro.christmasTree.SetPreStage(2)

	// Update state to armed
	ro.mu.Lock()
	ro.status.State = RaceStateArmed
	ro.mu.Unlock()

	// Simulate vehicles entering stage
	time.Sleep(500 * time.Millisecond)
	ro.christmasTree.SetStage(1)

	time.Sleep(300 * time.Millisecond)
	ro.christmasTree.SetStage(2)

	// Wait briefly, then start the tree sequence
	time.Sleep(500 * time.Millisecond)

	if ro.christmasTree.AllStaged() {
		ro.mu.Lock()
		ro.status.State = RaceStateRunning
		ro.mu.Unlock()

		// Arm the Christmas tree sequence and get green light time
		err := ro.christmasTree.StartSequence(config.TreeSequencePro)
		if err != nil {
			fmt.Printf("❌ Failed to start tree sequence: %v\n", err)
			return
		}

		// Wait for sequence to complete and get green light time
		// In a real implementation, the tree would return the green light time
		time.Sleep(500 * time.Millisecond) // Wait for sequence
		greenTime := time.Now()

		ro.timingSystem.SetGreenLight(greenTime)

		// Simulate vehicle race
		ro.simulateVehicleRun(greenTime)
	}
}

func (ro *RaceOrchestrator) simulateVehicleRun(greenTime time.Time) {
	// Simulate realistic reaction times and race progression

	// Lane 1 vehicle starts (good reaction time)
	reactionTime1 := 400 * time.Millisecond
	startTime1 := greenTime.Add(reactionTime1)
	ro.timingSystem.TriggerBeam("stage", 1, startTime1)

	// Lane 2 vehicle starts (slightly slower)
	reactionTime2 := 450 * time.Millisecond
	startTime2 := greenTime.Add(reactionTime2)
	ro.timingSystem.TriggerBeam("stage", 2, startTime2)

	// Simulate 60-foot times
	time.Sleep(50 * time.Millisecond) // Fast simulation
	ro.timingSystem.TriggerBeam("60_foot", 1, startTime1.Add(950*time.Millisecond))
	ro.timingSystem.TriggerBeam("60_foot", 2, startTime2.Add(980*time.Millisecond))

	// Simulate eighth-mile times
	time.Sleep(50 * time.Millisecond)
	ro.timingSystem.TriggerBeam("660_foot", 1, startTime1.Add(4200*time.Millisecond))
	ro.timingSystem.TriggerBeam("660_foot", 2, startTime2.Add(4350*time.Millisecond))

	// Simulate quarter-mile finish
	time.Sleep(50 * time.Millisecond)
	ro.timingSystem.TriggerBeam("1320_foot", 1, startTime1.Add(7300*time.Millisecond))
	ro.timingSystem.TriggerBeam("1320_foot", 2, startTime2.Add(7500*time.Millisecond))

	// Race complete
	ro.mu.Lock()
	ro.status.State = RaceStateComplete
	ro.mu.Unlock()

	// Publish race complete event
	if ro.eventBus != nil {
		ro.eventBus.Publish(
			events.NewEvent(events.EventRaceComplete).
				WithRaceID(ro.raceID).
				Build(),
		)
	}

	fmt.Println("🏁 libdrag Race Orchestrator: Race complete!")
}

func (ro *RaceOrchestrator) GetRaceStatus() RaceStatus {
	ro.mu.RLock()
	defer ro.mu.RUnlock()
	return ro.status
}

func (ro *RaceOrchestrator) GetResults() map[int]*timing.TimingResults {
	if ro.timingSystem == nil {
		return make(map[int]*timing.TimingResults)
	}
	return ro.timingSystem.GetAllResults()
}

func (ro *RaceOrchestrator) GetTimingSystem() *timing.TimingSystem {
	return ro.timingSystem
}

func (ro *RaceOrchestrator) GetTreeStatus() *tree.Status {
	if ro.christmasTree == nil {
		return nil
	}
	status := ro.christmasTree.GetTreeStatus()
	return &status
}

func (ro *RaceOrchestrator) Stop() error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.status.State = RaceStateIdle
	return nil
}

func (ro *RaceOrchestrator) IsRaceComplete() bool {
	ro.mu.RLock()
	defer ro.mu.RUnlock()
	return ro.status.State == RaceStateComplete
}

// SetEventBus sets the event bus for the orchestrator
func (ro *RaceOrchestrator) SetEventBus(eventBus *events.EventBus) {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	ro.eventBus = eventBus
}

// SetRaceID sets the race ID for the orchestrator
func (ro *RaceOrchestrator) SetRaceID(raceID string) {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	ro.raceID = raceID
}
