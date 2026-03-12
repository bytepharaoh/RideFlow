package repository

import (
	"context"
	"sync"

	"github.com/bytepharoh/rideflow/internal/driver/domain"
)

type InMemoryDriverRepository struct {
	mu      sync.RWMutex
	drivers map[string]*domain.Driver
}

func NewInMemoryDriverRepository() *InMemoryDriverRepository {
	return &InMemoryDriverRepository{
		drivers: make(map[string]*domain.Driver),
	}
}

func (r *InMemoryDriverRepository) Create(ctx context.Context, driver *domain.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drivers[driver.ID]; exists {
		return ErrNotFound
	}

	copy := *driver
	r.drivers[driver.ID] = &copy
	return nil
}

func (r *InMemoryDriverRepository) FindByID(ctx context.Context, id string) (*domain.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	driver, exists := r.drivers[id]
	if !exists {
		return nil, ErrNotFound
	}

	copy := *driver
	return &copy, nil
}

func (r *InMemoryDriverRepository) Update(ctx context.Context, driver *domain.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drivers[driver.ID]; !exists {
		return ErrNotFound
	}

	copy := *driver
	r.drivers[driver.ID] = &copy
	return nil
}

func (r *InMemoryDriverRepository) FindAvailable(ctx context.Context) ([]*domain.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var available []*domain.Driver
	for _, d := range r.drivers {
		if d.IsAvailable() {
			copy := *d
			available = append(available, &copy)
		}
	}

	return available, nil
}
