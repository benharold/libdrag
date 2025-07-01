package tree

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// LightType defines different lights on the christmas tree
type LightType string

const (
	LightPreStage LightType = "pre_stage"
	LightStage    LightType = "stage"
	LightAmber1   LightType = "amber_1"
	LightAmber2   LightType = "amber_2"
	LightAmber3   LightType = "amber_3"
	LightGreen    LightType = "green"
	LightRed      LightType = "red"
)

// LightState defines light states
type LightState string

const (
	LightOff   LightState = "off"
	LightOn    LightState = "on"
	LightBlink LightState = "blink"
)

// TreeStatus represents christmas tree state
type TreeStatus struct {
	IsArmed      bool                             `json:"is_armed"`
	IsRunning    bool                             `json:"is_running"`
	SequenceType config.TreeSequenceType          `json:"sequence_type"`
	CurrentStep  int                              `json:"current_step"`
	LightStates  map[int]map[LightType]LightState `json:"light_states"` // lane -> light -> state
	LastSequence time.Time                        `json:"last_sequence,omitempty"`
}

// ChristmasTree implements the christmas tree component
type ChristmasTree struct {
	id             string
	config         config.Config
	mu             sync.RWMutex
	status         TreeStatus
	compStatus     component.ComponentStatus
	running        bool
	lanesPreStaged map[int]bool
	lanesStaged    map[int]bool
	eventBus       *events.EventBus
	raceID         string
}

func NewChristmasTree() *ChristmasTree {
	return &ChristmasTree{
		id: "christmas_tree",
		status: TreeStatus{
			IsArmed:     false,
			IsRunning:   false,
			LightStates: make(map[int]map[LightType]LightState),
		},
		compStatus: component.ComponentStatus{
			ID:       "christmas_tree",
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
		lanesPreStaged: make(map[int]bool),
		lanesStaged:    make(map[int]bool),
	}
}

func (ct *ChristmasTree) GetID() string {
	return ct.id
}

func (ct *ChristmasTree) Initialize(ctx context.Context, cfg config.Config) error {
	ct.config = cfg

	// Initialize light states for all lanes
	trackConfig := cfg.Track()
	for lane := 1; lane <= trackConfig.LaneCount; lane++ {
		ct.status.LightStates[lane] = make(map[LightType]LightState)
		for _, lightType := range []LightType{LightPreStage, LightStage, LightAmber1, LightAmber2, LightAmber3, LightGreen, LightRed} {
			ct.status.LightStates[lane][lightType] = LightOff
		}
	}

	ct.compStatus.Status = "ready"
	return nil
}

func (ct *ChristmasTree) Start(ctx context.Context) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.running = true
	ct.compStatus.Status = "running"
	fmt.Println("🎄 libdrag Christmas Tree: Started")
	return nil
}

func (ct *ChristmasTree) Stop() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.running = false
	ct.compStatus.Status = "stopped"
	return nil
}

func (ct *ChristmasTree) GetStatus() component.ComponentStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.compStatus
}

func (ct *ChristmasTree) GetTreeStatus() TreeStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.status
}

// SetEventBus sets the event bus for publishing events
func (ct *ChristmasTree) SetEventBus(eventBus *events.EventBus) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.eventBus = eventBus
}

// SetRaceID sets the race ID for event context
func (ct *ChristmasTree) SetRaceID(raceID string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.raceID = raceID
}

func (ct *ChristmasTree) SetPreStage(lane int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.LightStates[lane][LightPreStage] = LightOn
	ct.lanesPreStaged[lane] = true

	fmt.Printf("🟡 libdrag: Pre-stage light ON for lane %d\n", lane)
	
	// Publish pre-stage event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreePreStage).
				WithRaceID(ct.raceID).
				WithLane(lane).
				Build(),
		)
	}

	// Check if both lanes are pre-staged to arm the race
	trackConfig := ct.config.Track()
	allPreStaged := true
	for laneNum := 1; laneNum <= trackConfig.LaneCount; laneNum++ {
		if !ct.lanesPreStaged[laneNum] {
			allPreStaged = false
			break
		}
	}

	if allPreStaged && !ct.status.IsArmed {
		ct.status.IsArmed = true
		fmt.Println("🔥 libdrag Christmas Tree: ARMED - Both lanes pre-staged")
		
		// Publish armed event
		if ct.eventBus != nil {
			ct.eventBus.Publish(
				events.NewEvent(events.EventTreeArmed).
					WithRaceID(ct.raceID).
					Build(),
			)
		}
	}
}

func (ct *ChristmasTree) SetStage(lane int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.LightStates[lane][LightStage] = LightOn
	ct.lanesStaged[lane] = true

	fmt.Printf("🟡 libdrag: Stage light ON for lane %d\n", lane)
	
	// Publish stage event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeStage).
				WithRaceID(ct.raceID).
				WithLane(lane).
				Build(),
		)
	}
}

