# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **CompuLink Auto-Start System**: Complete implementation of professional drag racing auto-start functionality
  - Three-light rule implementation (2 pre-stage + 1 stage = auto-start trigger)
  - Configurable staging timeouts (7-20 seconds) for different racing classes
  - Random delay generation (0.6-1.4 seconds with variation) matching CompuLink specifications
  - Guard beam violation detection with configurable rollout limits
  - Professional vs Sportsman tree sequence support
  - Manual starter override and emergency stop capabilities
  - Real-time parameter adjustment for different racing classes (Top Fuel, Bracket, Junior Dragster)
  - Comprehensive test suite with 8 test scenarios covering all functionality
- **Auto-Start Integration Layer**: Seamless integration with existing timing and Christmas tree components
  - Real-time beam monitoring and staging status updates
  - Automatic Christmas tree sequence triggering
  - Event-driven architecture for component coordination
  - Beam mapping for pre-stage, stage, and guard beam detection
- **Enhanced Makefile**: Improved developer experience with comprehensive help system
  - Default help target displays available commands when running `make` without arguments
  - Automatic target discovery from comment documentation
  - Additional utility targets: `check`, `run`, `dev-deps`
  - Complete `clean` target for proper artifact removal
  - Professional formatting with libdrag branding
- Initial release of libdrag library
- Cross-platform drag racing simulation
- NHRA/IHRA timing system support
- Christmas tree light sequence simulation
- Vehicle performance simulation
- JSON API for easy integration
- Race orchestration system
- High-precision timing components
- Configurable racing formats

### Features
- **Professional Auto-Start System**: Industry-standard CompuLink-compatible auto-start functionality
  - State machine management (Idle → Armed → Staging → Triggered → Fault)
  - Safety violation handling and fault reporting
  - Class-specific timing configurations
  - Thread-safe operation with proper synchronization
- Complete drag racing event simulation
- Real-time race monitoring
- Results and status reporting
- Clean shutdown procedures

### Technical Improvements
- **Code Organization**: New `/pkg/autostart/` package with modular design
  - Core auto-start engine (`autostart.go`)
  - Integration layer (`integration.go`) 
  - Comprehensive test suite (`autostart_test.go`)
- **Build System**: Enhanced Makefile with developer-friendly help system
- **Documentation**: Added CompuLink specifications and implementation notes

### Testing
- **Auto-Start Test Coverage**: 8 comprehensive test scenarios
  - Three-light rule validation
  - Staging timeout enforcement
  - Guard beam violation detection
  - Full staging sequence with random delays
  - Manual override functionality
  - Configuration management
  - Event handling system
  - Class-specific parameter validation

## [0.1.0] - 2025-06-27

### Added
- Initial library structure and core functionality
- Basic API interface for race simulation
- Component-based architecture
- Example demonstration application
