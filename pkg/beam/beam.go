// Package beam provides timing beam simulation for drag racing systems.
//
// This package models the physical timing infrastructure found on drag strips,
// including staging beams, 60-foot timers, 330-foot timers, and finish line beams.
// Each beam can detect when a vehicle passes through and publishes events
// to notify other systems of timing changes.
//
// Example usage:
//
//	beamSystem := beam.NewBeamSystem(eventBus, "race-123")
//	beamSystem.AddBeam(1, 60.0)  // Add 60ft beam to lane 1
//	beamSystem.BreakBeam(1, 60.0) // Simulate car breaking beam
package beam

import (
	"github.com/benharold/libdrag/pkg/events"
	"time"
)

// BeamSystem manages all timing beams for a drag racing event.
//
// The system organizes beams by lane and position, allowing efficient
// lookup and management of timing infrastructure. It publishes events
// when beam states change to notify other race systems.
//
// The nested map structure allows for quick access:
//   - beams[lane][position] returns the beam at that specific location
//   - Multiple beams per lane support staging, 60ft, 330ft, and finish timing
type BeamSystem struct {
	// beams maps lane number to position to beam pointer.
	// Lane numbers typically start at 1 (not 0-indexed).
	// Positions are measured in feet from the starting line.
	// Common positions: 0.0 (staging), 60.0, 330.0, 1320.0 (quarter mile)
	beams map[int]map[float64]*Beam // lane -> position -> beam

	// eventBus handles publishing beam break/restore events to other systems.
	// Events notify timing systems, scoreboards, and race management.
	eventBus *events.EventBus

	// raceID identifies which race this beam system is monitoring.
	// Allows multiple concurrent races with separate beam systems.
	raceID string
}

// Beam represents a single timing beam at a specific track position.
//
// In drag racing, beams use infrared or laser technology to detect
// when vehicles pass through. The beam tracks its current state
// and the time of the last state change for precise timing.
type Beam struct {
	// Position is the distance in feet from the starting line.
	// Standard NHRA positions:
	//   0.0   - Pre-stage beam
	//   7.0   - Stage beam
	//   60.0  - 60-foot timer
	//   330.0 - 330-foot timer
	//   660.0 - 660-foot timer (eighth mile)
	//   1000.0- 1000-foot timer
	//   1320.0- Quarter mile finish line
	Position float64

	// Lane identifies which racing lane this beam monitors.
	// Typically 1-2 for heads-up racing, but can support more lanes.
	Lane int

	// IsBroken indicates if the beam is currently broken (blocked).
	// true = vehicle is currently blocking the beam
	// false = beam path is clear
	IsBroken bool

	// LastChange records when the beam state last changed.
	// Used for precise timing calculations and event ordering.
	// Critical for determining elapsed times between beam breaks.
	LastChange time.Time
}

// NewBeamSystem creates a new beam management system for a drag racing event.
func NewBeamSystem(eventBus *events.EventBus, raceID string) *BeamSystem {
	return &BeamSystem{
		beams:    make(map[int]map[float64]*Beam),
		eventBus: eventBus,
		raceID:   raceID,
	}
}

// AddBeam creates a new timing beam at the specified lane and position.
func (bs *BeamSystem) AddBeam(lane int, position float64) {
	// Initialize lane map if it doesn't exist
	if bs.beams[lane] == nil {
		bs.beams[lane] = make(map[float64]*Beam)
	}

	// Create new beam
	beam := &Beam{
		Position:   position,
		Lane:       lane,
		IsBroken:   false,
		LastChange: time.Now(),
	}

	// Add beam to system
	bs.beams[lane][position] = beam

	// Publish beam added event
	if bs.eventBus != nil {
		event := events.Event{
			Type: "beam_added",
			Data: map[string]interface{}{
				"race_id":  bs.raceID,
				"lane":     lane,
				"position": position,
				"time":     beam.LastChange,
			},
		}
		bs.eventBus.Publish(event)
	}
}

// BreakBeam simulates a timing beam being broken by a vehicle.
func (bs *BeamSystem) BreakBeam(lane int, position float64) {
	// Check if beam exists
	if bs.beams[lane] == nil || bs.beams[lane][position] == nil {
		return
	}

	beam := bs.beams[lane][position]

	// Only break if not already broken
	if !beam.IsBroken {
		beam.IsBroken = true
		beam.LastChange = time.Now()

		// Publish beam break event
		if bs.eventBus != nil {
			event := events.Event{
				Type: "beam_broken",
				Data: map[string]interface{}{
					"race_id":  bs.raceID,
					"lane":     lane,
					"position": position,
					"time":     beam.LastChange,
				},
			}
			bs.eventBus.Publish(event)
		}
	}
}

// RestoreBeam simulates a timing beam being restored after a vehicle passes.
func (bs *BeamSystem) RestoreBeam(lane int, position float64) {
	// Check if beam exists
	if bs.beams[lane] == nil || bs.beams[lane][position] == nil {
		return
	}

	beam := bs.beams[lane][position]

	// Only restore if currently broken
	if beam.IsBroken {
		beam.IsBroken = false
		beam.LastChange = time.Now()

		// Publish beam restore event
		if bs.eventBus != nil {
			event := events.Event{
				Type: "beam_restored",
				Data: map[string]interface{}{
					"race_id":  bs.raceID,
					"lane":     lane,
					"position": position,
					"time":     beam.LastChange,
				},
			}
			bs.eventBus.Publish(event)
		}
	}
}

// GetBeam returns the beam at the specified lane and position.
func (bs *BeamSystem) GetBeam(lane int, position float64) *Beam {
	if bs.beams[lane] == nil {
		return nil
	}
	return bs.beams[lane][position]
}

// GetBeamsForLane returns all beams for a specific lane.
func (bs *BeamSystem) GetBeamsForLane(lane int) map[float64]*Beam {
	return bs.beams[lane]
}

// IsBeamBroken checks if a beam at the specified position is currently broken.
func (bs *BeamSystem) IsBeamBroken(lane int, position float64) bool {
	beam := bs.GetBeam(lane, position)
	return beam != nil && beam.IsBroken
}
