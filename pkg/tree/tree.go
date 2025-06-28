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
	id             events.ComponentID
	bus            events.EventBus
	config         config.Config
	mu             sync.RWMutex
	status         TreeStatus
	compStatus     component.ComponentStatus
	running        bool
	lanesPreStaged map[int]bool
	lanesStaged    map[int]bool
}

func NewChristmasTree() *ChristmasTree {
	return &ChristmasTree{
		id: events.ComponentChristmasTree,
		status: TreeStatus{
			IsArmed:     false,
			IsRunning:   false,
			LightStates: make(map[int]map[LightType]LightState),
		},
		compStatus: component.ComponentStatus{
			ID:       events.ComponentChristmasTree,
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
		lanesPreStaged: make(map[int]bool),
		lanesStaged:    make(map[int]bool),
	}
}

func (ct *ChristmasTree) GetID() events.ComponentID {
	return ct.id
}

func (ct *ChristmasTree) Initialize(ctx context.Context, bus events.EventBus, cfg config.Config) error {
	ct.bus = bus
	ct.config = cfg

	// Subscribe to timing system events
	ct.bus.Subscribe(events.EventPreStageOn, ct.handlePreStage)
	ct.bus.Subscribe(events.EventStageOn, ct.handleStage)
	ct.bus.Subscribe(events.EventRaceStarted, ct.handleRaceStart)

	// Initialize light states for all lanes
	trackConfig := cfg.GetTrackConfig()
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

func (ct *ChristmasTree) HandleEvent(ctx context.Context, event events.Event) error {
	switch event.GetType() {
	case events.EventPreStageOn:
		return ct.handlePreStage(ctx, event)
	case events.EventStageOn:
		return ct.handleStage(ctx, event)
	case events.EventRaceStarted:
		return ct.handleRaceStart(ctx, event)
	}
	return nil
}

func (ct *ChristmasTree) handlePreStage(ctx context.Context, event events.Event) error {
	data := event.GetData()
	lane, ok := data["lane"].(int)
	if !ok {
		return fmt.Errorf("invalid lane data")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.LightStates[lane][LightPreStage] = LightOn
	ct.lanesPreStaged[lane] = true

	fmt.Printf("🟡 libdrag: Pre-stage light ON for lane %d\n", lane)

	// Check if both lanes are pre-staged to arm the race
	trackConfig := ct.config.GetTrackConfig()
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

		// Publish race armed event
		event := &events.BaseEvent{
			Type:      events.EventRaceArmed,
			Timestamp: time.Now(),
			Source:    ct.id,
			Data:      map[string]interface{}{"all_lanes_pre_staged": true},
		}
		ct.bus.Publish(ctx, event)
	}

	return nil
}

func (ct *ChristmasTree) handleStage(ctx context.Context, event events.Event) error {
	data := event.GetData()
	lane, ok := data["lane"].(int)
	if !ok {
		return fmt.Errorf("invalid lane data")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.LightStates[lane][LightStage] = LightOn
	ct.lanesStaged[lane] = true

	fmt.Printf("🟡 libdrag: Stage light ON for lane %d\n", lane)

	// Check if both lanes are staged to auto-start sequence
	trackConfig := ct.config.GetTrackConfig()
	allStaged := true
	for laneNum := 1; laneNum <= trackConfig.LaneCount; laneNum++ {
		if !ct.lanesStaged[laneNum] {
			allStaged = false
			break
		}
	}

	if allStaged && ct.status.IsArmed && !ct.status.IsRunning {
		fmt.Println("🚀 libdrag: Auto-starting sequence - both lanes staged")

		// Auto-start sequence after short delay
		go func() {
			time.Sleep(500 * time.Millisecond) // Brief pause
			ct.handleRaceStart(ctx, &events.BaseEvent{
				Type:      events.EventRaceStarted,
				Timestamp: time.Now(),
				Source:    events.ComponentStarterControl,
				Data:      map[string]interface{}{"sequence_type": config.TreeSequencePro},
			})
		}()
	}

	return nil
}

func (ct *ChristmasTree) handleRaceStart(ctx context.Context, event events.Event) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if !ct.status.IsArmed {
		return fmt.Errorf("tree is not armed")
	}

	if ct.status.IsRunning {
		return fmt.Errorf("sequence already running")
	}

	data := event.GetData()
	sequenceType, ok := data["sequence_type"].(config.TreeSequenceType)
	if !ok {
		sequenceType = config.TreeSequencePro // Default
	}

	ct.status.IsRunning = true
	ct.status.SequenceType = sequenceType
	ct.status.LastSequence = time.Now()

	fmt.Printf("🎄 libdrag: Starting %s sequence\n", sequenceType)

	// Start the sequence in a goroutine
	go ct.runSequence(ctx, sequenceType)

	// Publish sequence start event
	event = &events.BaseEvent{
		Type:      events.EventTreeSequenceStart,
		Timestamp: time.Now(),
		Source:    ct.id,
		Data: map[string]interface{}{
			"sequence_type": sequenceType,
		},
	}
	ct.bus.Publish(ctx, event)

	return nil
}

func (ct *ChristmasTree) runSequence(ctx context.Context, sequenceType config.TreeSequenceType) {
	defer func() {
		ct.mu.Lock()
		ct.status.IsRunning = false
		ct.mu.Unlock()
	}()

	treeConfig := ct.config.GetTreeConfig()

	switch sequenceType {
	case config.TreeSequencePro:
		ct.runProSequence(ctx, treeConfig)
	case config.TreeSequenceSportsman:
		ct.runSportsmanSequence(ctx, treeConfig)
	default:
		ct.runProSequence(ctx, treeConfig)
	}
}

func (ct *ChristmasTree) runProSequence(ctx context.Context, cfg config.TreeSequenceConfig) {
	fmt.Println("🟡🟡🟡 libdrag: All three ambers ON")

	// All three ambers simultaneously
	ct.setAllLights(LightAmber1, LightOn)
	ct.setAllLights(LightAmber2, LightOn)
	ct.setAllLights(LightAmber3, LightOn)

	ct.publishLightEvent(ctx, events.EventAmberLight, map[string]interface{}{
		"lights": []LightType{LightAmber1, LightAmber2, LightAmber3},
		"state":  LightOn,
	})

	// Wait for green delay
	time.Sleep(cfg.GreenDelay)

	// Turn off ambers and turn on green
	ct.setAllLights(LightAmber1, LightOff)
	ct.setAllLights(LightAmber2, LightOff)
	ct.setAllLights(LightAmber3, LightOff)
	ct.setAllLights(LightGreen, LightOn)

	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")

	ct.publishLightEvent(ctx, events.EventGreenLight, map[string]interface{}{
		"all_lanes": true,
	})
}

func (ct *ChristmasTree) runSportsmanSequence(ctx context.Context, cfg config.TreeSequenceConfig) {
	// Sequential ambers
	amberLights := []LightType{LightAmber1, LightAmber2, LightAmber3}

	for i, light := range amberLights {
		fmt.Printf("🟡 libdrag: Amber %d ON\n", i+1)
		ct.setAllLights(light, LightOn)
		ct.publishLightEvent(ctx, events.EventAmberLight, map[string]interface{}{
			"light": light,
			"state": LightOn,
		})

		time.Sleep(cfg.AmberDelay)
	}

	// Wait for green delay
	time.Sleep(cfg.GreenDelay)

	// Turn off all ambers and turn on green
	for _, light := range amberLights {
		ct.setAllLights(light, LightOff)
	}
	ct.setAllLights(LightGreen, LightOn)

	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")

	ct.publishLightEvent(ctx, events.EventGreenLight, map[string]interface{}{
		"all_lanes": true,
	})
}

func (ct *ChristmasTree) setAllLights(lightType LightType, state LightState) {
	for lane := range ct.status.LightStates {
		ct.status.LightStates[lane][lightType] = state
	}
}

func (ct *ChristmasTree) publishLightEvent(ctx context.Context, eventType events.EventType, data map[string]interface{}) {
	event := &events.BaseEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    ct.id,
		Data:      data,
	}
	ct.bus.Publish(ctx, event)
}

func (ct *ChristmasTree) GetTreeStatus() TreeStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.status
}
