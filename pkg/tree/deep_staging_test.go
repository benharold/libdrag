package tree

import (
	"context"
	"testing"

	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/events"
)

// TestConfig for testing deep staging scenarios
type TestConfig struct {
	trackConfig  config.TrackConfig
	timingConfig config.TimingConfig
	treeConfig   config.TreeSequenceConfig
	safetyConfig config.SafetyConfig
	racingClass  string
}

func (tc *TestConfig) Track() config.TrackConfig                { return tc.trackConfig }
func (tc *TestConfig) Timing() config.TimingConfig             { return tc.timingConfig }
func (tc *TestConfig) Tree() config.TreeSequenceConfig         { return tc.treeConfig }
func (tc *TestConfig) Safety() config.SafetyConfig             { return tc.safetyConfig }
func (tc *TestConfig) RacingClass() string                     { return tc.racingClass }

func newTestConfig(class string) *TestConfig {
	return &TestConfig{
		trackConfig: config.TrackConfig{
			LaneCount: 2,
		},
		timingConfig: config.TimingConfig{},
		treeConfig:   config.TreeSequenceConfig{},
		safetyConfig: config.SafetyConfig{},
		racingClass:  class,
	}
}

// Test 1: Pre-stage and stage lights should follow beam states
func TestLightsFollowBeamStates(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)
	
	config := newTestConfig("Super Gas")
	err := tree.Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Initially, no lights should be on
	status := tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOff {
		t.Error("Pre-stage light should be off initially")
	}
	if status.LightStates[1][LightStage] != LightOff {
		t.Error("Stage light should be off initially")
	}

	// FAIL FIRST: This will fail because SetPreStage doesn't take a bool parameter yet
	tree.SetPreStage(1, true) // Beam is broken
	status = tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOn {
		t.Error("Pre-stage light should be on when beam is broken")
	}

	// Clear pre-stage beam (vehicle moves forward)
	tree.SetPreStage(1, false) // Beam is no longer broken
	status = tree.GetTreeStatus()
	if status.LightStates[1][LightPreStage] != LightOff {
		t.Error("Pre-stage light should be off when beam is not broken")
	}

	// Stage beam broken
	tree.SetStage(1, true)
	status = tree.GetTreeStatus()
	if status.LightStates[1][LightStage] != LightOn {
		t.Error("Stage light should be on when beam is broken")
	}
}

// Test 2: Deep staging detection - prohibited class
func TestDeepStagingDetection_ProhibitedClass(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)
	
	// Track violation events
	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeDeepStageViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Super Gas") // Deep staging prohibited
	err := tree.Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Simulate normal staging first
	tree.SetPreStage(1, true)  // Pre-stage beam broken
	tree.SetStage(1, true)     // Stage beam broken
	
	// Now simulate deep staging (vehicle moves past pre-stage)
	tree.SetPreStage(1, false) // Pre-stage beam no longer broken (deep staging!)

	// Should detect deep staging violation
	if len(violationEvents) != 1 {
		t.Errorf("Expected 1 deep staging violation event, got %d", len(violationEvents))
	}

	if len(violationEvents) > 0 {
		violation := violationEvents[0]
		if violation.Lane != 1 {
			t.Errorf("Expected violation for lane 1, got lane %d", violation.Lane)
		}
		if violation.Data["class"] != "Super Gas" {
			t.Errorf("Expected class 'Super Gas', got %v", violation.Data["class"])
		}
	}
}

// Test 3: Deep staging allowed - professional class
func TestDeepStagingDetection_AllowedClass(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	// Track deep staging events (informational)
	var deepStagingEvents []events.Event
	eventBus.Subscribe(events.EventTreeDeepStage, func(e events.Event) {
		deepStagingEvents = append(deepStagingEvents, e)
	})

	// Track violation events (should be none)
	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeDeepStageViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Top Fuel") // Deep staging allowed
	err := tree.Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to initialize tree: %v", err)
	}

	// Simulate deep staging sequence
	tree.SetPreStage(1, true)  // Pre-stage beam broken
	tree.SetStage(1, true)     // Stage beam broken
	tree.SetPreStage(1, false) // Deep staging (pre-stage clear, stage still on)

	// Should NOT create violation for Top Fuel
	if len(violationEvents) != 0 {
		t.Errorf("Expected 0 violation events for Top Fuel, got %d", len(violationEvents))
	}

	// Should create informational deep staging event
	if len(deepStagingEvents) != 1 {
		t.Errorf("Expected 1 deep staging event, got %d", len(deepStagingEvents))
	}

	if len(deepStagingEvents) > 0 {
		event := deepStagingEvents[0]
		if event.Lane != 1 {
			t.Errorf("Expected deep staging for lane 1, got lane %d", event.Lane)
		}
		if event.Data["deep_staged"] != true {
			t.Errorf("Expected deep_staged=true, got %v", event.Data["deep_staged"])
		}
	}
}

