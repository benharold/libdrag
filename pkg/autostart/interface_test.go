package autostart

import (
	"context"
	"testing"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/interfaces"
)

// Test that AutoStartSystem implements the AutoStart interface
func TestAutoStartSystemImplementsInterface(t *testing.T) {
	system := NewAutoStartSystem(nil)
	
	// This should compile if the interface is properly implemented
	var _ interfaces.AutoStart = system
}

// Test the KISS implementation
func TestAutoStartInterfaceMethods(t *testing.T) {
	system := NewAutoStartSystem(nil)
	ctx := context.Background()
	
	// Test basic state management
	if system.IsArmed() {
		t.Error("Should not be armed initially")
	}
	
	err := system.Arm(ctx)
	if err != nil {
		t.Fatalf("Failed to arm: %v", err)
	}
	
	if !system.IsArmed() {
		t.Error("Should be armed after Arm()")
	}
	
	// Test hold/release
	err = system.Hold(ctx)
	if err != nil {
		t.Fatalf("Failed to hold: %v", err)
	}
	
	if !system.IsHeld() {
		t.Error("Should be held after Hold()")
	}
	
	if system.GetActivationStatus() != "held" {
		t.Errorf("Expected status 'held', got '%s'", system.GetActivationStatus())
	}
	
	err = system.Release(ctx)
	if err != nil {
		t.Fatalf("Failed to release: %v", err)
	}
	
	if system.IsHeld() {
		t.Error("Should not be held after Release()")
	}
	
	// Test starter override
	err = system.SetStarterOverride(true)
	if err != nil {
		t.Fatalf("Failed to set starter override: %v", err)
	}
	
	if !system.IsStarterOverrideActive() {
		t.Error("Starter override should be active")
	}
	
	if system.GetActivationStatus() != "held" {
		t.Error("Should show held status when starter override is active")
	}
	
	// Test disarm
	err = system.Disarm(ctx)
	if err != nil {
		t.Fatalf("Failed to disarm: %v", err)
	}
	
	if system.IsArmed() {
		t.Error("Should not be armed after Disarm()")
	}
}

func TestAutoStartThreeLightCount(t *testing.T) {
	system := NewAutoStartSystem(nil)
	ctx := context.Background()
	
	// Initialize the system first
	cfg := config.NewDefaultConfig()
	err := system.Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	system.Arm(ctx)
	
	// Initially should have 0 lights
	if system.GetThreeLightCount() != 0 {
		t.Errorf("Expected 0 lights initially, got %d", system.GetThreeLightCount())
	}
	
	// Add staging states using interface method
	lane1State := interfaces.LaneStagingState{
		PreStaged: true,
		Staged:    false,
	}
	
	err = system.UpdateStagingState(1, lane1State)
	if err != nil {
		t.Fatalf("Failed to update staging state: %v", err)
	}
	
	if system.GetThreeLightCount() != 1 {
		t.Errorf("Expected 1 light with pre-stage, got %d", system.GetThreeLightCount())
	}
	
	// Add second pre-stage
	err = system.UpdateStagingState(2, lane1State)
	if err != nil {
		t.Fatalf("Failed to update staging state: %v", err)
	}
	
	if system.GetThreeLightCount() != 2 {
		t.Errorf("Expected 2 lights with two pre-stages, got %d", system.GetThreeLightCount())
	}
	
	// Add stage light (should trigger three-light rule)
	lane1Staged := interfaces.LaneStagingState{
		PreStaged: true,
		Staged:    true,
	}
	
	err = system.UpdateStagingState(1, lane1Staged)
	if err != nil {
		t.Fatalf("Failed to update staging state: %v", err)
	}
	
	if system.GetThreeLightCount() != 3 {
		t.Errorf("Expected 3 lights with three-light rule, got %d", system.GetThreeLightCount())
	}
}

func TestAutoStartDeepStagingInterface(t *testing.T) {
	system := NewAutoStartSystem(nil)
	ctx := context.Background()
	
	// Initialize the system first
	cfg := config.NewDefaultConfig()
	err := system.Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	system.Arm(ctx)
	
	// Test deep staging scenario
	deepStageState := interfaces.LaneStagingState{
		PreStaged:  false, // Rolled past pre-stage beam
		Staged:     false,
		DeepStaged: true,
		Position:   2.0, // 2 inches past stage beam
	}
	
	err = system.UpdateStagingState(1, deepStageState)
	if err != nil {
		t.Fatalf("Failed to update deep staging state: %v", err)
	}
	
	// Deep staging should count as 1 light (the stage light)
	if system.GetThreeLightCount() != 1 {
		t.Errorf("Expected 1 light for deep staging, got %d", system.GetThreeLightCount())
	}
	
	// Check that the underlying system received the correct staged=true
	status := system.GetAutoStartStatus()
	if len(status.VehicleStaging) == 0 {
		t.Fatal("Expected vehicle staging data")
	}
	
	lane1Status := status.VehicleStaging[1]
	if lane1Status == nil {
		t.Fatal("Expected lane 1 staging status")
	}
	
	if !lane1Status.Staged {
		t.Error("Deep staging should result in Staged=true in underlying system")
	}
	if lane1Status.PreStaged {
		t.Error("Deep staging should result in PreStaged=false in underlying system")
	}
}