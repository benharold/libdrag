# Drag Racing Staging Bulb States Matrix

## Matrix Overview

This 4x4 matrix shows all 16 possible combinations of pre-stage and stage bulb states for both lanes in a drag racing setup. Each cell represents a unique staging scenario that can occur during the staging process.

**Lane 1 (Left Lane):** Rows represent different combinations of Lane 1's pre-stage and stage bulbs  
**Lane 2 (Right Lane):** Columns represent different combinations of Lane 2's pre-stage and stage bulbs

## Staging States Matrix

| Lane 1 \ Lane 2                            | Pre: OFF<br>Stage: OFF | Pre: ON<br>Stage: OFF | Pre: ON<br>Stage: ON | Pre: OFF<br>Stage: ON<br>(Deep Staged) |
|--------------------------------------------|------------------------|-----------------------|----------------------|----------------------------------------|
| **Pre: OFF<br>Stage: OFF**                 | 0,0 - 0,0              | 0,0 - 1,0             | 0,0 - 1,1            | 0,0 - 0,1                              |
| **Pre: ON<br>Stage: OFF**                  | 1,0 - 0,0              | **1,0 - 1,0** ⚡       | **1,0 - 1,1** ⚡      | 1,0 - 0,1                              |
| **Pre: ON<br>Stage: ON**                   | 1,1 - 0,0              | **1,1 - 1,0** ⚡       | **1,1 - 1,1** ⚡      | **1,1 - 0,1** ⚡                        |
| **Pre: OFF<br>Stage: ON<br>(Deep Staged)** | 0,1 - 0,0              | 0,1 - 1,0             | **0,1 - 1,1** ⚡      | 0,1 - 0,1                              |

## Legend

- **⚡ AutoStart Conditions Met** - Three or more total bulbs are lit
- **Normal text** - AutoStart conditions NOT met
- **Notation:** Each cell shows "L1_Pre,L1_Stage - L2_Pre,L2_Stage" where 1 = ON and 0 = OFF

## Key Staging Scenarios

### Basic Staging States
- **Both Cars Not Staged (0,0 - 0,0):** No cars in staging area
- **Both Cars Pre-Staged (1,0 - 1,0):** Both cars approximately 7 inches from starting line
- **Both Cars Fully Staged (1,1 - 1,1):** Both cars ready to race, tree will activate
- **One Car Staged, One Pre-Staged (1,1 - 1,0):** AutoStart timeout begins for Lane 2

### Deep Staging Scenarios
- **Deep Staged Scenarios (0,1 - X,X):** Pre-stage bulb turned off by rolling forward past normal staging position
- **Both Deep Staged (0,1 - 0,1):** Rare scenario where both cars are deep staged

## AutoStart System Behavior

**⚡ Marked cells** indicate conditions where the AutoStart system's timeout mechanism would be active or the tree would start automatically:

### AutoStart Activation Conditions
1. **Three Total Bulbs Lit:** Two pre-stage bulbs + one stage bulb from either lane
2. **Four Total Bulbs Lit:** Both cars fully staged (race ready)
3. **Mixed Staging:** One car staged, one car pre-staged

### Timeout Scenarios
When AutoStart conditions are met (⚡ cells), the system behavior depends on the specific combination:

- **1,0 - 1,0:** Both pre-staged - no timeout, waiting for one car to stage
- **1,0 - 1,1:** Lane 1 pre-staged, Lane 2 staged - Lane 1 has limited time to stage
- **1,1 - 1,0:** Lane 1 staged, Lane 2 pre-staged - Lane 2 has limited time to stage
- **1,1 - 1,1:** Both fully staged - tree starts automatically after brief delay
- **1,1 - 0,1:** Lane 1 staged, Lane 2 deep staged - tree starts automatically
- **0,1 - 1,1:** Lane 1 deep staged, Lane 2 staged - tree starts automatically

## Practical Applications

### For Drivers
- Understanding which combinations trigger timeout conditions
- Knowing when you have limited time to complete staging
- Recognizing when the tree will start automatically

### For Starters
- Monitoring which bulb combinations require AutoStart activation
- Understanding when manual override might be needed for deep staging
- Recognizing problem scenarios that require intervention

### For Track Operations
- Configuring AutoStart timeout values based on class requirements
- Understanding when to honor deep staging requests
- Troubleshooting staging-related issues

## Technical Notes

- The matrix assumes standard beam placement (7 inches between pre-stage and stage)
- Deep staging scenarios show pre-stage OFF because the car has rolled past the pre-stage beam
- AutoStart timeout periods vary by class and track configuration (typically 7-15 seconds)
- Some classes prohibit deep staging entirely (Super Gas, Super Stock, Super Street)