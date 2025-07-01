# libdrag API Configuration Guide

## Configuration Overview

The libdrag API uses a flexible configuration system that supports both default settings and custom configurations for different racing scenarios.

## Default Configuration

When you call `Initialize()`, the API creates a default configuration suitable for NHRA-style quarter-mile drag racing:

```go
dragAPI := api.NewLibDragAPI()
err := dragAPI.Initialize() // Uses default config
```

### Default Settings

- **Track Length**: 1320 feet (quarter mile)
- **Lane Count**: 2
- **Lane Width**: 12 feet
- **Tree Sequence**: Professional (Pro Tree)
- **Timing Precision**: Microsecond
- **Auto-Start**: Enabled
- **Max Concurrent Races**: 10

## Custom Configuration

### Track Configuration

```go
import "github.com/benharold/libdrag/pkg/config"

// Create custom track config
trackConfig := config.TrackConfig{
    Length:    660,  // Eighth mile
    LaneCount: 2,
    LaneWidth: 12,
    BeamLayout: map[string]config.BeamConfig{
        "pre_stage": {
            Name:     "Pre-Stage",
            Position: -7,
            Height:   8,
            Lane:     0, // Both lanes
        },
        "stage": {
            Name:     "Stage", 
            Position: 0,
            Height:   8,
            Lane:     0,
        },
        "60_foot": {
            Name:     "60 Foot",
            Position: 60,
            Height:   8,
            Lane:     0,
        },
        "330_foot": {
            Name:     "330 Foot",
            Position: 330,
            Height:   8,
            Lane:     0,
        },
        "660_foot": {
            Name:     "660 Foot (Eighth Mile)",
            Position: 660,
            Height:   8,
            Lane:     0,
        },
    },
}
```

### Christmas Tree Configuration

```go
// Professional Tree (All ambers simultaneously)
treeConfig := config.TreeSequenceConfig{
    Type:            config.TreeSequencePro,
    AmberDelay:      0 * time.Millisecond, // No delay for pro
    GreenDelay:      400 * time.Millisecond, // 0.4 seconds
    PreStageTimeout: 30 * time.Second,
    StageTimeout:    10 * time.Second,
}

// Sportsman Tree (Sequential ambers)
sportsmanConfig := config.TreeSequenceConfig{
    Type:            config.TreeSequenceSportsman,
    AmberDelay:      500 * time.Millisecond, // 0.5 seconds between ambers
    GreenDelay:      400 * time.Millisecond,
    PreStageTimeout: 45 * time.Second,
    StageTimeout:    15 * time.Second,
}
```

### Timing System Configuration

```go
timingConfig := config.TimingConfig{
    Precision:       time.Microsecond,
    SpeedTrapLength: 66, // 66 feet for speed calculation
    AutoStart:       true,
}
```

### Safety Configuration

```go
safetyConfig := config.SafetyConfig{
    EmergencyStopEnabled: true,
    MaxReactionTime:      2 * time.Second,
    MinStagingTime:       500 * time.Millisecond,
}
```

## Racing Class Configurations

### Professional Classes

#### Top Fuel / Funny Car
```go
config := &config.DefaultConfig{
    TrackConfig: config.TrackConfig{
        Length:    1320, // Quarter mile
        LaneCount: 2,
    },
    TreeConfig: config.TreeSequenceConfig{
        Type:       config.TreeSequencePro,
        GreenDelay: 400 * time.Millisecond,
    },
    TimingConfig: config.TimingConfig{
        Precision: time.Microsecond,
        AutoStart: true,
    },
}
```

#### Pro Stock
```go
config := &config.DefaultConfig{
    TrackConfig: config.TrackConfig{
        Length:    1320,
        LaneCount: 2,
    },
    TreeConfig: config.TreeSequenceConfig{
        Type:            config.TreeSequencePro,
        GreenDelay:      400 * time.Millisecond,
        StageTimeout:    7 * time.Second, // Shorter for pros
    },
}
```

### Sportsman Classes

#### Bracket Racing
```go
config := &config.DefaultConfig{
    TreeConfig: config.TreeSequenceConfig{
        Type:            config.TreeSequenceSportsman,
        AmberDelay:      500 * time.Millisecond,
        GreenDelay:      400 * time.Millisecond,
        StageTimeout:    15 * time.Second,
    },
}
```

