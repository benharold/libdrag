package main

import (
	"fmt"
	"time"

	"github.com/benharold/libdrag/pkg/api"
)

// NewLibDrag creates a new libdrag instance (for mobile bindings)
func NewLibDrag() *api.LibDragAPI {
	return api.NewLibDragAPI()
}

func main() {
	fmt.Println("🏁 LIBDRAG - DRAG RACING LIBRARY")
	fmt.Println("=================================")

	// Create the libdrag API
	libdragAPI := api.NewLibDragAPI()

	// Initialize system
	fmt.Println("📊 Initializing libdrag system...")
	if err := libdragAPI.Initialize(); err != nil {
		fmt.Printf("❌ Failed to initialize libdrag: %v\n", err)
		return
	}

	fmt.Println("✅ libdrag system initialized successfully")

	// Start race
	fmt.Println("\n🚗 Starting race with libdrag...")
	raceID, err := libdragAPI.StartRaceWithID()
	if err != nil {
		fmt.Printf("❌ Failed to start race: %v\n", err)
		return
	}

	fmt.Printf("🏁 Race started with ID: %s\n", raceID)

	// Monitor race progress
	fmt.Println("🔄 Monitoring race progress...")

	// Wait for race to complete
	for i := 0; i < 100; i++ { // Max 10 seconds
		if libdragAPI.IsRaceCompleteByID(raceID) {
			break
		}

		// Show status updates
		if i%10 == 0 {
			status := libdragAPI.GetRaceStatusJSONByID(raceID)
			fmt.Printf("📊 Race Status: %s\n", status)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Display final results
	fmt.Println("\n🏆 LIBDRAG FINAL RESULTS")
	fmt.Println("========================")

	resultsJSON := libdragAPI.GetResultsJSONByID(raceID)
	fmt.Printf("Results JSON:\n%s\n", resultsJSON)

	treeStatusJSON := libdragAPI.GetTreeStatusJSONByID(raceID)
	fmt.Printf("\nTree Status JSON:\n%s\n", treeStatusJSON)

	// Clean shutdown
	fmt.Println("\n🔧 Shutting down libdrag system...")
	if err := libdragAPI.Stop(); err != nil {
		fmt.Printf("❌ Failed to shutdown cleanly: %v\n", err)
	} else {
		fmt.Println("✅ libdrag system shutdown complete")
	}
}
