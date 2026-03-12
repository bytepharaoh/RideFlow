package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusOffline   Status = "offline"
	StatusAvailable Status = "available"
	StatusOnTrip    Status = "on_trip"
)

var validTransitions = map[Status][]Status{
	StatusOffline:   {StatusAvailable},
	StatusAvailable: {StatusOffline, StatusOnTrip},
	StatusOnTrip:    {StatusAvailable, StatusOffline},
}

type Location struct {
	Latitude  float64
	Longitude float64
	UpdatedAt time.Time
}

type Driver struct {
	ID            string
	Name          string
	Vehicle       string
	Status        Status
	Location      *Location
	CurrentTripID string
	Rating        float64
	TotalTrips    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func New(id, name, vehicle string) (*Driver, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if vehicle == "" {
		return nil, errors.New("vehicle is required")
	}

	now := time.Now().UTC()
	return &Driver{
		ID:        id,
		Name:      name,
		Vehicle:   vehicle,
		Status:    StatusOffline,
		Rating:    5.0,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *Driver) Transition(to Status) error {
	allowed, ok := validTransitions[d.Status]
	if !ok {
		return errors.New("unknown status")
	}

	for _, s := range allowed {
		if s == to {
			d.Status = to
			d.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return &InvalidTransitionError{From: d.Status, To: to}
}

func (d *Driver) GoOnline() error {
	return d.Transition(StatusAvailable)
}

func (d *Driver) GoOffline() error {
	return d.Transition(StatusOffline)
}

// AssignTrip puts the driver on a trip.
func (d *Driver) AssignTrip(tripID string) error {
	if tripID == "" {
		return errors.New("trip id is required")
	}
	if err := d.Transition(StatusOnTrip); err != nil {
		return err
	}
	d.CurrentTripID = tripID
	return nil
}

// CompleteTrip marks the driver as available again after a trip ends.
func (d *Driver) CompleteTrip() error {
	if err := d.Transition(StatusAvailable); err != nil {
		return err
	}
	d.CurrentTripID = ""
	d.TotalTrips++
	return nil
}

// UpdateLocation updates the driver's last known position.
func (d *Driver) UpdateLocation(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return errors.New("longitude must be between -180 and 180")
	}

	d.Location = &Location{
		Latitude:  lat,
		Longitude: lng,
		UpdatedAt: time.Now().UTC(),
	}
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// IsAvailable returns true if the driver can receive trip offers.
func (d *Driver) IsAvailable() bool {
	return d.Status == StatusAvailable && d.Location != nil
}

type InvalidTransitionError struct {
	From Status
	To   Status
}

func (e *InvalidTransitionError) Error() string {
	return "invalid transition from " + string(e.From) + " to " + string(e.To)
}
