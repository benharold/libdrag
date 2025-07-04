package beam

import (
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/events"
	"github.com/stretchr/testify/assert"
)

func TestNewBeamSystem(t *testing.T) {
	// Arrange
	eventBus := events.NewEventBus(false) // Add required boolean parameter

	// Act
	beamSystem := NewBeamSystem(eventBus)

	// Assert
	assert.NotNil(t, beamSystem)
	assert.NotNil(t, beamSystem.beams)
	assert.Equal(t, eventBus, beamSystem.eventBus)
}

func TestBeamSystem_AddBeam(t *testing.T) {
	// Arrange
	eventBus := events.NewEventBus(false)
	beamSystem := NewBeamSystem(eventBus)

	// Act
	lane := 1
	position := 60.0
	beamSystem.AddBeam(lane, position)

	// Assert
	beam := beamSystem.GetBeam(lane, position)
	assert.NotNil(t, beam)
	assert.Equal(t, position, beam.Position)
	assert.Equal(t, lane, beam.Lane)
	assert.False(t, beam.IsBroken)
	assert.WithinDuration(t, time.Now(), beam.LastChange, time.Second)
}

func TestBeamSystem_BreakBeam(t *testing.T) {
	// Arrange
	eventBus := events.NewEventBus(false)
	beamSystem := NewBeamSystem(eventBus)
	lane := 1
	position := 60.0
	beamSystem.AddBeam(lane, position)

	// Act
	beamSystem.BreakBeam(lane, position)

	// Assert
	beam := beamSystem.GetBeam(lane, position)
	assert.True(t, beam.IsBroken)
	assert.True(t, beamSystem.IsBeamBroken(lane, position))
}

func TestBeamSystem_RestoreBeam(t *testing.T) {
	// Arrange
	eventBus := events.NewEventBus(false)
	beamSystem := NewBeamSystem(eventBus)
	lane := 1
	position := 60.0
	beamSystem.AddBeam(lane, position)
	beamSystem.BreakBeam(lane, position)

	// Act
	beamSystem.RestoreBeam(lane, position)

	// Assert
	beam := beamSystem.GetBeam(lane, position)
	assert.False(t, beam.IsBroken)
	assert.False(t, beamSystem.IsBeamBroken(lane, position))
}

func TestBeamSystem_GetBeamsForLane(t *testing.T) {
	// Arrange
	eventBus := events.NewEventBus(false)
	beamSystem := NewBeamSystem(eventBus)
	lane := 1
	beamSystem.AddBeam(lane, 60.0)
	beamSystem.AddBeam(lane, 330.0)

	// Act
	beams := beamSystem.GetBeamsForLane(lane)

	// Assert
	assert.Len(t, beams, 2)
	assert.NotNil(t, beams[60.0])
	assert.NotNil(t, beams[330.0])
}