func (ct *ChristmasTree) IsArmed() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.status.IsArmed
}

func (ct *ChristmasTree) AllStaged() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if !ct.status.IsArmed {
		return false
	}

	trackConfig := ct.config.Track()
	for laneNum := 1; laneNum <= trackConfig.LaneCount; laneNum++ {
		if !ct.lanesStaged[laneNum] {
			return false
		}
	}
	return true
}

func (ct *ChristmasTree) StartSequence(sequenceType config.TreeSequenceType) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if !ct.status.IsArmed {
		return fmt.Errorf("tree is not armed")
	}

	if ct.status.IsRunning {
		return fmt.Errorf("sequence already running")
	}

	ct.status.IsRunning = true
	ct.status.SequenceType = sequenceType
	ct.status.LastSequence = time.Now()

	fmt.Printf("🎄 libdrag: Starting %s sequence\n", sequenceType)
	
	// Publish sequence start event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeSequenceStart).
				WithRaceID(ct.raceID).
				WithData("sequence_type", string(sequenceType)).
				Build(),
		)
	}

	// Start the sequence in a goroutine
	go ct.runSequence(sequenceType)

	return nil
}

func (ct *ChristmasTree) runSequence(sequenceType config.TreeSequenceType) time.Time {
	defer func() {
		ct.mu.Lock()
		ct.status.IsRunning = false
		ct.mu.Unlock()
		
		// Publish sequence end event
		if ct.eventBus != nil {
			ct.eventBus.Publish(
				events.NewEvent(events.EventTreeSequenceEnd).
					WithRaceID(ct.raceID).
					WithData("sequence_type", string(sequenceType)).
					Build(),
			)
		}
	}()

	treeConfig := ct.config.Tree()

	switch sequenceType {
	case config.TreeSequencePro:
		return ct.runProSequence(treeConfig)
	case config.TreeSequenceSportsman:
		return ct.runSportsmanSequence(treeConfig)
	default:
		return ct.runProSequence(treeConfig)
	}
}

func (ct *ChristmasTree) runProSequence(cfg config.TreeSequenceConfig) time.Time {
	fmt.Println("🟡🟡🟡 libdrag: All three ambers ON")

	// All three ambers simultaneously
	ct.setAllLights(LightAmber1, LightOn)
	ct.setAllLights(LightAmber2, LightOn)
	ct.setAllLights(LightAmber3, LightOn)
	
	// Publish amber event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeAmberOn).
				WithRaceID(ct.raceID).
				WithData("count", 3).
				WithData("sequence", "pro").
				Build(),
		)
	}

	// Wait for green delay
	time.Sleep(cfg.GreenDelay)

	// Turn off ambers and turn on green
	ct.setAllLights(LightAmber1, LightOff)
	ct.setAllLights(LightAmber2, LightOff)
	ct.setAllLights(LightAmber3, LightOff)
	ct.setAllLights(LightGreen, LightOn)

	greenTime := time.Now()
	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")
	
	// Publish green light event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeGreenOn).
				WithRaceID(ct.raceID).
				WithData("green_time", greenTime).
				Build(),
		)
	}

	return greenTime
}

func (ct *ChristmasTree) runSportsmanSequence(cfg config.TreeSequenceConfig) time.Time {
	// Sequential ambers
	amberLights := []LightType{LightAmber1, LightAmber2, LightAmber3}

	for i, light := range amberLights {
		fmt.Printf("🟡 libdrag: Amber %d ON\n", i+1)
		ct.setAllLights(light, LightOn)
		
		// Publish amber event for each light
		if ct.eventBus != nil {
			ct.eventBus.Publish(
				events.NewEvent(events.EventTreeAmberOn).
					WithRaceID(ct.raceID).
					WithData("amber_number", i+1).
					WithData("sequence", "sportsman").
					Build(),
			)
		}

		if i < len(amberLights)-1 {
			time.Sleep(cfg.AmberDelay)
		}
	}

	// Wait for green delay after last amber
	time.Sleep(cfg.GreenDelay)

	// Turn off ambers and turn on green
	for _, light := range amberLights {
		ct.setAllLights(light, LightOff)
	}
	ct.setAllLights(LightGreen, LightOn)

	greenTime := time.Now()
	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")
	
	// Publish green light event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeGreenOn).
				WithRaceID(ct.raceID).
				WithData("green_time", greenTime).
				Build(),
		)
	}

	return greenTime
}

func (ct *ChristmasTree) setAllLights(lightType LightType, state LightState) {
	trackConfig := ct.config.Track()
	for lane := 1; lane <= trackConfig.LaneCount; lane++ {
		ct.status.LightStates[lane][lightType] = state
	}
}
