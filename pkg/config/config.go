package config

import "time"

// Config holds system-wide configuration
type Config interface {
	GetTrackConfig() TrackConfig
	GetTimingConfig() TimingConfig
	GetTreeConfig() TreeSequenceConfig
	GetSafetyConfig() SafetyConfig
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
	track  TrackConfig
	timing TimingConfig
	tree   TreeSequenceConfig
	safety SafetyConfig
}

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		track: TrackConfig{
			Length:    1320, // Quarter mile
			LaneCount: 2,
			LaneWidth: 12,
			BeamLayout: map[string]BeamConfig{
				"pre_stage_L":    {Name: "Pre-Stage Left", Position: -0.583, Lane: 1},
				"stage_L":        {Name: "Stage Left", Position: 0, Lane: 1},
				"sixty_foot_L":   {Name: "60-Foot Left", Position: 60, Lane: 1},
				"eighth_mile_L":  {Name: "1/8 Mile Left", Position: 660, Lane: 1},
				"quarter_mile_L": {Name: "1/4 Mile Left", Position: 1320, Lane: 1},
				"pre_stage_R":    {Name: "Pre-Stage Right", Position: -0.583, Lane: 2},
				"stage_R":        {Name: "Stage Right", Position: 0, Lane: 2},
				"sixty_foot_R":   {Name: "60-Foot Right", Position: 60, Lane: 2},
				"eighth_mile_R":  {Name: "1/8 Mile Right", Position: 660, Lane: 2},
				"quarter_mile_R": {Name: "1/4 Mile Right", Position: 1320, Lane: 2},
			},
		},
		timing: TimingConfig{
			Precision:       time.Millisecond,
			SpeedTrapLength: 66,
			AutoStart:       true,
		},
		tree: TreeSequenceConfig{
			Type:            TreeSequencePro,
			AmberDelay:      500 * time.Millisecond,
			GreenDelay:      400 * time.Millisecond,
			PreStageTimeout: 30 * time.Second,
			StageTimeout:    10 * time.Second,
		},
		safety: SafetyConfig{
			EmergencyStopEnabled: true,
			MaxReactionTime:      2 * time.Second,
			MinStagingTime:       100 * time.Millisecond,
		},
	}
}

func (dc *DefaultConfig) GetTrackConfig() TrackConfig       { return dc.track }
func (dc *DefaultConfig) GetTimingConfig() TimingConfig     { return dc.timing }
func (dc *DefaultConfig) GetTreeConfig() TreeSequenceConfig { return dc.tree }
func (dc *DefaultConfig) GetSafetyConfig() SafetyConfig     { return dc.safety }
