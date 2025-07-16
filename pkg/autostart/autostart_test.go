package autostart

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
	"github.com/benharold/libdrag/pkg/tree"
)

func TestAutoStartSystem_ThreeLightRule(t *testing.T) {
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
	christmasTree := tree.NewChristmasTree()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Connect tree and arm it (required for auto-start to work)
	system.SetTreeComponent(christmasTree)
	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
	}

	// Test that system starts in idle state
	status := system.GetAutoStartStatus()
	if status.State != StateIdle {
		t.Errorf("Expected StateIdle, got %v", status.State)
	}

	// Stage first vehicle pre-stage only - should not trigger
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

	// Stage one vehicle fully - should trigger three-light rule (tree is armed) and start timeout
	system.UpdateVehicleStaging(1, true, true, 0)
	time.Sleep(10 * time.Millisecond) // Allow processing
	status = system.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Errorf("Expected StateActivated after three lights, got %v", status.State)
	}
}

func TestAutoStartSystem_StagingTimeout(t *testing.T) {
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
	christmasTree := tree.NewChristmasTree()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Set test mode AFTER initialization to override the loaded config
	system.SetTestMode(true)

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Connect tree and arm it (required for auto-start to work)
	system.SetTreeComponent(christmasTree)
	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
	}

	// Trigger auto-start with three-light rule (both pre + one stage) – timeout starts on first stage
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0) // Triggers activation + starts timeout for lane 2

	// Verify we're in activated state
	time.Sleep(10 * time.Millisecond)
	status := system.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Fatalf("Expected StateActivated after three-light rule, got %v", status.State)
	}

	// Don't stage the second vehicle - let it timeout
	// Wait for timeout (in test mode, this should be 50ms)
	time.Sleep(80 * time.Millisecond) // Wait longer than the 50ms timeout

	status = system.GetAutoStartStatus()
	if status.State != StateFault {
		t.Errorf("Expected StateFault after timeout, got %v", status.State)
	}
	if !strings.Contains(status.LastFaultReason, "Staging timeout for lane 2") { // Updated to lane-specific
		t.Errorf("Expected lane 2 timeout fault, got: %v", status.LastFaultReason)
	}
}

func TestAutoStartSystem_GuardBeamViolation(t *testing.T) {
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
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
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
	christmasTree := tree.NewChristmasTree()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Set test mode AFTER initialization to override the loaded config
	system.SetTestMode(true)

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Connect tree and arm it (required for auto-start to work)
	system.SetTreeComponent(christmasTree)
	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
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

	// Step 2: One vehicle stages (triggers auto-start + timeout start)
	system.UpdateVehicleStaging(1, true, true, 0)
	time.Sleep(10 * time.Millisecond) // Allow processing
	status := system.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Errorf("Expected StateActivated, got %v", status.State)
	}

	// Step 3: Second vehicle stages (both staged – should cancel timeout)
	system.UpdateVehicleStaging(2, true, true, 0)
	time.Sleep(5 * time.Millisecond) // Shorter wait for staging detection

	// Step 4: Wait for the full sequence to complete
	// In test mode: MinStagingDuration=5ms, RandomDelay=1-3ms, so max ~8ms
	time.Sleep(15 * time.Millisecond) // Wait for complete sequence

	// At this point, the system should have progressed through staging and triggered the tree (no fault, as timeout canceled)
	status = system.GetAutoStartStatus()

	// The system might be in StateStaging or StateTriggered depending on timing
	if status.State != StateStaging && status.State != StateTriggered {
		t.Errorf("Expected StateStaging or StateTriggered, got %v", status.State)
	}
	if status.State == StateFault { // Explicit check: No fault should occur
		t.Errorf("Unexpected fault on timely second stage: %v", status.LastFaultReason)
	}

	// Verify tree was triggered
	if !treeTriggerCalled {
		t.Error("Expected tree trigger to be called")
	}

	// Wait a bit more to ensure we reach the triggered state
	time.Sleep(10 * time.Millisecond)
	status = system.GetAutoStartStatus()
	if status.State != StateTriggered {
		t.Errorf("Expected StateTriggered after tree activation, got %v", status.State)
	}
}

