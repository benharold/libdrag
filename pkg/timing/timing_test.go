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

// Test beam initialization with drag racing standard positions
func TestTimingBeamInitialization(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify standard drag racing beam positions are initialized
	expectedBeams := []string{"pre_stage", "stage", "60_foot", "330_foot", "660_foot", "1000_foot", "1320_foot"}

	ts.mu.RLock()
	beams := ts.beams
	ts.mu.RUnlock()

	for _, beamID := range expectedBeams {
		if beam, exists := beams[beamID]; !exists {
			t.Fatalf("Expected beam %s not found", beamID)
		} else {
			// Verify beam is active and not triggered initially
			if !beam.IsActive {
				t.Fatalf("Beam %s should be active initially", beamID)
			}
			if beam.IsTriggered {
				t.Fatalf("Beam %s should not be triggered initially", beamID)
			}
		}
	}
}

// Test reaction time calculation (green light to vehicle movement)
func TestReactionTimeCalculation(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	// Enable test mode for predictable timing
	ts.SetTestMode(true)

	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = ts.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := context.Background()

	// First, start the race (this resets timing results)
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{},
	}
	ts.HandleEvent(ctx, raceStartEvent)

	// Add vehicles to lanes (this creates timing results)
	vehicleEvent := &events.BaseEvent{
		Type:      events.EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{"lanes": []int{1, 2}},
	}
	ts.HandleEvent(ctx, vehicleEvent)

	// Simulate green light event
	greenLightTime := time.Now()
	greenEvent := &events.BaseEvent{
		Type:      events.EventGreenLight,
		Timestamp: greenLightTime,
		Source:    events.ComponentChristmasTree,
		Data:      map[string]interface{}{"all_lanes": true},
	}

	err = ts.HandleEvent(ctx, greenEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Simulate vehicle leaving starting line (stage beam break) 0.5 seconds later
	vehicleStartTime := greenLightTime.Add(500 * time.Millisecond)
	stageBreakEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: vehicleStartTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "stage",
			"lane":     1,
			"position": 0.0,
		},
	}

	err = ts.HandleEvent(ctx, stageBreakEvent)
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// Get timing results for lane 1
	laneResult := ts.GetResults(1)
	if laneResult == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if laneResult.ReactionTime == nil {
		t.Fatal("Reaction time should be calculated")
	}

	expectedRT := 0.5 // 500ms reaction time
	if *laneResult.ReactionTime != expectedRT {
		t.Fatalf("Expected reaction time %f, got %f", expectedRT, *laneResult.ReactionTime)
	}
}

// Test 60-foot time calculation
func TestSixtyFootTiming(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	ts.SetTestMode(true)

	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = ts.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := context.Background()

	// Start race and add vehicles
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{},
	}
	ts.HandleEvent(ctx, raceStartEvent)

	vehicleEvent := &events.BaseEvent{
		Type:      events.EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{"lanes": []int{1, 2}},
	}
	ts.HandleEvent(ctx, vehicleEvent)

	// Simulate race start
	startTime := time.Now()
	greenEvent := &events.BaseEvent{
		Type:      events.EventGreenLight,
		Timestamp: startTime,
		Source:    events.ComponentChristmasTree,
		Data:      map[string]interface{}{"all_lanes": true},
	}
	ts.HandleEvent(ctx, greenEvent)

	// Vehicle leaves starting line
	stageTime := startTime.Add(400 * time.Millisecond) // 0.4s reaction time
	stageEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: stageTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "stage",
			"lane":     1,
			"position": 0.0,
		},
	}
	ts.HandleEvent(ctx, stageEvent)

	// Vehicle hits 60-foot beam 1.0 seconds after leaving start line
	sixtyFootTime := stageTime.Add(1000 * time.Millisecond)
	sixtyFootEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: sixtyFootTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "60_foot",
			"lane":     1,
			"position": 60.0,
		},
	}
	ts.HandleEvent(ctx, sixtyFootEvent)

	// Check 60-foot time for lane 1
	laneResult := ts.GetResults(1)
	if laneResult == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if laneResult.SixtyFootTime == nil {
		t.Fatal("60-foot time should be calculated")
	}

	expected60ft := 1.0 // 1.0 second from start line to 60 feet
	if *laneResult.SixtyFootTime != expected60ft {
		t.Fatalf("Expected 60-foot time %f, got %f", expected60ft, *laneResult.SixtyFootTime)
	}
}

