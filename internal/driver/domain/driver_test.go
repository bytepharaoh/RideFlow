package domain_test

import (
	"errors"
	"testing"

	"github.com/bytepharoh/rideflow/internal/driver/domain"
)

func TestNewDriver_Valid(t *testing.T) {
	d, err := domain.New("id-1", "Ziad", "Toyota Camry")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if d.Status != domain.StatusOffline {
		t.Errorf("Status = %v, want offline", d.Status)
	}
	if d.Rating != 5.0 {
		t.Errorf("Rating = %v, want 5.0", d.Rating)
	}
}

func TestNewDriver_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dname   string
		vehicle string
	}{
		{"missing id", "", "Ziad", "Toyota"},
		{"missing name", "id-1", "", "Toyota"},
		{"missing vehicle", "id-1", "Ziad", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.New(tc.id, tc.dname, tc.vehicle)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDriver_Transitions(t *testing.T) {
	t.Run("offline to available", func(t *testing.T) {
		d, _ := domain.New("id-1", "Ziad", "Toyota")
		if err := d.GoOnline(); err != nil {
			t.Fatalf("GoOnline() error = %v", err)
		}
		if d.Status != domain.StatusAvailable {
			t.Errorf("Status = %v, want available", d.Status)
		}
	})

	t.Run("cannot go on trip from offline", func(t *testing.T) {
		d, _ := domain.New("id-1", "Ziad", "Toyota")
		err := d.AssignTrip("trip-1")
		if err == nil {
			t.Error("expected error, got nil")
		}
		var transErr *domain.InvalidTransitionError
		if !errors.As(err, &transErr) {
			t.Errorf("error type = %T, want *InvalidTransitionError", err)
		}
	})

	t.Run("assign trip when available", func(t *testing.T) {
		d, _ := domain.New("id-1", "Ziad", "Toyota")
		_ = d.GoOnline()
		if err := d.AssignTrip("trip-1"); err != nil {
			t.Fatalf("AssignTrip() error = %v", err)
		}
		if d.Status != domain.StatusOnTrip {
			t.Errorf("Status = %v, want on_trip", d.Status)
		}
		if d.CurrentTripID != "trip-1" {
			t.Errorf("CurrentTripID = %v, want trip-1", d.CurrentTripID)
		}
	})
}

func TestDriver_UpdateLocation(t *testing.T) {
	t.Run("valid coordinates", func(t *testing.T) {
		d, _ := domain.New("id-1", "Ziad", "Toyota")
		if err := d.UpdateLocation(30.04, 31.23); err != nil {
			t.Fatalf("UpdateLocation() error = %v", err)
		}
		if d.Location == nil {
			t.Fatal("Location is nil")
		}
	})

	t.Run("invalid latitude", func(t *testing.T) {
		d, _ := domain.New("id-1", "Ziad", "Toyota")
		if err := d.UpdateLocation(91.0, 31.23); err == nil {
			t.Error("expected error for invalid latitude")
		}
	})
}

func TestDriver_IsAvailable(t *testing.T) {
	d, _ := domain.New("id-1", "Ziad", "Toyota")

	if d.IsAvailable() {
		t.Error("offline driver should not be available")
	}

	_ = d.GoOnline()
	if d.IsAvailable() {
		t.Error("online driver without location should not be available")
	}

	_ = d.UpdateLocation(30.04, 31.23)
	if !d.IsAvailable() {
		t.Error("online driver with location should be available")
	}
}
