package autostart

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/tree"
)

// AutoStartState represents the current state of the auto-start system
type AutoStartState string

const (
	StateIdle       AutoStartState = "idle"       // Not monitoring, waiting for tree to be armed
	StateMonitoring AutoStartState = "monitoring" // Tree armed, monitoring for three beam rule
	StateActivated  AutoStartState = "activated"  // Three beams detected, countdown started
	StateStaging    AutoStartState = "staging"    // Both vehicles staged, final checks
	StateTriggered  AutoStartState = "triggered"  // Tree sequence initiated
	StateFault      AutoStartState = "fault"      // Safety violation or timeout
)

// StagingStatus represents vehicle staging state
type StagingStatus struct {
	Lane       int       `json:"lane"`
	PreStaged  bool      `json:"pre_staged"`
	Staged     bool      `json:"staged"`
	LastUpdate time.Time `json:"last_update"`
	GuardTrip  bool      `json:"guard_trip"` // Guard beam violation
	Rollout    float64   `json:"rollout"`    // Distance past stage beam
}

// AutoStartConfig holds configuration for the auto-start system
type AutoStartConfig struct {
	// Core timing parameters
	StagingTimeout     time.Duration `json:"staging_timeout"`      // Total time allowed for staging (7-20 seconds)
	MinStagingDuration time.Duration `json:"min_staging_duration"` // Minimum time both cars must be staged (0.5-1.0 seconds)
	RandomDelayMin     time.Duration `json:"random_delay_min"`     // Minimum random delay (0.6 seconds)
	RandomDelayMax     time.Duration `json:"random_delay_max"`     // Maximum random delay (1.4 seconds)
	RandomVariation    time.Duration `json:"random_variation"`     // Additional random variation (0.2 seconds)

	// Safety parameters
	GuardBeamDistance  float64 `json:"guard_beam_distance"`  // Distance to guard beam (13.375 inches)
	MaxRolloutDistance float64 `json:"max_rollout_distance"` // Maximum allowed rollout
	PreStageDistance   float64 `json:"pre_stage_distance"`   // Distance from start line (-7 inches)

	// Operational modes
	EnabledForElims      bool                    `json:"enabled_for_elims"`      // Auto-start for eliminations
	EnabledForTimeTrials bool                    `json:"enabled_for_timetrials"` // Auto-start for time trials
	TreeSequenceType     config.TreeSequenceType `json:"tree_sequence_type"`     // Pro or Sportsman tree

	// IHRA/NHRA class-specific settings
	RacingClass string `json:"racing_class"` // e.g., "Top Fuel", "Pro Stock", "Bracket"
}

// AutoStartStatus represents the current system status
type AutoStartStatus struct {
	State              AutoStartState         `json:"state"`
	IsEnabled          bool                   `json:"is_enabled"`
	VehicleStaging     map[int]*StagingStatus `json:"vehicle_staging"`
	CountdownStarted   time.Time              `json:"countdown_started,omitempty"`
	CountdownRemaining time.Duration          `json:"countdown_remaining"`
	BothVehiclesStaged time.Time              `json:"both_vehicles_staged,omitempty"`
	TreeTriggerTime    time.Time              `json:"tree_trigger_time,omitempty"`
	LastFaultReason    string                 `json:"last_fault_reason,omitempty"`
	OverrideActive     bool                   `json:"override_active"`
	StarterControl     bool                   `json:"starter_control"`
}

// AutoStartSystem implements the CompuLink-style auto-start functionality
type AutoStartSystem struct {
	id         string
	config     AutoStartConfig
	mu         sync.RWMutex
	status     AutoStartStatus
	compStatus component.ComponentStatus
	running    bool
	testMode   bool

	// Component integration
	tree *tree.ChristmasTree // Reference to tree component for automatic arming

	// Event handlers
	onTreeTrigger func() error
	onFault       func(reason string)
	onStateChange func(oldState, newState AutoStartState)

	// Internal timing
	countdownTimer *time.Timer
	stagingTimer   *time.Timer
	randomSeed     *rand.Rand
}

