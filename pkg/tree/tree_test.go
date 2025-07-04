package tree

import (
	"context"
	"github.com/benharold/libdrag/pkg/config"
	"testing"
)

func TestNewChristmasTree(t *testing.T) {
	tree := NewChristmasTree()
	if tree == nil {
		t.Fatal("NewChristmasTree returned nil")
	}

	status := tree.GetTreeStatus()
	if status.Armed {
		t.Fatal("Tree should not be armed initially")
	}
	if status.Activated {
		t.Fatal("Tree should not be activated initially")
	}
}

func TestChristmasTreeInitialize(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	status := tree.GetStatus()
	if status.Status != "ready" {
		t.Fatalf("Expected status 'ready', got '%s'", status.Status)
	}
}

func TestChristmasTreeLightStates(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Get initial tree status
	status := tree.GetTreeStatus()

	// Verify light states are initialized
	if status.LightStates == nil {
		t.Fatal("LightStates should be initialized")
	}

	// Check that we have light states for at least one lane
	if len(status.LightStates) == 0 {
		t.Fatal("Should have light states for at least one lane")
	}

	// Make sure no staging lights are on
	if status.LightStates[1][LightPreStage] != LightOff {
		t.Fatal("Pre-stage light should be off for lane 1")
	}
	if status.LightStates[2][LightPreStage] != LightOff {
		t.Fatal("Pre-stage light should be off for lane 2")
	}
	if status.LightStates[1][LightStage] != LightOff {
		t.Fatal("Stage light should be off for lane 1")
	}
	if status.LightStates[2][LightStage] != LightOff {
		t.Fatal("Stage light should be off for lane 1")
	}
}

// Test Pre-Stage Light Logic using direct method calls
func TestPreStage(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test single lane pre-stage
	tree.SetPreStage(1)

	// Verify pre-stage light is on for lane 1
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOn {
		t.Fatal("Pre-stage light should be on for lane 1")
	}

	// Pre-stage lane 2
	tree.SetPreStage(2)

	// Verify pre-stage light is on for lane 2
	if status.LightStates[2][LightPreStage] != LightOn {
		t.Fatal("Pre-stage light should be on for lane 2")
	}
}

// Test Stage Light Logic using direct method calls
func TestStage(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Stage both lanes
	tree.SetStage(1)
	tree.SetStage(2)

	// Verify stage light is on for lane 1
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightStage] != LightOn {
		t.Fatal("Stage light should be on for lane 1")
	}

	// Verify stage light is on for lane 2
	if status.LightStates[2][LightStage] != LightOn {
		t.Fatal("Stage light should be on for lane 2")
	}
}

// Test Tree Not Armed Error
func TestTreeNotArmedError(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Try to start sequence without arming tree
	err = tree.StartSequence(config.TreeSequencePro)
	if err == nil {
		t.Fatal("Expected error when starting sequence with unarmed tree")
	}

	if err.Error() != "tree is not armed" {
		t.Fatalf("Expected 'tree is not armed' error, got: %v", err)
	}
}
