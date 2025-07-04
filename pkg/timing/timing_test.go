package timing

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
)

func TestNewTimingSystem(t *testing.T) {
	ts := NewTimingSystem()
	if ts == nil {
		t.Fatal("NewTimingSystem returned nil")
	}
	if ts.GetID() != "timing_system" {
		t.Fatalf("Expected ID %s, got %s", "timing_system", ts.GetID())
	}
}

func TestTimingSystemInitialize(t *testing.T) {
	ts := NewTimingSystem()
	cfg := config.NewDefaultConfig()

	err := ts.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	status := ts.GetStatus()
	if status.Status != "ready" {
		t.Fatalf("Expected status 'ready', got '%s'", status.Status)
	}
}

func TestTimingSystemTestMode(t *testing.T) {
	ts := NewTimingSystem()

	// Test mode should be false by default
	ts.mu.RLock()
	defaultTestMode := ts.testMode
	ts.mu.RUnlock()

	if defaultTestMode {
		t.Fatal("Test mode should be false by default")
	}

	// Enable test mode
	ts.SetTestMode(true)

	ts.mu.RLock()
	testMode := ts.testMode
	ts.mu.RUnlock()

	if !testMode {
		t.Fatal("Test mode should be enabled")
	}

	// Disable test mode
	ts.SetTestMode(false)
	ts.mu.RLock()
	disabled := ts.testMode
	ts.mu.RUnlock()

	if disabled {
		t.Fatal("Test mode should be disabled after SetTestMode(false)")
	}
}

func TestTimingSystemStartStop(t *testing.T) {
	ts := NewTimingSystem()
	cfg := config.NewDefaultConfig()

	// Initialize first
	err := ts.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm the timing system
	err = ts.Arm(context.Background())
	if err != nil {
		t.Fatalf("Arm failed: %v", err)
	}

	status := ts.GetStatus()
	if status.Status != "running" {
		t.Fatalf("Expected status 'running', got '%s'", status.Status)
	}

	// EmergencyStop the timing system
	err = ts.EmergencyStop()
	if err != nil {
		t.Fatalf("EmergencyStop failed: %v", err)
	}

	status = ts.GetStatus()
	if status.Status != "stopped" {
		t.Fatalf("Expected status 'stopped', got '%s'", status.Status)
	}
}

func TestTimingSystemResults(t *testing.T) {
	ts := NewTimingSystem()

	// Should return nil for non-existent lane
	result := ts.GetResults(1)
	if result != nil {
		t.Fatal("Expected nil result for non-existent lane")
	}

	// Add a result and verify we can retrieve it
	ts.mu.Lock()
	ts.results[1] = &TimingResults{
		Lane:         1,
		StartTime:    time.Now(),
		IsComplete:   false,
		BeamTriggers: make(map[string]time.Time),
	}
	ts.mu.Unlock()

	result = ts.GetResults(1)
	if result == nil {
		t.Fatal("Expected result for lane 1")
	}
	if result.Lane != 1 {
		t.Fatalf("Expected lane 1, got %d", result.Lane)
	}
}

// Test reaction time calculation using direct method calls
func TestReactionTimeCalculation(t *testing.T) {
	ts := NewTimingSystem()
	cfg := config.NewDefaultConfig()

	err := ts.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm race and add vehicles
	ts.StartRace()
	ts.AddVehicles([]int{1, 2})

	// Simulate green light
	greenLightTime := time.Now()
	ts.SetGreenLight(greenLightTime)

	// Vehicle leaves starting line 0.5 seconds later
	vehicleStartTime := greenLightTime.Add(500 * time.Millisecond)
	ts.TriggerBeam("stage", 1, vehicleStartTime)

	// Check results
	result := ts.GetResults(1)
	if result == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if result.ReactionTime == nil {
		t.Fatal("Reaction time should be calculated")
	}

	expectedRT := 0.5
	if *result.ReactionTime != expectedRT {
		t.Fatalf("Expected reaction time %f, got %f", expectedRT, *result.ReactionTime)
	}
}

// Test red light detection
func TestRedLightDetection(t *testing.T) {
	ts := NewTimingSystem()
	cfg := config.NewDefaultConfig()

	err := ts.Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm race and add vehicles
	ts.StartRace()
	ts.AddVehicles([]int{1, 2})

	startTime := time.Now()

	// Vehicle leaves BEFORE green light (red light foul)
	earlyStartTime := startTime.Add(-100 * time.Millisecond)
	ts.TriggerBeam("stage", 1, earlyStartTime)

	// Green light comes after vehicle already left
	ts.SetGreenLight(startTime)

	// Check for red light foul
	result := ts.GetResults(1)
	if result == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if !result.IsFoul {
		t.Fatal("Should detect red light foul")
	}

	if result.FoulReason != "red_light" {
		t.Fatalf("Expected foul reason 'red_light', got '%s'", result.FoulReason)
	}
}