// NewAutoStartSystem creates a new auto-start system
func NewAutoStartSystem() *AutoStartSystem {
	return &AutoStartSystem{
		id:         "autostart_system",
		randomSeed: rand.New(rand.NewSource(time.Now().UnixNano())),
		status: AutoStartStatus{
			State:          StateIdle,
			IsEnabled:      true,
			VehicleStaging: make(map[int]*StagingStatus),
			StarterControl: true, // Default to starter having control
		},
		compStatus: component.ComponentStatus{
			ID:       "autostart_system",
			Status:   "ready",
			Metadata: make(map[string]interface{}),
		},
	}
}

// GetID returns the component ID
func (as *AutoStartSystem) GetID() string {
	return as.id
}

// Initialize initializes the auto-start system with configuration
func (as *AutoStartSystem) Initialize(ctx context.Context, cfg config.Config) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Load configuration and set defaults based on CompuLink specifications
	as.config = as.loadConfigFromSystem(cfg)

	// Initialize vehicle staging status for configured lanes
	trackConfig := cfg.Track()
	for i := 1; i <= trackConfig.LaneCount; i++ {
		as.status.VehicleStaging[i] = &StagingStatus{
			Lane: i,
		}
	}

	as.compStatus.Status = "initialized"
	return nil
}

// loadConfigFromSystem extracts auto-start configuration from system config
func (as *AutoStartSystem) loadConfigFromSystem(cfg config.Config) AutoStartConfig {
	treeConfig := cfg.Tree()

	// Default to professional settings (Top Fuel/Funny Car/Pro Stock)
	autoConfig := AutoStartConfig{
		StagingTimeout:       7 * time.Second,
		MinStagingDuration:   500 * time.Millisecond,
		RandomDelayMin:       600 * time.Millisecond,
		RandomDelayMax:       1100 * time.Millisecond, // Pro tree range
		RandomVariation:      200 * time.Millisecond,
		GuardBeamDistance:    13.375, // 13 3/8 inches
		MaxRolloutDistance:   6.0,    // Reasonable rollout limit
		PreStageDistance:     -7.0,   // 7 inches behind start line
		EnabledForElims:      true,
		EnabledForTimeTrials: false,
		TreeSequenceType:     treeConfig.Type,
		RacingClass:          "Professional",
	}

	// Adjust for sportsman classes if configured
	if treeConfig.Type == config.TreeSequenceSportsman {
		autoConfig.StagingTimeout = 10 * time.Second
		autoConfig.MinStagingDuration = 600 * time.Millisecond
		autoConfig.RandomDelayMin = 600 * time.Millisecond
		autoConfig.RandomDelayMax = 1400 * time.Millisecond // Sportsman range
		autoConfig.RacingClass = "Sportsman"
	}

	return autoConfig
}

// Start starts the auto-start system
func (as *AutoStartSystem) Start(ctx context.Context) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.running {
		return fmt.Errorf("auto-start system already running")
	}

	as.running = true
	as.compStatus.Status = "running"
	as.status.State = StateIdle

	return nil
}

// Stop stops the auto-start system
func (as *AutoStartSystem) Stop(ctx context.Context) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.running = false
	as.compStatus.Status = "stopped"
	as.status.State = StateIdle

	// Cancel any active timers
	if as.countdownTimer != nil {
		as.countdownTimer.Stop()
		as.countdownTimer = nil
	}
	if as.stagingTimer != nil {
		as.stagingTimer.Stop()
		as.stagingTimer = nil
	}

	return nil
}

// GetStatus returns the current component status
func (as *AutoStartSystem) GetStatus() component.ComponentStatus {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.compStatus
}

// GetAutoStartStatus returns detailed auto-start status
func (as *AutoStartSystem) GetAutoStartStatus() AutoStartStatus {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.status
}

