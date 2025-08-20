package flashcard

import (
	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Config holds FSRS configuration parameters.
type Config struct {
	// FSRS algorithm parameters
	RequestRetention float64 `json:"request_retention" validate:"min=0.7,max=0.99"`
	MaximumInterval  float64 `json:"maximum_interval" validate:"min=1,max=36500"`
	EnableShortTerm  bool    `json:"enable_short_term"`
	EnableFuzz       bool    `json:"enable_fuzz"`
	
	// Custom weights (optional - if nil, default weights are used)
	Weights []float64 `json:"weights,omitempty" validate:"omitempty,len=19"`
}

// DefaultConfig returns a default FSRS configuration.
func DefaultConfig() Config {
	defaultParams := fsrs.DefaultParam()
	return Config{
		RequestRetention: defaultParams.RequestRetention,
		MaximumInterval:  defaultParams.MaximumInterval,
		EnableShortTerm:  defaultParams.EnableShortTerm,
		EnableFuzz:       defaultParams.EnableFuzz,
		Weights:          nil, // Use default weights
	}
}

// ToFSRSParameters converts Config to FSRS Parameters.
func (c Config) ToFSRSParameters() fsrs.Parameters {
	params := fsrs.DefaultParam()
	
	// Set custom values
	params.RequestRetention = c.RequestRetention
	params.MaximumInterval = c.MaximumInterval
	params.EnableShortTerm = c.EnableShortTerm
	params.EnableFuzz = c.EnableFuzz
	
	// Use custom weights if provided
	if len(c.Weights) == 19 {
		copy(params.W[:], c.Weights)
		// Recalculate dependent parameters
		params.Decay = -0.5
		params.Factor = 19.0 / 81.0
	}
	
	return params
}

// OptimizedConfig creates a configuration with default FSRS parameters.
func OptimizedConfig(cards []Card) Config {
	return DefaultConfig()
}

// ValidateConfig ensures the configuration values are within acceptable ranges.
func ValidateConfig(config Config) error {
	if config.RequestRetention < 0.7 || config.RequestRetention > 0.99 {
		return ErrInvalidScore // Reuse existing error for simplicity
	}
	
	if config.MaximumInterval < 1 || config.MaximumInterval > 36500 {
		return ErrInvalidScore
	}
	
	if len(config.Weights) != 0 && len(config.Weights) != 19 {
		return ErrInvalidScore
	}
	
	return nil
}