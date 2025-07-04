package tree

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
)

func TestNewChristmasTree(t *testing.T) {
	tree := NewChristmasTree()
	if tree == nil {
		t.Fatal("NewChristmasTree returned nil")
	}

	status := tree.GetTreeStatus()
	if status.Armed {
		t.Fatal("Auto-start should not be activated initially")
	}
	if status.Activated {
		t.Fatal("Tree should not be running initially")
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
	treeStatus := tree.GetTreeStatus()

	// Verify light states are initialized
	if treeStatus.LightStates == nil {
		t.Fatal("LightStates should be initialized")
	}

	// Check that we have light states for at least one lane
	if len(treeStatus.LightStates) == 0 {
		t.Fatal("Should have light states for at least one lane")
	}
}

// Test Pre-Stage Light Logic using direct method calls
func TestPreStageSequence(t *testing.T) {
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

	// Tree should not be armed with only one lane pre-staged
	if status.Armed {
		t.Fatal("Auto-start should not be activated with only one lane pre-staged")
	}

	// Pre-stage lane 2
	tree.SetPreStage(2)
}

// Test Stage Light Logic using direct method calls
func TestStageSequence(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize and start
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	// Pre-stage both lanes first
	tree.SetPreStage(1)
	tree.SetPreStage(2)

	// Stage lane 1
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

// Test Pro Tree Sequence Timing using direct method calls
func TestProTreeSequence(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Pre-stage both lanes
	tree.SetPreStage(1)
	tree.SetPreStage(2)

	// Arm Pro sequence
	err = tree.StartSequence(config.TreeSequencePro)
	if err != nil {
		t.Fatalf("StartSequence failed: %v", err)
	}

	// Verify sequence is running
	status := tree.GetTreeStatus()
	if !status.Activated {
		t.Fatal("Tree sequence should be running")
	}

	if status.SequenceType != config.TreeSequencePro {
		t.Fatalf("Expected Pro sequence, got %v", status.SequenceType)
	}

	// Wait for sequence to complete (green delay is 400ms for pro tree)
	time.Sleep(500 * time.Millisecond)

	// Verify green light is on
	status = tree.GetTreeStatus()
	if status.LightStates[1][LightGreen] != LightOn {
		t.Fatal("Green light should be on after Pro sequence")
	}
	if status.LightStates[2][LightGreen] != LightOn {
		t.Fatal("Green light should be on for both lanes after Pro sequence")
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
