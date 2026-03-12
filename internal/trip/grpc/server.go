package grpc

import (
	"context"
	"log/slog"
	"strings"

	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	tripv1.UnimplementedTripServiceServer
	logger *slog.Logger
}

func New(logger *slog.Logger) *Server {
	return &Server{
		logger: logger,
	}
}
func (s *Server) PreviewTrip(
	ctx context.Context, req *tripv1.PreviewTripRequest,
) (*tripv1.PreviewTripResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is required")
	}

	origin := strings.TrimSpace(req.GetOrigin())
	destination := strings.TrimSpace(req.GetDestination())

	if origin == "" {
		return nil, status.Errorf(codes.InvalidArgument, "origin is required")

	}
	if destination == "" {
		return nil, status.Errorf(codes.InvalidArgument, "destination is required")
	}
	if strings.EqualFold(origin, destination) {
		return nil, status.Errorf(codes.InvalidArgument, "origin and destination must be different")
	}
	s.logger.Info("preview trip requested",
		"origin", origin,
		"destination", destination,
	)
	// TODO Phase 6: replace with real route and fare calculation.
	return &tripv1.PreviewTripResponse{
		DistanceKm:   25.0,
		FareEstimate: 45.00,
		EtaMinutes:   35,
	}, nil

}
