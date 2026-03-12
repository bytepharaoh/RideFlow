package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
	"github.com/bytepharoh/rideflow/internal/trip/repository"
)

type IDGenerator func() string
type TripService struct {
	repo       repository.TripRepository
	fareCalc   *domain.Calculator
	generateID IDGenerator
	logger     *slog.Logger
}

func New(
	repo repository.TripRepository,
	fareCalc *domain.Calculator,
	generateID IDGenerator,
	logger *slog.Logger,
) *TripService {
	return &TripService{
		repo:       repo,
		fareCalc:   fareCalc,
		generateID: generateID,
		logger:     logger,
	}
}

// PreviewTripResult holds the result of a trip preview calculation.
// The gRPC server converts this to a proto response.
type PreviewTripResult struct {
	DistanceKM   float64
	FareEstimate float64
	ETAMinutes   int
}

func (s *TripService) PreviewTrip(
	ctx context.Context,
	origin, destination string,
) (*PreviewTripResult, error) {
	if origin == "" {
		return nil, errors.New("origin is required")
	}
	if destination == "" {
		return nil, errors.New("destination is required")
	}
	const placeholderDistanceKM = 10.0
	const placeholderETAMinutes = 20
	fare := s.fareCalc.Calculate(placeholderDistanceKM)

	s.logger.Info("trip previewed",
		"origin", origin,
		"destination", destination,
		"distance_km", placeholderDistanceKM,
		"fare_estimate", fare,
	)

	return &PreviewTripResult{
		DistanceKM:   placeholderDistanceKM,
		FareEstimate: fare,
		ETAMinutes:   placeholderETAMinutes,
	}, nil
}

// CreateTripInput holds the data needed to create a trip.
type CreateTripInput struct {
	RiderID     string
	Origin      string
	Destination string
}

func (s *TripService) CreateTrip(
	ctx context.Context,
	input CreateTripInput,
) (*domain.Trip, error) {
	// Rule: one active trip per rider
	existing, err := s.repo.FindActiveByRiderID(ctx, input.RiderID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check active trip: %w", err)
	}
	if existing != nil {
		return nil, errors.New("rider already has an active trip")
	}

	// Calculate fare for the new trip
	// TODO: use real distance from routing API
	const placeholderDistanceKM = 10.0
	fare := s.fareCalc.Calculate(placeholderDistanceKM)

	// Create the domain object — enforces domain invariants
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

	// Persist the trip
	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, fmt.Errorf("save trip: %w", err)
	}

	s.logger.Info("trip created",
		"trip_id", trip.ID,
		"rider_id", trip.RiderID,
		"origin", trip.Origin,
		"destination", trip.Destination,
		"fare_estimate", trip.FareEstimate,
	)

	return trip, nil
}

// GetTrip retrieves a trip by ID.
// Returns a not found error if the trip does not exist.
func (s *TripService) GetTrip(
	ctx context.Context,
	tripID string,
) (*domain.Trip, error) {
	trip, err := s.repo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("trip %s not found", tripID)
		}
		return nil, fmt.Errorf("get trip: %w", err)
	}
	return trip, nil
}
