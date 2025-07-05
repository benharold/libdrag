package beam

import (
	"testing"

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
