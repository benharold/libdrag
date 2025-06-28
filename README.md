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

## Installation

```bash
go get github.com/benharold/libdrag
```

## Quick Start

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
    libdragAPI.Stop()
}
```

## API Reference

### Core API

- `NewLibDragAPI()` - Create a new libdrag API instance
- `Initialize()` - Initialize the racing system
- `StartRace()` - Start a new drag race
- `IsRaceComplete()` - Check if the current race is finished
- `Stop()` - Shutdown the system

### Data Retrieval

- `GetResultsJSON()` - Get race results as JSON
- `GetTreeStatusJSON()` - Get Christmas tree status as JSON
- `GetRaceStatusJSON()` - Get current race status as JSON

## Architecture

The library is structured with clear separation of concerns:

- **pkg/api**: Public API interface
- **pkg/orchestrator**: Race orchestration and coordination
- **pkg/timing**: High-precision timing system
- **pkg/tree**: Christmas tree light sequence
- **pkg/component**: Component system architecture
- **pkg/config**: Configuration management
- **internal/vehicle**: Vehicle simulation (internal implementation)

## Racing Formats Supported

- NHRA Pro Tree (0.4 second intervals)
- NHRA Sportsman Tree (0.5 second intervals)
- IHRA formats
- Custom timing configurations

## Use Cases

- **Racing Games**: Integrate realistic drag racing into gaming applications
- **Training Simulators**: Practice timing and reaction for real racers
- **Mobile Apps**: Build drag racing apps for iOS/Android
- **Analysis Tools**: Analyze racing data and performance
- **Educational**: Learn about drag racing timing and procedures

## Contributing

Contributions are welcome! Please feel free to submit pull requests, create issues, or suggest new features.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions, issues, or feature requests, please create an issue on GitHub.
