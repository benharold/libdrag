# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0-alpha] - 2025-07-16

### Added
- Initial public release of libdrag library
- CompuLink auto-start system implementation with three-light rule
- Christmas tree sequencing (Pro and Sportsman)
- High-precision timing system with beam integration
- Event-driven architecture with comprehensive event bus
- Concurrent race support with UUID tracking
- JSON API for race management and monitoring
- Component-based architecture with clean interfaces
- Comprehensive test suite with integration tests
- Official NHRA/IHRA rule book compliance documentation
- Professional drag racing terminology and research documentation

### Technical Implementation
- Auto-start system with staging timeouts and random delays
- Pro tree (0.4s) and Sportsman tree (0.5s) sequences
- Reaction time calculation and red light detection
- Complete race lifecycle management
- Type-safe event bus for component communication
- NHRA-standard track and timing configurations

### Documentation
- Complete API reference with examples
- NHRA/IHRA rule compliance matrix
- CompuLink auto-start system research
- Drag racing terminology and specifications

### Current Limitations
- Partial NHRA/IHRA compliance (see docs/nhra-ihra-compliance.md)
- Missing deep staging restrictions for certain classes
- No centerline violation detection
- Not production-ready for professional racing events