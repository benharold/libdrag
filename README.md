# libdrag

A cross-platform Go library for accurately simulating NHRA and IHRA drag racing events, including timing systems, Christmas tree sequencing, and race orchestration.

## Features

- 🏁 **Accurate Race Simulation**: Simulates NHRA/IHRA drag racing with realistic timing
- 🚦 **Christmas Tree**: Full Christmas tree light sequence simulation
- ⏱️ **Precision Timing**: High-precision timing system for accurate race results
- 🚗 **Vehicle Simulation**: Configurable vehicle performance characteristics
- 🎮 **Cross-Platform**: Works on Windows, macOS, Linux, and mobile platforms
- 📊 **JSON API**: Clean JSON interface for easy integration
- 🔧 **Configurable**: Flexible configuration system for different racing formats
- 🏆 **Concurrent Races**: Support for multiple simultaneous races with unique IDs

## Installation

```bash
go get github.com/benharold/libdrag
```

## Quick Start

### Basic Single Race (Legacy API)

```go
package main

import (
    "fmt"
    "time"
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    // Create and initialize the libdrag API
    libdragAPI := api.NewLibDragAPI()
    
    if err := libdragAPI.Initialize(); err != nil {
        panic(err)
    }
    
    // Start a race
    if err := libdragAPI.StartRace(); err != nil {
        panic(err)
    }
    
    // Wait for race completion
    for !libdragAPI.IsRaceComplete() {
        time.Sleep(100 * time.Millisecond)
    }
    
    // Get results
    results := libdragAPI.GetResultsJSON()
    fmt.Println("Race Results:", results)
    
    // Clean shutdown
    if err := libdragAPI.Stop(); err != nil {
        panic(err)
    }
}
```

### Multiple Concurrent Races (New API)

```go
package main

import (
    "fmt"
    "time"
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    // Create and initialize the libdrag API
    libdragAPI := api.NewLibDragAPI()
    
    if err := libdragAPI.Initialize(); err != nil {
        panic(err)
    }
    
    // Start multiple races
    raceID1, err := libdragAPI.StartRaceWithID()
    if err != nil {
        panic(err)
    }
    
    raceID2, err := libdragAPI.StartRaceWithID()
    if err != nil {
        panic(err)
    }
    
    // Monitor races independently
    for !libdragAPI.IsRaceCompleteByID(raceID1) || !libdragAPI.IsRaceCompleteByID(raceID2) {
        time.Sleep(100 * time.Millisecond)
        
        // Get status for each race
        fmt.Println("Race 1 Status:", libdragAPI.GetRaceStatusJSONByID(raceID1))
        fmt.Println("Race 2 Status:", libdragAPI.GetRaceStatusJSONByID(raceID2))
    }
    
    // Get results for each race
    fmt.Println("Race 1 Results:", libdragAPI.GetResultsJSONByID(raceID1))
    fmt.Println("Race 2 Results:", libdragAPI.GetResultsJSONByID(raceID2))
    
    // Clean shutdown
    if err := libdragAPI.Stop(); err != nil {
        panic(err)
    }
}
```

## API Reference

### Core API

- `NewLibDragAPI()` - Create a new libdrag API instance
- `Initialize() error` - Initialize the racing system
- `Stop() error` - Shutdown the system and cleanup all races

### Race Management

#### Legacy Single Race API
- `StartRace() error` - Start a new drag race (legacy method)
- `IsRaceComplete() bool` - Check if the current race is finished
- `GetResultsJSON() string` - Get race results as JSON
- `GetTreeStatusJSON() string` - Get Christmas tree status as JSON
- `GetRaceStatusJSON() string` - Get current race status as JSON

#### Multi-Race API with IDs
- `StartRaceWithID() (string, error)` - Start a new race and return unique race ID
- `IsRaceCompleteByID(raceID string) bool` - Check if a specific race is finished
- `GetResultsJSONByID(raceID string) string` - Get race results as JSON for specific race
- `GetTreeStatusJSONByID(raceID string) string` - Get Christmas tree status for specific race
- `GetRaceStatusJSONByID(raceID string) string` - Get race status for specific race
- `CompleteRace(raceID string) error` - Manually complete and cleanup a race

### Configuration & Management

- `Reset() error` - Clear all active races but keep API initialized
- `GetActiveRaceCount() int` - Get number of currently active races
- `GetMaxConcurrentRaces() int` - Get maximum allowed concurrent races
- `SetMaxConcurrentRaces(max int)` - Set maximum allowed concurrent races

## Architecture

The library is structured with clear separation of concerns:

- **pkg/api**: Public API interface with support for concurrent races
- **pkg/orchestrator**: Race orchestration and coordination
- **pkg/timing**: High-precision timing system
- **pkg/tree**: Christmas tree light sequence
- **pkg/component**: Component system architecture
- **pkg/config**: Configuration management
- **pkg/events**: Event bus system for component communication
- **internal/vehicle**: Vehicle simulation (internal implementation)

## Racing Formats Supported

- NHRA Pro Tree (0.4 second intervals)
- NHRA Sportsman Tree (0.5 second intervals)
- IHRA formats
- Custom timing configurations

## Concurrent Racing

The library supports multiple simultaneous races with these features:

- **Race IDs**: Each race gets a unique UUID for tracking
- **Configurable Limits**: Set maximum concurrent races (default: 10)
- **Automatic Cleanup**: Races are automatically cleaned up when complete
- **Resource Management**: Efficient memory usage with proper cleanup
- **Timeout Protection**: Races automatically timeout after 30 seconds

## Use Cases

- **Racing Games**: Integrate realistic drag racing into gaming applications
- **Training Simulators**: Practice timing and reaction for real racers
- **Mobile Apps**: Build drag racing apps for iOS/Android
- **Analysis Tools**: Analyze racing data and performance
- **Educational**: Learn about drag racing timing and procedures
- **Multi-User Platforms**: Support multiple simultaneous users/races

## Contributing

Contributions are welcome! Please feel free to submit pull requests, create issues, or suggest new features.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions, issues, or feature requests, please create an issue on GitHub.
