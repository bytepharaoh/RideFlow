package domain_test

import (
	"testing"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
)

func TestCalculator_Calculate(t *testing.T) {
	cfg := domain.FareConfig{
		BaseFare:        5.00,
		RatePerKM:       2.50,
		SurgeMultiplier: 1.0,
	}
	calc := domain.NewCalculator(cfg)

	tests := []struct {
		name       string
		distanceKM float64
		want       float64
	}{
		// fare = (5.00 + 10 * 2.50) * 1.0 = 30.00
		{"10km no surge", 10.0, 30.00},
		// fare = (5.00 + 0 * 2.50) * 1.0 = 5.00
		{"zero distance returns base fare", 0, 5.00},
		// fare = (5.00 + 5 * 2.50) * 1.0 = 17.50
		{"5km", 5.0, 17.50},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calc.Calculate(tc.distanceKM)
			if got != tc.want {
				t.Errorf("Calculate(%v) = %v, want %v", tc.distanceKM, got, tc.want)
			}
		})
	}
}

func TestCalculator_WithSurge(t *testing.T) {
	cfg := domain.FareConfig{
		BaseFare:        5.00,
		RatePerKM:       2.50,
		SurgeMultiplier: 1.5,
	}
	calc := domain.NewCalculator(cfg)

	// fare = (5.00 + 10 * 2.50) * 1.5 = 45.00
	got := calc.Calculate(10.0)
	want := 45.00

	if got != want {
		t.Errorf("Calculate(10.0) with surge = %v, want %v", got, want)
	}
}
