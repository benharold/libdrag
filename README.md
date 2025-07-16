# libdrag

[![Go Reference](https://pkg.go.dev/badge/github.com/benharold/libdrag.svg)](https://pkg.go.dev/github.com/benharold/libdrag)
[![Go Report Card](https://goreportcard.com/badge/github.com/benharold/libdrag)](https://goreportcard.com/report/github.com/benharold/libdrag)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **⚠️ Work in Progress**: This library is under active development. While functional for basic drag racing simulation, it is not yet fully compliant with all NHRA/IHRA standards. See [Compliance Status](#compliance-status) below.
A cross-platform Go library for accurately simulating NHRA and IHRA drag racing events, including professional auto-start systems, Christmas tree sequencing, high-precision timing, and race orchestration.
## 🚧 Project Status

**Current Version**: v0.1.0-alpha (In Development)

This library currently implements core drag racing simulation with:
- ✅ Basic auto-start system (three-light rule)
- ✅ Christmas tree sequencing (Pro/Sportsman)
- ✅ High-precision timing system
- ✅ Concurrent race support
- ✅ Event-driven architecture
- ⚠️ **Partial NHRA/IHRA compliance** (see [docs/nhra-ihra-compliance.md](docs/nhra-ihra-compliance.md))

**Not yet production-ready for professional racing events.**

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

### Basic Usage

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

## Testing & Development

### Running Tests

libdrag includes comprehensive unit tests that validate all core drag racing functionality. The tests cover timing calculations, Christmas tree sequences, configuration, and race orchestration.

#### Basic Test Execution

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific packages
go test ./pkg/timing ./pkg/tree ./pkg/config

# Run a specific test
go test -v ./pkg/timing -run TestReactionTimeCalculation
```

#### Test Coverage Reports

##### Generate Coverage Profile
```bash
# Generate coverage profile for all packages
go test -coverprofile=coverage.out ./...

# Generate coverage for specific packages only
go test -coverprofile=coverage.out ./pkg/timing ./pkg/tree ./pkg/config
```

##### View Coverage in Terminal
```bash
# Show coverage summary
go test -cover ./...

# Show detailed function-by-function coverage
go tool cover -func=coverage.out

# Show total coverage percentage
go tool cover -func=coverage.out | grep "total"
```

##### Generate HTML Coverage Report
```bash
# Generate interactive HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser (macOS)
open coverage.html

# Open coverage report in browser (Linux)
xdg-open coverage.html

# Open coverage report in browser (Windows)
start coverage.html
```

#### Coverage Targets by Component

The library maintains high test coverage for core drag racing functionality:

- **🚦 Christmas Tree (`pkg/tree`)**: **80%+ coverage**
  - Pre-stage/stage light sequences
  - Pro vs Sportsman tree timing (0.4s vs 0.5s)
  - Tree arming and error handling

- **⏱️ Timing System (`pkg/timing`)**: **55%+ coverage**
  - Reaction time calculations
  - 60-foot, 1/8 mile, 1/4 mile splits
  - Red light detection (jumping the start)
  - Speed trap calculations

- **⚙️ Configuration (`pkg/config`)**: **60%+ coverage**
  - NHRA-standard defaults
  - Track and beam layout validation
  - Tree sequence configurations

#### Test Categories

**Unit Tests**: Test individual components in isolation
```bash
go test ./pkg/timing -v    # Timing calculations
go test ./pkg/tree -v      # Christmas tree logic
go test ./pkg/config -v    # Configuration validation
```

**Integration Tests**: Test component interactions
```bash
go test ./pkg/orchestrator -v  # Race coordination
go test ./pkg/api -v           # End-to-end API
```

#### Continuous Integration

For CI/CD pipelines, use these commands:

```bash
# Run tests with coverage and fail if below threshold
go test -coverprofile=coverage.out ./... && \
go tool cover -func=coverage.out | grep "total" | \
awk '{print $3}' | sed 's/%//' | \
awk '{if($1<50) exit 1}'

# Generate coverage badge data
go test -coverprofile=coverage.out ./... && \
go tool cover -func=coverage.out | grep "total" | \
awk '{print "Coverage: " $3}'
```

#### Performance Testing

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...

# Profile CPU usage during tests
go test -cpuprofile=cpu.prof ./pkg/timing
go tool pprof cpu.prof
```

### Development Guidelines

#### Running the Demo

```bash
# Build and run the main demo
go run main.go

# Build and run the command-line demo
go run cmd/libdrag/main.go

# Build standalone binary
go build -o libdrag-demo main.go
./libdrag-demo
```

#### Code Quality Checks

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code for issues
go vet ./...

# Check for unused dependencies
go mod tidy
```

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

## Compliance Status

### ✅ Currently Implemented (NHRA/IHRA Compliant)
- Three-light rule auto-start activation (NHRA 4.4.1, IHRA 3.2.2)
- Pro Tree (0.4s) and Sportsman Tree (0.5s) sequences
- Basic staging timeouts (7s Professional, 10s Sportsman)
- Random delay generation (0.6-1.1s Pro, 0.6-1.4s Sportsman)
- Guard beam violation detection
- Emergency stop protocols

### ❌ Not Yet Implemented (Critical for Full Compliance)
- **Deep staging restrictions enforcement** (Super Gas/Stock/Street classes)
- **Centerline violation detection** (Required for professional events)
- **Complete class configurations** (missing 10+ NHRA/IHRA classes)
- **Enhanced safety systems** (boundary beams, backup timing)
- **Race director override capabilities**
- **Anti-cheating systems** (delay box detection)

See [docs/nhra-ihra-compliance.md](docs/nhra-ihra-compliance.md) for detailed compliance status.

## Use Cases

- **Racing Games**: Integrate realistic drag racing into gaming applications
- **Training Simulators**: Practice timing and reaction for real racers
- **Mobile Apps**: Build drag racing apps for iOS/Android
- **Analysis Tools**: Analyze racing data and performance
- **Educational**: Learn about drag racing timing and procedures
- **Multi-User Platforms**: Support multiple simultaneous users/races
- **Development Reference**: Study CompuLink auto-start system implementation

## Future Possibilities

As a hobby project, development happens when time and interest allow. Potential future enhancements include:

### Enhanced Compliance Features
- Deep staging restriction enforcement
- Complete NHRA/IHRA class configurations
- Enhanced auto-start system features
- Centerline violation detection

### Advanced Features
- Race director control panel
- Hardware integration support
- Anti-cheating detection systems
- Enhanced safety protocols

### Community Contributions
- IHRA-specific rule variations
- Regional track configurations
- Performance optimizations
- Documentation improvements

No timelines are set - contributions and development happen organically based on community interest and available time.

## Contributing

🎯 **Contributions welcome!** This hobby project would benefit from:

- **Drag racing experts** with NHRA/IHRA rule knowledge
- **Timing system experience** (CompuLink, etc.)
- **Go developers** interested in motorsports
- **Documentation improvements**
- **Test coverage expansion**

### Getting Started
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure `make test` passes
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support & Community

- 🐛 **Bug Reports**: [Create an issue](https://github.com/benharold/libdrag/issues)
- 💡 **Feature Requests**: [Create an issue](https://github.com/benharold/libdrag/issues)
- 💬 **Questions**: [GitHub Discussions](https://github.com/benharold/libdrag/discussions)
- 📖 **Documentation**: [docs/](docs/) directory

## Acknowledgments

- **NHRA** and **IHRA** for drag racing standards and specifications
- **CompuLink** for auto-start system protocols
- The drag racing community for technical expertise and feedback
