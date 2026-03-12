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
