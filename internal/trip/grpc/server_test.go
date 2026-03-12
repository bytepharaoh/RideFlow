package grpc

import (
	"context"
	"io"
	"log/slog"
	"testing"

	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerPreviewTripSuccess(t *testing.T) {
	t.Parallel()

	server := New(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	resp, err := server.PreviewTrip(context.Background(), &tripv1.PreviewTripRequest{
		Origin:      " Cairo ",
		Destination: "Giza",
	})
	if err != nil {
		t.Fatalf("PreviewTrip() error = %v", err)
	}

	if resp.DistanceKm != 25.0 {
		t.Fatalf("DistanceKm = %v, want %v", resp.DistanceKm, 25.0)
	}

	if resp.FareEstimate != 45.0 {
		t.Fatalf("FareEstimate = %v, want %v", resp.FareEstimate, 45.0)
	}

	if resp.EtaMinutes != 35 {
		t.Fatalf("EtaMinutes = %d, want %d", resp.EtaMinutes, 35)
	}
}

func TestServerPreviewTripNilRequest(t *testing.T) {
	t.Parallel()

	assertInvalidArgument(t, nil, "request is required")
}

func TestServerPreviewTripValidationFailures(t *testing.T) {
	t.Parallel()

	t.Run("origin blank after trim", func(t *testing.T) {
		t.Parallel()
		assertInvalidArgument(t, &tripv1.PreviewTripRequest{
			Origin:      "   ",
			Destination: "Giza",
		}, "origin is required")
	})

	t.Run("destination blank after trim", func(t *testing.T) {
		t.Parallel()
		assertInvalidArgument(t, &tripv1.PreviewTripRequest{
			Origin:      "Cairo",
			Destination: "   ",
		}, "destination is required")
	})

	t.Run("same origin and destination ignoring case and spaces", func(t *testing.T) {
		t.Parallel()
		assertInvalidArgument(t, &tripv1.PreviewTripRequest{
			Origin:      " Cairo ",
			Destination: "cairo",
		}, "origin and destination must be different")
	})
}

func assertInvalidArgument(t *testing.T, request *tripv1.PreviewTripRequest, wantMsg string) {
	t.Helper()

	server := New(slog.New(slog.NewJSONHandler(io.Discard, nil)))
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
