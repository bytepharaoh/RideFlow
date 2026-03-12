package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusRequested  Status = "requested"
	StatusAccepted   Status = "accepted"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

var validTransitions = map[Status][]Status{
	StatusRequested:  {StatusAccepted, StatusCancelled},
	StatusAccepted:   {StatusInProgress, StatusCancelled},
	StatusInProgress: {StatusCompleted, StatusCancelled},
	StatusCompleted:  {}, // terminal state — no further transitions
	StatusCancelled:  {}, // terminal state — no further transitions
}

type Trip struct {
	ID string

	RiderID string

	DriverID string

	Origin string

	Destination string

	Status Status

	DistanceKM float64

	FareEstimate float64

	FinalFare float64

	CreatedAt time.Time

	UpdatedAt time.Time
}

func New(id, riderID, origin, destination string, distanceKM, fareEstimate float64) (*Trip, error) {
	if id == "" {
		return nil, errors.New("trip id is required")
	}
	if riderID == "" {
		return nil, errors.New("rider id is required")
	}
	if origin == "" {
		return nil, errors.New("origin is required")
	}
	if destination == "" {
		return nil, errors.New("destination is required")
	}
	if distanceKM <= 0 {
		return nil, errors.New("distance must be greater than zero")
	}
	if fareEstimate <= 0 {
		return nil, errors.New("fare estimate must be greater than zero")
	}
	now := time.Now().UTC()
	return &Trip{
		ID:           id,
		RiderID:      riderID,
		Origin:       origin,
		Destination:  destination,
		Status:       StatusRequested,
		DistanceKM:   distanceKM,
		FareEstimate: fareEstimate,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Transition attempts to move the trip to a new status.
// This method is the state machine enforcer. All status changes
// must go through here — never set trip.Status directly.
func (t *Trip) Transition(to Status) error {
	allowed, ok := validTransitions[t.Status]
	if !ok {
		return errors.New("trip is in an unknown state")
	}
	for _, s := range allowed {
		if s == to {
			t.Status = to
			t.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return &InvalidTransitionError{
		From: t.Status,
		To:   to,
	}

}
func (t *Trip) AssignDriver(driverID string) error {
	if driverID == "" {
		return errors.New("driver id is required")
	}
	if err := t.Transition(StatusAccepted); err != nil {
		return err
	}
	t.DriverID = driverID
	return nil

}

// Complete marks the trip as completed and sets the final fare.
// The final fare may differ from the estimate if the route changed.
func (t *Trip) Complete(finalFare float64) error {
	if finalFare <= 0 {
		return errors.New("final fare must be greater than zero")
	}
	if err := t.Transition(StatusCompleted); err != nil {
		return err
	}
	t.FinalFare = finalFare
	return nil
}

// Cancel marks the trip as cancelled.
// Can be called by either the rider or driver before completion.
func (t *Trip) Cancel() error {
	return t.Transition(StatusCancelled)
}

// IsTerminal returns true if the trip is in a state from which
// it cannot transition further.
func (t *Trip) IsTerminal() bool {
	return t.Status == StatusCompleted || t.Status == StatusCancelled
}

type InvalidTransitionError struct {
	From Status
	To   Status
}

func (e *InvalidTransitionError) Error() string {
	return "invalid transition from " + string(e.From) + " to " + string(e.To)
}
