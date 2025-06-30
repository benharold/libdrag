package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/benharold/libdrag/pkg/api"
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

	// Start race
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