// Test 4: Class-specific prohibition rules
func TestDeepStagingProhibitionRules(t *testing.T) {
	tests := []struct {
		class      string
		prohibited bool
	}{
		{"Super Gas", true},
		{"Super Stock", true},
		{"Super Street", true},
		{"Top Fuel", false},
		{"Funny Car", false},
		{"Pro Stock", false},
		{"Bracket", false}, // Typically allowed in bracket racing
	}

	for _, test := range tests {
		t.Run(test.class, func(t *testing.T) {
			tree := NewChristmasTree()
			eventBus := events.NewEventBus(false)
			tree.SetEventBus(eventBus)

			// Track events
			var violationEvents []events.Event
			var deepStagingEvents []events.Event
			
			eventBus.Subscribe(events.EventTreeDeepStageViolation, func(e events.Event) {
				violationEvents = append(violationEvents, e)
			})
			eventBus.Subscribe(events.EventTreeDeepStage, func(e events.Event) {
				deepStagingEvents = append(deepStagingEvents, e)
			})

			config := newTestConfig(test.class)
			tree.Initialize(context.Background(), config)

			// Simulate deep staging
			tree.SetPreStage(1, true)
			tree.SetStage(1, true)
			tree.SetPreStage(1, false) // Deep stage

			if test.prohibited {
				if len(violationEvents) != 1 {
					t.Errorf("Class %s should generate violation, got %d events", test.class, len(violationEvents))
				}
				if len(deepStagingEvents) != 0 {
					t.Errorf("Class %s should not generate deep staging event, got %d", test.class, len(deepStagingEvents))
				}
			} else {
				if len(violationEvents) != 0 {
					t.Errorf("Class %s should not generate violation, got %d events", test.class, len(violationEvents))
				}
				if len(deepStagingEvents) != 1 {
					t.Errorf("Class %s should generate deep staging event, got %d", test.class, len(deepStagingEvents))
				}
			}
		})
	}
}

// Test 5: Normal staging should not trigger deep staging detection
func TestNormalStagingNoDeepStagingDetection(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var allEvents []events.Event
	eventBus.SubscribeAll(func(e events.Event) {
		allEvents = append(allEvents, e)
	})

	config := newTestConfig("Super Gas")
	tree.Initialize(context.Background(), config)

	// Normal staging sequence - should not trigger deep staging
	tree.SetPreStage(1, true)  // Pre-stage beam broken
	tree.SetStage(1, true)     // Stage beam broken
	// Both lights stay on - no deep staging

	// Filter for deep staging related events
	deepStagingEvents := 0
	for _, event := range allEvents {
		if event.Type == events.EventTreeDeepStage || event.Type == events.EventTreeDeepStageViolation {
			deepStagingEvents++
		}
	}

	if deepStagingEvents != 0 {
		t.Errorf("Normal staging should not trigger deep staging events, got %d", deepStagingEvents)
	}
}

// Test 6: Multiple lanes should be handled independently  
func TestMultipleLanesIndependent(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeDeepStageViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Super Gas")
	tree.Initialize(context.Background(), config)

	// Lane 1: Deep stage (should violate)
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)
	tree.SetPreStage(1, false) // Deep staging violation

	// Lane 2: Normal staging (should not violate)
	tree.SetPreStage(2, true)
	tree.SetStage(2, true)
	// No deep staging for lane 2

	// Should have exactly 1 violation (lane 1 only)
	if len(violationEvents) != 1 {
		t.Errorf("Expected 1 violation (lane 1), got %d", len(violationEvents))
	}

	if len(violationEvents) > 0 && violationEvents[0].Lane != 1 {
		t.Errorf("Expected violation for lane 1, got lane %d", violationEvents[0].Lane)
	}
}

// Test 7: Vehicle backing out of stage should clear deep staging
func TestVehicleBackingOut(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeDeepStageViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Super Gas")
	tree.Initialize(context.Background(), config)

	// Deep stage sequence
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)
	tree.SetPreStage(1, false) // Deep staging violation

	// Vehicle backs out completely
	tree.SetStage(1, false)     // Stage beam no longer broken
	tree.SetPreStage(1, false)  // Still not in pre-stage

	// Should have generated violation when deep staging occurred
	if len(violationEvents) != 1 {
		t.Errorf("Expected 1 violation when deep staging occurred, got %d", len(violationEvents))
	}

	// Vehicle re-enters normally
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)

	// Should not generate additional violations for normal staging
	if len(violationEvents) != 1 {
		t.Errorf("Normal re-staging should not generate additional violations, got %d", len(violationEvents))
	}
}

