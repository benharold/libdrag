package tree

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

func TestNewChristmasTree(t *testing.T) {
	tree := NewChristmasTree()
	if tree == nil {
		t.Fatal("NewChristmasTree returned nil")
	}

	status := tree.GetTreeStatus()
	if status.State != TreeStateIdle {
		t.Fatalf("Expected initial state %s, got %s", TreeStateIdle, status.State)
	}
}

func TestChristmasTreeInitialize(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	status := tree.GetStatus()
	if status.Status != "ready" {
		t.Fatalf("Expected status 'ready', got '%s'", status.Status)
	}
}

func TestChristmasTreeSequence(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start
	err = tree.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify initial state
	status := tree.GetTreeStatus()
	if status.State != TreeStateIdle {
		t.Fatalf("Expected state %s after start, got %s", TreeStateIdle, status.State)
	}

	// Test that tree responds to vehicle staging events
	// This would require more complex event simulation
}

func TestTreeStates(t *testing.T) {
	// Test that all tree states are valid
	validStates := []TreeState{
		TreeStateIdle,
		TreeStateWaitingForVehicles,
		TreeStatePreStage,
		TreeStateStaged,
		TreeStateCountdown,
		TreeStateGreen,
		TreeStateRacing,
		TreeStateFoul,
		TreeStateComplete,
	}

	for _, state := range validStates {
		if string(state) == "" {
			t.Fatalf("Tree state should not be empty: %v", state)
		}
	}
}
