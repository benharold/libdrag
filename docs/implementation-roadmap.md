# libdrag Implementation Roadmap

## Overview

This document provides a detailed implementation roadmap for achieving full NHRA/IHRA compliance in the libdrag library. The roadmap is divided into three phases, prioritizing safety-critical features first.

---

## Phase 1: Safety & Compliance (Essential) - 1-2 weeks

**Goal**: Achieve minimum safety compliance for sanctioned sportsman racing events.

### 1.1 Deep Staging Enforcement Implementation ✅ COMPLETED

**Priority**: 🚨 CRITICAL - Safety requirement  
**Completed**: 2025-07-16  
**NHRA Rules**: 9.1.4, 9.2.4, 9.3.4  
**IHRA Rules**: 3.2.1, 8.1-8.3

**Implementation Summary**:
- ✅ Class-specific deep staging prohibition (Super Gas, Super Stock, Super Street)
- ✅ Professional class allowance (Top Fuel, Funny Car, Pro Stock)
- ✅ Forward motion staging rule enforcement
- ✅ Real-time violation detection and event publishing
- ✅ Complete TDD test coverage with 12 comprehensive test scenarios
- ✅ Motion history tracking for audit purposes

#### Implementation Steps:

1. **Add Class Configuration System**
   ```go
   // File: pkg/config/classes.go (new file)
   type ClassConfig struct {
       Name               string        `json:"name"`
       StagingTimeout     time.Duration `json:"staging_timeout"`
       DeepStagingAllowed bool          `json:"deep_staging_allowed"`
       TreeType           TreeSequenceType `json:"tree_type"`
       MaxRolloutInches   float64       `json:"max_rollout_inches"`
       RandomDelayMin     time.Duration `json:"random_delay_min"`
       RandomDelayMax     time.Duration `json:"random_delay_max"`
   }
   ```

2. **Modify Christmas Tree for Deep Staging Detection**
   ```go
   // File: pkg/tree/tree.go - Add to SetStage method
   func (ct *ChristmasTree) SetStage(lane int) {
       // Existing staging logic...
       
       // Check for deep staging violation
       if ct.config.RacingClass() != "" {
           if err := ct.validateDeepStaging(lane); err != nil {
               ct.handleDeepStagingViolation(lane, err)
               return
           }
       }
   }
   
   func (ct *ChristmasTree) validateDeepStaging(lane int) error {
       class := getClassConfig(ct.config.RacingClass())
       if !class.DeepStagingAllowed {
           if ct.status.LightStates[lane][LightPreStage] == LightOff {
               return fmt.Errorf("deep staging prohibited in class %s", class.Name)
           }
       }
       return nil
   }
   ```

3. **Add Class Definitions**
   ```go
   // File: pkg/config/nhra_classes.go (new file)
   var NHRAClasses = map[string]ClassConfig{
       "super_gas": {
           Name:               "Super Gas",
           StagingTimeout:     10 * time.Second,
           DeepStagingAllowed: false, // NHRA Rule 9.1.4
           TreeType:           TreeSequenceSportsman,
           MaxRolloutInches:   12.0,
           RandomDelayMin:     600 * time.Millisecond,
           RandomDelayMax:     1400 * time.Millisecond,
       },
       "super_stock": {
           Name:               "Super Stock", 
           StagingTimeout:     10 * time.Second,
           DeepStagingAllowed: false, // NHRA Rule 9.2.4
           TreeType:           TreeSequenceSportsman,
           MaxRolloutInches:   12.0,
       },
       "super_street": {
           Name:               "Super Street",
           StagingTimeout:     10 * time.Second, 
           DeepStagingAllowed: false, // NHRA Rule 9.3.4
           TreeType:           TreeSequenceSportsman,
           MaxRolloutInches:   12.0,
       },
       "top_fuel": {
           Name:               "Top Fuel",
           StagingTimeout:     7 * time.Second,
           DeepStagingAllowed: true,
           TreeType:           TreeSequencePro,
           MaxRolloutInches:   6.0,
           RandomDelayMin:     600 * time.Millisecond,
           RandomDelayMax:     1100 * time.Millisecond,
       },
   }
   ```

4. **Update Auto-Start System**
   ```go
   // File: pkg/autostart/autostart.go - Modify activation logic
   func (as *AutoStartSystem) validateStaging() error {
       for _, status := range as.stagingStatus {
           if status.Staged {
               if err := as.validateClassCompliance(status.Lane); err != nil {
                   return err
               }
           }
       }
       return nil
   }
   ```

