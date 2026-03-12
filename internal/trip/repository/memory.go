package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
)

type InMemoryTripRepository struct {
	// mu protects trips from concurrent access.
	// RWMutex allows multiple concurrent readers but only one writer —
	mu    sync.RWMutex
	trips map[string]*domain.Trip
}

// NewInMemoryTripRepository creates an empty in-memory repository.
func NewInMemoryTripRepository() *InMemoryTripRepository {
	return &InMemoryTripRepository{
		trips: make(map[string]*domain.Trip),
	}
}
func (r *InMemoryTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.trips[trip.ID]; exists {
		return errors.New("trip already exists: " + trip.ID)
	}
	// Store a copy, not the pointer.
	// If the caller modifies the trip after Create(), we do not
	// want the stored version to change unexpectedly.

	copy := *trip
	r.trips[trip.ID] = &copy
	return nil
}

func (r *InMemoryTripRepository) FindByID(ctx context.Context, id string) (*domain.Trip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	trip, exists := r.trips[id]
	if !exists {
		return nil, ErrNotFound
	}
	copy := *trip
	return &copy, nil
}
func (r *InMemoryTripRepository) Update(ctx context.Context, trip *domain.Trip) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.trips[trip.ID]; !exists {
		return ErrNotFound
	}

	copy := *trip
	r.trips[trip.ID] = &copy
	return nil
}
func (r *InMemoryTripRepository) FindActiveByRiderID(ctx context.Context, riderID string) (*domain.Trip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, trip := range r.trips {
		if trip.RiderID == riderID && !trip.IsTerminal() {
			copy := *trip
			return &copy, nil
		}
	}

	return nil, ErrNotFound
}
