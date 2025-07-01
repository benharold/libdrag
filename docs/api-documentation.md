# libdrag API Documentation

## Overview

The libdrag API provides a comprehensive drag racing timing system with support for NHRA/IHRA-style racing events, Christmas tree light sequences, precision timing, and concurrent race management.

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    // Create and initialize the API
    dragAPI := api.NewLibDragAPI()
    err := dragAPI.Initialize()
    if err != nil {
        panic(err)
    }
    defer dragAPI.Stop()

    // Start a new race
    raceID, err := dragAPI.StartRaceWithID()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Started race: %s\n", raceID)

    // Monitor race progress
    for !dragAPI.IsRaceCompleteByID(raceID) {
        status := dragAPI.GetRaceStatusJSONByID(raceID)
        fmt.Printf("Race Status: %s\n", status)
        time.Sleep(500 * time.Millisecond)
    }

    // Get final results
    results := dragAPI.GetResultsJSONByID(raceID)
    fmt.Printf("Final Results: %s\n", results)
}
```

## Core API Methods

### Initialization

#### `NewLibDragAPI() *LibDragAPI`
Creates a new instance of the libdrag API.

**Returns:**
- `*LibDragAPI`: New API instance

#### `Initialize() error`
Initializes the libdrag system with default configuration.

**Returns:**
- `error`: Error if initialization fails

### Race Management

#### `StartRaceWithID() (string, error)`
Starts a new drag race and returns a unique race identifier.

**Returns:**
- `string`: Unique race ID (UUID format)
- `error`: Error if race cannot be started

#### `StartRace() error` *(Legacy)*
Starts a new drag race without returning the race ID. Provided for backward compatibility.

**Returns:**
- `error`: Error if race cannot be started

#### `CompleteRace(raceID string) error`
Manually completes a race and cleans up resources.

**Parameters:**
- `raceID`: The unique identifier of the race to complete

**Returns:**
- `error`: Error if race doesn't exist or cleanup fails

### Race Status

#### `GetRaceStatusJSONByID(raceID string) string`
Returns the current race status as JSON for a specific race.

**Parameters:**
- `raceID`: The unique identifier of the race

**Returns:**
- `string`: JSON-formatted race status

**Example Response:**
```json
{
  "race_id": "123e4567-e89b-12d3-a456-426614174000",
  "state": "running",
  "start_time": "2025-06-30T21:54:51.939Z",
  "elapsed_time": 5.5,
  "lane_1": {
    "vehicle_id": "vehicle_1",
    "staged": true,
    "reaction_time": 0.045,
    "current_position": 800.5
  },
  "lane_2": {
    "vehicle_id": "vehicle_2", 
    "staged": true,
    "reaction_time": 0.067,
    "current_position": 750.2
  }
}
```

#### `GetRaceStatusJSON() string` *(Legacy)*
Returns race status for the first active race. For backward compatibility.

#### `IsRaceCompleteByID(raceID string) bool`
Checks if a specific race has finished.

**Parameters:**
- `raceID`: The unique identifier of the race

**Returns:**
- `bool`: True if race is complete, false if still running

#### `IsRaceComplete() bool` *(Legacy)*
Checks if any active race is complete. For backward compatibility.

### Christmas Tree Status

#### `GetTreeStatusJSONByID(raceID string) string`
Returns the Christmas tree light status as JSON for a specific race.

**Parameters:**
- `raceID`: The unique identifier of the race

**Returns:**
- `string`: JSON-formatted tree status

**Example Response:**
```json
{
  "is_armed": true,
  "is_running": true,
  "sequence_type": "pro",
  "current_step": 3,
  "light_states": {
    "1": {
      "pre_stage": "on",
      "stage": "on", 
      "amber_1": "off",
      "amber_2": "off",
      "amber_3": "off",
      "green": "off",
      "red": "off"
    },
    "2": {
      "pre_stage": "on",
      "stage": "on",
      "amber_1": "off", 
      "amber_2": "off",
      "amber_3": "off",
      "green": "off",
      "red": "off"
    }
  }
}
```

#### `GetTreeStatusJSON() string` *(Legacy)*
Returns tree status for the first active race.

### Race Results

#### `GetResultsJSONByID(raceID string) string`
Returns the final race results as JSON for a specific race.

**Parameters:**
- `raceID`: The unique identifier of the race

**Returns:**
- `string`: JSON-formatted race results

**Example Response:**
```json
{
  "race_id": "123e4567-e89b-12d3-a456-426614174000",
  "winner": 1,
  "margin_of_victory": 0.0234,
  "lane_1": {
    "reaction_time": 0.045,
    "sixty_foot_time": 0.987,
    "eighth_mile_time": 5.234,
    "quarter_mile_time": 8.123,
    "trap_speed": 145.67,
    "is_foul": false
  },
  "lane_2": {
    "reaction_time": 0.067,
    "sixty_foot_time": 1.012,
    "eighth_mile_time": 5.456,
    "quarter_mile_time": 8.456,
    "trap_speed": 142.34,
    "is_foul": false
  }
}
```

#### `GetResultsJSON() string` *(Legacy)*
Returns results for the first active race.

### Race Management

#### `GetActiveRaceCount() int`
Returns the number of currently active races.

**Returns:**
- `int`: Number of active races

#### `GetActiveRaceIDs() []string`
Returns a list of all currently active race IDs.

**Returns:**
- `[]string`: Slice of active race IDs

#### `RaceExists(raceID string) bool`
Checks if a race with the given ID exists.

**Parameters:**
- `raceID`: The unique identifier to check

**Returns:**
- `bool`: True if race exists, false otherwise

#### `GetMaxConcurrentRaces() int`
Returns the maximum number of concurrent races allowed.

**Returns:**
- `int`: Maximum concurrent race limit

#### `SetMaxConcurrentRaces(max int)`
Sets the maximum number of concurrent races allowed.

**Parameters:**
- `max`: Maximum number of concurrent races (must be > 0)

### System Management

#### `Reset() error`
Clears all active races but keeps the API initialized.

**Returns:**
- `error`: Error if API is not initialized

#### `Stop() error`
Shuts down the API and cleans up all resources.

**Returns:**
- `error`: Error if shutdown fails

## Error Handling

The API returns errors in the following situations:

- **API Not Initialized**: When methods are called before `Initialize()`
- **Race Limit Exceeded**: When trying to start more than max concurrent races
- **Race Not Found**: When referencing a non-existent race ID
- **Resource Cleanup**: When cleanup operations fail

## Race States

Races progress through the following states:

1. **`idle`** - Race not started
2. **`staging`** - Vehicles positioning at starting line
3. **`armed`** - Both vehicles staged, tree sequence ready
4. **`running`** - Tree sequence started, vehicles racing
5. **`complete`** - Race finished, results available

## Thread Safety

The libdrag API is thread-safe and supports concurrent access. All methods use appropriate locking mechanisms to ensure data consistency across multiple goroutines.

## Performance Considerations

- **Concurrent Races**: Default limit is 10 concurrent races. Adjust based on system resources.
- **Monitoring**: Race completion monitoring runs automatically in background goroutines.
- **Memory Management**: Completed races are automatically cleaned up after a brief delay.
- **JSON Serialization**: Status and results are cached and serialized on-demand.

## Next Steps

- See [Examples](api-examples.md) for detailed usage scenarios
- Check [Configuration](api-configuration.md) for advanced setup options
- Review [Auto-Start System](auto-start-integration.md) for CompuLink functionality
