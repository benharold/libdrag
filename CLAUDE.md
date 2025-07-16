# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `make build` - Build the libdrag binary
- `make run` - Build and run the application
- `go run main.go` - Run the main demo
- `go run cmd/libdrag/main.go` - Run the command-line demo

### Testing
- `make test` - Run all tests with verbose output
- `make coverage` - Generate coverage report (creates coverage.html)
- `go test ./pkg/timing -run TestReactionTimeCalculation` - Run specific test
- `go test ./pkg/timing ./pkg/tree ./pkg/config` - Test specific packages

### Code Quality
- `make check` - Run all checks (fmt, vet, lint, test)
- `make fmt` - Format code using gofmt
- `make vet` - Vet code for potential issues  
- `make lint` - Lint code using golangci-lint (requires `make dev-deps`)
- `make dev-deps` - Install development dependencies

### Dependencies
- `make deps` - Download and tidy Go modules
- `make clean` - Clean build artifacts and coverage files

## System Architecture

LibDrag is a professional drag racing simulation library implementing CompuLink auto-start systems and NHRA/IHRA protocols with a component-based architecture.

### Core Component System
- **Component Interface**: All components implement `Initialize()`, `Arm()`, `EmergencyStop()`, `GetStatus()`, `GetID()`
- **Event-Driven**: Components communicate through a central event bus with type-safe subscriptions
- **Orchestrator Pattern**: Race orchestrator coordinates component lifecycle through direct method calls

### Key Components
- **pkg/api**: Public JSON API supporting concurrent races with unique UUIDs
- **pkg/orchestrator**: Race state management (Idle → Preparing → Staging → Armed → Running → Complete)
- **pkg/autostart**: CompuLink auto-start system with three-light rule and staging timeout
- **pkg/tree**: Christmas tree with Armed/Activated states, Pro vs Sportsman sequences
- **pkg/timing**: High-precision timing system with beam integration and foul detection
- **pkg/beam**: Beam state management for pre-stage, stage, and timing beams
- **pkg/events**: Event bus with comprehensive race event taxonomy
- **pkg/config**: NHRA-standard track and timing configurations
- **pkg/component**: Base component interface and event-aware components

### Auto-Start System Workflow
1. Starter arms tree (manual action) → Tree enters Armed state
2. Vehicles enter pre-stage → Tree lights illuminate  
3. Three-light rule (2 pre-stage + 1 stage) → Auto-start system Activates
4. Both vehicles stage → Minimum staging timer starts
5. Random delay expires → Tree sequence begins
6. Green light → Race timing begins

### Event Flow Architecture
```
Beam triggers → Timing System → Events → Auto-Start System
Auto-Start activation → Christmas Tree → Sequence execution  
Tree state changes → Events → Orchestrator coordination
Component status → Orchestrator → API status reporting
```

### Concurrent Racing
- Multiple simultaneous races supported with UUID tracking
- Configurable race limits (default: 10 concurrent races)
- Automatic cleanup of completed races with 30-second timeout protection
- Resource-efficient memory management

## Testing Strategy

### Coverage Targets
- **Christmas Tree (pkg/tree)**: 80%+ coverage - Pre-stage/stage sequences, Pro vs Sportsman timing
- **Timing System (pkg/timing)**: 55%+ coverage - Reaction times, splits, red light detection  
- **Configuration (pkg/config)**: 60%+ coverage - NHRA defaults, validation

### Test Categories
- **Unit Tests**: Individual component testing in isolation
- **Integration Tests**: Component interaction testing (`pkg/orchestrator`, `pkg/api`)
- **Auto-Start Integration**: Real-world auto-start system behavior (`pkg/autostart/integration.go`)

## Project Structure

- **cmd/libdrag/**: Command-line demo application
- **pkg/**: All public library packages following Go conventions
- **internal/vehicle/**: Internal vehicle simulation (not public API)
- **examples/**: Usage examples and race monitor
- **docs/**: API documentation and drag racing terminology references

## Key Terminology

- **Armed**: Starter has manually enabled the auto-start system
- **Activated**: Auto-start system has automatically detected staging conditions
- **Three-Light Rule**: Auto-start activates when 2 pre-stage + 1 stage light achieved
- **Deep Staging**: Vehicle positioned past stage beam (advanced technique)
- **Red Light**: Foul start - vehicle left before green light