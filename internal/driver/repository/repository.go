package repository

import (
	"context"
	"errors"

	"github.com/bytepharoh/rideflow/internal/driver/domain"
)

var ErrNotFound = errors.New("driver not found")

type DriverRepository interface {
	Create(ctx context.Context, driver *domain.Driver) error
	FindByID(ctx context.Context, id string) (*domain.Driver, error)
	Update(ctx context.Context, driver *domain.Driver) error
	// FindAvailable returns all drivers that are online and have a location.
	// Used for matching when a trip is created.
	FindAvailable(ctx context.Context) ([]*domain.Driver, error)
}
