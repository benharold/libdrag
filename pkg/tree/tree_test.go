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
	if status.IsArmed {
		t.Fatal("Tree should not be armed initially")
	}
	if status.IsRunning {
		t.Fatal("Tree should not be running initially")
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

func TestChristmasTreeLightStates(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Get initial tree status
	treeStatus := tree.GetTreeStatus()

	// Verify light states are initialized
	if treeStatus.LightStates == nil {
		t.Fatal("LightStates should be initialized")
	}

	// Check that we have light states for both lanes
	if len(treeStatus.LightStates) == 0 {
		t.Fatal("Should have light states for at least one lane")
	}
}

// Test Pre-Stage Light Logic
func TestPreStageSequence(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ctx := context.Background()

	// Test single lane pre-stage
	preStageEvent := &events.BaseEvent{
		Type:      events.EventPreStageOn,
		Timestamp: time.Now(),
		Source:    events.ComponentTimingSystem,
		Data:      map[string]interface{}{"lane": 1},
	}

	err = tree.HandleEvent(ctx, preStageEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Verify pre-stage light is on for lane 1
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOn {
		t.Fatal("Pre-stage light should be on for lane 1")
	}

	// Tree should not be armed with only one lane pre-staged
	if status.IsArmed {
		t.Fatal("Tree should not be armed with only one lane pre-staged")
	}

	// Pre-stage lane 2
	preStageEvent2 := &events.BaseEvent{
		Type:      events.EventPreStageOn,
		Timestamp: time.Now(),
		Source:    events.ComponentTimingSystem,
		Data:      map[string]interface{}{"lane": 2},
	}

	err = tree.HandleEvent(ctx, preStageEvent2)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Now tree should be armed
	status = tree.GetTreeStatus()
	if !status.IsArmed {
		t.Fatal("Tree should be armed when both lanes are pre-staged")
	}
}

// Test Stage Light Logic
func TestStageSequence(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize and start
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = tree.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := context.Background()

	// Pre-stage both lanes first
	for lane := 1; lane <= 2; lane++ {
		preStageEvent := &events.BaseEvent{
			Type:      events.EventPreStageOn,
			Timestamp: time.Now(),
			Source:    events.ComponentTimingSystem,
			Data:      map[string]interface{}{"lane": lane},
		}
		tree.HandleEvent(ctx, preStageEvent)
	}

	// Stage lane 1 (using correct event type)
	stageEvent := &events.BaseEvent{
		Type:      events.EventStageOn,
		Timestamp: time.Now(),
		Source:    events.ComponentTimingSystem,
		Data:      map[string]interface{}{"lane": 1},
	}

	err = tree.HandleEvent(ctx, stageEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Verify stage light is on for lane 1
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightStage] != LightOn {
		t.Fatal("Stage light should be on for lane 1")
	}
}

// Test Pro Tree Sequence Timing
func TestProTreeSequence(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ctx := context.Background()

	// Pre-stage and stage both lanes
	for lane := 1; lane <= 2; lane++ {
		preStageEvent := &events.BaseEvent{
			Type:      events.EventPreStageOn,
			Timestamp: time.Now(),
			Source:    events.ComponentTimingSystem,
			Data:      map[string]interface{}{"lane": lane},
		}
		tree.HandleEvent(ctx, preStageEvent)
	}

	// Start Pro sequence
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentStarterControl,
		Data:      map[string]interface{}{"sequence_type": config.TreeSequencePro},
	}

	err = tree.HandleEvent(ctx, raceStartEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Verify sequence is running
	status := tree.GetTreeStatus()
	if !status.IsRunning {
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

// Test Sportsman Tree Sequence Timing
func TestSportsmanTreeSequence(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ctx := context.Background()

	// Pre-stage and stage both lanes
	for lane := 1; lane <= 2; lane++ {
		preStageEvent := &events.BaseEvent{
			Type:      events.EventPreStageOn,
			Timestamp: time.Now(),
			Source:    events.ComponentTimingSystem,
			Data:      map[string]interface{}{"lane": lane},
		}
		tree.HandleEvent(ctx, preStageEvent)
	}

	// Start Sportsman sequence
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentStarterControl,
		Data:      map[string]interface{}{"sequence_type": config.TreeSequenceSportsman},
	}

	err = tree.HandleEvent(ctx, raceStartEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Verify sequence type
	status := tree.GetTreeStatus()
	if status.SequenceType != config.TreeSequenceSportsman {
		t.Fatalf("Expected Sportsman sequence, got %v", status.SequenceType)
	}
}

// Test Tree Not Armed Error
func TestTreeNotArmedError(t *testing.T) {
	tree := NewChristmasTree()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize
	err := tree.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ctx := context.Background()

	// Try to start sequence without arming tree
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentStarterControl,
		Data:      map[string]interface{}{"sequence_type": config.TreeSequencePro},
	}

	err = tree.HandleEvent(ctx, raceStartEvent)
	if err == nil {
		t.Fatal("Expected error when starting sequence with unarmed tree")
	}

	if err.Error() != "tree is not armed" {
		t.Fatalf("Expected 'tree is not armed' error, got: %v", err)
	}
}
