package autostart

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/timing"
	"github.com/benharold/libdrag/pkg/tree"
)

// AutoStartIntegration coordinates the auto-start system with timing and tree components
type AutoStartIntegration struct {
	autoStart     *AutoStartSystem
	timingSystem  *timing.TimingSystem
	christmasTree *tree.ChristmasTree
	mu            sync.RWMutex
	running       bool

	// Beam monitoring
	beamStates    map[string]*BeamState
	preStageBeams map[int]string // lane -> beam ID
	stageBeams    map[int]string // lane -> beam ID
	guardBeams    map[int]string // lane -> beam ID
}

// BeamState tracks individual beam status for auto-start logic
type BeamState struct {
	ID          string
	Lane        int
	IsTriggered bool
	LastChange  time.Time
	Position    float64
}

// NewAutoStartIntegration creates a new integration instance
func NewAutoStartIntegration(timingSystem *timing.TimingSystem, christmasTree *tree.ChristmasTree) *AutoStartIntegration {
	integration := &AutoStartIntegration{
		autoStart:     NewAutoStartSystem(),
		timingSystem:  timingSystem,
		christmasTree: christmasTree,
		beamStates:    make(map[string]*BeamState),
		preStageBeams: make(map[int]string),
		stageBeams:    make(map[int]string),
		guardBeams:    make(map[int]string),
	}

	// Set up auto-start event handlers
	integration.setupEventHandlers()

	return integration
}

// Initialize initializes the integration with configuration
func (asi *AutoStartIntegration) Initialize(ctx context.Context, cfg config.Config) error {
	// Initialize auto-start system
	if err := asi.autoStart.Initialize(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialize auto-start: %w", err)
	}

	// Map beams based on track configuration
	trackConfig := cfg.Track()
	for beamID, beamConfig := range trackConfig.BeamLayout {
		asi.beamStates[beamID] = &BeamState{
			ID:       beamID,
			Lane:     beamConfig.Lane,
			Position: beamConfig.Position,
		}

		// Map beams to their functions based on position
		if beamConfig.Position == -7.0 { // Pre-stage beam
			if beamConfig.Lane == 0 { // Both lanes
				asi.preStageBeams[1] = beamID + "_lane1"
				asi.preStageBeams[2] = beamID + "_lane2"
			} else {
				asi.preStageBeams[beamConfig.Lane] = beamID
			}
		} else if beamConfig.Position == 0.0 { // Stage beam
			if beamConfig.Lane == 0 { // Both lanes
				asi.stageBeams[1] = beamID + "_lane1"
				asi.stageBeams[2] = beamID + "_lane2"
			} else {
				asi.stageBeams[beamConfig.Lane] = beamID
			}
		} else if beamConfig.Position > 0 && beamConfig.Position < 20 { // Guard beam area
			if beamConfig.Lane == 0 { // Both lanes
				asi.guardBeams[1] = beamID + "_lane1"
				asi.guardBeams[2] = beamID + "_lane2"
			} else {
				asi.guardBeams[beamConfig.Lane] = beamID
			}
		}
	}

	return nil
}

// Start starts the integration
func (asi *AutoStartIntegration) Start(ctx context.Context) error {
	asi.mu.Lock()
	defer asi.mu.Unlock()

	if asi.running {
		return fmt.Errorf("auto-start integration already running")
	}

	// Start auto-start system
	if err := asi.autoStart.Start(ctx); err != nil {
		return fmt.Errorf("failed to start auto-start system: %w", err)
	}

	// Start monitoring timing beam triggers
	go asi.monitorTimingBeams(ctx)

	asi.running = true
	return nil
}

// Stop stops the integration
func (asi *AutoStartIntegration) Stop(ctx context.Context) error {
	asi.mu.Lock()
	defer asi.mu.Unlock()

	if !asi.running {
		return nil
	}

	asi.running = false
	return asi.autoStart.Stop(ctx)
}

// setupEventHandlers configures auto-start event callbacks
func (asi *AutoStartIntegration) setupEventHandlers() {
	// Handle tree trigger requests
	asi.autoStart.SetTreeTriggerHandler(func() error {
		return asi.triggerChristmasTree()
	})

	// Handle fault conditions
	asi.autoStart.SetFaultHandler(func(reason string) {
		asi.handleAutoStartFault(reason)
	})

	// Handle state changes
	asi.autoStart.SetStateChangeHandler(func(oldState, newState AutoStartState) {
		asi.handleStateChange(oldState, newState)
	})
}

// monitorTimingBeams watches for beam state changes and updates auto-start
func (asi *AutoStartIntegration) monitorTimingBeams(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Millisecond) // High frequency monitoring
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !asi.running {
				return
			}

			asi.updateBeamStates()
		}
	}
}

// updateBeamStates checks timing system for beam changes and updates auto-start
func (asi *AutoStartIntegration) updateBeamStates() {
	// Get current beam statuses from timing system
	timingStatus := asi.timingSystem.GetStatus()
	if timingStatus.Status != "running" {
		return
	}

	// Check each lane for staging changes
	for lane := 1; lane <= 2; lane++ {
		asi.updateLaneStaging(lane)
	}
}

// updateLaneStaging calculates staging status for a specific lane
func (asi *AutoStartIntegration) updateLaneStaging(lane int) {
	preStageBeamID := asi.preStageBeams[lane]
	stageBeamID := asi.stageBeams[lane]

	if preStageBeamID == "" || stageBeamID == "" {
		return // No beams configured for this lane
	}

	// Get beam statuses from timing system
	preStageTriggered := asi.isBeamTriggered(preStageBeamID)
	stageTriggered := asi.isBeamTriggered(stageBeamID)

	// Calculate vehicle position for rollout detection
	position := asi.calculateVehiclePosition(lane, stageTriggered)

	// Update auto-start system with current staging status
	asi.autoStart.UpdateVehicleStaging(lane, preStageTriggered, stageTriggered, position)
}