// SetEnabled enables or disables the auto-start system
func (as *AutoStartSystem) SetEnabled(enabled bool) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.status.IsEnabled = enabled
	if !enabled && as.status.State != StateIdle {
		as.resetToIdle("Manual disable")
	}
}

// UpdateVehicleStaging updates staging status for a vehicle (called by beam triggers)
func (as *AutoStartSystem) UpdateVehicleStaging(lane int, preStaged, staged bool, position float64) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if !as.running || !as.status.IsEnabled {
		return nil
	}

	stagingStatus, exists := as.status.VehicleStaging[lane]
	if !exists {
		return fmt.Errorf("invalid lane: %d", lane)
	}

	// Update staging status
	oldPreStaged := stagingStatus.PreStaged
	oldStaged := stagingStatus.Staged

	stagingStatus.PreStaged = preStaged
	stagingStatus.Staged = staged
	stagingStatus.LastUpdate = time.Now()
	stagingStatus.Rollout = position // Track rollout distance

	// Check for guard beam violation (excessive rollout)
	if position > as.config.MaxRolloutDistance {
		stagingStatus.GuardTrip = true
		as.triggerFault(fmt.Sprintf("Lane %d guard beam violation: rollout %.2f inches", lane, position))
		return nil
	}

	// Check courtesy staging violation (staged without both pre-staged)
	preCount := as.countPreStaged()
	if staged && preCount < 2 {
		// Courtesy violation: Staged without both pre-staged
		// Could fault or just log/warn per regs (encouraged, not enforced)
		fmt.Println("Courtesy staging violation: Staged before both pre-staged")
		// Optional: if config.CourtesyEnforced { as.triggerFault("Courtesy staging violation") }
	}

	// Check if this triggers the three-light rule
	if as.autoStartShouldActivateCountdownSequence(oldPreStaged, oldStaged, preStaged, staged) {
		as.triggerAutoStart()
	}

	// If activated and this update caused countStaged to become 1, start timeout
	if as.status.State == StateActivated && as.countStaged() == 1 && !oldStaged && staged {
		as.startSecondStageTimeout()
	}

	return nil
}

// autoStartShouldActivateCountdownSequence implements the "three-light rule"
// Only triggers when tree is already armed
func (as *AutoStartSystem) autoStartShouldActivateCountdownSequence(oldPreStaged, oldStaged, newPreStaged, newStaged bool) bool {
	// Auto-start can only activate the countdown sequence if tree is armed
	if as.tree == nil || !as.tree.IsArmed() {
		return false
	}

	if as.status.State != StateIdle {
		return false
	}

	// Activate on >=3 total bulbs lit ⚡
	return as.totalBulbsOn() >= 3
}

// triggerAutoStart activates the auto-start countdown sequence (tree must already be armed)
func (as *AutoStartSystem) triggerAutoStart() {
	oldState := as.status.State
	as.status.State = StateActivated
	as.status.CountdownStarted = time.Now()

	// Activate the auto-start system on the tree (tree must already be armed)
	if as.tree != nil {
		err := as.tree.ActivateAutoStart()
		if err != nil {
			as.triggerFault(fmt.Sprintf("Cannot activate auto-start: %v", err))
			return
		}
	}

	if as.onStateChange != nil {
		go as.onStateChange(oldState, StateActivated)
	}

	// Monitor for both vehicles fully staged (and start timeout if already one staged)
	go as.monitorForFullStaging()
	if as.countStaged() == 1 { // If activation happened on first stage
		as.startSecondStageTimeout()
	}

	// Monitor for both vehicles fully staged
	go as.monitorForFullStaging()
}

