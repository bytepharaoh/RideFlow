package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	driverv1 "github.com/bytepharoh/rideflow/internal/driver/gen/proto/driver"
	"github.com/bytepharoh/rideflow/internal/driver/service"
)

type Server struct {
	driverv1.UnimplementedDriverServiceServer
	svc    *service.DriverService
	logger *slog.Logger
}

func New(svc *service.DriverService, logger *slog.Logger) *Server {
	return &Server{svc: svc, logger: logger}
}

func (s *Server) RegisterDriver(ctx context.Context, req *driverv1.RegisterDriverRequest) (*driverv1.RegisterDriverResponse, error) {
	driver, err := s.svc.RegisterDriver(ctx, service.RegisterDriverInput{
		Name:    req.GetName(),
		Vehicle: req.GetVehicle(),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &driverv1.RegisterDriverResponse{
		DriverId: driver.ID,
		Name:     driver.Name,
		Vehicle:  driver.Vehicle,
		Status:   string(driver.Status),
	}, nil
}

func (s *Server) UpdateLocation(ctx context.Context, req *driverv1.UpdateLocationRequest) (*driverv1.UpdateLocationResponse, error) {
	err := s.svc.UpdateLocation(ctx,
		req.GetDriverId(),
		req.GetLatitude(),
		req.GetLongitude(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &driverv1.UpdateLocationResponse{Success: true}, nil
}

func (s *Server) SetAvailability(ctx context.Context, req *driverv1.SetAvailabilityRequest) (*driverv1.SetAvailabilityResponse, error) {
	err := s.svc.SetAvailability(ctx, req.GetDriverId(), req.GetOnline())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &driverv1.SetAvailabilityResponse{Status: "updated"}, nil
}

func (s *Server) FindNearestAvailable(ctx context.Context, req *driverv1.FindNearestAvailableRequest) (*driverv1.FindNearestAvailableResponse, error) {
	driver, err := s.svc.FindNearestAvailable(ctx, req.GetLatitude(), req.GetLongitude())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &driverv1.FindNearestAvailableResponse{
		DriverId:  driver.ID,
		Name:      driver.Name,
		Latitude:  driver.Location.Latitude,
		Longitude: driver.Location.Longitude,
	}, nil
}
