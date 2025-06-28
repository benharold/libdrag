package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// =====================================
// CORE TYPES AND INTERFACES
// =====================================

// EventType defines the types of events in the system
type EventType string

const (
	// Timing Events
	EventBeamTriggered   EventType = "beam_triggered"
	EventBeamReleased    EventType = "beam_released"
	EventTimingComplete  EventType = "timing_complete"
	
	// Christmas Tree Events
	EventPreStageOn      EventType = "pre_stage_on"
	EventStageOn         EventType = "stage_on"
	EventTreeSequenceStart EventType = "tree_sequence_start"
	EventAmberLight      EventType = "amber_light"
	EventGreenLight      EventType = "green_light"
	EventRedLight        EventType = "red_light"
	
	// Race Events
	EventRaceArmed       EventType = "race_armed"
	EventRaceStarted     EventType = "race_started"
	EventRaceAborted     EventType = "race_aborted"
	EventRunComplete     EventType = "run_complete"
	EventSystemReady     EventType = "system_ready"
	EventSystemError     EventType = "system_error"
	
	// Vehicle Events
	EventVehicleEntered  EventType = "vehicle_entered"
	EventVehicleStaged   EventType = "vehicle_staged"
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

func (e *BaseEvent) GetType() EventType                 { return e.Type }
func (e *BaseEvent) GetTimestamp() time.Time            { return e.Timestamp }
func (e *BaseEvent) GetSource() ComponentID             { return e.Source }
func (e *BaseEvent) GetData() map[string]interface{}    { return e.Data }

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

// ComponentStatus represents the current state of a component
type ComponentStatus struct {
	ID        ComponentID            `json:"id"`
	Status    string                 `json:"status"` // ready, running, error, stopped
	LastError error                  `json:"last_error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Component represents any system component
type Component interface {
	GetID() ComponentID
	Initialize(ctx context.Context, bus EventBus, config Config) error
	Start(ctx context.Context) error
	Stop() error
	GetStatus() ComponentStatus
	HandleEvent(ctx context.Context, event Event) error
}

// Config holds system-wide configuration
type Config interface {
	GetTrackConfig() TrackConfig
	GetTimingConfig() TimingConfig
	GetTreeConfig() TreeSequenceConfig
	GetSafetyConfig() SafetyConfig
}

// TrackConfig defines track specifications
type TrackConfig struct {
	Length      float64           `json:"length"`        // Track length in feet
	LaneCount   int               `json:"lane_count"`    // Number of lanes
	LaneWidth   float64           `json:"lane_width"`    // Width of each lane
	BeamLayout  map[string]BeamConfig `json:"beam_layout"` // Beam positions
}

// BeamConfig defines timing beam specifications
type BeamConfig struct {
	Name     string  `json:"name"`
	Position float64 `json:"position"` // Distance from starting line
	Height   float64 `json:"height"`   // Height above track
	Lane     int     `json:"lane"`     // Which lane (0 = both)
}

// TimingConfig defines timing system parameters
type TimingConfig struct {
	Precision       time.Duration      `json:"precision"`     // Timing precision
	SpeedTrapLength float64            `json:"speed_trap_length"` // Speed trap distance
	AutoStart       bool               `json:"auto_start"`    // Auto-start timing on stage
}

// TreeSequenceType defines different starting sequences
type TreeSequenceType string

const (
	TreeSequencePro       TreeSequenceType = "pro"        // All ambers simultaneously
	TreeSequenceSportsman TreeSequenceType = "sportsman"  // Sequential ambers
)

// TreeSequenceConfig defines timing for tree sequences
type TreeSequenceConfig struct {
	Type           TreeSequenceType `json:"type"`
	AmberDelay     time.Duration    `json:"amber_delay"`     // Time between ambers (sportsman)
	GreenDelay     time.Duration    `json:"green_delay"`     // Time from last amber to green
	PreStageTimeout time.Duration   `json:"pre_stage_timeout"`
	StageTimeout   time.Duration    `json:"stage_timeout"`
}

// SafetyConfig defines safety system parameters
type SafetyConfig struct {
	EmergencyStopEnabled bool          `json:"emergency_stop_enabled"`
	MaxReactionTime      time.Duration `json:"max_reaction_time"`
	MinStagingTime       time.Duration `json:"min_staging_time"`
}

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
	LightOff     LightState = "off"
	LightOn      LightState = "on"
	LightBlink   LightState = "blink"
)

// TimingResults holds race timing data
type TimingResults struct {
	Lane             int                    `json:"lane"`
	StartTime        time.Time              `json:"start_time"`
	ReactionTime     *float64               `json:"reaction_time,omitempty"`
	SixtyFootTime    *float64               `json:"sixty_foot_time,omitempty"`
	EighthMileTime   *float64               `json:"eighth_mile_time,omitempty"`
	QuarterMileTime  *float64               `json:"quarter_mile_time,omitempty"`
	TrapSpeed        *float64               `json:"trap_speed,omitempty"`
	IsComplete       bool                   `json:"is_complete"`
	IsFoul           bool                   `json:"is_foul"`
	FoulReason       string                 `json:"foul_reason,omitempty"`
	BeamTriggers     map[string]time.Time   `json:"beam_triggers"`
}

// BeamStatus represents the state of a timing beam
type BeamStatus struct {
	ID          string    `json:"id"`
	Position    float64   `json:"position"`
	IsTriggered bool      `json:"is_triggered"`
	LastTrigger time.Time `json:"last_trigger,omitempty"`
	IsActive    bool      `json:"is_active"`
	Lane        int       `json:"lane"`
}

// TreeStatus represents christmas tree state
type TreeStatus struct {
	IsArmed        bool                         `json:"is_armed"`
	IsRunning      bool                         `json:"is_running"`
	SequenceType   TreeSequenceType             `json:"sequence_type"`
	CurrentStep    int                          `json:"current_step"`
	LightStates    map[int]map[LightType]LightState `json:"light_states"` // lane -> light -> state
	LastSequence   time.Time                    `json:"last_sequence,omitempty"`
}

// RaceState defines race progression states
type RaceState string

const (
	RaceStateIdle     RaceState = "idle"
	RaceStatePreparing RaceState = "preparing"
	RaceStateStaging  RaceState = "staging"
	RaceStateArmed    RaceState = "armed"
	RaceStateRunning  RaceState = "running"
	RaceStateComplete RaceState = "complete"
	RaceStateAborted  RaceState = "aborted"
	RaceStateError    RaceState = "error"
)

// RaceStatus represents overall race state
type RaceStatus struct {
	State       RaceState              `json:"state"`
	StartTime   time.Time              `json:"start_time,omitempty"`
	Components  map[ComponentID]ComponentStatus `json:"components"`
	ActiveLanes []int                  `json:"active_lanes"`
	LastError   error                  `json:"last_error,omitempty"`
}

// VehicleInterface defines vehicle monitoring
type VehicleInterface interface {
	Component
	GetLane() int
	IsStaged() bool
	GetPosition() float64
}

// =====================================
// EVENT BUS IMPLEMENTATION
// =====================================

// SimpleEventBus implements EventBus with in-memory message passing
type SimpleEventBus struct {
	mu           sync.RWMutex
	subscribers  map[EventType][]subscription
	allSubs      []subscription
	running      bool
	eventChannel chan Event
	ctx          context.Context
	cancel       context.CancelFunc
}

type subscription struct {
	id      string
	handler EventHandler
	active  bool
}

func (s *subscription) Unsubscribe() error {
	s.active = false
	return nil
}

func (s *subscription) IsActive() bool {
	return s.active
}

// NewSimpleEventBus creates a new event bus
func NewSimpleEventBus() *SimpleEventBus {
	return &SimpleEventBus{
		subscribers:  make(map[EventType][]subscription),
		allSubs:      make([]subscription, 0),
		eventChannel: make(chan Event, 1000), // Buffered channel
	}
}

func (bus *SimpleEventBus) Publish(ctx context.Context, event Event) error {
	select {
	case bus.eventChannel <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event bus full, dropping event: %s", event.GetType())
	}
}

func (bus *SimpleEventBus) Subscribe(eventType EventType, handler EventHandler) Subscription {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	sub := subscription{
		id:      fmt.Sprintf("%s_%d", eventType, time.Now().UnixNano()),
		handler: handler,
		active:  true,
	}
	
	bus.subscribers[eventType] = append(bus.subscribers[eventType], sub)
	return &sub
}

func (bus *SimpleEventBus) SubscribeAll(handler EventHandler) Subscription {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	sub := subscription{
		id:      fmt.Sprintf("all_%d", time.Now().UnixNano()),
		handler: handler,
		active:  true,
	}
	
	bus.allSubs = append(bus.allSubs, sub)
	return &sub
}

func (bus *SimpleEventBus) Start(ctx context.Context) error {
	bus.ctx, bus.cancel = context.WithCancel(ctx)
	bus.running = true
	
	go bus.eventLoop()
	return nil
}

func (bus *SimpleEventBus) Stop() error {
	if bus.cancel != nil {
		bus.cancel()
	}
	bus.running = false
	return nil
}

func (bus *SimpleEventBus) eventLoop() {
	for {
		select {
		case event := <-bus.eventChannel:
			bus.deliverEvent(event)
		case <-bus.ctx.Done():
			return
		}
	}
}

func (bus *SimpleEventBus) deliverEvent(event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	// Deliver to specific event type subscribers
	if subs, exists := bus.subscribers[event.GetType()]; exists {
		for _, sub := range subs {
			if sub.active {
				go func(s subscription) {
					if err := s.handler(bus.ctx, event); err != nil {
						fmt.Printf("⚠️  Error handling event %s: %v\n", event.GetType(), err)
					}
				}(sub)
			}
		}
	}
	
	// Deliver to "all events" subscribers
	for _, sub := range bus.allSubs {
		if sub.active {
			go func(s subscription) {
				if err := s.handler(bus.ctx, event); err != nil {
					fmt.Printf("⚠️  Error handling event %s: %v\n", event.GetType(), err)
				}
			}(sub)
		}
	}
}

// =====================================
// CONFIG IMPLEMENTATION
// =====================================

// DefaultConfig implements Config interface
type DefaultConfig struct {
	track  TrackConfig
	timing TimingConfig
	tree   TreeSequenceConfig
	safety SafetyConfig
}

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		track: TrackConfig{
			Length:    1320, // Quarter mile
			LaneCount: 2,
			LaneWidth: 12,
			BeamLayout: map[string]BeamConfig{
				"pre_stage_L":    {Name: "Pre-Stage Left", Position: -0.583, Lane: 1},
				"stage_L":        {Name: "Stage Left", Position: 0, Lane: 1},
				"sixty_foot_L":   {Name: "60-Foot Left", Position: 60, Lane: 1},
				"eighth_mile_L":  {Name: "1/8 Mile Left", Position: 660, Lane: 1},
				"quarter_mile_L": {Name: "1/4 Mile Left", Position: 1320, Lane: 1},
				"pre_stage_R":    {Name: "Pre-Stage Right", Position: -0.583, Lane: 2},
				"stage_R":        {Name: "Stage Right", Position: 0, Lane: 2},
				"sixty_foot_R":   {Name: "60-Foot Right", Position: 60, Lane: 2},
				"eighth_mile_R":  {Name: "1/8 Mile Right", Position: 660, Lane: 2},
				"quarter_mile_R": {Name: "1/4 Mile Right", Position: 1320, Lane: 2},
			},
		},
		timing: TimingConfig{
			Precision:       time.Millisecond,
			SpeedTrapLength: 66,
			AutoStart:       true,
		},
		tree: TreeSequenceConfig{
			Type:            TreeSequencePro,
			AmberDelay:      500 * time.Millisecond,
			GreenDelay:      400 * time.Millisecond,
			PreStageTimeout: 30 * time.Second,
			StageTimeout:    10 * time.Second,
		},
		safety: SafetyConfig{
			EmergencyStopEnabled: true,
			MaxReactionTime:      2 * time.Second,
			MinStagingTime:       100 * time.Millisecond,
		},
	}
}

func (dc *DefaultConfig) GetTrackConfig() TrackConfig       { return dc.track }
func (dc *DefaultConfig) GetTimingConfig() TimingConfig     { return dc.timing }
func (dc *DefaultConfig) GetTreeConfig() TreeSequenceConfig { return dc.tree }
func (dc *DefaultConfig) GetSafetyConfig() SafetyConfig     { return dc.safety }

// =====================================
// TIMING SYSTEM IMPLEMENTATION
// =====================================

// TimingBeam represents a physical timing beam
type TimingBeam struct {
	ID          string
	Position    float64
	Lane        int
	IsTriggered bool
	LastTrigger time.Time
	IsActive    bool
}

// TimingSystem implements the timing system component
type TimingSystem struct {
	id       ComponentID
	bus      EventBus
	config   Config
	mu       sync.RWMutex
	beams    map[string]*TimingBeam
	results  map[int]*TimingResults
	running  bool
	status   ComponentStatus
}

func NewTimingSystem() *TimingSystem {
	return &TimingSystem{
		id:      ComponentTimingSystem,
		beams:   make(map[string]*TimingBeam),
		results: make(map[int]*TimingResults),
		status: ComponentStatus{
			ID:       ComponentTimingSystem,
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
	}
}

func (ts *TimingSystem) GetID() ComponentID {
	return ts.id
}

func (ts *TimingSystem) Initialize(ctx context.Context, bus EventBus, config Config) error {
	ts.bus = bus
	ts.config = config
	
	// Initialize beams from config
	trackConfig := config.GetTrackConfig()
	for beamID, beamConfig := range trackConfig.BeamLayout {
		ts.beams[beamID] = &TimingBeam{
			ID:       beamID,
			Position: beamConfig.Position,
			Lane:     beamConfig.Lane,
			IsActive: true,
		}
	}
	
	// Subscribe to relevant events
	ts.bus.Subscribe(EventRaceStarted, ts.handleRaceStart)
	ts.bus.Subscribe(EventVehicleEntered, ts.handleVehicleEnter)
	
	ts.status.Status = "ready"
	return nil
}

func (ts *TimingSystem) Start(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	ts.running = true
	ts.status.Status = "running"
	
	// Start beam monitoring simulation
	go ts.simulateBeamTriggers(ctx)
	
	return nil
}

func (ts *TimingSystem) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	ts.running = false
	ts.status.Status = "stopped"
	return nil
}

func (ts *TimingSystem) GetStatus() ComponentStatus {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.status
}

func (ts *TimingSystem) HandleEvent(ctx context.Context, event Event) error {
	switch event.GetType() {
	case EventRaceStarted:
		return ts.handleRaceStart(ctx, event)
	case EventVehicleEntered:
		return ts.handleVehicleEnter(ctx, event)
	}
	return nil
}

func (ts *TimingSystem) handleRaceStart(ctx context.Context, event Event) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	fmt.Println("🔍 libdrag Timing System: Race started, resetting timers")
	
	// Reset timing results
	ts.results = make(map[int]*TimingResults)
	
	// Reset beam states
	for _, beam := range ts.beams {
		beam.IsTriggered = false
		beam.LastTrigger = time.Time{}
	}
	
	return nil
}

func (ts *TimingSystem) handleVehicleEnter(ctx context.Context, event Event) error {
	data := event.GetData()
	if lanes, ok := data["lanes"].([]int); ok {
		for _, lane := range lanes {
			ts.mu.Lock()
			ts.results[lane] = &TimingResults{
				Lane:         lane,
				StartTime:    time.Now(),
				BeamTriggers: make(map[string]time.Time),
				IsComplete:   false,
				IsFoul:       false,
			}
			ts.mu.Unlock()
		}
	}
	return nil
}

// simulateBeamTriggers simulates realistic beam triggers for demonstration
func (ts *TimingSystem) simulateBeamTriggers(ctx context.Context) {
	time.Sleep(2 * time.Second) // Wait for race to start
	
	if !ts.running {
		return
	}
	
	fmt.Println("🚗 libdrag: Simulating vehicle beam triggers...")
	
	// Simulate pre-stage triggers
	ts.triggerBeam(ctx, "pre_stage_L", 1)
	time.Sleep(200 * time.Millisecond)
	ts.triggerBeam(ctx, "pre_stage_R", 2)
	
	time.Sleep(500 * time.Millisecond)
	
	// Simulate stage triggers
	ts.triggerBeam(ctx, "stage_L", 1)
	time.Sleep(100 * time.Millisecond)
	ts.triggerBeam(ctx, "stage_R", 2)
	
	// Wait for green light (approximately)
	time.Sleep(1 * time.Second)
	
	// Simulate race progression with realistic times
	beamSequence := []struct {
		beamL, beamR string
		delayL, delayR time.Duration
	}{
		{"sixty_foot_L", "sixty_foot_R", 950 * time.Millisecond, 980 * time.Millisecond},
		{"eighth_mile_L", "eighth_mile_R", 4200 * time.Millisecond, 4350 * time.Millisecond},
		{"quarter_mile_L", "quarter_mile_R", 3100 * time.Millisecond, 3250 * time.Millisecond},
	}
	
	for _, seq := range beamSequence {
		go func(beamL string, delayL time.Duration, lane int) {
			time.Sleep(delayL)
			if ts.running {
				ts.triggerBeam(ctx, beamL, lane)
			}
		}(seq.beamL, seq.delayL, 1)
		
		go func(beamR string, delayR time.Duration, lane int) {
			time.Sleep(delayR)
			if ts.running {
				ts.triggerBeam(ctx, beamR, lane)
			}
		}(seq.beamR, seq.delayR, 2)
	}
}

func (ts *TimingSystem) triggerBeam(ctx context.Context, beamID string, lane int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if beam, exists := ts.beams[beamID]; exists {
		beam.IsTriggered = true
		beam.LastTrigger = time.Now()
		
		fmt.Printf("⚡ libdrag: Beam triggered: %s (Lane %d) at %.3fs\n", 
			beamID, lane, time.Since(ts.results[lane].StartTime).Seconds())
		
		// Record timing event
		if result, exists := ts.results[lane]; exists {
			result.BeamTriggers[beamID] = beam.LastTrigger
			ts.updateTimingMilestones(beamID, result)
		}
		
		// Publish beam trigger event
		event := &BaseEvent{
			Type:      EventBeamTriggered,
			Timestamp: time.Now(),
			Source:    ts.id,
			Data: map[string]interface{}{
				"beam_id":  beamID,
				"lane":     lane,
				"position": beam.Position,
			},
		}
		
		// Special handling for staging beams
		var eventType EventType
		switch beamID {
		case "pre_stage_L", "pre_stage_R":
			eventType = EventPreStageOn
		case "stage_L", "stage_R":
			eventType = EventStageOn
		default:
			eventType = EventBeamTriggered
		}
		
		if eventType != EventBeamTriggered {
			stageEvent := &BaseEvent{
				Type:      eventType,
				Timestamp: time.Now(),
				Source:    ts.id,
				Data: map[string]interface{}{
					"lane": lane,
				},
			}
			ts.bus.Publish(ctx, stageEvent)
		}
		
		ts.bus.Publish(ctx, event)
	}
}

func (ts *TimingSystem) updateTimingMilestones(beamID string, result *TimingResults) {
	elapsed := time.Since(result.StartTime).Seconds()
	
	switch beamID {
	case "stage_L", "stage_R":
		if result.ReactionTime == nil {
			result.ReactionTime = &elapsed
		}
	case "sixty_foot_L", "sixty_foot_R":
		if result.SixtyFootTime == nil {
			result.SixtyFootTime = &elapsed
		}
	case "eighth_mile_L", "eighth_mile_R":
		if result.EighthMileTime == nil {
			result.EighthMileTime = &elapsed
		}
	case "quarter_mile_L", "quarter_mile_R":
		if result.QuarterMileTime == nil {
			result.QuarterMileTime = &elapsed
			// Calculate trap speed (simplified)
			speed := 120.0 + rand.Float64()*40.0 // Random speed between 120-160 mph
			result.TrapSpeed = &speed
			result.IsComplete = true
			
			// Publish run complete
			event := &BaseEvent{
				Type:      EventRunComplete,
				Timestamp: time.Now(),
				Source:    ts.id,
				Data: map[string]interface{}{
					"lane":    result.Lane,
					"results": result,
				},
			}
			ts.bus.Publish(context.Background(), event)
		}
	}
}

func (ts *TimingSystem) GetResults(lane int) *TimingResults {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	if result, exists := ts.results[lane]; exists {
		return result
	}
	return nil
}

// =====================================
// CHRISTMAS TREE IMPLEMENTATION
// =====================================

// ChristmasTree implements the christmas tree component
type ChristmasTree struct {
	id       ComponentID
	bus      EventBus
	config   Config
	mu       sync.RWMutex
	status   TreeStatus
	compStatus ComponentStatus
	running  bool
	lanesPreStaged map[int]bool
	lanesStaged    map[int]bool
}

func NewChristmasTree() *ChristmasTree {
	return &ChristmasTree{
		id: ComponentChristmasTree,
		status: TreeStatus{
			IsArmed:     false,
			IsRunning:   false,
			LightStates: make(map[int]map[LightType]LightState),
		},
		compStatus: ComponentStatus{
			ID:       ComponentChristmasTree,
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
		lanesPreStaged: make(map[int]bool),
		lanesStaged:    make(map[int]bool),
	}
}

func (ct *ChristmasTree) GetID() ComponentID {
	return ct.id
}

func (ct *ChristmasTree) Initialize(ctx context.Context, bus EventBus, config Config) error {
	ct.bus = bus
	ct.config = config
	
	// Subscribe to timing system events
	ct.bus.Subscribe(EventPreStageOn, ct.handlePreStage)
	ct.bus.Subscribe(EventStageOn, ct.handleStage)
	ct.bus.Subscribe(EventRaceStarted, ct.handleRaceStart)
	
	// Initialize light states for all lanes
	trackConfig := config.GetTrackConfig()
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

func (ct *ChristmasTree) GetStatus() ComponentStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.compStatus
}

func (ct *ChristmasTree) HandleEvent(ctx context.Context, event Event) error {
	switch event.GetType() {
	case EventPreStageOn:
		return ct.handlePreStage(ctx, event)
	case EventStageOn:
		return ct.handleStage(ctx, event)
	case EventRaceStarted:
		return ct.handleRaceStart(ctx, event)
	}
	return nil
}

func (ct *ChristmasTree) handlePreStage(ctx context.Context, event Event) error {
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
		event := &BaseEvent{
			Type:      EventRaceArmed,
			Timestamp: time.Now(),
			Source:    ct.id,
			Data:      map[string]interface{}{"all_lanes_pre_staged": true},
		}
		ct.bus.Publish(ctx, event)
	}
	
	return nil
}

func (ct *ChristmasTree) handleStage(ctx context.Context, event Event) error {
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
			ct.handleRaceStart(ctx, &BaseEvent{
				Type:      EventRaceStarted,
				Timestamp: time.Now(),
				Source:    ComponentStarterControl,
				Data:      map[string]interface{}{"sequence_type": TreeSequencePro},
			})
		}()
	}
	
	return nil
}

func (ct *ChristmasTree) handleRaceStart(ctx context.Context, event Event) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	
	if !ct.status.IsArmed {
		return fmt.Errorf("tree is not armed")
	}
	
	if ct.status.IsRunning {
		return fmt.Errorf("sequence already running")
	}
	
	data := event.GetData()
	sequenceType, ok := data["sequence_type"].(TreeSequenceType)
	if !ok {
		sequenceType = TreeSequencePro // Default
	}
	
	ct.status.IsRunning = true
	ct.status.SequenceType = sequenceType
	ct.status.LastSequence = time.Now()
	
	fmt.Printf("🎄 libdrag: Starting %s sequence\n", sequenceType)
	
	// Start the sequence in a goroutine
	go ct.runSequence(ctx, sequenceType)
	
	// Publish sequence start event
	event = &BaseEvent{
		Type:      EventTreeSequenceStart,
		Timestamp: time.Now(),
		Source:    ct.id,
		Data: map[string]interface{}{
			"sequence_type": sequenceType,
		},
	}
	ct.bus.Publish(ctx, event)
	
	return nil
}

func (ct *ChristmasTree) runSequence(ctx context.Context, sequenceType TreeSequenceType) {
	defer func() {
		ct.mu.Lock()
		ct.status.IsRunning = false
		ct.mu.Unlock()
	}()
	
	config := ct.config.GetTreeConfig()
	
	switch sequenceType {
	case TreeSequencePro:
		ct.runProSequence(ctx, config)
	case TreeSequenceSportsman:
		ct.runSportsmanSequence(ctx, config)
	default:
		ct.runProSequence(ctx, config)
	}
}

func (ct *ChristmasTree) runProSequence(ctx context.Context, config TreeSequenceConfig) {
	fmt.Println("🟡🟡🟡 libdrag: All three ambers ON")
	
	// All three ambers simultaneously
	ct.setAllLights(LightAmber1, LightOn)
	ct.setAllLights(LightAmber2, LightOn)
	ct.setAllLights(LightAmber3, LightOn)
	
	ct.publishLightEvent(ctx, EventAmberLight, map[string]interface{}{
		"lights": []LightType{LightAmber1, LightAmber2, LightAmber3},
		"state":  LightOn,
	})
	
	// Wait for green delay
	time.Sleep(config.GreenDelay)
	
	// Turn off ambers and turn on green
	ct.setAllLights(LightAmber1, LightOff)
	ct.setAllLights(LightAmber2, LightOff)
	ct.setAllLights(LightAmber3, LightOff)
	ct.setAllLights(LightGreen, LightOn)
	
	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")
	
	ct.publishLightEvent(ctx, EventGreenLight, map[string]interface{}{
		"all_lanes": true,
	})
}

func (ct *ChristmasTree) runSportsmanSequence(ctx context.Context, config TreeSequenceConfig) {
	// Sequential ambers
	amberLights := []LightType{LightAmber1, LightAmber2, LightAmber3}
	
	for i, light := range amberLights {
		fmt.Printf("🟡 libdrag: Amber %d ON\n", i+1)
		ct.setAllLights(light, LightOn)
		ct.publishLightEvent(ctx, EventAmberLight, map[string]interface{}{
			"light": light,
			"state": LightOn,
		})
		
		time.Sleep(config.AmberDelay)
	}
	
	// Wait for green delay
	time.Sleep(config.GreenDelay)
	
	// Turn off all ambers and turn on green
	for _, light := range amberLights {
		ct.setAllLights(light, LightOff)
	}
	ct.setAllLights(LightGreen, LightOn)
	
	fmt.Println("🟢 libdrag: GREEN LIGHT! GO GO GO!")
	
	ct.publishLightEvent(ctx, EventGreenLight, map[string]interface{}{
		"all_lanes": true,
	})
}

func (ct *ChristmasTree) setAllLights(lightType LightType, state LightState) {
	for lane := range ct.status.LightStates {
		ct.status.LightStates[lane][lightType] = state
	}
}

func (ct *ChristmasTree) publishLightEvent(ctx context.Context, eventType EventType, data map[string]interface{}) {
	event := &BaseEvent{
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

// =====================================
// SIMPLE VEHICLE IMPLEMENTATION
// =====================================

// SimpleVehicle implements a basic vehicle for testing
type SimpleVehicle struct {
	id     ComponentID
	lane   int
	staged bool
	position float64
	status ComponentStatus
}

func NewSimpleVehicle(lane int) *SimpleVehicle {
	return &SimpleVehicle{
		id:   ComponentID(fmt.Sprintf("vehicle_%d", lane)),
		lane: lane,
		status: ComponentStatus{
			ID:     ComponentID(fmt.Sprintf("vehicle_%d", lane)),
			Status: "ready",
			Metadata: make(map[string]interface{}),
		},
	}
}

func (sv *SimpleVehicle) GetID() ComponentID { return sv.id }
func (sv *SimpleVehicle) GetLane() int { return sv.lane }
func (sv *SimpleVehicle) IsStaged() bool { return sv.staged }
func (sv *SimpleVehicle) GetPosition() float64 { return sv.position }
func (sv *SimpleVehicle) GetStatus() ComponentStatus { return sv.status }

func (sv *SimpleVehicle) Initialize(ctx context.Context, bus EventBus, config Config) error {
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

func (sv *SimpleVehicle) HandleEvent(ctx context.Context, event Event) error {
	return nil
}

// =====================================
// RACE ORCHESTRATOR IMPLEMENTATION
// =====================================

// RaceOrchestrator coordinates all race components
type RaceOrchestrator struct {
	mu          sync.RWMutex
	bus         EventBus
	config      Config
	components  map[ComponentID]Component
	status      RaceStatus
	eventLog    []Event
	timingSystem *TimingSystem
	christmasTree *ChristmasTree
}

func NewRaceOrchestrator() *RaceOrchestrator {
	return &RaceOrchestrator{
		components: make(map[ComponentID]Component),
		status: RaceStatus{
			State:       RaceStateIdle,
			Components:  make(map[ComponentID]ComponentStatus),
			ActiveLanes: make([]int, 0),
		},
		eventLog: make([]Event, 0),
	}
}

func (ro *RaceOrchestrator) Initialize(ctx context.Context, components []Component, config Config) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	
	ro.config = config
	ro.bus = NewSimpleEventBus()
	
	// Start event bus
	if err := ro.bus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %v", err)
	}
	
	// Subscribe to all events for logging
	ro.bus.SubscribeAll(ro.logEvent)
	
	// Initialize all components
	for _, component := range components {
		ro.components[component.GetID()] = component
		
		// Store references to key components
		switch comp := component.(type) {
		case *TimingSystem:
			ro.timingSystem = comp
		case *ChristmasTree:
			ro.christmasTree = comp
		}
		
		if err := component.Initialize(ctx, ro.bus, config); err != nil {
			return fmt.Errorf("failed to initialize component %s: %v", component.GetID(), err)
		}
		
		if err := component.Start(ctx); err != nil {
			return fmt.Errorf("failed to start component %s: %v", component.GetID(), err)
		}
		
		ro.status.Components[component.GetID()] = component.GetStatus()
	}
	
	// Subscribe to race completion events
	ro.bus.Subscribe(EventRunComplete, ro.handleRunComplete)
	
	ro.status.State = RaceStateIdle
	
	// Publish system ready event
	event := &BaseEvent{
		Type:      EventSystemReady,
		Timestamp: time.Now(),
		Source:    ComponentOrchestrator,
		Data:      map[string]interface{}{"components_count": len(components)},
	}
	ro.bus.Publish(ctx, event)
	
	return nil
}

func (ro *RaceOrchestrator) StartRace(leftVehicle, rightVehicle VehicleInterface) error {
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
	event := &BaseEvent{
		Type:      EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    ComponentOrchestrator,
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
	status.Components = make(map[ComponentID]ComponentStatus)
	for id, component := range ro.components {
		status.Components[id] = component.GetStatus()
	}
	
	return status
}

func (ro *RaceOrchestrator) GetResults() map[int]*TimingResults {
	if ro.timingSystem == nil {
		return nil
	}
	
	results := make(map[int]*TimingResults)
	results[1] = ro.timingSystem.GetResults(1)
	results[2] = ro.timingSystem.GetResults(2)
	
	return results
}

func (ro *RaceOrchestrator) GetTreeStatus() TreeStatus {
	if ro.christmasTree == nil {
		return TreeStatus{}
	}
	return ro.christmasTree.GetTreeStatus()
}

func (ro *RaceOrchestrator) Stop() {
	for _, component := range ro.components {
		component.Stop()
	}
	if ro.bus != nil {
		ro.bus.Stop()
	}
}

func (ro *RaceOrchestrator) logEvent(ctx context.Context, event Event) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	
	ro.eventLog = append(ro.eventLog, event)
	return nil
}

func (ro *RaceOrchestrator) handleRunComplete(ctx context.Context, event Event) error {
	data := event.GetData()
	result, ok := data["results"].(*TimingResults)
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

// =====================================
// LIBDRAG API FOR MOBILE INTEGRATION
// =====================================

// LibDragAPI provides a mobile-friendly interface
type LibDragAPI struct {
	orchestrator *RaceOrchestrator
	mu           sync.RWMutex
}

func NewLibDragAPI() *LibDragAPI {
	return &LibDragAPI{
		orchestrator: NewRaceOrchestrator(),
	}
}

// Initialize the libdrag system
func (api *LibDragAPI) Initialize() error {
	// Create configuration
	config := NewDefaultConfig()
	
	// Create components
	timingSystem := NewTimingSystem()
	christmasTree := NewChristmasTree()
	
	components := []Component{
		timingSystem,
		christmasTree,
	}
	
	// Initialize system
	ctx := context.Background()
	return api.orchestrator.Initialize(ctx, components, config)
}

// StartRace starts a new drag race
func (api *LibDragAPI) StartRace() error {
	leftVehicle := NewSimpleVehicle(1)
	rightVehicle := NewSimpleVehicle(2)
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
	return status.State == RaceStateComplete
}

// =====================================
// MAIN DEMONSTRATION PROGRAM
// =====================================

func main() {
	fmt.Println("🏁 LIBDRAG - DRAG RACING LIBRARY DEMONSTRATION")
	fmt.Println("===============================================")
	
	// Create the libdrag API
	api := NewLibDragAPI()
	
	// Initialize system
	fmt.Println("📊 Initializing libdrag system...")
	if err := api.Initialize(); err != nil {
		fmt.Printf("❌ Failed to initialize libdrag: %v\n", err)
		return
	}
	
	fmt.Println("✅ libdrag system initialized successfully")
	
	// Start race
	fmt.Println("\n🚗 Starting race with libdrag...")
	if err := api.StartRace(); err != nil {
		fmt.Printf("❌ Failed to start race: %v\n", err)
		return
	}
	
	// Monitor race progress
	fmt.Println("🔄 Monitoring race progress...")
	
	// Wait for race to complete
	for i := 0; i < 100; i++ { // Max 10 seconds
		if api.IsRaceComplete() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// Display final results
	fmt.Println("\n🏆 LIBDRAG FINAL RESULTS")
	fmt.Println("========================")
	
	resultsJSON := api.GetResultsJSON()
	fmt.Printf("Results JSON:\n%s\n", resultsJSON)
	
	treeStatusJSON := api.GetTreeStatusJSON()
	fmt.Printf("\nChristmas Tree Status JSON:\n%s\n", treeStatusJSON)
	
	// Clean shutdown
	fmt.Println("🛑 Shutting down libdrag system...")
	api.Stop()
	
	fmt.Println("✨ libdrag demo completed successfully!")
}

// =====================================
// EXPORT FOR MOBILE BINDINGS
// =====================================

// These functions will be available when building mobile libraries

// NewLibDrag creates a new libdrag instance (for mobile)
func NewLibDrag() *LibDragAPI {
	return NewLibDragAPI()
}

// Version returns the libdrag version
func Version() string {
	return "libdrag v1.0.0 - Professional Drag Racing Library"
}