#### Testing Requirements:
- Add test cases for each class with deep staging scenarios
- Test violation detection and handling
- Verify class-specific timeout enforcement

---

### 1.2 Guard Beam Rollout Distance Enforcement

**Priority**: 🚨 CRITICAL - Automated safety system  
**Estimated Time**: 2-3 days  
**NHRA Specification**: Guard beam at 13.375" from stage beam

#### Implementation Steps:

1. **Add Guard Beam Configuration**
   ```go
   // File: pkg/config/config.go - Add guard beam specs
   type GuardBeamConfig struct {
       Position        float64 `json:"position"`         // 13.375 inches
       MaxRollout      float64 `json:"max_rollout"`      // Class-specific
       AutoEnforcement bool    `json:"auto_enforcement"` // Automatic detection
       FalseFilter     bool    `json:"false_filter"`     // Filter debris
   }
   ```

2. **Implement Automated Rollout Monitoring**
   ```go
   // File: pkg/timing/rollout.go (new file)
   func (ts *TimingSystem) MonitorRolloutDistance(lane int, position float64) error {
       guardBeam := ts.config.Track().GuardBeam
       maxRollout := ts.getClassMaxRollout()
       
       if position > guardBeam.Position + maxRollout {
           return ts.triggerRolloutViolation(lane, position)
       }
       return nil
   }
   ```

### 1.3 Class-Specific Staging Timeouts

**Priority**: 🚨 HIGH - Rule compliance  
**Estimated Time**: 1-2 days  
**Dependencies**: Class configuration system from 1.1

#### Implementation Steps:

1. **Modify Auto-Start Timeout Logic**
   ```go
   // File: pkg/autostart/autostart.go
   func (as *AutoStartSystem) getStagingTimeout() time.Duration {
       class := getClassConfig(as.config.RacingClass())
       if class != nil {
           return class.StagingTimeout
       }
       return 10 * time.Second // fallback
   }
   ```

2. **Update Configuration Interface**
   ```go
   // File: pkg/config/config.go
   type Config interface {
       Track() TrackConfig
       Timing() TimingConfig  
       Tree() TreeSequenceConfig
       Safety() SafetyConfig
       RacingClass() string
       ClassConfig() *ClassConfig // New method
   }
   ```

#### Testing Requirements:
- Test professional class 7-second timeouts
- Test sportsman class 10-second timeouts  
- Test bracket racing 15-second timeouts
- Verify timeout enforcement accuracy

---

### 1.4 Enhanced Error Handling & Logging

**Priority**: 🔧 MEDIUM - Code quality  
**Estimated Time**: 1 day

#### Implementation Steps:

1. **Replace Print Statements with Structured Logging**
   ```go
   // File: pkg/tree/tree.go - Replace fmt.Println calls
   import "log/slog"
   
   func (ct *ChristmasTree) SetPreStage(lane int) {
       // Replace: fmt.Printf("🟡 libdrag: Pre-stage light ON for lane %d\n", lane)
       slog.Info("Pre-stage light activated", 
           "lane", lane, 
           "race_id", ct.raceID,
           "component", "christmas_tree")
   }
   ```

2. **Add Context Cancellation for Goroutines**
   ```go
   // File: pkg/tree/tree.go
   func (ct *ChristmasTree) StartSequence(ctx context.Context, sequenceType config.TreeSequenceType) error {
       // Add context parameter and use in goroutines
       go ct.runSequenceWithContext(ctx, sequenceType)
   }
   ```

---

## Phase 2: Enhanced Features (1-2 weeks)

**Goal**: Add professional-grade features and complete class support.

### 2.1 Manual Violation Recording Interface

**Priority**: 🚨 HIGH - Professional event integration  
**Estimated Time**: 2-3 days  
**NHRA Rule**: 6.1.2

#### Implementation Steps:

1. **Create Officials Interface Component**
   ```go
   // File: pkg/officials/violations.go (new package)
   type ViolationRecorder struct {
       RaceID     string
       Violations []OfficialViolation
       EventBus   *events.EventBus
   }
   
   type OfficialViolation struct {
       Type        ViolationType // centerline, unsporting, equipment
       Lane        int
       Official    string        // name of calling official
       Timestamp   time.Time
       Description string
   }
   ```

2. **Add Violation Recording API**
   ```go
   // File: pkg/api/api.go - Add violation recording endpoints
   func (a *API) RecordOfficialViolation(raceID string, violation OfficialViolation) error {
       // Record manual violation detected by track officials
       // Update race results accordingly
   }
   ```

