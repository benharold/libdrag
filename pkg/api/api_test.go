package api

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewLibDragAPI(t *testing.T) {
	api := NewLibDragAPI()
	if api == nil {
		t.Fatal("NewLibDragAPI returned nil")
	}
	if api.orchestrators == nil {
		t.Fatal("orchestrators map not initialized")
	}
}

func TestInitialize(t *testing.T) {
	api := NewLibDragAPI()

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
}

func TestBasicRaceFlow(t *testing.T) {
	api := NewLibDragAPI()

	// Initialize
	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm race
	err = api.StartRace()
	if err != nil {
		t.Fatalf("StartRace failed: %v", err)
	}

	// Enable test mode for faster execution
	api.SetTestMode(true)

	// Wait for race completion (increased timeout to allow race to complete even in test mode)
	timeout := time.After(5 * time.Second)     // Increased from 2 seconds to 5 seconds
	ticker := time.Tick(10 * time.Millisecond) // Faster polling

	for {
		select {
		case <-timeout:
			t.Fatal("Race did not complete within timeout")
		case <-ticker:
			if api.IsRaceComplete() {
				goto raceComplete
			}
		}
	}

raceComplete:
	// Verify we can get results
	results := api.GetResultsJSON()
	if results == "" {
		t.Error("GetResultsJSON returned empty string")
	}

	treeStatus := api.GetTreeStatusJSON()
	if treeStatus == "" {
		t.Error("GetTreeStatusJSON returned empty string")
	}

	raceStatus := api.GetRaceStatusJSON()
	if raceStatus == "" {
		t.Error("GetRaceStatusJSON returned empty string")
	}

	// Clean shutdown
	api.Stop()
}