// Test 8: Forward motion rule - legal sequence
func TestForwardMotionRule_LegalSequence(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeStagingViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Top Fuel") // Use allowed class to focus on motion rule
	tree.Initialize(context.Background(), config)

	// Legal forward motion sequence: pre-stage -> stage -> deep stage
	tree.SetPreStage(1, true)  // Enter pre-stage
	tree.SetStage(1, true)     // Enter stage (forward motion)  
	tree.SetPreStage(1, false) // Deep stage (forward motion)

	// Should not generate any staging motion violations
	if len(violationEvents) != 0 {
		t.Errorf("Legal forward motion sequence should not generate violations, got %d", len(violationEvents))
	}
}

// Test 9: Forward motion rule - illegal backing and re-staging
func TestForwardMotionRule_IllegalBackingReStaging(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeStagingViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Top Fuel")
	tree.Initialize(context.Background(), config)

	// Start with legal forward staging
	tree.SetPreStage(1, true)  // Enter pre-stage
	tree.SetStage(1, true)     // Enter stage

	// Now illegal sequence: back out of stage, then re-stage
	tree.SetStage(1, false)    // Back out of stage (illegal backing motion)
	tree.SetStage(1, true)     // Re-enter stage (VIOLATION - last motion not forward)

	// Should generate staging motion violation
	if len(violationEvents) != 1 {
		t.Errorf("Expected 1 staging motion violation for illegal backing and re-staging, got %d", len(violationEvents))
	}

	if len(violationEvents) > 0 {
		violation := violationEvents[0]
		if violation.Lane != 1 {
			t.Errorf("Expected violation for lane 1, got lane %d", violation.Lane)
		}
		if violation.Data["violation_type"] != "backward_staging_motion" {
			t.Errorf("Expected backward_staging_motion violation, got %v", violation.Data["violation_type"])
		}
	}
}

// Test 10: Forward motion rule - multiple backing violations
func TestForwardMotionRule_MultipleBackingViolations(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeStagingViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Pro Stock")
	tree.Initialize(context.Background(), config)

	// Legal start
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)

	// First illegal backing sequence
	tree.SetStage(1, false)   // Back out
	tree.SetStage(1, true)    // Re-stage (violation #1)

	// Second illegal backing sequence  
	tree.SetStage(1, false)   // Back out again
	tree.SetStage(1, true)    // Re-stage again (violation #2)

	// Should generate multiple violations for repeated backing
	if len(violationEvents) != 2 {
		t.Errorf("Expected 2 staging motion violations for repeated backing, got %d", len(violationEvents))
	}
}

// Test 11: Forward motion rule - backing out completely then re-entering legally
func TestForwardMotionRule_CompleteBackoutThenLegalReentry(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeStagingViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Funny Car")
	tree.Initialize(context.Background(), config)

	// Initial staging
	tree.SetPreStage(1, true)
	tree.SetStage(1, true)

	// Complete back-out (both beams clear) - this should reset the staging sequence
	tree.SetStage(1, false)     // Back out of stage
	tree.SetPreStage(1, false)  // Back out of pre-stage completely

	// Legal re-entry from scratch (motion history should reset)
	tree.SetPreStage(1, true)   // Fresh pre-stage entry
	tree.SetStage(1, true)      // Fresh stage entry (should be legal)

	// Should not generate violations for fresh legal sequence after complete back-out
	if len(violationEvents) != 0 {
		t.Errorf("Complete back-out and fresh re-entry should not generate violations, got %d", len(violationEvents))
	}
}

// Test 12: Forward motion rule - pre-stage backing violations
func TestForwardMotionRule_PreStageBacking(t *testing.T) {
	tree := NewChristmasTree()
	eventBus := events.NewEventBus(false)
	tree.SetEventBus(eventBus)

	var violationEvents []events.Event
	eventBus.Subscribe(events.EventTreeStagingViolation, func(e events.Event) {
		violationEvents = append(violationEvents, e)
	})

	config := newTestConfig("Super Stock")
	tree.Initialize(context.Background(), config)

	// Enter pre-stage
	tree.SetPreStage(1, true)

	// Back out of pre-stage and re-enter
	tree.SetPreStage(1, false)  // Back out of pre-stage  
	tree.SetPreStage(1, true)   // Re-enter pre-stage (should be legal - no stage beam crossed yet)

	// Continue with legal forward motion
	tree.SetStage(1, true)      // Enter stage (legal forward motion)

	// Should not generate violations for pre-stage backing before reaching stage
	if len(violationEvents) != 0 {
		t.Errorf("Pre-stage backing before stage beam should not generate violations, got %d", len(violationEvents))
	}
}