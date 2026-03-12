package grpc

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/bytepharoh/rideflow/internal/trip/domain"
	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	"github.com/bytepharoh/rideflow/internal/trip/repository"
	tripservice "github.com/bytepharoh/rideflow/internal/trip/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerPreviewTripSuccess(t *testing.T) {
	t.Parallel()

	server := newTestServer()
	resp, err := server.PreviewTrip(context.Background(), &tripv1.PreviewTripRequest{
		Origin:      " Cairo ",
		Destination: "Giza",
	})
	if err != nil {
		t.Fatalf("PreviewTrip() error = %v", err)
	}

	if resp.DistanceKm != 10.0 {
		t.Fatalf("DistanceKm = %v, want %v", resp.DistanceKm, 10.0)
	}

	if resp.FareEstimate != 30.0 {
		t.Fatalf("FareEstimate = %v, want %v", resp.FareEstimate, 30.0)
	}

	if resp.EtaMinutes != 20 {
		t.Fatalf("EtaMinutes = %d, want %d", resp.EtaMinutes, 20)
	}
}

func TestServerPreviewTripNilRequest(t *testing.T) {
	t.Parallel()

	assertInvalidArgument(t, nil, "origin is required")
}

func TestServerPreviewTripValidationFailures(t *testing.T) {
	t.Parallel()

	t.Run("origin empty", func(t *testing.T) {
		t.Parallel()
		assertInvalidArgument(t, &tripv1.PreviewTripRequest{
			Origin:      "",
			Destination: "Giza",
		}, "origin is required")
	})

	t.Run("destination empty", func(t *testing.T) {
		t.Parallel()
		assertInvalidArgument(t, &tripv1.PreviewTripRequest{
			Origin:      "Cairo",
			Destination: "",
		}, "destination is required")
	})
}

func assertInvalidArgument(t *testing.T, request *tripv1.PreviewTripRequest, wantMsg string) {
	t.Helper()

	server := newTestServer()
	_, err := server.PreviewTrip(context.Background(), request)
	if err == nil {
		t.Fatalf("PreviewTrip() error = nil, want %q", wantMsg)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("status.FromError() ok = false, want true")
	}

	if st.Code() != codes.InvalidArgument {
		t.Fatalf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}

	if st.Message() != wantMsg {
		t.Fatalf("status message = %q, want %q", st.Message(), wantMsg)
	}
}

func newTestServer() *Server {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	svc := tripservice.New(
		repository.NewInMemoryTripRepository(),
		domain.NewCalculator(domain.DefaultFareConfig()),
		func() string { return "trip-test-id" },
		nil,
		logger,
	)

	return New(svc, logger)
}
