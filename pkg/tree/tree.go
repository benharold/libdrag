package tree

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// LightType defines different lights on the Christmas tree
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

// Status represents Christmas tree state
type Status struct {
	Armed          bool                             `json:"armed"`     // starter has enabled auto-start system to take control
	Activated      bool                             `json:"activated"` // auto-start system detected staging conditions and started sequence
	SequenceType   config.TreeSequenceType          `json:"sequence_type"`
	CurrentStep    int                              `json:"current_step"`
	LightStates    map[int]map[LightType]LightState `json:"light_states"` // lane -> light -> state
	LastSequence   time.Time                        `json:"last_sequence,omitempty"`
	ArmedTime      time.Time                        `json:"armed_time,omitempty"`      // when starter armed the tree
	ActivationTime time.Time                        `json:"activation_time,omitempty"` // when auto-start activated sequence
	StabilityTimer time.Time                        `json:"stability_timer,omitempty"` // for 0.6s stability requirement
}

// StagingMotionState tracks the staging motion sequence for a lane
type StagingMotionState struct {
	ReachedStage    bool // Has this lane ever reached the stage beam?
	LastStageState  bool // Last state of stage beam (to detect backing)
	MotionHistory   []string // Track sequence of motions for debugging
}

// ChristmasTree implements the Christmas tree component
type ChristmasTree struct {
	id             string
	config         config.Config
	mu             sync.RWMutex
	status         Status
	compStatus     component.ComponentStatus
	lanesPreStaged map[int]bool
	lanesStaged    map[int]bool
	stagingMotion  map[int]*StagingMotionState // Track staging motion per lane
	eventBus       *events.EventBus
	raceID         string
}

func NewChristmasTree() *ChristmasTree {
	id := uuid.New().String()
	return &ChristmasTree{
		id: id,
		status: Status{
			Armed:       false,
			Activated:   false,
			LightStates: make(map[int]map[LightType]LightState),
		},
		compStatus: component.ComponentStatus{
			ID:       id,
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
		lanesPreStaged: make(map[int]bool),
		lanesStaged:    make(map[int]bool),
		stagingMotion:  make(map[int]*StagingMotionState),
	}
}

func (ct *ChristmasTree) GetID() string {
	return ct.id
}

func (ct *ChristmasTree) Initialize(_ context.Context, cfg config.Config) error {
	ct.config = cfg

	// Initialize light states for all lanes
	trackConfig := cfg.Track()
	for lane := 1; lane <= trackConfig.LaneCount; lane++ {
		ct.status.LightStates[lane] = make(map[LightType]LightState)
		for _, lightType := range []LightType{LightPreStage, LightStage, LightAmber1, LightAmber2, LightAmber3, LightGreen, LightRed} {
			ct.status.LightStates[lane][lightType] = LightOff
		}
		
		// Initialize staging motion tracking for each lane
		ct.stagingMotion[lane] = &StagingMotionState{
			ReachedStage:  false,
			LastStageState: false,
			MotionHistory: make([]string, 0),
		}
	}

	ct.compStatus.Status = "ready"
	return nil
}

func (ct *ChristmasTree) Arm(_ context.Context) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.Armed = true
	ct.status.ArmedTime = time.Now()
	ct.compStatus.Status = "armed"
	fmt.Println("💪 libdrag Christmas Tree: Armed by starter - Auto-start system enabled")

	// Publish armed event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeArmed).
				WithRaceID(ct.raceID).
				WithData("armed_by", "starter").
				Build(),
		)
	}

	return nil
}

// DisarmTree disarms the tree (starter control only)
func (ct *ChristmasTree) DisarmTree() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if !ct.status.Armed {
		return
	}

	ct.status.Armed = false
	ct.status.Activated = false
	ct.status.ArmedTime = time.Time{}
	ct.status.ActivationTime = time.Time{}
	ct.status.StabilityTimer = time.Time{}
	ct.compStatus.Status = "ready"
	fmt.Println("💪 libdrag Christmas Tree: DISARMED by starter")

	// Publish disarmed event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeDisarmed).
				WithRaceID(ct.raceID).
				Build(),
		)
	}
}

func (ct *ChristmasTree) ActivateAutoStart() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if !ct.status.Armed {
		return fmt.Errorf("tree must be armed before auto-start can activate")
	}

	if ct.status.Activated {
		return fmt.Errorf("auto-start system already activated")
	}

	ct.status.Activated = true
	ct.status.ActivationTime = time.Now()
	ct.compStatus.Status = "activated"
	fmt.Println("⏳ libdrag Christmas Tree: Auto-start system activated - staging conditions detected")

	// Publish activation event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeActivated).
				WithRaceID(ct.raceID).
				WithData("activation_time", ct.status.ActivationTime).
				Build(),
		)
	}

	return nil
}

