package autostart

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
)

func TestAutoStartSystem_ThreeLightRule(t *testing.T) {
	system := NewAutoStartSystem()
	system.SetTestMode(true)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Test that system starts in idle state
	status := system.GetAutoStartStatus()
	if status.State != StateIdle {
		t.Errorf("Expected StateIdle, got %v", status.State)
	}

	// Stage one vehicle pre-stage only - should not trigger
	system.UpdateVehicleStaging(1, true, false, 0)
	status = system.GetAutoStartStatus()
	if status.State != StateIdle {
		t.Errorf("Expected StateIdle after one pre-stage, got %v", status.State)
	}

	// Stage second vehicle pre-stage only - should not trigger
	system.UpdateVehicleStaging(2, true, false, 0)
	status = system.GetAutoStartStatus()
	if status.State != StateIdle {
		t.Errorf("Expected StateIdle after two pre-stages, got %v", status.State)
	}

	// Stage one vehicle fully - should trigger three-light rule
	system.UpdateVehicleStaging(1, true, true, 0)
	time.Sleep(10 * time.Millisecond) // Allow processing
	status = system.GetAutoStartStatus()
	if status.State != StateArmed {
		t.Errorf("Expected StateArmed after three lights, got %v", status.State)
	}
}

func TestAutoStartSystem_StagingTimeout(t *testing.T) {
	system := NewAutoStartSystem()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Set test mode AFTER initialization to override the loaded config
	system.SetTestMode(true)

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Trigger auto-start with three-light rule
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0) // This triggers auto-start

	// Verify we're in armed state
	time.Sleep(10 * time.Millisecond)
	status := system.GetAutoStartStatus()
	if status.State != StateArmed {
		t.Fatalf("Expected StateArmed after three-light rule, got %v", status.State)
	}

	// Don't stage the second vehicle - let it timeout
	// Wait for timeout (in test mode, this should be 50ms)
	time.Sleep(80 * time.Millisecond) // Wait longer than the 50ms timeout

	status = system.GetAutoStartStatus()
	if status.State != StateFault {
		t.Errorf("Expected StateFault after timeout, got %v", status.State)
	}
	if status.LastFaultReason != "Staging timeout exceeded" {
		t.Errorf("Expected timeout fault, got: %v", status.LastFaultReason)
	}
}

func TestAutoStartSystem_GuardBeamViolation(t *testing.T) {
	system := NewAutoStartSystem()
	system.SetTestMode(true)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Trigger guard beam violation (excessive rollout)
	system.UpdateVehicleStaging(1, true, true, 10.0) // 10 inches rollout exceeds limit

	status := system.GetAutoStartStatus()
	if status.State != StateFault {
		t.Errorf("Expected StateFault after guard beam violation, got %v", status.State)
	}
	if !status.VehicleStaging[1].GuardTrip {
		t.Error("Expected guard trip to be set")
	}
}

func TestAutoStartSystem_FullStagingSequence(t *testing.T) {
	system := NewAutoStartSystem()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Set test mode AFTER initialization to override the loaded config
	system.SetTestMode(true)

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Track tree trigger
	treeTriggerCalled := false
	system.SetTreeTriggerHandler(func() error {
		treeTriggerCalled = true
		return nil
	})

	// Step 1: Both vehicles pre-stage
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)

	// Step 2: One vehicle stages (triggers auto-start)
	system.UpdateVehicleStaging(1, true, true, 0)
	time.Sleep(10 * time.Millisecond) // Allow processing
	status := system.GetAutoStartStatus()
	if status.State != StateArmed {
		t.Errorf("Expected StateArmed, got %v", status.State)
	}

	// Step 3: Second vehicle stages (both staged)
	system.UpdateVehicleStaging(2, true, true, 0)
	time.Sleep(15 * time.Millisecond) // Allow staging detection
	status = system.GetAutoStartStatus()
	if status.State != StateStaging {
		t.Errorf("Expected StateStaging after both staged, got %v", status.State)
	}

	// Step 4: Wait for minimum staging duration + random delay
	// In test mode: MinStagingDuration=5ms, RandomDelay=1-3ms, so max ~8ms
	time.Sleep(20 * time.Millisecond) // Wait longer than max possible delay

	// Verify tree was triggered
	if !treeTriggerCalled {
		t.Error("Expected tree trigger to be called")
	}

	status = system.GetAutoStartStatus()
	if status.State != StateTriggered {
		t.Errorf("Expected StateTriggered after tree activation, got %v", status.State)
	}
}

