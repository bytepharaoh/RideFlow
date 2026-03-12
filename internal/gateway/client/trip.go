package client

import (
	"context"
	"fmt"

	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type TripClient struct {
	client tripv1.TripServiceClient
}

func NewTripClient(addr string) (*TripClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(
		insecure.NewCredentials(),
	))
	if err != nil {
		return nil, fmt.Errorf("trip client: connect to %s: %w", addr, err)
	}
	return &TripClient{
		client: tripv1.NewTripServiceClient(conn),
	}, nil
}

type PreviewTripResult struct {
	DistanceKM   float64 `json:"distance_km"`
	FareEstimate float64 `json:"fare_estimate"`
	ETAMinutes   int32   `json:"eta_minutes"`
}

type CreateTripInput struct {
	RiderID     string
	Origin      string
	Destination string
	OriginLat   float64
	OriginLng   float64
}

type CreateTripResult struct {
	TripID       string  `json:"trip_id"`
	Status       string  `json:"status"`
	FareEstimate float64 `json:"fare_estimate"`
}

type GetTripResult struct {
	TripID       string  `json:"trip_id"`
	RiderID      string  `json:"rider_id"`
	DriverID     string  `json:"driver_id"`
	Origin       string  `json:"origin"`
	Destination  string  `json:"destination"`
	Status       string  `json:"status"`
	FareEstimate float64 `json:"fare_estimate"`
	FinalFare    float64 `json:"final_fare"`
}

func (c *TripClient) PreviewTrip(
	ctx context.Context,
	origin, destination string,
) (*PreviewTripResult, error) {
	resp, err := c.client.PreviewTrip(ctx, &tripv1.PreviewTripRequest{
		Origin:      origin,
		Destination: destination,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return nil, fmt.Errorf("validation: %s", st.Message())
			case codes.Unavailable:
				return nil, fmt.Errorf("service unavailable: %s", st.Message())

			}
		}
		return nil, fmt.Errorf("preview trip: %w", err)

	}
	return &PreviewTripResult{
		DistanceKM:   resp.DistanceKm,
		FareEstimate: resp.FareEstimate,
		ETAMinutes:   resp.EtaMinutes,
	}, nil
}

func (c *TripClient) CreateTrip(ctx context.Context, input CreateTripInput) (*CreateTripResult, error) {
	resp, err := c.client.CreateTrip(ctx, &tripv1.CreateTripRequest{
		RiderId:     input.RiderID,
		Origin:      input.Origin,
		Destination: input.Destination,
		OriginLat:   input.OriginLat,
		OriginLng:   input.OriginLng,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unavailable:
				return nil, fmt.Errorf("service unavailable: %s", st.Message())
			case codes.InvalidArgument:
				return nil, fmt.Errorf("validation: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("create trip: %w", err)
	}

	return &CreateTripResult{
		TripID:       resp.GetTripId(),
		Status:       resp.GetStatus(),
		FareEstimate: resp.GetFareEstimate(),
	}, nil
}

func (c *TripClient) GetTrip(ctx context.Context, tripID string) (*GetTripResult, error) {
	resp, err := c.client.GetTrip(ctx, &tripv1.GetTripRequest{
		TripId: tripID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("not found: %s", st.Message())
			case codes.Unavailable:
				return nil, fmt.Errorf("service unavailable: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("get trip: %w", err)
	}

	return &GetTripResult{
		TripID:       resp.GetTripId(),
		RiderID:      resp.GetRiderId(),
		DriverID:     resp.GetDriverId(),
		Origin:       resp.GetOrigin(),
		Destination:  resp.GetDestination(),
		Status:       resp.GetStatus(),
		FareEstimate: resp.GetFareEstimate(),
		FinalFare:    resp.GetFinalFare(),
	}, nil
}