func (ct *ChristmasTree) Activate() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.Activated = true
	ct.compStatus.Status = "activated"
	fmt.Println("⏳ libdrag Christmas Tree: Activated")
	return nil
}

func (ct *ChristmasTree) EmergencyStop() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.status.Armed = false
	ct.status.Activated = false
	ct.compStatus.Status = "emergency_stopped"

	// Clear all lights first
	trackConfig := ct.config.Track()
	for lane := 1; lane <= trackConfig.LaneCount; lane++ {
		for _, lightType := range []LightType{LightPreStage, LightStage, LightAmber1, LightAmber2, LightAmber3, LightGreen, LightRed} {
			ct.status.LightStates[lane][lightType] = LightOff
			ct.status.LightStates[lane][LightRed] = LightBlink
		}
	}

	fmt.Println("🚨 libdrag Christmas Tree: EMERGENCY STOP")

	// Publish emergency stop event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeEmergencyStop).
				WithRaceID(ct.raceID).
				Build(),
		)
	}

	return nil
}

func (ct *ChristmasTree) GetStatus() component.ComponentStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.compStatus
}

func (ct *ChristmasTree) GetTreeStatus() Status {
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

func (ct *ChristmasTree) SetPreStage(lane int, beamBroken bool) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if beamBroken {
		ct.status.LightStates[lane][LightPreStage] = LightOn
		ct.lanesPreStaged[lane] = true
		fmt.Printf("🟡 libdrag: Pre-stage light ON for lane %d\n", lane)
	} else {
		ct.status.LightStates[lane][LightPreStage] = LightOff
		ct.lanesPreStaged[lane] = false
		fmt.Printf("⚫ libdrag: Pre-stage light OFF for lane %d\n", lane)
		
		// Check if vehicle has completely backed out (both beams clear)
		stageBeamClear := ct.status.LightStates[lane][LightStage] == LightOff
		if stageBeamClear {
			// Complete back-out - reset staging motion tracking
			ct.resetStagingMotion(lane)
		}
		
		// Check for deep staging when pre-stage turns off
		ct.checkDeepStaging(lane)
	}

	// Publish pre-stage event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreePreStage).
				WithRaceID(ct.raceID).
				WithLane(lane).
				WithData("beam_broken", beamBroken).
				Build(),
		)
	}
}

func (ct *ChristmasTree) SetStage(lane int, beamBroken bool) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Track staging motion before updating state
	ct.trackStagingMotion(lane, beamBroken)

	if beamBroken {
		ct.status.LightStates[lane][LightStage] = LightOn
		ct.lanesStaged[lane] = true
		fmt.Printf("🟡 libdrag: Stage light ON for lane %d\n", lane)
	} else {
		ct.status.LightStates[lane][LightStage] = LightOff
		ct.lanesStaged[lane] = false
		fmt.Printf("⚫ libdrag: Stage light OFF for lane %d\n", lane)
	}

	// Check for deep staging when stage changes
	ct.checkDeepStaging(lane)
	
	// Publish stage event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeStage).
				WithRaceID(ct.raceID).
				WithLane(lane).
				WithData("beam_broken", beamBroken).
				Build(),
		)
	}
}

// checkDeepStaging detects deep staging and handles class-specific rules
func (ct *ChristmasTree) checkDeepStaging(lane int) {
	preStageOn := ct.status.LightStates[lane][LightPreStage] == LightOn
	stageOn := ct.status.LightStates[lane][LightStage] == LightOn
	
	isDeepStaged := !preStageOn && stageOn
	
	if isDeepStaged {
		ct.handleDeepStaging(lane)
	}
}

// handleDeepStaging processes deep staging based on class rules
func (ct *ChristmasTree) handleDeepStaging(lane int) {
	if ct.config == nil {
		return // Can't check class rules without config
	}
	
	racingClass := ct.config.RacingClass()
	
	if ct.isDeepStagingProhibited(racingClass) {
		ct.handleDeepStagingViolation(lane, racingClass)
	} else {
		ct.handleDeepStagingAllowed(lane)
	}
}

// isDeepStagingProhibited checks if deep staging is prohibited for the given class
func (ct *ChristmasTree) isDeepStagingProhibited(class string) bool {
	prohibitedClasses := map[string]bool{
		"Super Gas":    true,
		"Super Stock":  true,
		"Super Street": true,
	}
	return prohibitedClasses[class]
}

// handleDeepStagingViolation processes a deep staging violation
func (ct *ChristmasTree) handleDeepStagingViolation(lane int, class string) {
	fmt.Printf("⚠️  libdrag: Deep staging detected in lane %d (Class: %s - PROHIBITED)\n", lane, class)
	
	// Publish event for starter/officials to decide
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeDeepStageViolation).
				WithRaceID(ct.raceID).
				WithLane(lane).
				WithData("class", class).
				WithData("action_required", "starter_decision").
				Build(),
		)
	}
}

