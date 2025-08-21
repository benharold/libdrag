package interfaces

import (
	"context"
	"testing"
)

// TestStarter is a test implementation of the Starter interface
type TestStarter struct {
	treeArmed     bool
	treeHeld      bool
	lastDecision  StartDecision
	holdReason    string
	abortReason   string
	decisions     map[string]bool // For configuring decision behavior
}

func NewTestStarter() *TestStarter {
	return &TestStarter{
		decisions: make(map[string]bool),
		lastDecision: StartDecision{
			CanStart: false,
			ShouldHold: false,
			Reason: "initial_state",
		},
	}
}

// Starter interface implementation
func (ts *TestStarter) ArmTree(ctx context.Context) error {
	ts.treeArmed = true
	ts.treeHeld = false
	return nil
}

func (ts *TestStarter) HoldTree(ctx context.Context, reason string) error {
	ts.treeHeld = true
	ts.holdReason = reason
	return nil
}

func (ts *TestStarter) ReleaseTree(ctx context.Context) error {
	ts.treeHeld = false
	ts.holdReason = ""
	return nil
}

func (ts *TestStarter) ManualStart(ctx context.Context) error {
	if !ts.treeArmed {
		ts.treeArmed = true
	}
	ts.lastDecision = StartDecision{
		CanStart: true,
		ShouldHold: false,
		Reason: "manual_start",
		ManualOverride: true,
	}
	return nil
}

func (ts *TestStarter) AbortRace(ctx context.Context, reason string) error {
	ts.abortReason = reason
	ts.treeArmed = false
	ts.treeHeld = false
	return nil
}

func (ts *TestStarter) ShouldHonorDeepStaging(lane int, class string) bool {
	// Test implementation: honor for most classes except Super Gas/Stock/Street
	switch class {
	case "Super Gas", "Super Stock", "Super Street":
		return false
	default:
		return true
	}
}

func (ts *TestStarter) ShouldWaitForStaging(conditions TrackConditions) bool {
	return conditions.SafeToRace
}

func (ts *TestStarter) CanStartRace(stagingState StagingState) StartDecision {
	// Test implementation of starter decision logic
	
	// Check if held by starter
	if ts.treeHeld {
		return StartDecision{
			CanStart: false,
			ShouldHold: true,
			Reason: ts.holdReason,
		}
	}
	
	// Check basic staging requirements
	lane1Staged := stagingState.Lane1.Staged || stagingState.Lane1.DeepStaged
	lane2Staged := stagingState.Lane2.Staged || stagingState.Lane2.DeepStaged
	
	// Normal race - both cars must be staged
	if lane1Staged && lane2Staged {
		decision := StartDecision{
			CanStart: true,
			ShouldHold: false,
			Reason: "both_staged",
		}
		ts.lastDecision = decision
		return decision
	}
	
	// Bye run - only one car staged
	if (lane1Staged && !lane2Staged) || (!lane1Staged && lane2Staged) {
		decision := StartDecision{
			CanStart: true,
			ShouldHold: false,
			Reason: "bye_run",
			ManualOverride: true, // Bye runs require manual control
		}
		ts.lastDecision = decision
		return decision
	}
	
	// Neither car staged
	decision := StartDecision{
		CanStart: false,
		ShouldHold: false,
		Reason: "waiting_for_staging",
	}
	ts.lastDecision = decision
	return decision
}

func (ts *TestStarter) IsTreeArmed() bool {
	return ts.treeArmed
}

func (ts *TestStarter) IsTreeHeld() bool {
	return ts.treeHeld
}

func (ts *TestStarter) GetLastDecision() StartDecision {
	return ts.lastDecision
}

// Test cases for Starter interface behavior

func TestStarterTreeControl(t *testing.T) {
	starter := NewTestStarter()
	ctx := context.Background()
	
	// Test initial state
	if starter.IsTreeArmed() {
		t.Error("Tree should not be armed initially")
	}
	if starter.IsTreeHeld() {
		t.Error("Tree should not be held initially")
	}
	
	// Test arming tree
	err := starter.ArmTree(ctx)
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
	}
	if !starter.IsTreeArmed() {
		t.Error("Tree should be armed after ArmTree()")
	}
	if starter.IsTreeHeld() {
		t.Error("Tree should not be held after ArmTree()")
	}
	
	// Test holding tree
	holdReason := "track_conditions"
	err = starter.HoldTree(ctx, holdReason)
	if err != nil {
		t.Fatalf("Failed to hold tree: %v", err)
	}
	if !starter.IsTreeHeld() {
		t.Error("Tree should be held after HoldTree()")
	}
	if starter.holdReason != holdReason {
		t.Errorf("Expected hold reason '%s', got '%s'", holdReason, starter.holdReason)
	}
	
	// Test releasing tree
	err = starter.ReleaseTree(ctx)
	if err != nil {
		t.Fatalf("Failed to release tree: %v", err)
	}
	if starter.IsTreeHeld() {
		t.Error("Tree should not be held after ReleaseTree()")
	}
	
	// Test manual start
	starter.treeArmed = false // Reset
	err = starter.ManualStart(ctx)
	if err != nil {
		t.Fatalf("Failed to manual start: %v", err)
	}
	if !starter.IsTreeArmed() {
		t.Error("Tree should be armed after ManualStart()")
	}
	decision := starter.GetLastDecision()
	if !decision.CanStart {
		t.Error("Decision should allow start after ManualStart()")
	}
	if !decision.ManualOverride {
		t.Error("Decision should indicate manual override after ManualStart()")
	}
	
	// Test abort race
	abortReason := "oil_on_track"
	err = starter.AbortRace(ctx, abortReason)
	if err != nil {
		t.Fatalf("Failed to abort race: %v", err)
	}
	if starter.IsTreeArmed() {
		t.Error("Tree should not be armed after AbortRace()")
	}
	if starter.abortReason != abortReason {
		t.Errorf("Expected abort reason '%s', got '%s'", abortReason, starter.abortReason)
	}
}