#### Junior Dragster
```go
config := &config.DefaultConfig{
    TrackConfig: config.TrackConfig{
        Length:    660, // Eighth mile for juniors
        LaneCount: 2,
    },
    TreeConfig: config.TreeSequenceConfig{
        Type:            config.TreeSequenceSportsman,
        AmberDelay:      500 * time.Millisecond,
        PreStageTimeout: 60 * time.Second, // More time for young drivers
        StageTimeout:    20 * time.Second,
    },
}
```

## Auto-Start Configuration

The libdrag system includes a CompuLink-compatible auto-start system:

### Professional Auto-Start
```go
autoStartConfig := autostart.AutoStartConfig{
    StagingTimeout:     7 * time.Second,    // Quick staging for pros
    MinStagingDuration: 500 * time.Millisecond,
    RandomDelayMin:     600 * time.Millisecond,
    RandomDelayMax:     1100 * time.Millisecond,
    MaxRolloutDistance: 6.0, // Inches
    EnabledForElims:    true,
    TreeSequenceType:   config.TreeSequencePro,
    RacingClass:        "Professional",
}
```

### Sportsman Auto-Start
```go
autoStartConfig := autostart.AutoStartConfig{
    StagingTimeout:     15 * time.Second,   // More time for sportsman
    MinStagingDuration: 600 * time.Millisecond,
    RandomDelayMin:     600 * time.Millisecond,
    RandomDelayMax:     1400 * time.Millisecond,
    EnabledForElims:    true,
    EnabledForTimeTrials: true,
    TreeSequenceType:   config.TreeSequenceSportsman,
    RacingClass:        "Sportsman",
}
```

## Performance Tuning

### Concurrent Race Management
```go
// Adjust based on server capacity
dragAPI.SetMaxConcurrentRaces(20) // Higher for powerful servers
dragAPI.SetMaxConcurrentRaces(5)  // Lower for limited resources
```

### Memory Optimization
```go
// Custom configuration for memory-constrained environments
config := &config.DefaultConfig{
    TimingConfig: config.TimingConfig{
        Precision: time.Millisecond, // Lower precision = less memory
    },
}
```

### High-Frequency Racing
```go
// Configuration for high-volume racing events
config := &config.DefaultConfig{
    TreeConfig: config.TreeSequenceConfig{
        PreStageTimeout: 15 * time.Second, // Faster staging
        StageTimeout:    5 * time.Second,
    },
}
```

## Environment-Specific Configurations

### Development Environment
```go
config := &config.DefaultConfig{
    TreeConfig: config.TreeSequenceConfig{
        Type:            config.TreeSequencePro,
        GreenDelay:      100 * time.Millisecond, // Faster for testing
        PreStageTimeout: 5 * time.Second,
        StageTimeout:    3 * time.Second,
    },
}
```

### Production Racing
```go
config := &config.DefaultConfig{
    TreeConfig: config.TreeSequenceConfig{
        Type:            config.TreeSequencePro,
        GreenDelay:      400 * time.Millisecond, // Official timing
        PreStageTimeout: 30 * time.Second,
        StageTimeout:    10 * time.Second,
    },
    SafetyConfig: config.SafetyConfig{
        EmergencyStopEnabled: true,
        MaxReactionTime:      2 * time.Second,
    },
}
```

## Integration with External Systems

### Track Management System Integration
```go
// Custom beam layout for existing track hardware
customBeamLayout := map[string]config.BeamConfig{
    "start_line": {
        Name:     "Starting Line",
        Position: 0,
        Height:   8,
        Lane:     0,
    },
    "finish_line": {
        Name:     "Finish Line", 
        Position: 1320,
        Height:   8,
        Lane:     0,
    },
    // Add more beams as needed for existing hardware
}
```

### Timing System Integration
```go
timingConfig := config.TimingConfig{
    Precision:       time.Nanosecond, // Ultra-high precision
    SpeedTrapLength: 132, // Custom speed trap distance
    AutoStart:       false, // Manual control for some events
}
```

## Configuration Best Practices

1. **Start with Defaults**: Use `config.NewDefaultConfig()` as a base
2. **Class-Specific**: Create configurations for each racing class
3. **Environment Separation**: Different configs for dev/test/production
4. **Safety First**: Always enable safety features in production
5. **Performance Testing**: Test concurrent race limits under load
6. **Documentation**: Document custom configurations for your track

## Validation

The configuration system includes built-in validation:

```go
config := config.NewDefaultConfig()

// Validation happens automatically during initialization
err := dragAPI.InitializeWithConfig(config)
if err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

Common validation errors:
- Track length must be positive
- Lane count must be 1-8
- Timing precision must be valid duration
- Beam positions must be sequential
- Tree delays must be reasonable (0-2000ms)