// handleDeepStagingAllowed processes allowed deep staging
func (ct *ChristmasTree) handleDeepStagingAllowed(lane int) {
	fmt.Printf("🔵 libdrag: Deep staging detected in lane %d (Allowed)\n", lane)
	
	// Informational only
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeDeepStage).
				WithRaceID(ct.raceID).
				WithLane(lane).
				WithData("deep_staged", true).
				Build(),
		)
	}
}

// trackStagingMotion monitors staging beam state changes to enforce forward motion rule
func (ct *ChristmasTree) trackStagingMotion(lane int, beamBroken bool) {
	motionState := ct.stagingMotion[lane]
	if motionState == nil {
		return // Safety check
	}

	// If vehicle has never reached stage beam and is now breaking it, mark as reached
	if !motionState.ReachedStage && beamBroken {
		motionState.ReachedStage = true
		motionState.LastStageState = true
		motionState.MotionHistory = append(motionState.MotionHistory, "enter_stage")
		return
	}

	// If vehicle has reached stage beam previously, check for backward motion violations
	if motionState.ReachedStage {
		// Detect backing out of stage beam
		if motionState.LastStageState && !beamBroken {
			motionState.LastStageState = false
			motionState.MotionHistory = append(motionState.MotionHistory, "back_out_stage")
			return
		}
		
		// Detect re-entering stage beam after backing out (VIOLATION)
		if !motionState.LastStageState && beamBroken {
			motionState.LastStageState = true
			motionState.MotionHistory = append(motionState.MotionHistory, "re_enter_stage_VIOLATION")
			ct.handleStagingMotionViolation(lane)
			return
		}
	}
}

// handleStagingMotionViolation processes backward staging motion violations
func (ct *ChristmasTree) handleStagingMotionViolation(lane int) {
	fmt.Printf("⚠️  libdrag: Staging motion violation in lane %d - vehicle backed out and re-entered stage beam\n", lane)
	
	// Publish staging violation event
	if ct.eventBus != nil {
		motionState := ct.stagingMotion[lane]
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeStagingViolation).
				WithRaceID(ct.raceID).
				WithLane(lane).
				WithData("violation_type", "backward_staging_motion").
				WithData("motion_history", motionState.MotionHistory).
				WithData("rule", "last_motion_must_be_forward").
				Build(),
		)
	}
}

// resetStagingMotion resets the staging motion tracking for a lane (when completely backing out)
func (ct *ChristmasTree) resetStagingMotion(lane int) {
	if ct.stagingMotion[lane] != nil {
		ct.stagingMotion[lane].ReachedStage = false
		ct.stagingMotion[lane].LastStageState = false
		ct.stagingMotion[lane].MotionHistory = make([]string, 0)
	}
}

func (ct *ChristmasTree) IsArmed() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.status.Armed
}

func (ct *ChristmasTree) AllStaged() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if !ct.status.Armed {
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

	if !ct.status.Armed {
		return fmt.Errorf("tree is not armed")
	}

	if ct.status.Activated {
		return fmt.Errorf("tree is not activated")
	}

	ct.status.Activated = true
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

	// run the sequence in a goroutine
	go ct.runSequence(sequenceType)

	return nil
}

func (ct *ChristmasTree) runSequence(sequenceType config.TreeSequenceType) time.Time {
	defer func() {
		ct.mu.Lock()
		ct.status.Activated = false
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

// StartStagingProcess starts the staging process for the Christmas tree
func (ct *ChristmasTree) StartStagingProcess(sequenceType config.TreeSequenceType) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if !ct.status.Armed {
		return fmt.Errorf("tree is not armed")
	}

	if !ct.status.Activated {
		return fmt.Errorf("auto-start system is not activated")
	}

	ct.status.SequenceType = sequenceType
	ct.status.LastSequence = time.Now()
	ct.compStatus.Status = "staging_process"

	fmt.Printf("🎄 libdrag: Starting staging process - %s sequence\n", sequenceType)

	// Publish staging process start event
	if ct.eventBus != nil {
		ct.eventBus.Publish(
			events.NewEvent(events.EventTreeSequenceStart).
				WithRaceID(ct.raceID).
				WithData("sequence_type", string(sequenceType)).
				Build(),
		)
	}

	// run the sequence in a goroutine
	go ct.runStagingSequence(sequenceType)

	return nil
}

func (ct *ChristmasTree) runStagingSequence(sequenceType config.TreeSequenceType) time.Time {
	defer func() {
		ct.mu.Lock()
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
