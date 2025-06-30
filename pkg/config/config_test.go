package config

import (
	"testing"
	"time"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig returned nil")
	}

	// Test that track config is properly initialized
	trackConfig := cfg.Track()
	if trackConfig.BeamLayout == nil {
		t.Fatal("BeamLayout should be initialized")
	}

	// Verify essential beams exist
	expectedBeams := []string{"pre_stage", "stage", "60_foot", "330_foot", "660_foot", "1000_foot", "1320_foot"}

	for _, beamID := range expectedBeams {
		if _, exists := trackConfig.BeamLayout[beamID]; !exists {
			t.Fatalf("Expected beam %s not found in BeamLayout", beamID)
		}
	}
}

func TestTreeConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	treeConfig := cfg.Tree()

	// Verify tree config has reasonable defaults
	if treeConfig.PreStageTimeout <= 0 {
		t.Fatal("PreStageTimeout should be positive")
	}

	if treeConfig.StageTimeout <= 0 {
		t.Fatal("StageTimeout should be positive")
	}

	// Verify timing values are reasonable
	if treeConfig.GreenDelay != 400*time.Millisecond {
		t.Fatal("GreenDelay should be 400ms for pro tree")
	}
}

func TestBeamConfigValidation(t *testing.T) {
	cfg := NewDefaultConfig()
	trackConfig := cfg.Track()

	// Verify beam positions are logical
	beams := trackConfig.BeamLayout

	// Pre-stage should be before starting line
	if beams["pre_stage"].Position >= 0 {
		t.Fatal("Pre-stage beam should be before starting line (negative position)")
	}

	// Stage beam should be at starting line
	if beams["stage"].Position != 0 {
		t.Fatal("Stage beam should be at starting line (position 0)")
	}

	// Quarter mile should be at 1320 feet
	if beams["1320_foot"].Position != 1320 {
		t.Fatal("Quarter mile beam should be at 1320 feet")
	}
}
