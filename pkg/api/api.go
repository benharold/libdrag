package api

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/benharold/libdrag/internal/vehicle"
	"github.com/benharold/libdrag/pkg/component"
	"github.com/benharold/libdrag/pkg/config"
	"github.com/benharold/libdrag/pkg/orchestrator"
	"github.com/benharold/libdrag/pkg/timing"
	"github.com/benharold/libdrag/pkg/tree"
	"github.com/google/uuid"
	"github.com/speps/go-hashids/v2"
)

// LibDragAPI provides a mobile-friendly interface
type LibDragAPI struct {
	orchestrators      map[string]*orchestrator.RaceOrchestrator
	mu                 sync.RWMutex
	maxConcurrentRaces int
	globalConfig       config.Config
	initialized        bool
}

func NewLibDragAPI() *LibDragAPI {
	return &LibDragAPI{
		orchestrators:      make(map[string]*orchestrator.RaceOrchestrator),
		maxConcurrentRaces: 10, // Default limit
	}
}

// Initialize the libdrag system
func (api *LibDragAPI) Initialize() error {
	api.mu.Lock()
	defer api.mu.Unlock()

	// Create global configuration
	api.globalConfig = config.NewDefaultConfig()
	api.initialized = true

	return nil
}

// StartRaceWithID starts a new drag race and returns a unique race ID
func (api *LibDragAPI) StartRaceWithID() (string, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if !api.initialized {
		return "", fmt.Errorf("API not initialized")
	}

	// Check concurrent race limit
	if len(api.orchestrators) >= api.maxConcurrentRaces {
		return "", fmt.Errorf("maximum concurrent races (%d) reached", api.maxConcurrentRaces)
	}

	// Generate unique race ID
	raceID := uuid.New().String()

	// Create new orchestrator for this race
	raceOrchestrator := orchestrator.NewRaceOrchestrator()

	// Create components for this race with race ID context
	timingSystem := timing.NewTimingSystemWithRaceID(raceID)
	christmasTree := tree.NewChristmasTree()

	components := []component.Component{
		timingSystem,
		christmasTree,
	}

	// Initialize the race orchestrator
	ctx := context.Background()
	if err := raceOrchestrator.Initialize(ctx, components, api.globalConfig); err != nil {
		return "", fmt.Errorf("failed to initialize race orchestrator: %v", err)
	}

	// Store the orchestrator
	api.orchestrators[raceID] = raceOrchestrator

	// Start the race
	leftVehicle := vehicle.NewSimpleVehicle(1)
	rightVehicle := vehicle.NewSimpleVehicle(2)

	if err := raceOrchestrator.StartRace(leftVehicle, rightVehicle); err != nil {
		// Clean up on failure
		delete(api.orchestrators, raceID)
		return "", err
	}

	// Start goroutine to clean up completed races
	go api.monitorRaceCompletion(raceID)

	return raceID, nil
}

// monitorRaceCompletion monitors a race and cleans up when complete
func (api *LibDragAPI) monitorRaceCompletion(raceID string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second) // Maximum race duration

	for {
		select {
		case <-timeout:
			// Race timed out, force cleanup
			api.CompleteRace(raceID)
			return
		case <-ticker.C:
			if api.IsRaceCompleteByID(raceID) {
				// Wait a bit longer to allow final status updates
				time.Sleep(1 * time.Second)
				return // Race completed naturally
			}
		}
	}
}

// GetRaceStatusJSON returns race status as JSON (legacy method)
// GetRaceStatusJSONByID returns race status as JSON for a specific race
func (api *LibDragAPI) GetRaceStatusJSONByID(raceID string) string {
	api.mu.RLock()
	defer api.mu.RUnlock()

	orchestrator, exists := api.orchestrators[raceID]
	if !exists {
		return "{\"error\":\"race not found\"}"
	}

	status := orchestrator.GetRaceStatus()
	jsonData, _ := json.Marshal(status)
	return string(jsonData)
}

// GetTreeStatusJSON returns christmas tree status as JSON (legacy method)
// GetTreeStatusJSONByID returns christmas tree status as JSON for a specific race
func (api *LibDragAPI) GetTreeStatusJSONByID(raceID string) string {
	api.mu.RLock()
	defer api.mu.RUnlock()

	orchestrator, exists := api.orchestrators[raceID]
	if !exists {
		return "{\"error\":\"race not found\"}"
	}

	status := orchestrator.GetTreeStatus()
	jsonData, _ := json.Marshal(status)
	return string(jsonData)
}

// GetResultsJSON returns race results as JSON (legacy method)
// GetResultsJSONByID returns race results as JSON for a specific race
func (api *LibDragAPI) GetResultsJSONByID(raceID string) string {
	api.mu.RLock()
	defer api.mu.RUnlock()

	orchestrator, exists := api.orchestrators[raceID]
	if !exists {
		return "{\"error\":\"race not found\"}"
	}

	results := orchestrator.GetResults()
	jsonData, _ := json.Marshal(results)
	return string(jsonData)
}

// IsRaceComplete checks if the current race is finished (legacy method)
// IsRaceCompleteByID checks if a specific race is finished
func (api *LibDragAPI) IsRaceCompleteByID(raceID string) bool {
	api.mu.RLock()
	defer api.mu.RUnlock()

	orch, exists := api.orchestrators[raceID]
	if !exists {
		return true // Race doesn't exist, consider it complete
	}

	status := orch.GetRaceStatus()
	return status.State == orchestrator.RaceStateComplete
}