3. **Integrate with Event System**
   ```go
   // File: pkg/events/events.go - Add new event types
   const (
       EventOfficialViolation   EventType = "official_violation"
       EventCenterlineViolation EventType = "centerline_violation" // manual detection
   )
   ```

---

### 2.2 Complete NHRA/IHRA Class Support

**Priority**: 🎯 MEDIUM - Feature completeness  
**Estimated Time**: 2-3 days

#### Classes to Add:
- **NHRA Professional**: Funny Car, Pro Stock, Pro Stock Motorcycle, Pro Modified
- **NHRA Sportsman**: Super Comp, Comp Eliminator, Stock Eliminator, Super Street
- **IHRA Classes**: Pro Modified (different specs), Top Sportsman, Junior Dragster
- **Regional Variations**: Track-specific rule modifications

---

### 2.3 Professional Precision Timing

**Priority**: 🎯 MEDIUM - Professional readiness  
**Estimated Time**: 2-3 days

#### Implementation Steps:

1. **Upgrade Timing Precision**
   ```go
   // File: pkg/timing/timing.go
   type TimingSystem struct {
       precision time.Duration // Change from millisecond to microsecond
   }
   ```

2. **Add Redundant Timing Support**
   ```go
   // File: pkg/timing/redundant.go (new file)
   type RedundantTimingSystem struct {
       Primary   *TimingSystem
       Secondary *TimingSystem
       Validator func(primary, secondary *TimingResults) bool
   }
   ```

---

## Phase 3: Professional Features (2-3 weeks)

**Goal**: Production-ready system for professional racing events.

### 3.1 Race Director Control Panel

**Priority**: 🏁 LOW - Professional feature  
**Estimated Time**: 1 week

#### Features:
- Real-time race control interface
- Manual override capabilities  
- Live configuration changes
- Emergency stop controls
- Race replay and analysis

---

### 3.2 Data Integrity & Audit Systems

**Priority**: 🏁 LOW - Professional feature  
**Estimated Time**: 1 week

#### Features:
- Cryptographic signing of results
- Complete audit trail logging
- Tamper detection
- Integration with sanctioning body systems

---

### 3.3 Hardware Integration

**Priority**: 🏁 LOW - Professional feature  
**Estimated Time**: 1 week

#### Features:
- CompuLink system compatibility
- Professional timing hardware support
- Track infrastructure integration
- Sensor validation and calibration

---

## Implementation Guidelines

### Development Workflow

1. **Feature Branch Strategy**
   ```bash
   git checkout -b feature/deep-staging-enforcement
   # Implement feature
   git commit -m "feat: add deep staging enforcement for NHRA classes"
   git push origin feature/deep-staging-enforcement
   # Create PR for review
   ```

2. **Testing Requirements**
   - Minimum 80% test coverage for new features
   - Integration tests for rule compliance
   - Performance tests for timing precision
   - Safety scenario testing

3. **Documentation Updates**
   - Update compliance status in `docs/nhra-ihra-compliance.md`
   - Add API documentation for new features
   - Update examples with new class configurations

### Code Quality Standards

1. **Error Handling**
   - All errors must be properly wrapped with context
   - Use structured logging instead of print statements
   - Implement graceful degradation for non-critical failures

2. **Concurrency Safety**
   - All shared state must be protected with mutexes
   - Use context for goroutine lifecycle management
   - Implement proper cleanup for resources

3. **Configuration Management**
   - All timing constants must be configurable
   - Support for hot-reloading of non-safety-critical configs
   - Validation of all configuration values on startup

---

## Success Metrics

### Phase 1 Completion Criteria:
- [x] Deep staging enforcement for Super Gas/Stock/Street
- [x] Forward motion staging rule enforcement
- [ ] Guard beam rollout distance enforcement operational
- [ ] Class-specific staging timeouts implemented
- [ ] Test coverage >60% for all core packages
- [ ] All print statements replaced with structured logging

### Phase 2 Completion Criteria:
- [ ] Manual violation recording interface operational
- [ ] All 15+ NHRA classes configured and tested
- [ ] Professional timing precision achieved
- [ ] Integration tests passing for all compliance scenarios

### Phase 3 Completion Criteria:
- [ ] Race director control panel functional
- [ ] Data integrity systems operational  
- [ ] Hardware integration tested
- [ ] Full NHRA/IHRA certification ready

---

**Document Created**: 2025-07-16  
**Last Updated**: 2025-07-16  
**Version**: 1.0  
**Next Review**: After Phase 1 completion