func TestAutoStartSystem_ManualOverride(t *testing.T) {
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
	christmasTree := tree.NewChristmasTree()
	system.SetTestMode(true)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Connect tree and arm it (required for auto-start to work)
	system.SetTreeComponent(christmasTree)
	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
	}

	// Arm auto-start sequence
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0)

	// Verify activated
	time.Sleep(10 * time.Millisecond) // Allow processing
	status := system.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Errorf("Expected StateActivated, got %v", status.State)
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
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Get initial config (should be Sportsman default)
	initialConfig := system.GetConfiguration()
	if initialConfig.StagingTimeout != 10*time.Second {
		t.Errorf("Expected 10s timeout for sportsman config, got %v", initialConfig.StagingTimeout)
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
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
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
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)

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
			expectedTimeout: 10 * time.Second,
			expectedClass:   "Sportsman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := events.NewEventBus(false)
			system := NewAutoStartSystem(eventBus)

			cfg := config.NewDefaultConfig()
			cfg.TreeConfig.Type = tt.treeType
			// Set racing class to match expected behavior
			if tt.treeType == config.TreeSequencePro {
				cfg.SetRacingClass("ProFourTenths")
			} else {
				cfg.SetRacingClass("Sportsman")
			}

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

func TestAutoStartSystem_SecondStageTimeoutAndCancel(t *testing.T) {
	eventBus := events.NewEventBus(false)
	system := NewAutoStartSystem(eventBus)
	christmasTree := tree.NewChristmasTree()

	cfg := config.NewDefaultConfig()
	err := system.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	err = christmasTree.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	system.SetTestMode(true)

	err = system.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	system.SetTreeComponent(christmasTree)
	err = christmasTree.Arm(context.Background())
	if err != nil {
		t.Fatalf("Failed to arm tree: %v", err)
	}

	// Trigger activation (three-light: both pre + one stage) – starts timeout for lane 2
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0) // Activation + first stage → timeout starts

	time.Sleep(10 * time.Millisecond)
	status := system.GetAutoStartStatus()
	if status.State != StateActivated {
		t.Fatalf("Expected StateActivated, got %v", status.State)
	}

	// Subtest 1: Timeout expiry fouls lane 2
	t.Run("Timeout Fouls Second Lane", func(t *testing.T) {
		time.Sleep(80 * time.Millisecond) // >50ms test timeout
		status = system.GetAutoStartStatus()
		if status.State != StateFault {
			t.Errorf("Expected StateFault, got %v", status.State)
		}
		if !strings.Contains(status.LastFaultReason, "Staging timeout for lane 2") {
			t.Errorf("Expected lane 2 fault, got: %v", status.LastFaultReason)
		}
	})

	// Reset for cancel subtest: Reset autostart and fully reset/disarm tree
	system.resetToIdle("Test reset")
	christmasTree.DisarmTree() // Reset tree Activated state

	// Restart and re-arm for second subtest
	system.Start(context.Background())
	christmasTree.Arm(context.Background())

	// Retrigger activation + first stage
	system.UpdateVehicleStaging(1, true, false, 0)
	system.UpdateVehicleStaging(2, true, false, 0)
	system.UpdateVehicleStaging(1, true, true, 0)
	time.Sleep(10 * time.Millisecond)

	// Subtest 2: Timely second stage cancels timeout, no fault
	t.Run("Timely Second Stage Cancels Timeout", func(t *testing.T) {
		system.UpdateVehicleStaging(2, true, true, 0) // Both staged within time → cancel
		time.Sleep(80 * time.Millisecond)             // Wait past would-be timeout
		status = system.GetAutoStartStatus()
		if status.State == StateFault {
			t.Errorf("Unexpected fault on timely stage: %v", status.LastFaultReason)
		}
		if status.State != StateStaging && status.State != StateTriggered {
			t.Errorf("Expected Staging/Triggered, got %v", status.State)
		}
	})
}
