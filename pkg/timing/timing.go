package timing

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// TimingResults holds race timing data
type TimingResults struct {
	Lane            int                  `json:"lane"`
	StartTime       time.Time            `json:"start_time"`
	ReactionTime    *float64             `json:"reaction_time,omitempty"`
	SixtyFootTime   *float64             `json:"sixty_foot_time,omitempty"`
	EighthMileTime  *float64             `json:"eighth_mile_time,omitempty"`
	QuarterMileTime *float64             `json:"quarter_mile_time,omitempty"`
	TrapSpeed       *float64             `json:"trap_speed,omitempty"`
	IsComplete      bool                 `json:"is_complete"`
	IsFoul          bool                 `json:"is_foul"`
	FoulReason      string               `json:"foul_reason,omitempty"`
	BeamTriggers    map[string]time.Time `json:"beam_triggers"`
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
	id      events.ComponentID
	bus     events.EventBus
	config  config.Config
	mu      sync.RWMutex
	beams   map[string]*TimingBeam
	results map[int]*TimingResults
	running bool
	status  component.ComponentStatus
}

func NewTimingSystem() *TimingSystem {
	return &TimingSystem{
		id:      events.ComponentTimingSystem,
		beams:   make(map[string]*TimingBeam),
		results: make(map[int]*TimingResults),
		status: component.ComponentStatus{
			ID:       events.ComponentTimingSystem,
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
	}
}

func (ts *TimingSystem) GetID() events.ComponentID {
	return ts.id
}

func (ts *TimingSystem) Initialize(ctx context.Context, bus events.EventBus, cfg config.Config) error {
	ts.bus = bus
	ts.config = cfg

	// Initialize beams from config
	trackConfig := cfg.GetTrackConfig()
	for beamID, beamConfig := range trackConfig.BeamLayout {
		ts.beams[beamID] = &TimingBeam{
			ID:       beamID,
			Position: beamConfig.Position,
			Lane:     beamConfig.Lane,
			IsActive: true,
		}
	}

	// Subscribe to relevant events
	ts.bus.Subscribe(events.EventRaceStarted, ts.handleRaceStart)
	ts.bus.Subscribe(events.EventVehicleEntered, ts.handleVehicleEnter)

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

func (ts *TimingSystem) GetStatus() component.ComponentStatus {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.status
}

func (ts *TimingSystem) HandleEvent(ctx context.Context, event events.Event) error {
	switch event.GetType() {
	case events.EventRaceStarted:
		return ts.handleRaceStart(ctx, event)
	case events.EventVehicleEntered:
		return ts.handleVehicleEnter(ctx, event)
	}
	return nil
}

func (ts *TimingSystem) handleRaceStart(ctx context.Context, event events.Event) error {
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

func (ts *TimingSystem) handleVehicleEnter(ctx context.Context, event events.Event) error {
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
		beamL, beamR   string
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
		event := &events.BaseEvent{
			Type:      events.EventBeamTriggered,
			Timestamp: time.Now(),
			Source:    ts.id,
			Data: map[string]interface{}{
				"beam_id":  beamID,
				"lane":     lane,
				"position": beam.Position,
			},
		}

		// Special handling for staging beams
		var eventType events.EventType
		switch beamID {
		case "pre_stage_L", "pre_stage_R":
			eventType = events.EventPreStageOn
		case "stage_L", "stage_R":
			eventType = events.EventStageOn
		default:
			eventType = events.EventBeamTriggered
		}

		if eventType != events.EventBeamTriggered {
			stageEvent := &events.BaseEvent{
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
			event := &events.BaseEvent{
				Type:      events.EventRunComplete,
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
