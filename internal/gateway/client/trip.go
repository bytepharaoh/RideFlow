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
	DistanceKM   float64
	FareEstimate float64
	ETAMinutes   int32
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
