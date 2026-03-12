package domain

type FareConfig struct {
	BaseFare  float64
	RatePerKM float64

	// SurgeMutiplier is applied during high-demand periods.
	// 1.0 means no surge. 1.5 means 50% more expensive.
	SurgeMultiplier float64
}

var BaseFare float64 = 5.00
var RatePerKM float64 = 2.5
var SurgeMultiplier float64 = 1.0

// DefaultFareConfig returns the standard fare configuration.
func DefaultFareConfig() FareConfig {
	return FareConfig{
		BaseFare:        BaseFare,
		RatePerKM:       RatePerKM,
		SurgeMultiplier: SurgeMultiplier,
	}
}

type Calculator struct {
	config FareConfig
}

// NewCalculator creates a Calculator with the given configuration.
func NewCalculator(cfg FareConfig) *Calculator {
	return &Calculator{config: cfg}
}

// Calculate computes the fare for a trip of the given distance.
//
// Formula:
//
//	fare = (base_fare + distance_km * rate_per_km) * surge_multiplier
func (c *Calculator) Calculate(distanceKM float64) float64 {
	if distanceKM <= 0 {
		return c.config.BaseFare
	}

	raw := (c.config.BaseFare + distanceKM*c.config.RatePerKM) * c.config.SurgeMultiplier
	// This technique: multiply by 100, add 0.5, truncate, divide by 100.
	return float64(int(raw*100+0.5)) / 100
}
