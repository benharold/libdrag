package interfaces

import (
	"context"
	"time"
)

// StagingState represents the current staging status of both lanes
type StagingState struct {
	Lane1 LaneStagingState `json:"lane1"`
	Lane2 LaneStagingState `json:"lane2"`
}

// LaneStagingState represents staging status for a single lane
type LaneStagingState struct {
	PreStaged   bool    `json:"pre_staged"`
	Staged      bool    `json:"staged"`
	DeepStaged  bool    `json:"deep_staged"`
	Position    float64 `json:"position"`    // Distance from stage beam
	LastChanged time.Time `json:"last_changed"`
}

// TrackConditions represents current track and environmental conditions
type TrackConditions struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	TrackTemp   float64 `json:"track_temp"`
	SafeToRace  bool    `json:"safe_to_race"`
}

// RaceState represents the current state of a race
type RaceState string

const (
	RaceStateIdle       RaceState = "idle"
	RaceStateStaging    RaceState = "staging"
	RaceStateArmed      RaceState = "armed"
	RaceStateActivated  RaceState = "activated"
	RaceStateRunning    RaceState = "running"
	RaceStateComplete   RaceState = "complete"
	RaceStateAborted    RaceState = "aborted"
)

// StartDecision represents a starter's decision about race initiation
type StartDecision struct {
	CanStart      bool   `json:"can_start"`
	ShouldHold    bool   `json:"should_hold"`
	Reason        string `json:"reason"`
	ManualOverride bool  `json:"manual_override"`
}

// Starter interface defines the operations a race starter can perform
type Starter interface {
	// Tree Control - starter's primary responsibilities
	ArmTree(ctx context.Context) error
	HoldTree(ctx context.Context, reason string) error
	ReleaseTree(ctx context.Context) error
	ManualStart(ctx context.Context) error
	AbortRace(ctx context.Context, reason string) error
	
	// Decision Making - starter judgment calls
	ShouldHonorDeepStaging(lane int, class string) bool
	ShouldWaitForStaging(conditions TrackConditions) bool
	CanStartRace(stagingState StagingState) StartDecision
	
	// State Management
	IsTreeArmed() bool
	IsTreeHeld() bool
	GetLastDecision() StartDecision
}

// AutoStart interface defines auto-start system operations
type AutoStart interface {
	// State Management
	Arm(ctx context.Context) error
	Disarm(ctx context.Context) error
	Hold(ctx context.Context) error
	Release(ctx context.Context) error
	
	// Staging Monitoring
	UpdateStagingState(lane int, state LaneStagingState) error
	IsActivationReady() bool
	GetThreeLightCount() int
	
	// Integration with Starter
	SetStarterOverride(enabled bool) error
	IsStarterOverrideActive() bool
	
	// State Queries
	IsArmed() bool
	IsHeld() bool
	GetActivationStatus() string
}

// RaceControl interface defines race orchestration operations
type RaceControl interface {
	// Race Management
	StartRace(ctx context.Context, participants []int) error
	HandleByeRun(ctx context.Context, lane int) error
	AbortRace(ctx context.Context, reason string) error
	
	// Mode Control
	SetManualMode(ctx context.Context, enabled bool) error
	SetAutoStartMode(ctx context.Context, enabled bool) error
	
	// State Queries
	GetRaceState() RaceState
	IsManualMode() bool
	IsAutoStartEnabled() bool
	
	// Component Coordination
	GetStagingState() StagingState
	GetTrackConditions() TrackConditions
}

// BeamSensor interface defines operations for individual beam sensors
type BeamSensor interface {
	// State Management
	IsTriggered() bool
	GetLastTrigger() time.Time
	GetPosition() float64
	
	// Calibration
	Calibrate(ctx context.Context) error
	Test(ctx context.Context) error
	
	// State Changes
	OnTrigger(callback func(timestamp time.Time)) error
	OnRestore(callback func(timestamp time.Time)) error
}

// StagingAnalyzer interface defines staging state analysis
type StagingAnalyzer interface {
	// State Analysis
	AnalyzeStagingState(preStageBeam, stageBeam BeamSensor) LaneStagingState
	DetectDeepStaging(preStageBeam, stageBeam BeamSensor) bool
	ValidateStageSequence(currentState, newState LaneStagingState) error
	
	// Rule Enforcement
	IsDeepStagingAllowed(class string) bool
	GetMaxStagingTime(class string) time.Duration
	GetMinStagingTime(class string) time.Duration
}

// TimingControl interface defines timing system operations
type TimingControl interface {
	// Race Timing
	StartTiming(ctx context.Context, greenLightTime time.Time) error
	StopTiming(ctx context.Context) error
	ResetTiming(ctx context.Context) error
	
	// Beam Integration
	RegisterBeamTrigger(beamID string, lane int, timestamp time.Time) error
	
	// Results
	GetReactionTime(lane int) (float64, error)
	GetElapsedTime(lane int) (float64, error)
	IsComplete(lane int) bool
	HasFoul(lane int) bool
}