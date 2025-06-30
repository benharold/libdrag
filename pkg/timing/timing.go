package timing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
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
	id             string
	config         config.Config
	mu             sync.RWMutex
	beams          map[string]*TimingBeam
	results        map[int]*TimingResults
	running        bool
	status         component.ComponentStatus
	raceID         string
	testMode       bool
	greenLightTime time.Time
}

func NewTimingSystem() *TimingSystem {
	return NewTimingSystemWithRaceID("")
}

func NewTimingSystemWithRaceID(raceID string) *TimingSystem {
	return &TimingSystem{
		id:       "timing_system",
		beams:    make(map[string]*TimingBeam),
		results:  make(map[int]*TimingResults),
		raceID:   raceID,
		testMode: false,
		status: component.ComponentStatus{
			ID:       "timing_system",
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
	}
}

// SetTestMode enables or disables test mode (fast execution)
func (ts *TimingSystem) SetTestMode(enabled bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.testMode = enabled
}

func (ts *TimingSystem) GetID() string {
	return ts.id
}

func (ts *TimingSystem) Initialize(ctx context.Context, cfg config.Config) error {
	ts.config = cfg

	// Initialize beams from config
	trackConfig := cfg.Track()
	for beamID, beamConfig := range trackConfig.BeamLayout {
		ts.beams[beamID] = &TimingBeam{
			ID:       beamID,
			Position: beamConfig.Position,
			Lane:     beamConfig.Lane,
			IsActive: true,
		}
	}

	ts.status.Status = "ready"
	return nil
}

func (ts *TimingSystem) Start(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.running = true
	ts.status.Status = "running"
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

// Direct methods to replace event handling
func (ts *TimingSystem) StartRace() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	fmt.Println("🔍 libdrag Timing System: Race started, resetting timers")

	// Reset timing results
	ts.results = make(map[int]*TimingResults)
	ts.greenLightTime = time.Time{}

	// Reset beam states
	for _, beam := range ts.beams {
		beam.IsTriggered = false
		beam.LastTrigger = time.Time{}
	}
}

func (ts *TimingSystem) AddVehicles(lanes []int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, lane := range lanes {
		ts.results[lane] = &TimingResults{
			Lane:         lane,
			StartTime:    time.Time{}, // Will be set when vehicle actually starts
			BeamTriggers: make(map[string]time.Time),
			IsComplete:   false,
			IsFoul:       false,
		}
	}
}

func (ts *TimingSystem) SetGreenLight(greenTime time.Time) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.greenLightTime = greenTime
	fmt.Printf("🟢 libdrag Timing System: Green light at %v\n", ts.greenLightTime)

	// Check for existing early starts (red light fouls)
	for _, result := range ts.results {
		if !result.StartTime.IsZero() {
			// Vehicle already left starting line before green light
			reactionTime := result.StartTime.Sub(ts.greenLightTime).Seconds()
			result.ReactionTime = &reactionTime

			if reactionTime < 0 {
				result.IsFoul = true
				result.FoulReason = "red_light"
				fmt.Printf("🚨 libdrag: Red light foul detected for lane %d (RT: %.3fs)\n", result.Lane, reactionTime)
			}
		}
	}
}

func (ts *TimingSystem) TriggerBeam(beamID string, lane int, triggerTime time.Time) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Update beam state
	if beam, exists := ts.beams[beamID]; exists {
		beam.IsTriggered = true
		beam.LastTrigger = triggerTime
	}

	// Update timing results if lane exists
	if result, exists := ts.results[lane]; exists {
		result.BeamTriggers[beamID] = triggerTime

		// Calculate timing splits based on beam
		switch beamID {
		case "stage":
			// Vehicle left starting line - calculate reaction time
			if !ts.greenLightTime.IsZero() {
				reactionTime := triggerTime.Sub(ts.greenLightTime).Seconds()
				result.ReactionTime = &reactionTime
				result.StartTime = triggerTime

				// Check for red light (negative reaction time)
				if reactionTime < 0 {
					result.IsFoul = true
					result.FoulReason = "red_light"
				}
			} else {
				// No green light yet - set start time for later calculation
				result.StartTime = triggerTime
			}

		case "60_foot":
			// Calculate 60-foot time from start line
			if !result.StartTime.IsZero() {
				sixtyFootTime := triggerTime.Sub(result.StartTime).Seconds()
				result.SixtyFootTime = &sixtyFootTime
			}

		case "330_foot":
			// Calculate 330-foot time from start line
			if !result.StartTime.IsZero() {
				time330 := triggerTime.Sub(result.StartTime).Seconds()
				// Note: Could add ThreeThirtyFootTime field if needed
				_ = time330
			}

		case "660_foot":
			// Calculate eighth-mile time from start line
			if !result.StartTime.IsZero() {
				eighthMileTime := triggerTime.Sub(result.StartTime).Seconds()
				result.EighthMileTime = &eighthMileTime
			}

		case "1320_foot":
			// Calculate quarter-mile time from start line
			if !result.StartTime.IsZero() {
				quarterMileTime := triggerTime.Sub(result.StartTime).Seconds()
				result.QuarterMileTime = &quarterMileTime
				result.IsComplete = true

				// Calculate trap speed (simplified calculation)
				trapSpeed := 1320.0 / quarterMileTime * 0.681818 // Convert ft/s to mph
				result.TrapSpeed = &trapSpeed
			}
		}

		fmt.Printf("🏁 libdrag Timing: Lane %d triggered %s beam at %v\n", lane, beamID, triggerTime)
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

func (ts *TimingSystem) GetAllResults() map[int]*TimingResults {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Return a copy to avoid race conditions
	results := make(map[int]*TimingResults)
	for lane, result := range ts.results {
		results[lane] = result
	}
	return results
}
