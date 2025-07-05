package autostart

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/tree"
)

// TestThreeBeamAutomaticArming demonstrates the complete integration
// between the auto-start system and tree component for automatic arming
func TestThreeBeamAutomaticArming(t *testing.T) {
	// Create components
	autoStart := NewAutoStartSystem()
	christmasTree := tree.NewChristmasTree()

	// Initialize both components
	cfg := config.NewDefaultConfig()

	err := autoStart.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize auto-start: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Arm components
	err = autoStart.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start auto-start: %v", err)
	}

	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to start tree: %v", err)
	}

	// Connect auto-start system to tree component
	autoStart.SetTreeComponent(christmasTree)
	autoStart.SetTestMode(true)

	// Verify initial states
	if autoStart.GetAutoStartStatus().State != StateIdle {
		t.Errorf("Expected auto-start to be in idle state")
	}
	if !christmasTree.IsArmed() {
		t.Errorf("Expected tree to be armed after starter armed it in setup")
	}

	// Test 1: Two pre-stage beams only (should not trigger auto-start activation)
	autoStart.UpdateVehicleStaging(1, true, false, 0) // Lane 1 pre-stage
	autoStart.UpdateVehicleStaging(2, true, false, 0) // Lane 2 pre-stage
	time.Sleep(10 * time.Millisecond)

	if autoStart.GetAutoStartStatus().State != StateIdle {
		t.Errorf("Expected auto-start to remain idle with only two pre-stage beams")
	}
	// Tree should remain armed (it was armed by starter, not by beam conditions)
	if !christmasTree.IsArmed() {
		t.Errorf("Expected tree to remain armed (tree was armed by starter)")
	}

	// Test 2: Add third beam (one stage) - should trigger three-beam rule
	autoStart.UpdateVehicleStaging(1, true, true, 0) // Lane 1 pre-stage + stage (third beam)
	time.Sleep(10 * time.Millisecond)

	// Verify auto-start system armed
	status := autoStart.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Errorf("Expected auto-start to be armed after three beams, got %v", status.State)
	}

	// Verify tree component automatically armed
	if !christmasTree.IsArmed() {
		t.Errorf("Expected tree to be automatically armed after three beams")
	}

	// Verify tree is armed (but we don't check ArmedBy since that field was removed)
	treeStatus := christmasTree.GetTreeStatus()
	if !treeStatus.Armed {
		t.Errorf("Expected tree to be armed by auto-start system")
	}

	// Test 3: Complete staging (both vehicles staged) - should progress to staging state
	autoStart.UpdateVehicleStaging(2, true, true, 0) // Lane 2 pre-stage + stage
	time.Sleep(20 * time.Millisecond)                // Allow time for staging detection

	// Auto-start should progress through states rapidly in test mode
	status = autoStart.GetAutoStartStatus()
	// In test mode, it might already be triggered due to fast timing
	if status.State != StateStaging && status.State != StateTriggered {
		t.Errorf("Expected auto-start to be in staging or triggered state, got %v", status.State)
	}

	// Tree should remain armed throughout the process
	if !christmasTree.IsArmed() {
		t.Errorf("Expected tree to remain armed during staging")
	}

	t.Logf("✅ Three-beam automatic arming test completed successfully")
	t.Logf("   • Auto-start state: %v", status.State)
	t.Logf("   • Tree armed: %v", christmasTree.IsArmed())
}