// CompleteRace manually marks a race as complete and cleans up resources
func (api *LibDragAPI) CompleteRace(raceID string) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	_, exists := api.orchestrators[raceID]
	if !exists {
		return fmt.Errorf("race %s not found", raceID)
	}

	// Stop the orchestrator if it has a Stop method
	// Note: This assumes the orchestrator has cleanup methods
	// You may need to implement these in the orchestrator package

	// Remove from active races
	delete(api.orchestrators, raceID)
	return nil
}

// GetMaxConcurrentRaces returns the maximum number of concurrent races allowed
func (api *LibDragAPI) GetMaxConcurrentRaces() int {
	api.mu.RLock()
	defer api.mu.RUnlock()
	return api.maxConcurrentRaces
}

// SetMaxConcurrentRaces sets the maximum number of concurrent races allowed
func (api *LibDragAPI) SetMaxConcurrentRaces(max int) {
	api.mu.Lock()
	defer api.mu.Unlock()
	if max > 0 {
		api.maxConcurrentRaces = max
	}
}

// Stop shuts down the API and cleans up all active races
func (api *LibDragAPI) Stop() error {
	api.mu.Lock()
	defer api.mu.Unlock()

	// Stop all active races
	for raceID := range api.orchestrators {
		delete(api.orchestrators, raceID)
	}

	api.initialized = false
	return nil
}

// Reset clears all active races but keeps the API initialized
func (api *LibDragAPI) Reset() error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if !api.initialized {
		return fmt.Errorf("API not initialized")
	}

	// Clear all active races
	for raceID := range api.orchestrators {
		delete(api.orchestrators, raceID)
	}

	return nil
}

// GetActiveRaceCount returns the number of currently active races
func (api *LibDragAPI) GetActiveRaceCount() int {
	api.mu.RLock()
	defer api.mu.RUnlock()
	return len(api.orchestrators)
}

// GetActiveRaceIDs returns a list of all currently active race IDs
func (api *LibDragAPI) GetActiveRaceIDs() []string {
	api.mu.RLock()
	defer api.mu.RUnlock()

	raceIDs := make([]string, 0, len(api.orchestrators))
	for raceID := range api.orchestrators {
		raceIDs = append(raceIDs, raceID)
	}
	return raceIDs
}

// RaceExists checks if a race with the given ID exists
func (api *LibDragAPI) RaceExists(raceID string) bool {
	api.mu.RLock()
	defer api.mu.RUnlock()
	_, exists := api.orchestrators[raceID]
	return exists
}

// GetAllRaceStatuses returns status information for all active races
func (api *LibDragAPI) GetAllRaceStatuses() map[string]string {
	api.mu.RLock()
	defer api.mu.RUnlock()

	statuses := make(map[string]string)
	for raceID, orchestrator := range api.orchestrators {
		status := orchestrator.GetRaceStatus()
		jsonData, _ := json.Marshal(status)
		statuses[raceID] = string(jsonData)
	}
	return statuses
}

// GetShortRaceID returns a short identifier for logging purposes
func (api *LibDragAPI) GetShortRaceID(raceID string) string {
	shortID, err := encodeRaceID(raceID)
	if err != nil {
		// Fallback to first 8 characters of UUID if encoding fails
		if len(raceID) >= 8 {
			return raceID[:8]
		}
		return raceID
	}
	return shortID
}

// SetTestMode enables fast mode for all timing systems (for testing)
func (api *LibDragAPI) SetTestMode(enabled bool) {
	api.mu.Lock()
	defer api.mu.Unlock()

	for _, orchestrator := range api.orchestrators {
		// Get the timing system from the orchestrator and enable test mode
		if timingSystem := orchestrator.GetTimingSystem(); timingSystem != nil {
			timingSystem.SetTestMode(enabled)
		}
	}
}

// Version returns the libdrag version
func Version() string {
	return "libdrag v1.0.0 - Professional Drag Racing Library"
}

// hashids encoder
var hd *hashids.HashID

// Initialize the hashids package
func init() {
	// Use a static salt and min length for consistent short IDs
	hdata := hashids.NewData()
	hdata.Salt = "libdrag"
	hdata.MinLength = 6

	var err error
	hd, err = hashids.NewWithData(hdata)
	if err != nil {
		// Fallback to default if initialization fails
		hd, _ = hashids.New()
	}
}

// Encode a race ID to a short string
func encodeRaceID(raceID string) (string, error) {
	// Hash the race ID to a fixed size
	hash := md5.Sum([]byte(raceID))
	id := binary.BigEndian.Uint64(hash[:8])

	// Encode the hashed ID to a short string
	encoded, err := hd.Encode([]int{int(id)})
	if err != nil {
		return "", err
	}

	return encoded, nil
}

// Decode a short string to a race ID
func decodeRaceID(encodedID string) (string, error) {
	// Decode the short string to an int
	decoded := hd.Decode(encodedID)
	if len(decoded) == 0 {
		return "", fmt.Errorf("failed to decode ID")
	}

	// Convert the decoded int back to a race ID
	raceID := fmt.Sprintf("%x", decoded[0])
	return raceID, nil
}
