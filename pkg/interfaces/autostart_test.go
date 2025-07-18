package interfaces

import (
	"context"
	"testing"
)

// TestAutoStart is a test implementation of the AutoStart interface
type TestAutoStart struct {
	armed            bool
	held             bool
	starterOverride  bool
	stagingStates    map[int]LaneStagingState
	activationReady  bool
	threeLightCount  int
	activationStatus string
}

func NewTestAutoStart() *TestAutoStart {
	return &TestAutoStart{
		stagingStates:    make(map[int]LaneStagingState),
		activationStatus: "idle",
	}
}

// AutoStart interface implementation
func (tas *TestAutoStart) Arm(ctx context.Context) error {
	tas.armed = true
	tas.activationStatus = "armed"
	return nil
}

func (tas *TestAutoStart) Disarm(ctx context.Context) error {
	tas.armed = false
	tas.held = false
	tas.activationStatus = "disarmed"
	return nil
}

func (tas *TestAutoStart) Hold(ctx context.Context) error {
	tas.held = true
	tas.updateActivationReadiness()
	return nil
}

func (tas *TestAutoStart) Release(ctx context.Context) error {
	tas.held = false
	tas.updateActivationReadiness()
	return nil
}

func (tas *TestAutoStart) UpdateStagingState(lane int, state LaneStagingState) error {
	tas.stagingStates[lane] = state
	tas.calculateThreeLightCount()
	tas.updateActivationReadiness()
	return nil
}

func (tas *TestAutoStart) IsActivationReady() bool {
	return tas.activationReady
}

func (tas *TestAutoStart) GetThreeLightCount() int {
	return tas.threeLightCount
}

func (tas *TestAutoStart) SetStarterOverride(enabled bool) error {
	tas.starterOverride = enabled
	tas.updateActivationReadiness()
	return nil
}

func (tas *TestAutoStart) IsStarterOverrideActive() bool {
	return tas.starterOverride
}

func (tas *TestAutoStart) IsArmed() bool {
	return tas.armed
}

func (tas *TestAutoStart) IsHeld() bool {
	return tas.held
}

func (tas *TestAutoStart) GetActivationStatus() string {
	return tas.activationStatus
}

// Helper methods for test implementation
func (tas *TestAutoStart) calculateThreeLightCount() {
	count := 0
	
	// Count pre-stage lights
	for _, state := range tas.stagingStates {
		if state.PreStaged {
			count++
		}
	}
	
	// Count stage lights (including deep staging)
	for _, state := range tas.stagingStates {
		if state.Staged || state.DeepStaged {
			count++
		}
	}
	
	tas.threeLightCount = count
}

func (tas *TestAutoStart) updateActivationReadiness() {
	// Three-light rule: need 2 pre-stage + 1 stage minimum
	// Don't activate if starter override is active or system is held
	tas.activationReady = tas.armed && 
		!tas.held && 
		!tas.starterOverride && 
		tas.threeLightCount >= 3
	
	if tas.activationReady {
		tas.activationStatus = "ready_to_activate"
	} else if tas.held {
		tas.activationStatus = "held"
	} else if tas.starterOverride {
		tas.activationStatus = "starter_override"
	} else if !tas.armed {
		tas.activationStatus = "disarmed"
	} else {
		tas.activationStatus = "armed"
	}
}

// Test cases for AutoStart interface behavior

