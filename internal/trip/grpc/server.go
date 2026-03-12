package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	"github.com/bytepharoh/rideflow/internal/trip/service"
)

const maxInt32 = int(^uint32(0) >> 1)

type Server struct {
	tripv1.UnimplementedTripServiceServer
	svc    *service.TripService
	logger *slog.Logger
}

func New(svc *service.TripService, logger *slog.Logger) *Server {
	return &Server{
		svc:    svc,
		logger: logger,
	}
}
func (s *Server) PreviewTrip(
	ctx context.Context,
	req *tripv1.PreviewTripRequest,
) (*tripv1.PreviewTripResponse, error) {
	result, err := s.svc.PreviewTrip(ctx, req.GetOrigin(), req.GetDestination())
	if err != nil {
		s.logger.Error("preview trip failed", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if result.ETAMinutes < 0 || result.ETAMinutes > maxInt32 {
		s.logger.Error("preview trip failed", "eta_minutes", result.ETAMinutes)
		return nil, status.Error(codes.Internal, "eta minutes is out of range")
	}

	return &tripv1.PreviewTripResponse{
		DistanceKm:   result.DistanceKM,
		FareEstimate: result.FareEstimate,
		EtaMinutes:   int32(result.ETAMinutes),
	}, nil
}
func (s *Server) CreateTrip(ctx context.Context, req *tripv1.CreateTripRequest) (*tripv1.CreateTripResponse, error) {
	trip, err := s.svc.CreateTrip(ctx, service.CreateTripInput{
		RiderID:     req.GetRiderId(),
		Origin:      req.GetOrigin(),
		Destination: req.GetDestination(),
		OriginLat:   req.GetOriginLat(),
		OriginLng:   req.GetOriginLng(),
	})
	if err != nil {
		s.logger.Error("create trip failed", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &tripv1.CreateTripResponse{
		TripId:       trip.ID,
		Status:       string(trip.Status),
		FareEstimate: trip.FareEstimate,
	}, nil
}
func (s *Server) GetTrip(ctx context.Context, req *tripv1.GetTripRequest) (*tripv1.GetTripResponse, error) {
	trip, err := s.svc.GetTrip(ctx, req.GetTripId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &tripv1.GetTripResponse{
		TripId:       trip.ID,
		RiderId:      trip.RiderID,
		DriverId:     trip.DriverID,
		Origin:       trip.Origin,
		Destination:  trip.Destination,
		Status:       string(trip.Status),
		FareEstimate: trip.FareEstimate,
		FinalFare:    trip.FinalFare,
	}, nil
}
