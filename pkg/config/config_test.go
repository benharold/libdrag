package config

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig returned nil")
	}

	// Test that track config is properly initialized
	trackConfig := cfg.GetTrackConfig()
	if trackConfig.BeamLayout == nil {
		t.Fatal("BeamLayout should be initialized")
	}

	// Verify essential beams exist
	expectedBeams := []string{"pre_stage_L", "pre_stage_R", "stage_L", "stage_R",
		"sixty_foot_L", "sixty_foot_R", "eighth_mile_L", "eighth_mile_R",
		"quarter_mile_L", "quarter_mile_R"}

	for _, beamID := range expectedBeams {
		if _, exists := trackConfig.BeamLayout[beamID]; !exists {
			t.Fatalf("Expected beam %s not found in BeamLayout", beamID)
		}
	}
}

func TestTreeConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	treeConfig := cfg.GetTreeConfig()

	// Verify tree config has reasonable defaults
	if treeConfig.PreStageTimeoutSeconds <= 0 {
		t.Fatal("PreStageTimeoutSeconds should be positive")
	}

	if treeConfig.StageTimeoutSeconds <= 0 {
		t.Fatal("StageTimeoutSeconds should be positive")
	}
}

func TestBeamConfigValidation(t *testing.T) {
	cfg := NewDefaultConfig()
	trackConfig := cfg.GetTrackConfig()

	// Verify beam positions are logical
	beams := trackConfig.BeamLayout

	// Pre-stage should be before stage
	if beams["pre_stage_L"].Position >= beams["stage_L"].Position {
		t.Fatal("Pre-stage should be before stage position")
	}

	// Stage should be before sixty foot
	if beams["stage_L"].Position >= beams["sixty_foot_L"].Position {
		t.Fatal("Stage should be before sixty foot position")
	}

	// Sixty foot should be before eighth mile
	if beams["sixty_foot_L"].Position >= beams["eighth_mile_L"].Position {
		t.Fatal("Sixty foot should be before eighth mile position")
	}

	// Eighth mile should be before quarter mile
	if beams["eighth_mile_L"].Position >= beams["quarter_mile_L"].Position {
		t.Fatal("Eighth mile should be before quarter mile position")
	}
}
