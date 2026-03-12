package domain_test

import (
	"errors"
	"testing"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
)

func TestNewTrip_ValidInput(t *testing.T) {
	trip, err := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)

	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if trip.Status != domain.StatusRequested {
		t.Errorf("Status = %v, want %v", trip.Status, domain.StatusRequested)
	}
	if trip.ID != "id-1" {
		t.Errorf("ID = %v, want id-1", trip.ID)
	}
}

func TestNewTrip_MissingFields(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		riderID     string
		origin      string
		destination string
		distance    float64
		fare        float64
	}{
		{"missing id", "", "rider-1", "Cairo", "Giza", 10, 30},
		{"missing riderID", "id-1", "", "Cairo", "Giza", 10, 30},
		{"missing origin", "id-1", "rider-1", "", "Giza", 10, 30},
		{"missing destination", "id-1", "rider-1", "Cairo", "", 10, 30},
		{"zero distance", "id-1", "rider-1", "Cairo", "Giza", 0, 30},
		{"zero fare", "id-1", "rider-1", "Cairo", "Giza", 10, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.New(tc.id, tc.riderID, tc.origin, tc.destination, tc.distance, tc.fare)
			if err == nil {
				t.Error("New() error = nil, want non-nil")
			}
		})
	}
}

func TestTrip_ValidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from domain.Status
		to   domain.Status
	}{
		{"requested to accepted", domain.StatusRequested, domain.StatusAccepted},
		{"requested to cancelled", domain.StatusRequested, domain.StatusCancelled},
		{"accepted to in_progress", domain.StatusAccepted, domain.StatusInProgress},
		{"accepted to cancelled", domain.StatusAccepted, domain.StatusCancelled},
		{"in_progress to completed", domain.StatusInProgress, domain.StatusCompleted},
		{"in_progress to cancelled", domain.StatusInProgress, domain.StatusCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trip, _ := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)
			trip.Status = tc.from

			if err := trip.Transition(tc.to); err != nil {
				t.Errorf("Transition() error = %v, want nil", err)
			}
			if trip.Status != tc.to {
				t.Errorf("Status = %v, want %v", trip.Status, tc.to)
			}
		})
	}
}

func TestTrip_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from domain.Status
		to   domain.Status
	}{
		{"requested to in_progress", domain.StatusRequested, domain.StatusInProgress},
		{"requested to completed", domain.StatusRequested, domain.StatusCompleted},
		{"completed to anything", domain.StatusCompleted, domain.StatusCancelled},
		{"cancelled to anything", domain.StatusCancelled, domain.StatusRequested},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trip, _ := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)
			trip.Status = tc.from

			err := trip.Transition(tc.to)
			if err == nil {
				t.Error("Transition() error = nil, want non-nil")
			}

			// Verify it is the specific error type we expect
			var transErr *domain.InvalidTransitionError
			if !errors.As(err, &transErr) {
				t.Errorf("error type = %T, want *InvalidTransitionError", err)
			}
		})
	}
}

func TestTrip_AssignDriver(t *testing.T) {
	t.Run("assigns driver and transitions to accepted", func(t *testing.T) {
		trip, _ := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)

		if err := trip.AssignDriver("driver-1"); err != nil {
			t.Fatalf("AssignDriver() error = %v", err)
		}
		if trip.DriverID != "driver-1" {
			t.Errorf("DriverID = %v, want driver-1", trip.DriverID)
		}
		if trip.Status != domain.StatusAccepted {
			t.Errorf("Status = %v, want accepted", trip.Status)
		}
	})

	t.Run("rejects empty driver id", func(t *testing.T) {
		trip, _ := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)

		if err := trip.AssignDriver(""); err == nil {
			t.Error("AssignDriver() error = nil, want non-nil")
		}
	})
}

func TestTrip_IsTerminal(t *testing.T) {
	tests := []struct {
		status   domain.Status
		terminal bool
	}{
		{domain.StatusRequested, false},
		{domain.StatusAccepted, false},
		{domain.StatusInProgress, false},
		{domain.StatusCompleted, true},
		{domain.StatusCancelled, true},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			trip, _ := domain.New("id-1", "rider-1", "Cairo", "Giza", 10.0, 30.0)
			trip.Status = tc.status

			if got := trip.IsTerminal(); got != tc.terminal {
				t.Errorf("IsTerminal() = %v, want %v", got, tc.terminal)
			}
		})
	}
}
