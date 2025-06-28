package timing

import (
	"context"
	"testing"
	"time"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

func TestNewTimingSystem(t *testing.T) {
	ts := NewTimingSystem()
	if ts == nil {
		t.Fatal("NewTimingSystem returned nil")
	}
	if ts.GetID() != events.ComponentTimingSystem {
		t.Fatalf("Expected ID %s, got %s", events.ComponentTimingSystem, ts.GetID())
	}
}

func TestTimingSystemInitialize(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	err := ts.Initialize(context.Background(), bus, cfg)
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
	enabled := ts.testMode
	ts.mu.RUnlock()

	if !enabled {
		t.Fatal("Test mode should be enabled after SetTestMode(true)")
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
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Initialize first
	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start the timing system
	err = ts.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	status := ts.GetStatus()
	if status.Status != "running" {
		t.Fatalf("Expected status 'running', got '%s'", status.Status)
	}

	// Stop the timing system
	err = ts.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
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