// isBeamTriggered checks if a beam is currently triggered
func (asi *AutoStartIntegration) isBeamTriggered(beamID string) bool {
	// This would interface with the actual timing system beam status
	// For now, we'll simulate based on beam state tracking
	if beamState, exists := asi.beamStates[beamID]; exists {
		return beamState.IsTriggered
	}
	return false
}

// calculateVehiclePosition estimates vehicle position for rollout calculation
func (asi *AutoStartIntegration) calculateVehiclePosition(lane int, stageTriggered bool) float64 {
	if !stageTriggered {
		return 0.0
	}

	// This would calculate actual rollout distance based on beam geometry
	// For now, return a reasonable simulation
	guardBeamID := asi.guardBeams[lane]
	if guardBeamID != "" && asi.isBeamTriggered(guardBeamID) {
		return 15.0 // Simulate significant rollout if guard beam triggered
	}

	return 2.0 // Normal staging rollout
}

// triggerChristmasTree initiates the Christmas tree sequence
func (asi *AutoStartIntegration) triggerChristmasTree() error {
	if asi.christmasTree == nil {
		return fmt.Errorf("Christmas tree not available")
	}

	// Get auto-start configuration to determine tree type
	config := asi.autoStart.GetConfiguration()

	// Trigger appropriate tree sequence
	return asi.christmasTree.StartSequence(config.TreeSequenceType)
}

// handleAutoStartFault processes fault conditions
func (asi *AutoStartIntegration) handleAutoStartFault(reason string) {
	// Log the fault
	fmt.Printf("Auto-start fault: %s\n", reason)

	// Handle fault by resetting tree to safe state
	// The existing tree interface doesn't have red light methods,
	// so we'll handle this through state management
}

// handleStateChange processes auto-start state transitions
func (asi *AutoStartIntegration) handleStateChange(oldState, newState AutoStartState) {
	fmt.Printf("Auto-start state change: %s -> %s\n", oldState, newState)

	// The existing Christmas tree manages its own armed state
	// based on vehicle staging, so we don't need to call SetArmed
}

// Manual control methods

// GetAutoStartSystem returns the auto-start system for direct access
func (asi *AutoStartIntegration) GetAutoStartSystem() *AutoStartSystem {
	return asi.autoStart
}

// ManualTreeTrigger allows starter to manually trigger the tree
func (asi *AutoStartIntegration) ManualTreeTrigger() error {
	// Bypass auto-start and directly trigger tree
	return asi.triggerChristmasTree()
}

// SetAutoStartEnabled enables/disables auto-start functionality
func (asi *AutoStartIntegration) SetAutoStartEnabled(enabled bool) {
	asi.autoStart.SetEnabled(enabled)
}

// GetStatus returns comprehensive status information
func (asi *AutoStartIntegration) GetStatus() map[string]interface{} {
	autoStartStatus := asi.autoStart.GetAutoStartStatus()

	return map[string]interface{}{
		"auto_start": autoStartStatus,
		"running":    asi.running,
		"beams":      asi.beamStates,
		"beam_mapping": map[string]interface{}{
			"pre_stage": asi.preStageBeams,
			"stage":     asi.stageBeams,
			"guard":     asi.guardBeams,
		},
	}
}

// UpdateRacingClass adjusts auto-start parameters for different racing classes
func (asi *AutoStartIntegration) UpdateRacingClass(class string) {
	autoConfig := asi.autoStart.GetConfiguration()

	switch class {
	case "Top Fuel", "Funny Car", "Pro Stock":
		autoConfig.StagingTimeout = 7 * time.Second
		autoConfig.MinStagingDuration = 500 * time.Millisecond
		autoConfig.TreeSequenceType = config.TreeSequencePro
		autoConfig.RandomDelayMin = 600 * time.Millisecond
		autoConfig.RandomDelayMax = 1100 * time.Millisecond
	case "Pro Modified", "Pro Stock Motorcycle":
		autoConfig.StagingTimeout = 10 * time.Second
		autoConfig.MinStagingDuration = 500 * time.Millisecond
		autoConfig.TreeSequenceType = config.TreeSequencePro
	case "Bracket", "Super Class":
		autoConfig.StagingTimeout = 15 * time.Second
		autoConfig.MinStagingDuration = 600 * time.Millisecond
		autoConfig.TreeSequenceType = config.TreeSequenceSportsman
		autoConfig.RandomDelayMin = 600 * time.Millisecond
		autoConfig.RandomDelayMax = 1400 * time.Millisecond
	case "Junior Dragster":
		autoConfig.StagingTimeout = 15 * time.Second
		autoConfig.MinStagingDuration = 1000 * time.Millisecond
		autoConfig.TreeSequenceType = config.TreeSequenceSportsman
		autoConfig.EnabledForTimeTrials = true // More forgiving for learning
	}

	autoConfig.RacingClass = class
	asi.autoStart.UpdateConfiguration(autoConfig)
}

// SimulateBeamTrigger simulates a beam trigger for testing
func (asi *AutoStartIntegration) SimulateBeamTrigger(beamID string, triggered bool) {
	asi.mu.Lock()
	defer asi.mu.Unlock()

	if beamState, exists := asi.beamStates[beamID]; exists {
		beamState.IsTriggered = triggered
		beamState.LastChange = time.Now()
	}
}

// SetTestMode enables test mode for accelerated timing
func (asi *AutoStartIntegration) SetTestMode(enabled bool) {
	asi.autoStart.SetTestMode(enabled)
}