func TestAutoStartStateManagement(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	
	// Test initial state
	if autoStart.IsArmed() {
		t.Error("AutoStart should not be armed initially")
	}
	if autoStart.IsHeld() {
		t.Error("AutoStart should not be held initially")
	}
	if autoStart.GetActivationStatus() != "idle" {
		t.Errorf("Expected status 'idle', got '%s'", autoStart.GetActivationStatus())
	}
	
	// Test arming
	err := autoStart.Arm(ctx)
	if err != nil {
		t.Fatalf("Failed to arm autostart: %v", err)
	}
	if !autoStart.IsArmed() {
		t.Error("AutoStart should be armed after Arm()")
	}
	if autoStart.GetActivationStatus() != "armed" {
		t.Errorf("Expected status 'armed', got '%s'", autoStart.GetActivationStatus())
	}
	
	// Test holding
	err = autoStart.Hold(ctx)
	if err != nil {
		t.Fatalf("Failed to hold autostart: %v", err)
	}
	if !autoStart.IsHeld() {
		t.Error("AutoStart should be held after Hold()")
	}
	if autoStart.GetActivationStatus() != "held" {
		t.Errorf("Expected status 'held', got '%s'", autoStart.GetActivationStatus())
	}
	
	// Test releasing
	err = autoStart.Release(ctx)
	if err != nil {
		t.Fatalf("Failed to release autostart: %v", err)
	}
	if autoStart.IsHeld() {
		t.Error("AutoStart should not be held after Release()")
	}
	if autoStart.GetActivationStatus() != "armed" {
		t.Errorf("Expected status 'armed', got '%s'", autoStart.GetActivationStatus())
	}
	
	// Test disarming
	err = autoStart.Disarm(ctx)
	if err != nil {
		t.Fatalf("Failed to disarm autostart: %v", err)
	}
	if autoStart.IsArmed() {
		t.Error("AutoStart should not be armed after Disarm()")
	}
	if autoStart.IsHeld() {
		t.Error("AutoStart should not be held after Disarm()")
	}
	if autoStart.GetActivationStatus() != "disarmed" {
		t.Errorf("Expected status 'disarmed', got '%s'", autoStart.GetActivationStatus())
	}
}

func TestAutoStartThreeLightRule(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	
	// Arm the system
	autoStart.Arm(ctx)
	
	// Test no lights - should not be ready
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready with no lights")
	}
	if autoStart.GetThreeLightCount() != 0 {
		t.Errorf("Expected 0 lights, got %d", autoStart.GetThreeLightCount())
	}
	
	// Test one pre-stage light
	state1 := LaneStagingState{PreStaged: true, Staged: false}
	autoStart.UpdateStagingState(1, state1)
	
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready with only 1 light")
	}
	if autoStart.GetThreeLightCount() != 1 {
		t.Errorf("Expected 1 light, got %d", autoStart.GetThreeLightCount())
	}
	
	// Test two pre-stage lights
	state2 := LaneStagingState{PreStaged: true, Staged: false}
	autoStart.UpdateStagingState(2, state2)
	
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready with only 2 lights")
	}
	if autoStart.GetThreeLightCount() != 2 {
		t.Errorf("Expected 2 lights, got %d", autoStart.GetThreeLightCount())
	}
	
	// Test three lights (2 pre-stage + 1 stage) - should trigger three-light rule
	state1.Staged = true
	autoStart.UpdateStagingState(1, state1)
	
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready with 3 lights (three-light rule)")
	}
	if autoStart.GetThreeLightCount() != 3 {
		t.Errorf("Expected 3 lights, got %d", autoStart.GetThreeLightCount())
	}
	
	// Test both staged (4 lights total)
	state2.Staged = true
	autoStart.UpdateStagingState(2, state2)
	
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready with both cars staged")
	}
	if autoStart.GetThreeLightCount() != 4 {
		t.Errorf("Expected 4 lights, got %d", autoStart.GetThreeLightCount())
	}
}

