package beam

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// BeamID represents a specific beam identifier
type BeamID string

// Standard beam identifiers
const (
	BeamPreStage  BeamID = "pre_stage"
	BeamStage     BeamID = "stage"
	Beam60Foot    BeamID = "60_foot"
	Beam330Foot   BeamID = "330_foot"
	Beam660Foot   BeamID = "660_foot"   // 1/8 mile
	Beam1000Foot  BeamID = "1000_foot"  // 1/8 mile speed
	Beam1320Foot  BeamID = "1320_foot"  // 1/4 mile
	BeamSpeedTrap BeamID = "speed_trap" // 1/4 mile speed
)

// BeamState represents the current state of a beam
type BeamState struct {
	BeamID     BeamID    `json:"beam_id"`
	Lane       int       `json:"lane"`
	Position   float64   `json:"position"`
	IsBroken   bool      `json:"is_broken"`
	LastChange time.Time `json:"last_change"`
}

// BeamSystem manages all timing beams on the track
type BeamSystem struct {
	id       string
	mu       sync.RWMutex
	beams    map[int]map[BeamID]*BeamState // lane -> beamID -> state
	config   config.Config
	eventBus *events.EventBus
	raceID   string
	status   component.ComponentStatus
}

// NewBeamSystem creates a new beam system
func NewBeamSystem(eventBus *events.EventBus) *BeamSystem {
	return &BeamSystem{
		id:       "beam_system",
		beams:    make(map[int]map[BeamID]*BeamState),
		eventBus: eventBus,
		status: component.ComponentStatus{
			ID:       "beam_system",
			Status:   "stopped",
			Metadata: make(map[string]interface{}),
		},
	}
}

// GetID returns the component ID
func (bs *BeamSystem) GetID() string {
	return bs.id
}

// Initialize sets up the beam system with track configuration
func (bs *BeamSystem) Initialize(ctx context.Context, cfg config.Config) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.config = cfg
	trackConfig := cfg.Track()

	// Initialize beams for each lane based on beam layout
	for lane := 1; lane <= trackConfig.LaneCount; lane++ {
		bs.beams[lane] = make(map[BeamID]*BeamState)

		// Create beam states from configuration
		for beamID, beamConfig := range trackConfig.BeamLayout {
			// Convert string beamID from config to BeamID type
			bid := BeamID(beamID)
			bs.beams[lane][bid] = &BeamState{
				BeamID:   bid,
				Lane:     lane,
				Position: beamConfig.Position,
				IsBroken: false,
			}
		}
	}

	bs.status.Status = "ready"
	return nil
}

// Start begins beam system operation
func (bs *BeamSystem) Start(ctx context.Context) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.status.Status = "running"
	return nil
}

// Stop halts beam system operation
func (bs *BeamSystem) Stop() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.status.Status = "stopped"
	return nil
}

// GetStatus returns the component status
func (bs *BeamSystem) GetStatus() component.ComponentStatus {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.status
}

// SetEventBus sets the event bus for publishing events
func (bs *BeamSystem) SetEventBus(eventBus *events.EventBus) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.eventBus = eventBus
}

// SetRaceID sets the race ID for event context
func (bs *BeamSystem) SetRaceID(raceID string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.raceID = raceID
}

// TriggerBeam updates the state of a specific beam
func (bs *BeamSystem) TriggerBeam(lane int, beamID BeamID, isBroken bool) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Validate lane exists
	laneBeams, exists := bs.beams[lane]
	if !exists {
		return fmt.Errorf("lane %d does not exist", lane)
	}

	// Validate beam exists
	beam, exists := laneBeams[beamID]
	if !exists {
		return fmt.Errorf("beam %s does not exist in lane %d", beamID, lane)
	}

	// Check if state actually changed
	if beam.IsBroken == isBroken {
		return nil // No change
	}

	// Update beam state
	previousState := beam.IsBroken
	beam.IsBroken = isBroken
	beam.LastChange = time.Now()

	// Publish appropriate event
	if bs.eventBus != nil {
		eventType := events.EventBeamRestored
		if isBroken {
			eventType = events.EventBeamBroken
		}

		bs.eventBus.Publish(
			events.NewEvent(eventType).
				WithRaceID(bs.raceID).
				WithLane(lane).
				WithData("beam_id", string(beamID)).
				WithData("position", beam.Position).
				WithData("previous_state", previousState).
				WithData("timestamp", beam.LastChange).
				Build(),
		)
	}

	return nil
}

// GetBeamState returns the current state of a specific beam
func (bs *BeamSystem) GetBeamState(lane int, beamID BeamID) (*BeamState, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	laneBeams, exists := bs.beams[lane]
	if !exists {
		return nil, fmt.Errorf("lane %d does not exist", lane)
	}

	beam, exists := laneBeams[beamID]
	if !exists {
		return nil, fmt.Errorf("beam %s does not exist in lane %d", beamID, lane)
	}

	// Return a copy to prevent external modification
	stateCopy := *beam
	return &stateCopy, nil
}

// GetLaneBeamStates returns all beam states for a specific lane
func (bs *BeamSystem) GetLaneBeamStates(lane int) (map[BeamID]*BeamState, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	laneBeams, exists := bs.beams[lane]
	if !exists {
		return nil, fmt.Errorf("lane %d does not exist", lane)
	}

	// Return a copy of the map
	result := make(map[BeamID]*BeamState)
	for beamID, beam := range laneBeams {
		stateCopy := *beam
		result[beamID] = &stateCopy
	}

	return result, nil
}

// GetAllBeamStates returns all beam states for all lanes
func (bs *BeamSystem) GetAllBeamStates() map[int]map[BeamID]*BeamState {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	// Create a deep copy
	result := make(map[int]map[BeamID]*BeamState)
	for lane, laneBeams := range bs.beams {
		result[lane] = make(map[BeamID]*BeamState)
		for beamID, beam := range laneBeams {
			stateCopy := *beam
			result[lane][beamID] = &stateCopy
		}
	}

	return result
}

// ResetBeams resets all beams to unbroken state
func (bs *BeamSystem) ResetBeams() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	for _, laneBeams := range bs.beams {
		for _, beam := range laneBeams {
			if beam.IsBroken {
				beam.IsBroken = false
				beam.LastChange = time.Now()
			}
		}
	}

	// Publish reset event
	if bs.eventBus != nil {
		bs.eventBus.Publish(
			events.NewEvent(events.EventType("beam.reset_all")).
				WithRaceID(bs.raceID).
				Build(),
		)
	}
}

// ValidateBeamSequence checks if beams are being triggered in proper order
func (bs *BeamSystem) ValidateBeamSequence(lane int) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	laneBeams, exists := bs.beams[lane]
	if !exists {
		return fmt.Errorf("lane %d does not exist", lane)
	}

	// Check staging sequence - pre-stage should be broken before stage
	preStage, _ := laneBeams[BeamPreStage]
	stage, _ := laneBeams[BeamStage]

	if stage != nil && stage.IsBroken && preStage != nil && !preStage.IsBroken {
		return fmt.Errorf("invalid staging sequence: stage beam broken before pre-stage beam")
	}

	return nil
}
