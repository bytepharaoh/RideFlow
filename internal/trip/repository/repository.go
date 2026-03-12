package repository

import (
	"context"
	"errors"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
)

var ErrNotFound = errors.New("trip not found")

type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	FindByID(ctx context.Context, id string) (*domain.Trip, error)
	Update(ctx context.Context, trip *domain.Trip) error
	FindActiveByRiderID(ctx context.Context, riderID string) (*domain.Trip, error)
}
