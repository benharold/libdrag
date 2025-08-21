package interfaces

import (
	"context"
	"testing"
)

// TestRaceControl is a test implementation of the RaceControl interface
type TestRaceControl struct {
	raceState        RaceState
	manualMode       bool
	autoStartEnabled bool
	stagingState     StagingState
	trackConditions  TrackConditions
	participants     []int
	lastAbortReason  string
}

func NewTestRaceControl() *TestRaceControl {
	return &TestRaceControl{
		raceState:        RaceStateIdle,
		manualMode:       false,
		autoStartEnabled: true,
		stagingState: StagingState{
			Lane1: LaneStagingState{},
			Lane2: LaneStagingState{},
		},
		trackConditions: TrackConditions{
			Temperature: 75.0,
			Humidity:    45.0,
			WindSpeed:   5.0,
			TrackTemp:   120.0,
			SafeToRace:  true,
		},
		participants: make([]int, 0),
	}
}

// RaceControl interface implementation
func (trc *TestRaceControl) StartRace(ctx context.Context, participants []int) error {
	trc.participants = participants
	
	// Validate participants
	if len(participants) == 0 {
		return ErrNoParticipants
	}
	
	// Check if bye run
	if len(participants) == 1 {
		trc.raceState = RaceStateStaging // Bye runs go to staging immediately
	} else {
		trc.raceState = RaceStateStaging
	}
	
	return nil
}

func (trc *TestRaceControl) HandleByeRun(ctx context.Context, lane int) error {
	trc.participants = []int{lane}
	trc.raceState = RaceStateStaging
	trc.manualMode = true // Bye runs require manual control
	return nil
}

func (trc *TestRaceControl) AbortRace(ctx context.Context, reason string) error {
	trc.lastAbortReason = reason
	trc.raceState = RaceStateAborted
	trc.participants = make([]int, 0)
	return nil
}

func (trc *TestRaceControl) SetManualMode(ctx context.Context, enabled bool) error {
	trc.manualMode = enabled
	return nil
}

func (trc *TestRaceControl) SetAutoStartMode(ctx context.Context, enabled bool) error {
	trc.autoStartEnabled = enabled
	return nil
}

func (trc *TestRaceControl) GetRaceState() RaceState {
	return trc.raceState
}

func (trc *TestRaceControl) IsManualMode() bool {
	return trc.manualMode
}

func (trc *TestRaceControl) IsAutoStartEnabled() bool {
	return trc.autoStartEnabled
}

func (trc *TestRaceControl) GetStagingState() StagingState {
	return trc.stagingState
}

func (trc *TestRaceControl) GetTrackConditions() TrackConditions {
	return trc.trackConditions
}

// Test helper methods
func (trc *TestRaceControl) setStagingState(lane1, lane2 LaneStagingState) {
	trc.stagingState = StagingState{
		Lane1: lane1,
		Lane2: lane2,
	}
}

func (trc *TestRaceControl) setTrackConditions(conditions TrackConditions) {
	trc.trackConditions = conditions
}

// Custom errors for testing
var (
	ErrNoParticipants = &RaceControlError{"no participants provided"}
)

type RaceControlError struct {
	Message string
}

func (e *RaceControlError) Error() string {
	return e.Message
}

// Test cases for RaceControl interface behavior

func TestRaceControlBasicOperations(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test initial state
	if raceControl.GetRaceState() != RaceStateIdle {
		t.Errorf("Expected initial state %s, got %s", RaceStateIdle, raceControl.GetRaceState())
	}
	
	if raceControl.IsManualMode() {
		t.Error("Should not be in manual mode initially")
	}
	
	if !raceControl.IsAutoStartEnabled() {
		t.Error("AutoStart should be enabled initially")
	}
	
	// Test starting a normal race
	participants := []int{1, 2}
	err := raceControl.StartRace(ctx, participants)
	if err != nil {
		t.Fatalf("Failed to start race: %v", err)
	}
	
	if raceControl.GetRaceState() != RaceStateStaging {
		t.Errorf("Expected state %s after start, got %s", RaceStateStaging, raceControl.GetRaceState())
	}
	
	// Test aborting race
	abortReason := "oil_on_track"
	err = raceControl.AbortRace(ctx, abortReason)
	if err != nil {
		t.Fatalf("Failed to abort race: %v", err)
	}
	
	if raceControl.GetRaceState() != RaceStateAborted {
		t.Errorf("Expected state %s after abort, got %s", RaceStateAborted, raceControl.GetRaceState())
	}
	
	if raceControl.lastAbortReason != abortReason {
		t.Errorf("Expected abort reason '%s', got '%s'", abortReason, raceControl.lastAbortReason)
	}
}

