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

	// Check the ID exists (remove hardcoded check)
	if tree.GetID() == "" {
		t.Fatal("Tree ID should not be empty")
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
	tree.SetPreStage(1, true)

	// Verify pre-stage light is on for lane 1
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOn {
		t.Fatal("Pre-stage light should be on for lane 1")
	}

	// Pre-stage lane 2
	tree.SetPreStage(2, true)

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
	tree.SetStage(1, true)
	tree.SetStage(2, true)

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

func TestChristmasTreeArm(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify tree is not armed initially
	if tree.IsArmed() {
		t.Fatal("Tree should not be armed initially")
	}

	// Arm the tree
	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	// Verify tree is armed
	if !tree.IsArmed() {
		t.Fatal("Tree should be armed after calling Arm()")
	}

	// Verify component status
	status := tree.GetStatus()
	if status.Status != "armed" {
		t.Fatalf("Expected component status 'armed', got '%s'", status.Status)
	}

	// Verify tree status
	treeStatus := tree.GetTreeStatus()
	if !treeStatus.Armed {
		t.Fatal("Tree status should show armed")
	}
}

func TestChristmasTreeActivate(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify tree is not activated initially
	treeStatus := tree.GetTreeStatus()
	if treeStatus.Activated {
		t.Fatal("Tree should not be activated initially")
	}

	// Activate the tree
	err = tree.Activate()
	if err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	// Verify tree is activated
	treeStatus = tree.GetTreeStatus()
	if !treeStatus.Activated {
		t.Fatal("Tree should be activated after calling Activate()")
	}

	// Verify component status
	status := tree.GetStatus()
	if status.Status != "activated" {
		t.Fatalf("Expected component status 'activated', got '%s'", status.Status)
	}
}

func TestChristmasTreeArmAndActivate(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm then activate
	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	err = tree.Activate()
	if err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	// Verify both states
	treeStatus := tree.GetTreeStatus()
	if !treeStatus.Armed {
		t.Fatal("Tree should be armed")
	}
	if !treeStatus.Activated {
		t.Fatal("Tree should be activated")
	}

	status := tree.GetStatus()
	if status.Status != "activated" {
		t.Fatalf("Expected component status 'activated', got '%s'", status.Status)
	}
}

func TestNewChristmasTreeRandomID(t *testing.T) {
	tree1 := NewChristmasTree()
	tree2 := NewChristmasTree()

	// Verify IDs are not empty
	if tree1.GetID() == "" {
		t.Fatal("Tree ID should not be empty")
	}
	if tree2.GetID() == "" {
		t.Fatal("Tree ID should not be empty")
	}

	// Verify IDs are different (random)
	if tree1.GetID() == tree2.GetID() {
		t.Fatal("Tree IDs should be unique/random")
	}

	// Verify ID format (optional - adjust pattern as needed)
	if len(tree1.GetID()) < 8 {
		t.Fatal("Tree ID should be at least 8 characters")
	}
}

func TestChristmasTreeEmergencyStop(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm and activate the tree first
	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	err = tree.Activate()
	if err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	// Set some lights on to verify they get cleared
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)

	// Verify initial state
	treeStatus := tree.GetTreeStatus()
	if !treeStatus.Armed {
		t.Fatal("Tree should be armed before emergency stop")
	}
	if !treeStatus.Activated {
		t.Fatal("Tree should be activated before emergency stop")
	}

	// Call emergency stop
	err = tree.EmergencyStop()
	if err != nil {
		t.Fatalf("EmergencyStop failed: %v", err)
	}

	// Verify tree is disarmed and deactivated
	treeStatus = tree.GetTreeStatus()
	if treeStatus.Armed {
		t.Fatal("Tree should not be armed after emergency stop")
	}
	if treeStatus.Activated {
		t.Fatal("Tree should not be activated after emergency stop")
	}

	// Verify component status
	status := tree.GetStatus()
	if status.Status != "emergency_stopped" {
		t.Fatalf("Expected component status 'emergency_stopped', got '%s'", status.Status)
	}

	// Verify all lights are off except red lights should be blinking
	for lane := 1; lane <= cfg.Track().LaneCount; lane++ {
		for _, lightType := range []LightType{LightPreStage, LightStage, LightAmber1, LightAmber2, LightAmber3, LightGreen} {
			if treeStatus.LightStates[lane][lightType] != LightOff {
				t.Fatalf("Light %s for lane %d should be off after emergency stop", lightType, lane)
			}
		}

		// Red lights should be blinking
		if treeStatus.LightStates[lane][LightRed] != LightBlink {
			t.Fatalf("Red light for lane %d should be blinking after emergency stop", lane)
		}
	}
}

func TestChristmasTreeAllStaged(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Initially should not be all staged (tree not armed)
	if tree.AllStaged() {
		t.Fatal("Tree should not be all staged initially")
	}

	// Arm the tree
	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	// Still not all staged (no lanes staged yet)
	if tree.AllStaged() {
		t.Fatal("Tree should not be all staged when no lanes are staged")
	}

	// Stage only lane 1
	tree.SetStage(1, true)
	if tree.AllStaged() {
		t.Fatal("Tree should not be all staged when only lane 1 is staged")
	}

	// Stage lane 2 as well (assuming 2-lane track from config)
	tree.SetStage(2, true)
	if !tree.AllStaged() {
		t.Fatal("Tree should be all staged when both lanes are staged")
	}

	// Verify with unarmed tree (should return false even with all lanes staged)
	tree.DisarmTree()
	if tree.AllStaged() {
		t.Fatal("Tree should not be all staged when disarmed, even with lanes staged")
	}
}

func TestChristmasTreeDisarmTree(t *testing.T) {
	tree := NewChristmasTree()
	cfg := config.NewDefaultConfig()

	err := tree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test disarming when not armed (edge case)
	initialStatus := tree.GetTreeStatus()
	if initialStatus.Armed {
		t.Fatal("Tree should not be armed initially")
	}

	// Call DisarmTree when already not armed (should be safe)
	tree.DisarmTree()

	// Verify state remains unchanged
	status := tree.GetTreeStatus()
	if status.Armed {
		t.Fatal("Tree should still not be armed after calling DisarmTree on unarmed tree")
	}

	// Now arm the tree and test normal disarm
	err = tree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	// Verify tree is armed
	status = tree.GetTreeStatus()
	if !status.Armed {
		t.Fatal("Tree should be armed after calling Arm()")
	}

	// Test normal disarm
	tree.DisarmTree()

	// Verify tree is disarmed
	status = tree.GetTreeStatus()
	if status.Armed {
		t.Fatal("Tree should not be armed after calling DisarmTree()")
	}
}
