package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/benharold/libdrag/pkg/api"
	"github.com/benharold/libdrag/pkg/events"
)

// NewLibDrag creates a new libdrag instance (for mobile bindings)
func NewLibDrag() *api.LibDragAPI {
	return api.NewLibDragAPI()
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("🏁 LIBDRAG - DRAG RACING LIBRARY")
	slog.Info("=================================")

	// Create the libdrag API
	libdragAPI := api.NewLibDragAPI()

	// Initialize system
	slog.Info("📊 Initializing libdrag system...")
	if err := libdragAPI.Initialize(); err != nil {
		slog.Error("❌ Failed to initialize libdrag", "error", err)
		os.Exit(1)
	}

	slog.Info("✅ libdrag system initialized successfully")

	// Subscribe to events for demonstration
	slog.Info("📡 Setting up event listeners...")

	// Subscribe to tree events
	libdragAPI.Subscribe(events.EventTreePreStage, func(e events.Event) {
		slog.Info("EVENT: Pre-stage", "lane", e.Lane, "race_id", e.RaceID)
	})

	libdragAPI.Subscribe(events.EventTreeStage, func(e events.Event) {
		slog.Info("EVENT: Stage", "lane", e.Lane, "race_id", e.RaceID)
	})

	libdragAPI.Subscribe(events.EventTreeArmed, func(e events.Event) {
		slog.Info("EVENT: Tree Armed!", "race_id", e.RaceID)
	})

	libdragAPI.Subscribe(events.EventTreeGreenOn, func(e events.Event) {
		slog.Info("EVENT: GREEN LIGHT!", "race_id", e.RaceID)
	})

	// Subscribe to timing events
	libdragAPI.Subscribe(events.EventTimingReaction, func(e events.Event) {
		if rt, ok := e.Data["reaction_time"].(float64); ok {
			slog.Info("EVENT: Reaction Time", "lane", e.Lane, "time", fmt.Sprintf("%.3fs", rt))
		}
	})

	libdragAPI.Subscribe(events.EventTiming60Foot, func(e events.Event) {
		if time, ok := e.Data["time"].(float64); ok {
			slog.Info("EVENT: 60-foot", "lane", e.Lane, "time", fmt.Sprintf("%.3fs", time))
		}
	})

	libdragAPI.Subscribe(events.EventTimingQuarterMile, func(e events.Event) {
		if time, ok := e.Data["time"].(float64); ok {
			if speed, ok := e.Data["trap_speed"].(float64); ok {
				slog.Info("EVENT: Quarter Mile", "lane", e.Lane, "time", fmt.Sprintf("%.3fs", time), "speed", fmt.Sprintf("%.1f mph", speed))
			}
		}
	})

	// Subscribe to race events
	libdragAPI.Subscribe(events.EventRaceStart, func(e events.Event) {
		slog.Info("EVENT: Race Started", "race_id", e.RaceID)
	})

	libdragAPI.Subscribe(events.EventRaceComplete, func(e events.Event) {
		slog.Info("EVENT: Race Complete", "race_id", e.RaceID)
	})

	// Subscribe to foul events
	libdragAPI.Subscribe(events.EventTreeRedLight, func(e events.Event) {
		if rt, ok := e.Data["reaction_time"].(float64); ok {
			slog.Warn("EVENT: RED LIGHT FOUL!", "lane", e.Lane, "reaction_time", fmt.Sprintf("%.3fs", rt))
		}
	})

	// Arm race
	slog.Info("🚗 Starting race with libdrag...")
	raceID, err := libdragAPI.StartRaceWithID()
	if err != nil {
		slog.Error("❌ Failed to start race", "error", err)
		os.Exit(1)
	}

	slog.Info("🏁 Race started", "race_id", raceID)

	// Monitor race progress
	slog.Info("🔄 Monitoring race progress...")

	// Wait for race to complete
	for i := 0; i < 100; i++ { // Max 10 seconds
		if libdragAPI.IsRaceCompleteByID(raceID) {
			break
		}

		// Show status updates
		if i%10 == 0 {
			status := libdragAPI.GetRaceStatusJSONByID(raceID)
			slog.Info("📊 Race Status", "status", status)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Display final results
	slog.Info("🏆 LIBDRAG FINAL RESULTS")
	slog.Info("========================")

	resultsJSON := libdragAPI.GetResultsJSONByID(raceID)
	slog.Info("Results JSON", "results", resultsJSON)

	treeStatusJSON := libdragAPI.GetTreeStatusJSONByID(raceID)
	slog.Info("Tree Status JSON", "tree_status", treeStatusJSON)

	// Clean shutdown
	slog.Info("🔧 Shutting down libdrag system...")
	if err := libdragAPI.Stop(); err != nil {
		slog.Error("❌ Failed to shutdown cleanly", "error", err)
	} else {
		slog.Info("✅ libdrag system shutdown complete")
	}
}
