package main

import (
	"fmt"
	"time"

	"github.com/benharold/libdrag/pkg/api"
)

func main() {
	fmt.Println("🏁 LIBDRAG - DRAG RACING LIBRARY DEMONSTRATION")
	fmt.Println("===============================================")

	// Create the libdrag API
	libdragAPI := api.NewLibDragAPI()

	// Initialize system
	fmt.Println("📊 Initializing libdrag system...")
	if err := libdragAPI.Initialize(); err != nil {
		fmt.Printf("❌ Failed to initialize libdrag: %v\n", err)
		return
	}

	fmt.Println("✅ libdrag system initialized successfully")

	// Arm race
	fmt.Println("\n🚗 Starting race with libdrag...")
	if err := libdragAPI.StartRace(); err != nil {
		fmt.Printf("❌ Failed to start race: %v\n", err)
		return
	}

	// Monitor race progress
	fmt.Println("🔄 Monitoring race progress...")

	// Wait for race to complete
	for i := 0; i < 100; i++ { // Max 10 seconds
		if libdragAPI.IsRaceComplete() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Display final results
	fmt.Println("\n🏆 LIBDRAG FINAL RESULTS")
	fmt.Println("========================")

	resultsJSON := libdragAPI.GetResultsJSON()
	fmt.Printf("Results JSON:\n%s\n", resultsJSON)

	treeStatusJSON := libdragAPI.GetTreeStatusJSON()
	fmt.Printf("\nChristmas Tree Status JSON:\n%s\n", treeStatusJSON)

	// Clean shutdown
	fmt.Println("🛑 Shutting down libdrag system...")
	libdragAPI.Stop()

	fmt.Println("✨ libdrag demo completed successfully!")
}
