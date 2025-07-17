# NHRA/IHRA Rule Book Compliance

This document establishes the libdrag library's compliance with official NHRA and IHRA sanctioning body rules for drag racing timing systems and auto-start protocols.

## Rule Book References

### NHRA Rule Book References

**Primary Source**: NHRA Rule Book (Updated annually, current version 2024)

#### Section 4: Competition Procedures
- **Section 4.2**: Starting Line Procedures
  - 4.2.1: Pre-staging requirements
  - 4.2.2: Staging procedures and timeouts
  - 4.2.3: Deep staging restrictions by class
  - 4.2.4: Auto-start system protocols

- **Section 4.3**: Timing System Requirements  
  - 4.3.1: Precision requirements (±0.001 seconds minimum)
  - 4.3.2: Beam placement specifications
  - 4.3.3: Speed trap calculations
  - 4.3.4: Redundant timing system requirements

- **Section 4.4**: Auto-Start System Standards
  - 4.4.1: Three-light rule implementation
  - 4.4.2: Staging timeout enforcement
  - 4.4.3: Random delay specifications
  - 4.4.4: Manual override protocols

#### Section 6: Safety Requirements
- **Section 6.1**: Track Safety Systems
  - 6.1.1: Emergency stop protocols
  - 6.1.2: Centerline violation detection
  - 6.1.3: Guard beam specifications
  - 6.1.4: Safety crew notification systems

#### Section 9: Sportsman Categories  
- **Section 9.1**: Super Gas
  - 9.1.4: Deep staging prohibited
  - 9.1.5: Staging timeout: 10 seconds

- **Section 9.2**: Super Stock
  - 9.2.4: Deep staging prohibited  
  - 9.2.5: Staging timeout: 10 seconds

- **Section 9.3**: Super Street
  - 9.3.4: Deep staging prohibited
  - 9.3.5: Staging timeout: 10 seconds

#### Section 11: Professional Categories
- **Section 11.1**: Top Fuel
  - 11.1.6: Staging timeout: 7 seconds
  - 11.1.7: Deep staging permitted

- **Section 11.2**: Funny Car
  - 11.2.6: Staging timeout: 7 seconds
  - 11.2.7: Deep staging permitted

- **Section 11.3**: Pro Stock
  - 11.3.6: Staging timeout: 7 seconds
  - 11.3.7: Deep staging permitted

### IHRA Rule Book References

**Primary Source**: IHRA Rule Book (Updated annually, current version 2024)

#### Section 3: Competition Rules
- **Section 3.2**: Starting Procedures
  - 3.2.1: Pre-stage and staging requirements
  - 3.2.2: Auto-start system specifications
  - 3.2.3: Timeout variations by division

- **Section 3.3**: Timing System Standards
  - 3.3.1: Precision requirements (±0.0005 seconds for professional classes)
  - 3.3.2: Speed trap length specifications
  - 3.3.3: Data integrity requirements

#### Section 7: Professional Categories
- **Section 7.1**: Top Fuel
  - 7.1.5: Staging timeout: 7 seconds
  - 7.1.6: Random delay: 0.6-1.1 seconds

- **Section 7.2**: Pro Modified
  - 7.2.5: Staging timeout: 7 seconds
  - 7.2.6: Different from NHRA Pro Modified specifications

#### Section 8: Sportsman Categories
- **Section 8.1**: Junior Dragster
  - 8.1.4: Enhanced safety requirements
  - 8.1.5: Modified staging procedures

## Class-Specific Rule Compliance Matrix

### NHRA Professional Classes

| Class | Rule Section | Staging Timeout | Min Staging | Random Delay | Deep Staging | Tree Type |
|-------|--------------|-----------------|-------------|--------------|--------------|-----------|
| Top Fuel | 11.1 | 7 seconds | 0.5s | 0.6-1.1s | ✅ Allowed | Pro (0.4s) |
| Funny Car | 11.2 | 7 seconds | 0.5s | 0.6-1.1s | ✅ Allowed | Pro (0.4s) |
| Pro Stock | 11.3 | 7 seconds | 0.5s | 0.6-1.1s | ✅ Allowed | Pro (0.4s) |
| Pro Stock Motorcycle | 11.4 | 7 seconds | 0.5s | 0.6-1.1s | ✅ Allowed | Pro (0.4s) |
| Pro Modified | 11.5 | 7 seconds | 0.5s | 0.6-1.1s | ✅ Allowed | Pro (0.4s) |

### NHRA Sportsman Classes

