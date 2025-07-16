# Boundary Beam and Rollout Distance Monitoring Clarification

## Overview

Based on research into NHRA CompuLink timing systems and official drag racing protocols, this document clarifies the distinction between automated and manual enforcement systems in professional drag racing.

## CompuLink Guard Beam System (Automated)

### Current NHRA Implementation
- **Guard Beam Position**: 13 3/8 inches downstream from stage beam
- **Purpose**: Automatic rollout distance enforcement
- **Function**: Starts ET clock if stage beam fails to trigger
- **Enforcement**: Limits staging rollout to specified distance

### Technical Operation
- **Automated Detection**: Electronic beam system detects wheel position
- **Automatic Enforcement**: System triggers timing without human intervention
- **Rollout Limiting**: Prevents excessive rollout beyond allowed distance
- **False Trigger Protection**: Modern systems filter out debris/insects

### System Evolution
- **Original System**: Introduced by CompuLink in 1984
- **Recent Upgrades**: Reinstated ~4 years ago due to Pro Stock timing issues
- **Future Technology**: New infra-red wall technology under development
  - Works off car body rather than front wheel
  - Enhanced precision and reliability

## Manual vs Automated Enforcement Summary

### Automated Equipment-Based Enforcement
- **Rollout Distance**: Guard beam system automatically enforces maximum rollout
- **Staging Beams**: Pre-stage and stage beam detection is fully automated
- **Timing Systems**: All timing measurements are automated and electronic
- **Deep Staging Detection**: Can be automatically detected by monitoring beam states

### Manual Official-Based Enforcement
- **Centerline Violations**: Track officials manually observe and call violations
- **Unsportsmanlike Conduct**: Officials make judgment calls on behavior
- **Equipment Violations**: Officials inspect and approve/reject equipment
- **Lane Boundary Crossings**: Visual observation by track officials

## Implementation Implications for libdrag

### Phase 1: Automated Systems (Equipment-Based)
1. **Guard Beam Rollout Enforcement** - Can be fully automated
   - Monitor distance past stage beam
   - Trigger violations automatically
   - Class-specific rollout limits (6" professional, 12" sportsman)

2. **Deep Staging Detection** - Can be fully automated
   - Monitor pre-stage beam state when staged
   - Automatically detect and enforce class-specific restrictions
   - Immediate foul assessment for prohibited classes

### Phase 2: Manual Integration (Official-Based)
1. **Official Violation Recording Interface** - Requires human input
   - API for track officials to record centerline violations
   - Manual entry of unsportsmanlike conduct penalties
   - Integration with automated timing results

2. **Hybrid Systems** - Combination of automated detection and manual confirmation
   - Automated detection with official verification
   - Manual override capabilities for all systems

## Current libdrag Status

### ✅ Correctly Identified as Automated
- Rollout distance monitoring (guard beam system)
- Staging beam detection and enforcement
- Timing system precision and measurement

### ⚠️ Correctly Identified as Manual
- Centerline violation detection and enforcement
- Unsportsmanlike conduct penalties
- Equipment inspection and approval

### 🔧 Implementation Priority
1. **High Priority**: Automated rollout distance enforcement (guard beam)
2. **High Priority**: Deep staging detection and enforcement
3. **Medium Priority**: Manual violation recording interface
4. **Low Priority**: Advanced CompuLink integration features

## Technical Specifications

### Guard Beam Configuration
```go
type GuardBeamConfig struct {
    Position        float64       `json:"position"`         // 13.375 inches from stage beam
    MaxRollout      float64       `json:"max_rollout"`      // Class-specific limits
    AutoEnforcement bool          `json:"auto_enforcement"` // Automatic violation detection
    FalseFilter     bool          `json:"false_filter"`     // Filter debris/insects
}
```

### Class-Specific Rollout Limits
- **Professional Classes**: 6 inches maximum rollout
- **Sportsman Classes**: 12 inches maximum rollout
- **Junior Dragster**: Enhanced safety limits (varies by division)

---

**Sources**:
- CompuLink Timing Systems Technical Documentation
- NHRA Official Timing Equipment Specifications
- Competition Plus Drag Racing News Archives
- NHRA 101 Technical Reference Guide

**Created**: 2025-07-16
**Last Updated**: 2025-07-16
**Next Review**: After Phase 1 implementation