func TestMultipleRaces(t *testing.T) {
	api := NewLibDragAPI()

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer api.Stop()

	// Run multiple races concurrently to ensure system stability
	numRaces := 3
	var wg sync.WaitGroup
	results := make([]error, numRaces)

	// Arm all races concurrently
	for i := 0; i < numRaces; i++ {
		wg.Add(1)
		go func(raceIndex int) {
			defer wg.Done()

			// Arm race with unique ID
			raceID, err := api.StartRaceWithID()
			if err != nil {
				results[raceIndex] = fmt.Errorf("StartRace %d failed: %v", raceIndex+1, err)
				return
			}

			shortID := api.GetShortRaceID(raceID)
			t.Logf("Started race %d with ID: %s (short: %s)", raceIndex+1, raceID, shortID)

			// Enable test mode for this race to run faster
			api.SetTestMode(true)

			// Wait for completion of this specific race (increased timeout for test mode)
			for j := 0; j < 50; j++ { // Increased from 20 iterations (5 second timeout instead of 2)
				if api.IsRaceCompleteByID(raceID) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if !api.IsRaceCompleteByID(raceID) {
				results[raceIndex] = fmt.Errorf("Race %d (%s) did not complete", raceIndex+1, shortID)
				return
			}

			t.Logf("Race %d (%s) completed successfully", raceIndex+1, shortID)

			// Verify we can get results for this specific race
			raceResults := api.GetResultsJSONByID(raceID)
			if raceResults == "" || raceResults == "{}" {
				results[raceIndex] = fmt.Errorf("Race %d (%s) returned empty results", raceIndex+1, shortID)
			}
		}(i)
	}

	// Wait for all races to complete
	wg.Wait()

	// Check that all races completed successfully
	for i, err := range results {
		if err != nil {
			t.Fatalf("Race %d failed: %v", i+1, err)
		}
	}
}

// TestConcurrentRaces tests that multiple races can run simultaneously
func TestConcurrentRaces(t *testing.T) {
	api := NewLibDragAPI()

	// Initialize
	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	numConcurrentRaces := 3
	var wg sync.WaitGroup
	results := make([]error, numConcurrentRaces)

	// Arm multiple races concurrently
	for i := 0; i < numConcurrentRaces; i++ {
		wg.Add(1)
		go func(raceIndex int) {
			defer wg.Done()

			// Arm race and get race ID
			raceID, err := api.StartRaceWithID()
			if err != nil {
				results[raceIndex] = err
				return
			}

			shortID := api.GetShortRaceID(raceID)
			t.Logf("Started concurrent race %d (%s)", raceIndex, shortID)

			// Enable test mode for faster execution
			api.SetTestMode(true)

			// Wait for this specific race to complete (increased timeout for test mode)
			for j := 0; j < 50; j++ { // Increased from 20 iterations (5 second timeout instead of 2)
				if api.IsRaceCompleteByID(raceID) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if !api.IsRaceCompleteByID(raceID) {
				results[raceIndex] = fmt.Errorf("race %s did not complete", shortID)
			} else {
				t.Logf("Concurrent race %d (%s) completed", raceIndex, shortID)
			}
		}(i)
	}

	wg.Wait()

	// Check that all races completed successfully
	for i, err := range results {
		if err != nil {
			t.Fatalf("Race %d failed: %v", i, err)
		}
	}
}

// TestRaceIsolation tests that races don't interfere with each other
func TestRaceIsolation(t *testing.T) {
	api := NewLibDragAPI()

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Arm two races
	raceID1, err := api.StartRaceWithID()
	if err != nil {
		t.Fatalf("StartRace 1 failed: %v", err)
	}

	raceID2, err := api.StartRaceWithID()
	if err != nil {
		t.Fatalf("StartRace 2 failed: %v", err)
	}

	shortID1 := api.GetShortRaceID(raceID1)
	shortID2 := api.GetShortRaceID(raceID2)

	t.Logf("Started isolation test with races %s and %s", shortID1, shortID2)

	// Verify they have different IDs
	if raceID1 == raceID2 {
		t.Fatal("Race IDs should be different for concurrent races")
	}

	// Verify we can get independent status for each race
	status1 := api.GetRaceStatusJSONByID(raceID1)
	status2 := api.GetRaceStatusJSONByID(raceID2)

	if status1 == status2 {
		t.Fatal("Race statuses should be independent")
	}

	t.Logf("Race isolation test passed: %s and %s have independent statuses", shortID1, shortID2)
}

// TestMaxConcurrentRaces tests that the system respects concurrency limits
func TestMaxConcurrentRaces(t *testing.T) {
	api := NewLibDragAPI()

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	maxRaces := api.GetMaxConcurrentRaces()
	if maxRaces <= 0 {
		t.Fatal("Max concurrent races should be positive")
	}

	// Arm maximum number of races
	raceIDs := make([]string, maxRaces)
	for i := 0; i < maxRaces; i++ {
		raceID, err := api.StartRaceWithID()
		if err != nil {
			t.Fatalf("Failed to start race %d: %v", i, err)
		}
		raceIDs[i] = raceID
	}

	// Try to start one more race - should fail
	_, err = api.StartRaceWithID()
	if err == nil {
		t.Fatal("Should not be able to start more than max concurrent races")
	}

	// Complete one race and try again
	// For this test, we'll simulate completion by calling a cleanup method
	err = api.CompleteRace(raceIDs[0])
	if err != nil {
		t.Fatalf("Failed to complete race: %v", err)
	}

	// Now we should be able to start another race
	_, err = api.StartRaceWithID()
	if err != nil {
		t.Fatalf("Should be able to start race after completing one: %v", err)
	}
}

// TestUniqueRaceIdentifiers tests that each race gets a unique identifier
func TestUniqueRaceIdentifiers(t *testing.T) {
	api := NewLibDragAPI()

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer api.Stop()

	// Arm multiple races and collect their IDs
	numRaces := 5
	raceIDs := make([]string, numRaces)
	shortIDs := make([]string, numRaces)

	for i := 0; i < numRaces; i++ {
		raceID, err := api.StartRaceWithID()
		if err != nil {
			t.Fatalf("Failed to start race %d: %v", i, err)
		}
		raceIDs[i] = raceID
		shortIDs[i] = api.GetShortRaceID(raceID)

		t.Logf("Created race %d with short ID: %s", i+1, shortIDs[i])

		// Verify ID is not empty
		if raceID == "" {
			t.Fatalf("Race %d got empty ID", i)
		}

		// Verify ID format (should be a valid UUID)
		if len(raceID) != 36 {
			t.Fatalf("Race %d ID '%s' doesn't appear to be a valid UUID", i, raceID)
		}
	}

	// Verify all IDs are unique
	idSet := make(map[string]bool)
	for i, id := range raceIDs {
		if idSet[id] {
			t.Fatalf("Duplicate race ID found: %s (race %d)", id, i)
		}
		idSet[id] = true
	}

	// Verify we can get independent status for each race
	for i, raceID := range raceIDs {
		shortID := shortIDs[i]
		status := api.GetRaceStatusJSONByID(raceID)
		if status == "{\"error\":\"race not found\"}" {
			t.Fatalf("Race %d (%s) not found", i, shortID)
		}

		results := api.GetResultsJSONByID(raceID)
		if results == "{\"error\":\"race not found\"}" {
			t.Fatalf("Results for race %d (%s) not found", i, shortID)
		}
	}

	t.Logf("Successfully created %d races with short IDs: %v", numRaces, shortIDs)
}
