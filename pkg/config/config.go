package config

import "time"

// Config holds system-wide configuration
type Config interface {
	Track() TrackConfig
	Timing() TimingConfig
	Tree() TreeSequenceConfig
	Safety() SafetyConfig
	RacingClass() string // Added for class selection
}

// TrackConfig defines track specifications
type TrackConfig struct {
	Length     float64               `json:"length"`      // Track length in feet
	LaneCount  int                   `json:"lane_count"`  // Number of lanes
	LaneWidth  float64               `json:"lane_width"`  // Width of each lane
	BeamLayout map[string]BeamConfig `json:"beam_layout"` // Beam positions
}

// BeamConfig defines timing beam specifications
type BeamConfig struct {
	Name     string  `json:"name"`
	Position float64 `json:"position"` // Distance from starting line
	Height   float64 `json:"height"`   // Height above track
	Lane     int     `json:"lane"`     // Which lane (0 = both)
}

// TimingConfig defines timing system parameters
type TimingConfig struct {
	Precision       time.Duration `json:"precision"`         // Timing precision
	SpeedTrapLength float64       `json:"speed_trap_length"` // Speed trap distance
	AutoStart       bool          `json:"auto_start"`        // Auto-start timing on stage
}

// TreeSequenceType defines different starting sequences
type TreeSequenceType string

const (
	TreeSequencePro       TreeSequenceType = "pro"       // All ambers simultaneously
	TreeSequenceSportsman TreeSequenceType = "sportsman" // Sequential ambers
)

// TreeSequenceConfig defines timing for tree sequences
type TreeSequenceConfig struct {
	Type            TreeSequenceType `json:"type"`
	AmberDelay      time.Duration    `json:"amber_delay"` // Time between ambers (sportsman)
	GreenDelay      time.Duration    `json:"green_delay"` // Time from last amber to green
	PreStageTimeout time.Duration    `json:"pre_stage_timeout"`
	StageTimeout    time.Duration    `json:"stage_timeout"`
}

// SafetyConfig defines safety system parameters
type SafetyConfig struct {
	EmergencyStopEnabled bool          `json:"emergency_stop_enabled"`
	MaxReactionTime      time.Duration `json:"max_reaction_time"`
	MinStagingTime       time.Duration `json:"min_staging_time"`
}

// DefaultConfig implements Config interface
type DefaultConfig struct {
	TrackConfig  TrackConfig        `json:"track"`
	TimingConfig TimingConfig       `json:"timing"`
	TreeConfig   TreeSequenceConfig `json:"tree"`
	SafetyConfig SafetyConfig       `json:"safety"`
	racingClass  string             `json:"racing_class"` // Private field
}

func (c *DefaultConfig) Track() TrackConfig {
	return c.TrackConfig
}

func (c *DefaultConfig) Timing() TimingConfig {
	return c.TimingConfig
}

func (c *DefaultConfig) Tree() TreeSequenceConfig {
	return c.TreeConfig
}

func (c *DefaultConfig) Safety() SafetyConfig {
	return c.SafetyConfig
}

func (c *DefaultConfig) RacingClass() string {
	return c.racingClass
}

// NewDefaultConfig creates a default configuration for NHRA-style drag racing
func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		TrackConfig: TrackConfig{
			Length:    1320, // Quarter mile in feet
			LaneCount: 2,
			LaneWidth: 12, // 12 feet per lane
			BeamLayout: map[string]BeamConfig{
				"pre_stage": {
					Name:     "Pre-Stage",
					Position: -7, // 7 feet before starting line
					Height:   8,  // 8 inches above track
					Lane:     0,  // Both lanes
				},
				"stage": {
					Name:     "Stage",
					Position: 0, // Starting line
					Height:   8, // 8 inches above track
					Lane:     0, // Both lanes
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
				"1000_foot": {
					Name:     "1000 Foot",
					Position: 1000,
					Height:   8,
					Lane:     0,
				},
				"1320_foot": {
					Name:     "1320 Foot (Quarter Mile)",
					Position: 1320,
					Height:   8,
					Lane:     0,
				},
			},
		},
		TimingConfig: TimingConfig{
			Precision:       time.Microsecond,
			SpeedTrapLength: 66, // 66 feet for speed trap calculation
			AutoStart:       true,
		},
		TreeConfig: TreeSequenceConfig{
			Type:            TreeSequencePro,        // Default to Pro tree
			AmberDelay:      500 * time.Millisecond, // 0.5 seconds for sportsman
			GreenDelay:      400 * time.Millisecond, // 0.4 seconds for pro tree
			PreStageTimeout: 30 * time.Second,
			StageTimeout:    10 * time.Second,
		},
		SafetyConfig: SafetyConfig{
			EmergencyStopEnabled: true,
			MaxReactionTime:      2 * time.Second,
			MinStagingTime:       500 * time.Millisecond,
		},
		racingClass: "Sportsman", // Default class
	}
}

// SetRacingClass sets the racing class
func (c *DefaultConfig) SetRacingClass(class string) {
	c.racingClass = class
}
