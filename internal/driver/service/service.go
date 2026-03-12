package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/driver/domain"
	"github.com/bytepharoh/rideflow/internal/driver/repository"
)

type IDGenerator func() string

type DriverService struct {
	repo       repository.DriverRepository
	generateID IDGenerator
	logger     *slog.Logger
}

func New(
	repo repository.DriverRepository,
	generateID IDGenerator,
	logger *slog.Logger,
) *DriverService {
	return &DriverService{
		repo:       repo,
		generateID: generateID,
		logger:     logger,
	}
}

type RegisterDriverInput struct {
	Name    string
	Vehicle string
}

func (s *DriverService) RegisterDriver(ctx context.Context, input RegisterDriverInput) (*domain.Driver, error) {
	driver, err := domain.New(s.generateID(), input.Name, input.Vehicle)
	if err != nil {
		return nil, fmt.Errorf("create driver: %w", err)
	}

	if err := s.repo.Create(ctx, driver); err != nil {
		return nil, fmt.Errorf("save driver: %w", err)
	}

	s.logger.Info("driver registered", "driver_id", driver.ID, "name", driver.Name)
	return driver, nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error {
	driver, err := s.repo.FindByID(ctx, driverID)
	if err != nil {
		return fmt.Errorf("find driver: %w", err)
	}

	if err := driver.UpdateLocation(lat, lng); err != nil {
		return fmt.Errorf("update location: %w", err)
	}

	if err := s.repo.Update(ctx, driver); err != nil {
		return fmt.Errorf("save driver: %w", err)
	}

	return nil
}

func (s *DriverService) SetAvailability(ctx context.Context, driverID string, online bool) error {
	driver, err := s.repo.FindByID(ctx, driverID)
	if err != nil {
		return fmt.Errorf("find driver: %w", err)
	}

	if online {
		err = driver.GoOnline()
	} else {
		err = driver.GoOffline()
	}
	if err != nil {
		return fmt.Errorf("set availability: %w", err)
	}

	if err := s.repo.Update(ctx, driver); err != nil {
		return fmt.Errorf("save driver: %w", err)
	}

	s.logger.Info("driver availability updated",
		"driver_id", driverID,
		"online", online,
	)
	return nil
}

// FindNearestAvailable finds the closest available driver to the given coordinates.
// This is called when a trip is created — Phase 8 wires this into the event flow.
func (s *DriverService) FindNearestAvailable(ctx context.Context, lat, lng float64) (*domain.Driver, error) {
	available, err := s.repo.FindAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("find available drivers: %w", err)
	}

	if len(available) == 0 {
		return nil, errors.New("no available drivers")
	}

	nearest := domain.NearestDriver(available, lat, lng)
	if nearest == nil {
		return nil, errors.New("no drivers with location")
	}

	return nearest, nil
}

func (s *DriverService) AssignTrip(ctx context.Context, driverID, tripID string) error {
	driver, err := s.repo.FindByID(ctx, driverID)
	if err != nil {
		return fmt.Errorf("find driver: %w", err)
	}

	if err := driver.AssignTrip(tripID); err != nil {
		return fmt.Errorf("assign trip: %w", err)
	}

	if err := s.repo.Update(ctx, driver); err != nil {
		return fmt.Errorf("save driver: %w", err)
	}

	s.logger.Info("trip assigned to driver",
		"driver_id", driverID,
		"trip_id", tripID,
	)
	return nil
}