func TestRaceControlModeManagement(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test manual mode
	err := raceControl.SetManualMode(ctx, true)
	if err != nil {
		t.Fatalf("Failed to set manual mode: %v", err)
	}
	
	if !raceControl.IsManualMode() {
		t.Error("Should be in manual mode")
	}
	
	// Test autostart disable
	err = raceControl.SetAutoStartMode(ctx, false)
	if err != nil {
		t.Fatalf("Failed to disable autostart: %v", err)
	}
	
	if raceControl.IsAutoStartEnabled() {
		t.Error("AutoStart should be disabled")
	}
	
	// Test returning to auto mode
	err = raceControl.SetManualMode(ctx, false)
	if err != nil {
		t.Fatalf("Failed to disable manual mode: %v", err)
	}
	
	err = raceControl.SetAutoStartMode(ctx, true)
	if err != nil {
		t.Fatalf("Failed to enable autostart: %v", err)
	}
	
	if raceControl.IsManualMode() {
		t.Error("Should not be in manual mode")
	}
	
	if !raceControl.IsAutoStartEnabled() {
		t.Error("AutoStart should be enabled")
	}
}

func TestRaceControlByeRunHandling(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test bye run
	err := raceControl.HandleByeRun(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to handle bye run: %v", err)
	}
	
	if raceControl.GetRaceState() != RaceStateStaging {
		t.Errorf("Expected state %s for bye run, got %s", RaceStateStaging, raceControl.GetRaceState())
	}
	
	if !raceControl.IsManualMode() {
		t.Error("Bye run should enable manual mode")
	}
	
	if len(raceControl.participants) != 1 || raceControl.participants[0] != 1 {
		t.Error("Bye run should have single participant in lane 1")
	}
}

func TestRaceControlSingleCarRace(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test single car race (competition single)
	participants := []int{2}
	err := raceControl.StartRace(ctx, participants)
	if err != nil {
		t.Fatalf("Failed to start single car race: %v", err)
	}
	
	if raceControl.GetRaceState() != RaceStateStaging {
		t.Errorf("Expected state %s for single car race, got %s", RaceStateStaging, raceControl.GetRaceState())
	}
}

func TestRaceControlErrorHandling(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test starting race with no participants
	err := raceControl.StartRace(ctx, []int{})
	if err == nil {
		t.Error("Should fail to start race with no participants")
	}
	
	if raceControl.GetRaceState() != RaceStateIdle {
		t.Error("State should remain idle when start fails")
	}
}

func TestRaceControlStagingStateIntegration(t *testing.T) {
	raceControl := NewTestRaceControl()
	
	// Test getting initial staging state
	stagingState := raceControl.GetStagingState()
	if stagingState.Lane1.PreStaged || stagingState.Lane1.Staged {
		t.Error("Lane 1 should not be staged initially")
	}
	if stagingState.Lane2.PreStaged || stagingState.Lane2.Staged {
		t.Error("Lane 2 should not be staged initially")
	}
	
	// Test updating staging state
	lane1State := LaneStagingState{
		PreStaged:  true,
		Staged:     false,
		DeepStaged: false,
		Position:   -7.0,
	}
	
	lane2State := LaneStagingState{
		PreStaged:  true,
		Staged:     true,
		DeepStaged: false,
		Position:   0.0,
	}
	
	raceControl.setStagingState(lane1State, lane2State)
	
	updatedState := raceControl.GetStagingState()
	if !updatedState.Lane1.PreStaged {
		t.Error("Lane 1 should be pre-staged")
	}
	if updatedState.Lane1.Staged {
		t.Error("Lane 1 should not be staged")
	}
	if !updatedState.Lane2.PreStaged || !updatedState.Lane2.Staged {
		t.Error("Lane 2 should be pre-staged and staged")
	}
}

func TestRaceControlTrackConditionsIntegration(t *testing.T) {
	raceControl := NewTestRaceControl()
	
	// Test getting track conditions
	conditions := raceControl.GetTrackConditions()
	if !conditions.SafeToRace {
		t.Error("Track should be safe to race initially")
	}
	
	// Test unsafe conditions
	unsafeConditions := TrackConditions{
		Temperature: 100.0, // Very hot
		Humidity:    85.0,  // High humidity
		WindSpeed:   30.0,  // High wind
		TrackTemp:   150.0, // Very hot track
		SafeToRace:  false,
	}
	
	raceControl.setTrackConditions(unsafeConditions)
	
	updatedConditions := raceControl.GetTrackConditions()
	if updatedConditions.SafeToRace {
		t.Error("Track should not be safe to race with unsafe conditions")
	}
	if updatedConditions.WindSpeed != 30.0 {
		t.Errorf("Expected wind speed 30.0, got %f", updatedConditions.WindSpeed)
	}
}

func TestRaceControlStateMachine(t *testing.T) {
	raceControl := NewTestRaceControl()
	ctx := context.Background()
	
	// Test state progression
	states := []RaceState{RaceStateIdle}
	
	// Start race -> Staging
	raceControl.StartRace(ctx, []int{1, 2})
	states = append(states, raceControl.GetRaceState())
	
	// Abort race -> Aborted
	raceControl.AbortRace(ctx, "test_abort")
	states = append(states, raceControl.GetRaceState())
	
	expectedStates := []RaceState{RaceStateIdle, RaceStateStaging, RaceStateAborted}
	
	for i, expected := range expectedStates {
		if i < len(states) && states[i] != expected {
			t.Errorf("State transition %d: expected %s, got %s", i, expected, states[i])
		}
	}
}