func TestStarterDeepStagingDecisions(t *testing.T) {
	starter := NewTestStarter()
	
	tests := []struct {
		class    string
		expected bool
		reason   string
	}{
		{"Pro Stock", true, "Pro Stock allows deep staging"},
		{"Pro Modified", true, "Pro Modified allows deep staging"},
		{"Top Fuel", true, "Top Fuel allows deep staging"},
		{"Super Gas", false, "Super Gas prohibits deep staging"},
		{"Super Stock", false, "Super Stock prohibits deep staging"},
		{"Super Street", false, "Super Street prohibits deep staging"},
		{"Bracket", true, "Bracket racing typically allows deep staging"},
	}
	
	for _, test := range tests {
		t.Run(test.class, func(t *testing.T) {
			result := starter.ShouldHonorDeepStaging(1, test.class)
			if result != test.expected {
				t.Errorf("ShouldHonorDeepStaging(%s) = %v, expected %v (%s)", 
					test.class, result, test.expected, test.reason)
			}
		})
	}
}

func TestStarterRaceStartDecisions(t *testing.T) {
	starter := NewTestStarter()
	ctx := context.Background()
	
	// Test both cars staged - should allow start
	stagingState := StagingState{
		Lane1: LaneStagingState{PreStaged: true, Staged: true},
		Lane2: LaneStagingState{PreStaged: true, Staged: true},
	}
	
	decision := starter.CanStartRace(stagingState)
	if !decision.CanStart {
		t.Error("Should allow start when both cars are staged")
	}
	if decision.Reason != "both_staged" {
		t.Errorf("Expected reason 'both_staged', got '%s'", decision.Reason)
	}
	
	// Test bye run scenario - only lane 1 staged
	byeRunState := StagingState{
		Lane1: LaneStagingState{PreStaged: true, Staged: true},
		Lane2: LaneStagingState{PreStaged: false, Staged: false},
	}
	
	decision = starter.CanStartRace(byeRunState)
	if !decision.CanStart {
		t.Error("Should allow start for bye run")
	}
	if decision.Reason != "bye_run" {
		t.Errorf("Expected reason 'bye_run', got '%s'", decision.Reason)
	}
	if !decision.ManualOverride {
		t.Error("Bye run should require manual override")
	}
	
	// Test deep staging scenario
	deepStageState := StagingState{
		Lane1: LaneStagingState{PreStaged: false, Staged: true, DeepStaged: true},
		Lane2: LaneStagingState{PreStaged: true, Staged: true},
	}
	
	decision = starter.CanStartRace(deepStageState)
	if !decision.CanStart {
		t.Error("Should allow start with deep staging")
	}
	
	// Test held tree - should not allow start
	starter.HoldTree(ctx, "safety_concern")
	decision = starter.CanStartRace(stagingState)
	if decision.CanStart {
		t.Error("Should not allow start when tree is held")
	}
	if !decision.ShouldHold {
		t.Error("Decision should indicate tree should be held")
	}
	if decision.Reason != "safety_concern" {
		t.Errorf("Expected hold reason 'safety_concern', got '%s'", decision.Reason)
	}
	
	// Test neither car staged
	starter.ReleaseTree(ctx) // Release hold
	noStagingState := StagingState{
		Lane1: LaneStagingState{PreStaged: true, Staged: false},
		Lane2: LaneStagingState{PreStaged: false, Staged: false},
	}
	
	decision = starter.CanStartRace(noStagingState)
	if decision.CanStart {
		t.Error("Should not allow start when no cars are staged")
	}
	if decision.Reason != "waiting_for_staging" {
		t.Errorf("Expected reason 'waiting_for_staging', got '%s'", decision.Reason)
	}
}

func TestStarterTrackConditionsDecision(t *testing.T) {
	starter := NewTestStarter()
	
	// Test safe conditions
	safeConditions := TrackConditions{
		Temperature: 75.0,
		Humidity: 45.0,
		WindSpeed: 5.0,
		TrackTemp: 120.0,
		SafeToRace: true,
	}
	
	if !starter.ShouldWaitForStaging(safeConditions) {
		t.Error("Should wait for staging when conditions are safe")
	}
	
	// Test unsafe conditions
	unsafeConditions := TrackConditions{
		Temperature: 75.0,
		Humidity: 45.0,
		WindSpeed: 25.0, // High wind
		TrackTemp: 120.0,
		SafeToRace: false,
	}
	
	if starter.ShouldWaitForStaging(unsafeConditions) {
		t.Error("Should not wait for staging when conditions are unsafe")
	}
}