// monitorForFullStaging watches for both vehicles to be fully staged
func (as *AutoStartSystem) monitorForFullStaging() {
	ticker := time.NewTicker(5 * time.Millisecond) // Very frequent checking for test reliability
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			as.mu.Lock()
			if as.status.State != StateActivated {
				as.mu.Unlock()
				return
			}

			// Check if both vehicles are fully staged
			stagedCount := 0
			for _, staging := range as.status.VehicleStaging {
				if staging.Staged {
					stagedCount++
				}
			}

			// Only transition to staging if we have exactly 2 staged vehicles and haven't transitioned yet
			if stagedCount == 2 && as.status.BothVehiclesStaged.IsZero() {
				as.status.BothVehiclesStaged = time.Now()
				as.status.State = StateStaging

				// Cancel staging timeout since both are now staged
				if as.stagingTimer != nil {
					as.stagingTimer.Stop()
					as.stagingTimer = nil
				}

				// Arm minimum staging timer
				as.stagingTimer = time.AfterFunc(as.config.MinStagingDuration, func() {
					as.mu.Lock()
					defer as.mu.Unlock()
					if as.status.State == StateStaging {
						as.triggerTreeSequence()
					}
				})
			}
			as.mu.Unlock()
		}
	}
}

// triggerTreeSequence initiates the Christmas tree sequence with random delay
func (as *AutoStartSystem) triggerTreeSequence() {
	// In test mode, use minimal delay to ensure reliable testing
	var randomDelay time.Duration
	if as.testMode {
		randomDelay = 1 * time.Millisecond // Very short delay for tests
	} else {
		randomDelay = as.calculateRandomDelay()
	}

	// Schedule tree trigger
	time.AfterFunc(randomDelay, func() {
		as.mu.Lock()
		defer as.mu.Unlock()

		if as.status.State == StateStaging {
			as.status.State = StateTriggered
			as.status.TreeTriggerTime = time.Now()

			// Trigger the tree sequence immediately (don't use goroutine for test reliability)
			if as.onTreeTrigger != nil {
				err := as.onTreeTrigger()
				if err != nil {
					as.triggerFault(fmt.Sprintf("Tree trigger error: %v", err))
					return
				}
			}

			// Reset to idle after successful trigger
			time.AfterFunc(100*time.Millisecond, func() { // Shorter delay for tests
				as.mu.Lock()
				defer as.mu.Unlock()
				as.resetToIdle("Race completed")
			})
		}
	})
}

// calculateRandomDelay implements CompuLink's random delay algorithm
func (as *AutoStartSystem) calculateRandomDelay() time.Duration {
	// Base random delay between min and max
	baseRange := as.config.RandomDelayMax - as.config.RandomDelayMin
	baseDelay := as.config.RandomDelayMin + time.Duration(as.randomSeed.Float64()*float64(baseRange))

	// Add additional random variation
	variation := time.Duration(as.randomSeed.Float64() * float64(as.config.RandomVariation))

	return baseDelay + variation
}

// triggerFault handles safety violations and system faults
func (as *AutoStartSystem) triggerFault(reason string) {
	oldState := as.status.State
	as.status.State = StateFault
	as.status.LastFaultReason = reason

	// Cancel timer
	if as.stagingTimer != nil {
		as.stagingTimer.Stop()
		as.stagingTimer = nil
	}

	if as.onFault != nil {
		go as.onFault(reason)
	}

	if as.onStateChange != nil {
		go as.onStateChange(oldState, StateFault)
	}
}

// resetToIdle resets the system to idle state
func (as *AutoStartSystem) resetToIdle(reason string) {
	oldState := as.status.State
	as.status.State = StateIdle
	as.status.CountdownStarted = time.Time{}
	as.status.BothVehiclesStaged = time.Time{}
	as.status.TreeTriggerTime = time.Time{}
	as.status.CountdownRemaining = 0

	// Reset vehicle staging status
	for _, staging := range as.status.VehicleStaging {
		staging.PreStaged = false
		staging.Staged = false
		staging.GuardTrip = false
		staging.Rollout = 0
	}

	// Cancel timer
	if as.stagingTimer != nil {
		as.stagingTimer.Stop()
		as.stagingTimer = nil
	}

	if as.onStateChange != nil {
		go as.onStateChange(oldState, StateIdle)
	}
}

