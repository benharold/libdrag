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
- **pkg/tree**: Christmas tree with Armed/Activated states, Pro vs Sportsman sequences, deep staging detection
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
- **Christmas Tree (pkg/tree)**: 80%+ coverage - Pre-stage/stage sequences, Pro vs Sportsman timing, deep staging enforcement
- **Timing System (pkg/timing)**: 55%+ coverage - Reaction times, splits, red light detection  
- **Configuration (pkg/config)**: 60%+ coverage - NHRA defaults, validation

### Test Categories
- **Unit Tests**: Individual component testing in isolation
- **Integration Tests**: Component interaction testing (`pkg/orchestrator`, `pkg/api`)
- **Auto-Start Integration**: Real-world auto-start system behavior (`pkg/autostart/integration.go`)
- **Deep Staging Tests**: TDD implementation with comprehensive class-specific rule testing (`pkg/tree/deep_staging_test.go`)

## Standards Compliance

LibDrag implements official NHRA and IHRA sanctioning body requirements:

- **NHRA Rule Book Compliance**: Sections 4.2-4.4 (Competition Procedures), Section 6.1 (Safety), Sections 9.1-9.6 (Sportsman), Sections 11.1-11.5 (Professional)
- **IHRA Rule Book Compliance**: Sections 3.2-3.3 (Competition Rules), Section 7 (Professional), Section 8 (Sportsman)
- **CompuLink Auto-Start Protocol**: Full three-light rule implementation with class-specific timing
- **See**: `docs/nhra-ihra-compliance.md` for detailed rule compliance matrix and implementation status

### Critical Compliance Features
- ✅ Three-light rule implementation (NHRA 4.4.1, IHRA 3.2.2)
- ✅ Class-specific staging timeouts (7s Professional, 10s Sportsman, 15s Bracket)
- ✅ Random delay specifications (0.6-1.1s Pro, 0.6-1.4s Sportsman)
- ✅ Pro vs Sportsman tree sequences (0.4s/0.5s green delays)
- ✅ **Deep staging restrictions enforcement** (Super Gas/Stock/Street prohibited, Pro classes allowed)
- ✅ **Forward motion staging rule** (Last motion must be forward, no backing and re-staging)
- ❌ **Centerline violation detection** (Required for professional events)

## Project Structure

- **cmd/libdrag/**: Command-line demo application
- **pkg/**: All public library packages following Go conventions
- **internal/vehicle/**: Internal vehicle simulation (not public API)
- **examples/**: Usage examples and race monitor
- **docs/**: API documentation, rule compliance, and drag racing terminology references
  - **docs/nhra-ihra-compliance.md**: Official rule book compliance documentation
  - **docs/api-documentation.md**: Complete API reference
  - **docs/research/**: Technical research and implementation details

## Key Terminology

- **Armed**: Starter has manually enabled the auto-start system
- **Activated**: Auto-start system has automatically detected staging conditions
- **Three-Light Rule**: Auto-start activates when 2 pre-stage + 1 stage light achieved
- **Deep Staging**: Vehicle positioned past stage beam (advanced technique)
- **Deep Staging Violation**: Prohibited in Super Gas, Super Stock, and Super Street classes
- **Red Light**: Foul start - vehicle left before green light

## Deep Staging Implementation

### Class-Specific Rules
- **Prohibited Classes**: Super Gas, Super Stock, Super Street
- **Allowed Classes**: Top Fuel, Funny Car, Pro Stock, Bracket Racing
- **Detection**: Pre-stage light OFF + Stage light ON = Deep staging detected
- **Events**: `tree.deep_stage_violation` for prohibited classes, `tree.deep_stage` for allowed classes

### Forward Motion Staging Rule
- **Rule**: Last motion into staging area MUST be forward motion
- **Legal Sequence**: Pre-stage → Stage → Deep Stage (optional)
- **Violation**: Pre-stage → Stage → Back out of Stage → Re-enter Stage
- **Reset**: Complete back-out (both beams clear) resets motion tracking
- **Events**: `tree.staging_violation` with violation type `backward_staging_motion`

### Method Signatures
- `SetPreStage(lane int, beamBroken bool)` - Pre-stage beam state
- `SetStage(lane int, beamBroken bool)` - Stage beam state
- Light states follow beam states: broken beam = light ON, clear beam = light OFF