// Test quarter-mile ET calculation
func TestQuarterMileElapsedTime(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	ts.SetTestMode(true)

	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = ts.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := context.Background()

	// Start race and add vehicles
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{},
	}
	ts.HandleEvent(ctx, raceStartEvent)

	vehicleEvent := &events.BaseEvent{
		Type:      events.EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{"lanes": []int{1, 2}},
	}
	ts.HandleEvent(ctx, vehicleEvent)

	// Simulate complete race sequence
	startTime := time.Now()

	// Green light
	greenEvent := &events.BaseEvent{
		Type:      events.EventGreenLight,
		Timestamp: startTime,
		Source:    events.ComponentChristmasTree,
		Data:      map[string]interface{}{"all_lanes": true},
	}
	ts.HandleEvent(ctx, greenEvent)

	// Vehicle leaves starting line
	stageTime := startTime.Add(350 * time.Millisecond) // 0.35s reaction time
	stageEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: stageTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "stage",
			"lane":     1,
			"position": 0.0,
		},
	}
	ts.HandleEvent(ctx, stageEvent)

	// Vehicle finishes quarter-mile 10.5 seconds after leaving start line
	finishTime := stageTime.Add(10500 * time.Millisecond)
	finishEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: finishTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "1320_foot",
			"lane":     1,
			"position": 1320.0,
		},
	}
	ts.HandleEvent(ctx, finishEvent)

	// Check quarter-mile ET for lane 1
	laneResult := ts.GetResults(1)
	if laneResult == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if laneResult.QuarterMileTime == nil {
		t.Fatal("Quarter-mile time should be calculated")
	}

	expectedET := 10.5 // 10.5 second ET
	if *laneResult.QuarterMileTime != expectedET {
		t.Fatalf("Expected quarter-mile ET %f, got %f", expectedET, *laneResult.QuarterMileTime)
	}

	// Verify reaction time is also calculated
	if laneResult.ReactionTime == nil {
		t.Fatal("Reaction time should be calculated")
	}

	expectedRT := 0.35
	if *laneResult.ReactionTime != expectedRT {
		t.Fatalf("Expected reaction time %f, got %f", expectedRT, *laneResult.ReactionTime)
	}
}

// Test red light detection (jumping the start)
func TestRedLightDetection(t *testing.T) {
	ts := NewTimingSystem()
	bus := events.NewSimpleEventBus()
	cfg := config.NewDefaultConfig()

	ts.SetTestMode(true)

	err := ts.Initialize(context.Background(), bus, cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = ts.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := context.Background()

	// Start race and add vehicles
	raceStartEvent := &events.BaseEvent{
		Type:      events.EventRaceStarted,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{},
	}
	ts.HandleEvent(ctx, raceStartEvent)

	vehicleEvent := &events.BaseEvent{
		Type:      events.EventVehicleEntered,
		Timestamp: time.Now(),
		Source:    events.ComponentOrchestrator,
		Data:      map[string]interface{}{"lanes": []int{1, 2}},
	}
	ts.HandleEvent(ctx, vehicleEvent)

	startTime := time.Now()

	// Vehicle leaves before green light (red light foul)
	earlyStartTime := startTime.Add(-100 * time.Millisecond) // Leave 100ms before green
	earlyStageEvent := &events.BaseEvent{
		Type:      events.EventBeamTriggered,
		Timestamp: earlyStartTime,
		Source:    events.ComponentTimingSystem,
		Data: map[string]interface{}{
			"beam_id":  "stage",
			"lane":     1,
			"position": 0.0,
		},
	}
	ts.HandleEvent(ctx, earlyStageEvent)

	// Green light comes after vehicle already left
	greenEvent := &events.BaseEvent{
		Type:      events.EventGreenLight,
		Timestamp: startTime,
		Source:    events.ComponentChristmasTree,
		Data:      map[string]interface{}{"all_lanes": true},
	}
	ts.HandleEvent(ctx, greenEvent)

	// Check for red light foul on lane 1
	laneResult := ts.GetResults(1)
	if laneResult == nil {
		t.Fatal("No timing results found for lane 1")
	}

	if !laneResult.IsFoul {
		t.Fatal("Should detect red light foul")
	}

	if laneResult.FoulReason != "red_light" {
		t.Fatalf("Expected foul reason 'red_light', got '%s'", laneResult.FoulReason)
	}

	// Reaction time should be negative (early start)
	if laneResult.ReactionTime == nil {
		t.Fatal("Reaction time should still be calculated for red light")
	}

	if *laneResult.ReactionTime >= 0 {
		t.Fatal("Reaction time should be negative for red light foul")
	}
}
