package domain_test

import (
	"testing"
	"time"

	"github.com/bytepharoh/rideflow/internal/driver/domain"
)

func TestNearestDriver(t *testing.T) {
	drivers := []*domain.Driver{
		{
			ID:     "far-driver",
			Status: domain.StatusAvailable,
			Location: &domain.Location{
				Latitude:  31.0, // further away
				Longitude: 32.0,
				UpdatedAt: time.Now(),
			},
		},
		{
			ID:     "near-driver",
			Status: domain.StatusAvailable,
			Location: &domain.Location{
				Latitude:  30.05, // closer to the pickup
				Longitude: 31.24,
				UpdatedAt: time.Now(),
			},
		},
	}

	// Pickup at Cairo coordinates
	nearest := domain.NearestDriver(drivers, 30.04, 31.23)

	if nearest == nil {
		t.Fatal("NearestDriver() returned nil")
	}
	if nearest.ID != "near-driver" {
		t.Errorf("NearestDriver() = %v, want near-driver", nearest.ID)
	}
}

func TestNearestDriver_NoDrivers(t *testing.T) {
	nearest := domain.NearestDriver([]*domain.Driver{}, 30.04, 31.23)
	if nearest != nil {
		t.Error("expected nil for empty driver list")
	}
}