func TestAutoStartDeepStagingScenarios(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	autoStart.Arm(ctx)
	
	// Test deep staging scenario: pre-stage OFF, stage ON (counts as 1 light)
	deepStageState := LaneStagingState{
		PreStaged: false,
		Staged: false,
		DeepStaged: true,
	}
	
	normalStageState := LaneStagingState{
		PreStaged: true,
		Staged: true,
	}
	
	// Lane 1 deep staged, Lane 2 normal staged
	autoStart.UpdateStagingState(1, deepStageState)
	autoStart.UpdateStagingState(2, normalStageState)
	
	// Should count as 3 lights: 0 + 1 (deep) + 1 (pre) + 1 (stage) = 3
	if autoStart.GetThreeLightCount() != 3 {
		t.Errorf("Expected 3 lights with deep staging, got %d", autoStart.GetThreeLightCount())
	}
	
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready with deep staging scenario")
	}
	
	// Test both cars deep staged
	autoStart.UpdateStagingState(2, deepStageState)
	
	// Should count as 2 lights: 0 + 1 (deep) + 0 + 1 (deep) = 2
	if autoStart.GetThreeLightCount() != 2 {
		t.Errorf("Expected 2 lights with both deep staged, got %d", autoStart.GetThreeLightCount())
	}
	
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready with only 2 lights (both deep staged)")
	}
}

func TestAutoStartStarterOverride(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	
	// Set up three-light condition
	autoStart.Arm(ctx)
	state1 := LaneStagingState{PreStaged: true, Staged: true}
	state2 := LaneStagingState{PreStaged: true, Staged: false}
	autoStart.UpdateStagingState(1, state1)
	autoStart.UpdateStagingState(2, state2)
	
	// Should be ready normally
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready with three lights")
	}
	
	// Test starter override prevents activation
	err := autoStart.SetStarterOverride(true)
	if err != nil {
		t.Fatalf("Failed to set starter override: %v", err)
	}
	
	if !autoStart.IsStarterOverrideActive() {
		t.Error("Starter override should be active")
	}
	
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready when starter override is active")
	}
	
	// Test releasing starter override
	err = autoStart.SetStarterOverride(false)
	if err != nil {
		t.Fatalf("Failed to release starter override: %v", err)
	}
	
	if autoStart.IsStarterOverrideActive() {
		t.Error("Starter override should not be active")
	}
	
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready when starter override is released")
	}
}

func TestAutoStartHeldPreventsActivation(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	
	// Set up three-light condition
	autoStart.Arm(ctx)
	state1 := LaneStagingState{PreStaged: true, Staged: true}
	state2 := LaneStagingState{PreStaged: true, Staged: false}
	autoStart.UpdateStagingState(1, state1)
	autoStart.UpdateStagingState(2, state2)
	
	// Should be ready normally
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready with three lights")
	}
	
	// Test hold prevents activation
	err := autoStart.Hold(ctx)
	if err != nil {
		t.Fatalf("Failed to hold autostart: %v", err)
	}
	
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready when held")
	}
	
	// Test release restores activation readiness
	err = autoStart.Release(ctx)
	if err != nil {
		t.Fatalf("Failed to release autostart: %v", err)
	}
	
	if !autoStart.IsActivationReady() {
		t.Error("Should be ready when released")
	}
}

func TestAutoStartByeRunScenario(t *testing.T) {
	autoStart := NewTestAutoStart()
	ctx := context.Background()
	
	// Set up bye run scenario - only one car staged
	autoStart.Arm(ctx)
	state1 := LaneStagingState{PreStaged: true, Staged: true}
	autoStart.UpdateStagingState(1, state1)
	
	// Should only have 2 lights (pre-stage + stage for lane 1)
	if autoStart.GetThreeLightCount() != 2 {
		t.Errorf("Expected 2 lights for bye run, got %d", autoStart.GetThreeLightCount())
	}
	
	// Should not activate automatically (needs 3 lights minimum)
	if autoStart.IsActivationReady() {
		t.Error("AutoStart should not activate for bye run (needs manual control)")
	}
	
	// Starter override should be used for bye runs
	err := autoStart.SetStarterOverride(true)
	if err != nil {
		t.Fatalf("Failed to set starter override: %v", err)
	}
	
	// Even if we had enough lights, override should prevent activation
	if autoStart.IsActivationReady() {
		t.Error("Should not be ready when starter override is active for bye run")
	}
}