func TestAutoStartSystem_ManualOverride(t *testing.T) {
	system := NewAutoStartSystem()
	system.SetTestMode(true)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Start auto-start sequence
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0)

	// Verify armed
	status := system.GetAutoStartStatus()
	if status.State != StateArmed {
		t.Errorf("Expected StateArmed, got %v", status.State)
	}

	// Manual override
	system.ManualOverride()

	status = system.GetAutoStartStatus()
	if status.State != StateIdle {
		t.Errorf("Expected StateIdle after override, got %v", status.State)
	}
	if !status.OverrideActive {
		t.Error("Expected override to be active")
	}
	if status.IsEnabled {
		t.Error("Expected auto-start to be disabled")
	}
}

func TestAutoStartSystem_ConfigurationUpdate(t *testing.T) {
	system := NewAutoStartSystem()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Get initial config
	initialConfig := system.GetConfiguration()
	if initialConfig.StagingTimeout != 7*time.Second {
		t.Errorf("Expected 7s timeout for pro config, got %v", initialConfig.StagingTimeout)
	}

	// Update to sportsman configuration
	newConfig := initialConfig
	newConfig.StagingTimeout = 15 * time.Second
	newConfig.TreeSequenceType = config.TreeSequenceSportsman
	newConfig.RacingClass = "Bracket"

	system.UpdateConfiguration(newConfig)

	// Verify update
	updatedConfig := system.GetConfiguration()
	if updatedConfig.StagingTimeout != 15*time.Second {
		t.Errorf("Expected 15s timeout after update, got %v", updatedConfig.StagingTimeout)
	}
	if updatedConfig.TreeSequenceType != config.TreeSequenceSportsman {
		t.Errorf("Expected sportsman tree, got %v", updatedConfig.TreeSequenceType)
	}
}

func TestAutoStartSystem_EventHandlers(t *testing.T) {
	system := NewAutoStartSystem()
	system.SetTestMode(true)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Track events
	var stateChanges []AutoStartState
	var faultReasons []string

	system.SetStateChangeHandler(func(oldState, newState AutoStartState) {
		stateChanges = append(stateChanges, newState)
	})

	system.SetFaultHandler(func(reason string) {
		faultReasons = append(faultReasons, reason)
	})

	// Trigger a fault
	system.UpdateVehicleStaging(1, true, true, 15.0) // Excessive rollout

	time.Sleep(10 * time.Millisecond) // Allow event processing

	// Verify events were triggered
	if len(stateChanges) == 0 {
		t.Error("Expected state change events")
	}
	if len(faultReasons) == 0 {
		t.Error("Expected fault event")
	}

	// Check for fault state
	found := false
	for _, state := range stateChanges {
		if state == StateFault {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected StateFault in state changes")
	}
}

func TestAutoStartSystem_RandomDelayCalculation(t *testing.T) {
	system := NewAutoStartSystem()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test multiple delay calculations
	delays := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		delays[i] = system.calculateRandomDelay()
	}

	// Verify delays are within expected range
	config := system.GetConfiguration()
	minExpected := config.RandomDelayMin
	maxExpected := config.RandomDelayMax + config.RandomVariation

	for i, delay := range delays {
		if delay < minExpected {
			t.Errorf("Delay %d too short: %v < %v", i, delay, minExpected)
		}
		if delay > maxExpected {
			t.Errorf("Delay %d too long: %v > %v", i, delay, maxExpected)
		}
	}

	// Verify there's some variation in the delays
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("All delays are the same - randomization not working")
	}
}

func TestAutoStartSystem_ClassSpecificConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		treeType        config.TreeSequenceType
		expectedTimeout time.Duration
		expectedClass   string
	}{
		{
			name:            "Professional Tree",
			treeType:        config.TreeSequencePro,
			expectedTimeout: 7 * time.Second,
			expectedClass:   "Professional",
		},
		{
			name:            "Sportsman Tree",
			treeType:        config.TreeSequenceSportsman,
			expectedTimeout: 15 * time.Second,
			expectedClass:   "Sportsman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewAutoStartSystem()

			cfg := config.NewDefaultConfig()
			cfg.TreeConfig.Type = tt.treeType

			err := system.Initialize(context.Background(), cfg)
			if err != nil {
				t.Fatalf("Failed to initialize: %v", err)
			}

			config := system.GetConfiguration()
			if config.StagingTimeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %v, got %v", tt.expectedTimeout, config.StagingTimeout)
			}
			if config.RacingClass != tt.expectedClass {
				t.Errorf("Expected class %v, got %v", tt.expectedClass, config.RacingClass)
			}
			if config.TreeSequenceType != tt.treeType {
				t.Errorf("Expected tree type %v, got %v", tt.treeType, config.TreeSequenceType)
			}
		})
	}
}
