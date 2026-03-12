package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
	"github.com/bytepharoh/rideflow/internal/trip/repository"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

type IDGenerator func() string

type TripService struct {
	repo       repository.TripRepository
	fareCalc   *domain.Calculator
	generateID IDGenerator
	publisher  *rabbitmq.Publisher
	logger     *slog.Logger
}

func New(
	repo repository.TripRepository,
	fareCalc *domain.Calculator,
	generateID IDGenerator,
	publisher *rabbitmq.Publisher,
	logger *slog.Logger,
) *TripService {
	return &TripService{
		repo:       repo,
		fareCalc:   fareCalc,
		generateID: generateID,
		publisher:  publisher,
		logger:     logger,
	}
}

// PreviewTrip stays the same — no changes needed
type PreviewTripResult struct {
	DistanceKM   float64
	FareEstimate float64
	ETAMinutes   int
}

func (s *TripService) PreviewTrip(ctx context.Context, origin, destination string) (*PreviewTripResult, error) {
	if origin == "" {
		return nil, errors.New("origin is required")
	}
	if destination == "" {
		return nil, errors.New("destination is required")
	}

	const placeholderDistanceKM = 10.0
	const placeholderETAMinutes = 20

	fare := s.fareCalc.Calculate(placeholderDistanceKM)

	return &PreviewTripResult{
		DistanceKM:   placeholderDistanceKM,
		FareEstimate: fare,
		ETAMinutes:   placeholderETAMinutes,
	}, nil
}

type CreateTripInput struct {
	RiderID     string
	Origin      string
	Destination string
	OriginLat   float64
	OriginLng   float64
}

func (s *TripService) CreateTrip(ctx context.Context, input CreateTripInput) (*domain.Trip, error) {
	existing, err := s.repo.FindActiveByRiderID(ctx, input.RiderID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check active trip: %w", err)
	}
	if existing != nil {
		return nil, errors.New("rider already has an active trip")
	}

	const placeholderDistanceKM = 10.0
	fare := s.fareCalc.Calculate(placeholderDistanceKM)

	trip, err := domain.New(
		s.generateID(),
		input.RiderID,
		input.Origin,
		input.Destination,
		placeholderDistanceKM,
		fare,
	)
	if err != nil {
		return nil, fmt.Errorf("create trip: %w", err)
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, fmt.Errorf("save trip: %w", err)
	}

	// Publish TripCreated event — Driver Service will consume this.
	// We publish after saving so we never publish an event for a
	// trip that failed to persist.
	event := events.TripCreated{
		TripID:      trip.ID,
		RiderID:     trip.RiderID,
		Origin:      trip.Origin,
		Destination: trip.Destination,
		OriginLat:   input.OriginLat,
		OriginLng:   input.OriginLng,
	}

	if err := s.publisher.Publish(ctx, events.RoutingKeyTripCreated, event); err != nil {
		// Publishing failed but the trip was saved.
		// We log the error but do not fail the request —
		// the trip exists, we just need to handle the missing event.
		// In Phase 17 we add an outbox pattern to handle this reliably.
		s.logger.Error("failed to publish trip created event",
			"trip_id", trip.ID,
			"error", err,
		)
	}

	s.logger.Info("trip created",
		"trip_id", trip.ID,
		"rider_id", trip.RiderID,
	)

	return trip, nil
}

func (s *TripService) GetTrip(ctx context.Context, tripID string) (*domain.Trip, error) {
	trip, err := s.repo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("trip %s not found", tripID)
		}
		return nil, fmt.Errorf("get trip: %w", err)
	}
	return trip, nil
}

// AssignDriver is called when the Driver Service assigns a driver.
// This is triggered by consuming the DriverAssigned event.
func (s *TripService) AssignDriver(ctx context.Context, tripID, driverID string) error {
	trip, err := s.repo.FindByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("find trip: %w", err)
	}

	if err := trip.AssignDriver(driverID); err != nil {
		return fmt.Errorf("assign driver: %w", err)
	}

	if err := s.repo.Update(ctx, trip); err != nil {
		return fmt.Errorf("update trip: %w", err)
	}

	s.logger.Info("driver assigned to trip",
		"trip_id", tripID,
		"driver_id", driverID,
	)

	return nil
}