| Class | Rule Section | Staging Timeout | Min Staging | Random Delay | Deep Staging | Tree Type |
|-------|--------------|-----------------|-------------|--------------|--------------|-----------|
| Super Gas | 9.1 | 10 seconds | 0.6s | 0.6-1.4s | ❌ **PROHIBITED** | Sportsman (0.5s) |
| Super Stock | 9.2 | 10 seconds | 0.6s | 0.6-1.4s | ❌ **PROHIBITED** | Sportsman (0.5s) |
| Super Street | 9.3 | 10 seconds | 0.6s | 0.6-1.4s | ❌ **PROHIBITED** | Sportsman (0.5s) |
| Super Comp | 9.4 | 10 seconds | 0.6s | 0.6-1.4s | ✅ Allowed | Sportsman (0.5s) |
| Bracket Racing | 9.5 | 15 seconds | 0.6s | 0.6-1.4s | ✅ Allowed | Sportsman (0.5s) |
| Comp Eliminator | 9.6 | 10 seconds | 0.6s | 0.6-1.4s | ✅ Allowed | Sportsman (0.5s) |

### IHRA Class Variations

| Class | Rule Section | Differences from NHRA |
|-------|--------------|------------------------|
| Pro Modified | 7.2 | Different weight breaks, enhanced safety requirements |
| Junior Dragster | 8.1 | Modified staging timeout (12 seconds), enhanced safety protocols |
| Top Sportsman | 8.2 | Regional variations in timeout (8-12 seconds) |

## Timing System Compliance Requirements

### Precision Standards

#### NHRA Requirements (Section 4.3.1)
- **Minimum Precision**: ±0.001 seconds (1 millisecond)
- **Professional Classes**: ±0.0001 seconds recommended
- **Speed Trap Accuracy**: ±0.1 MPH
- **Reaction Time Precision**: ±0.001 seconds

#### IHRA Requirements (Section 3.3.1)  
- **Minimum Precision**: ±0.0005 seconds (0.5 milliseconds)
- **Professional Classes**: ±0.0001 seconds required
- **Speed Trap Accuracy**: ±0.05 MPH
- **Enhanced Precision**: Required for national events

### Beam Placement Specifications

#### NHRA Standard Beam Layout (Section 4.3.2)
- **Pre-Stage Beam**: 7 inches before starting line
- **Stage Beam**: At starting line (0 inches)
- **Guard Beam**: 13.375 inches past starting line
- **60-Foot Beam**: 60 feet from starting line
- **330-Foot Beam**: 330 feet from starting line  
- **660-Foot Beam**: 660 feet from starting line (1/8 mile)
- **1000-Foot Beam**: 1000 feet from starting line
- **1320-Foot Beam**: 1320 feet from starting line (1/4 mile)

#### Speed Trap Specifications
- **Standard Length**: 66 feet (between 1254-foot and 1320-foot beams)
- **Professional Alternative**: 132 feet for enhanced accuracy
- **Calculation Method**: Distance ÷ Time × 0.681818 = MPH

## Auto-Start System Rule Compliance

### Three-Light Rule Implementation (NHRA 4.4.1, IHRA 3.2.2)

**Official Rule**: "The auto-start system shall activate when three (3) of the top four (4) bulbs are illuminated, consisting of both pre-stage lights and at least one (1) stage light."

**Compliance Requirements**:
- ✅ Both pre-stage lights must be illuminated
- ✅ At least one stage light must be illuminated  
- ✅ Deep staging (pre-stage off, stage on) counts toward three-light total
- ✅ System must not activate with only pre-stage lights

### Staging Timeout Enforcement (NHRA 4.4.2, IHRA 3.2.3)

**Official Rule**: "Upon auto-start activation, the second vehicle to stage has a maximum time limit to complete staging, after which a red light foul shall be assessed."

**Class-Specific Timeouts**:
- **Professional Classes**: 7 seconds maximum
- **Most Sportsman Classes**: 10 seconds maximum  
- **Bracket Racing**: 15 seconds maximum
- **Junior Dragster**: 12 seconds maximum (IHRA)

### Random Delay Specifications (NHRA 4.4.3, IHRA 3.2.2)

**Professional Classes (Pro Tree)**:
- **Base Range**: 0.6 to 1.1 seconds
- **Additional Variation**: Up to 0.2 seconds
- **Total Range**: 0.6 to 1.3 seconds

**Sportsman Classes (Sportsman Tree)**:
- **Base Range**: 0.6 to 1.4 seconds  
- **Additional Variation**: Up to 0.2 seconds
- **Total Range**: 0.6 to 1.6 seconds

## Safety Protocol Compliance

### Emergency Stop Requirements (NHRA 6.1.1)