// Manual override and control methods

// ManualOverride disables auto-start and gives full control to starter
func (as *AutoStartSystem) ManualOverride() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.status.OverrideActive = true
	as.status.IsEnabled = false
	as.resetToIdle("Manual override activated")
}

// ClearOverride re-enables auto-start system
func (as *AutoStartSystem) ClearOverride() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.status.OverrideActive = false
	as.status.IsEnabled = true
}

// UpdateConfiguration allows real-time parameter adjustments
func (as *AutoStartSystem) UpdateConfiguration(newConfig AutoStartConfig) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.config = newConfig
}

// GetConfiguration returns current configuration
func (as *AutoStartSystem) GetConfiguration() AutoStartConfig {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.config
}

// Event handler setters

// SetTreeTriggerHandler sets the callback for when tree should be triggered
func (as *AutoStartSystem) SetTreeTriggerHandler(handler func() error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.onTreeTrigger = handler
}

// SetFaultHandler sets the callback for when faults occur
func (as *AutoStartSystem) SetFaultHandler(handler func(reason string)) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.onFault = handler
}

// SetStateChangeHandler sets the callback for state changes
func (as *AutoStartSystem) SetStateChangeHandler(handler func(oldState, newState AutoStartState)) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.onStateChange = handler
}

// SetTestMode enables fast execution for testing
func (as *AutoStartSystem) SetTestMode(enabled bool) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.testMode = enabled

	if enabled {
		// Accelerate timing for testing - make timeout much shorter
		as.config.StagingTimeout = 50 * time.Millisecond // Very short timeout for reliable testing
		as.config.MinStagingDuration = 5 * time.Millisecond
		as.config.RandomDelayMin = 1 * time.Millisecond
		as.config.RandomDelayMax = 3 * time.Millisecond
	}
}

// Component integration methods

// SetTreeComponent sets the tree component reference for automatic arming
func (as *AutoStartSystem) SetTreeComponent(treeComponent *tree.ChristmasTree) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.tree = treeComponent
}

// countPreStaged returns the number of pre-staged vehicles.
func (as *AutoStartSystem) countPreStaged() int {
	count := 0
	for _, staging := range as.status.VehicleStaging {
		if staging.PreStaged {
			count++
		}
	}
	return count
}

// countStaged returns the number of staged vehicles.
func (as *AutoStartSystem) countStaged() int {
	count := 0
	for _, staging := range as.status.VehicleStaging {
		if staging.Staged {
			count++
		}
	}
	return count
}

// totalBulbsOn returns the total number of lit top bulbs (pre + stage across lanes).
func (as *AutoStartSystem) totalBulbsOn() int {
	count := 0
	for _, staging := range as.status.VehicleStaging {
		if staging.PreStaged {
			count++
		}
		if staging.Staged {
			count++
		}
	}
	return count
}

// startSecondStageTimeout starts the timeout for the second vehicle to stage.
func (as *AutoStartSystem) startSecondStageTimeout() {
	as.stagingTimer = time.AfterFunc(as.config.StagingTimeout, func() {
		as.mu.Lock()
		defer as.mu.Unlock()
		if as.status.State != StateActivated { // Only fault if still waiting
			return
		}
		// Find unstaged lane
		var timedOutLane int
		for lane, staging := range as.status.VehicleStaging {
			if !staging.Staged {
				timedOutLane = lane
				break
			}
		}
		as.triggerFault(fmt.Sprintf("Staging timeout for lane %d", timedOutLane))
		// Publish foul event for timing system (red light)
		// if as.eventBus != nil { // Assume event bus added
		//     as.eventBus.Publish(events.NewEvent(events.EventStagingTimeoutFoul).WithLane(timedOutLane).Build())
		// }
	})
}