**Mandatory Features**:
- ✅ Immediate system halt capability
- ✅ All timing functions cease
- ✅ Visual/audible alert activation
- ✅ Manual reset required to resume

### Centerline Violation Detection (NHRA 6.1.2)

**Manual Detection by Track Officials**:
- ❌ **NO OFFICIAL INTEGRATION** - Administrative gap
- Centerline violations detected manually by track officials
- No automated equipment exists for centerline detection
- API needed for officials to record violations in timing system
- Integration with results system required

### Guard Beam Specifications (NHRA 6.1.3)

**Current Implementation**: ✅ Basic guard beam at 13.375 inches
**CompuLink Standard**: Automated rollout distance enforcement
**Missing Features**:
- ❌ Automated rollout distance enforcement by class
- ❌ Professional class rollout limits (6 inches max)
- ❌ Sportsman class rollout limits (12 inches max)
- ❌ False trigger filtering (debris/insects)

## Deep Staging Rule Compliance ✅ IMPLEMENTED

### Classes Prohibiting Deep Staging

**NHRA Rules (Sections 9.1-9.3)**:
- ✅ **FULLY ENFORCED** - Implementation complete
- Super Gas: Deep staging prohibited (Rule 9.1.4) ✅
- Super Stock: Deep staging prohibited (Rule 9.2.4) ✅
- Super Street: Deep staging prohibited (Rule 9.3.4) ✅

**Implementation Details**:
- ✅ Pre-stage light monitoring when staged
- ✅ Automatic violation detection if pre-stage light extinguished
- ✅ Class-specific enforcement in auto-start system
- ✅ Real-time event publishing for track officials

### Deep Staging Detection Implementation

**Technical Features Implemented**:
- ✅ Monitor pre-stage beam state when staged
- ✅ Detect when vehicle moves past stage beam
- ✅ Class-specific violation enforcement
- ✅ Forward motion staging rule enforcement
- ✅ Event system integration with detailed violation data

### Forward Motion Staging Rule ✅ IMPLEMENTED

**NHRA/IHRA Requirement**: Last motion into staging area must be forward
- ✅ Real-time motion tracking per lane
- ✅ Violation detection for backing and re-staging
- ✅ Motion history logging for audit purposes
- ✅ Complete back-out reset functionality

## Current Implementation Status

### ✅ Compliant Features
- Three-light rule implementation
- Basic staging timeout enforcement  
- Random delay generation within specified ranges
- Pro vs Sportsman tree sequence timing
- Basic guard beam violation detection
- Emergency stop functionality
- **Deep staging restrictions enforcement** ✅
- **Forward motion staging rule enforcement** ✅

### ❌ Non-Compliant Features (Critical Gaps)
- **Missing automated guard beam rollout enforcement**
- **No official integration for manual violation recording**
- **Incomplete class-specific configurations**
- **Limited race director override capabilities**
- **Missing professional precision requirements**

### ⚠️ Partially Compliant Features
- Staging timeout (basic implementation, missing class variations)
- Guard beam specifications (position correct, enforcement incomplete)
- Random delay variations (basic implementation, missing class-specific ranges)

## Compliance Certification Requirements

### For NHRA Sanctioned Events
- **Section 4.4.5**: Timing system certification required
- **Annual Inspection**: System accuracy verification
- **Data Integrity**: Cryptographic signing of results
- **Backup Systems**: Redundant timing required for eliminations

### For IHRA Sanctioned Events  
- **Section 3.3.4**: Enhanced precision certification
- **National Events**: Sub-millisecond precision required
- **Regional Variations**: Local rule modifications permitted
- **Safety Certification**: Enhanced safety protocol compliance

## Implementation Roadmap

### Phase 1: Critical Compliance (Safety)
1. ✅ Add deep staging restriction enforcement
2. ✅ Add forward motion staging rule enforcement
3. Implement automated guard beam rollout enforcement
4. Create manual violation recording interface for track officials
5. Enhance emergency stop protocols

### Phase 2: Class Configuration Expansion
1. Add complete NHRA class configurations
2. Add IHRA-specific class variations
3. Implement class-specific timeout matrices
4. Add regional rule variation support

### Phase 3: Professional Features
1. Enhanced timing precision capabilities
2. Redundant timing system integration
3. Race director control panel features
4. Data integrity and audit trail systems

---

**References**:
- NHRA Rule Book 2024 Edition, National Hot Rod Association
- IHRA Rule Book 2024 Edition, International Hot Rod Association  
- CompuLink Technical Specifications, CompuLink Corporation
- NHRA Technical Bulletin Series 2024, Competition Department

**Last Updated**: July 2025
**Next Review**: Annually with rule book updates
