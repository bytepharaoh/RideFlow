package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytepharoh/rideflow/internal/gateway/middleware"
)

func TestPreviewTripRequestValidateSuccess(t *testing.T) {
	t.Parallel()

	req := previewTripRequest{
		Origin:      "Cairo",
		Destination: "Alexandria",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestPreviewTripRequestValidateFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request previewTripRequest
		wantErr string
	}{
		{
			name: "origin is blank after trimming",
			request: previewTripRequest{
				Origin:      "   ",
				Destination: "Alexandria",
			},
			wantErr: "origin is required",
		},
		{
			name: "destination is blank after trimming",
			request: previewTripRequest{
				Origin:      "Cairo",
				Destination: "   ",
			},
			wantErr: "destination is required",
		},
		{
			name: "origin and destination match ignoring case and spaces",
			request: previewTripRequest{
				Origin:      " Cairo ",
				Destination: "cairo",
			},
			wantErr: "origin and destination must be different",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.request.Validate()
			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}

			if err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestTripHandlerPreviewTripSuccess(t *testing.T) {
	t.Parallel()

	recorder := executePreviewTripRequest(t, `{"origin":"Cairo","destination":"Alexandria"}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Data struct {
			DistanceKM   float64 `json:"distance_km"`
			FareEstimate float64 `json:"fare_estimate"`
			ETAMinutes   int     `json:"eta_minutes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Data.DistanceKM != 25.0 {
		t.Fatalf("distance_km = %v, want %v", response.Data.DistanceKM, 25.0)
	}

	if response.Data.FareEstimate != 45.0 {
		t.Fatalf("fare_estimate = %v, want %v", response.Data.FareEstimate, 45.0)
	}

	if response.Data.ETAMinutes != 35 {
		t.Fatalf("eta_minutes = %d, want %d", response.Data.ETAMinutes, 35)
	}
}

func TestTripHandlerPreviewTripInvalidJSON(t *testing.T) {
	t.Parallel()

	assertPreviewTripError(t, `{"origin":`, http.StatusBadRequest, "INVALID_REQUEST", "request body must be valid JSON")
}

func TestTripHandlerPreviewTripValidationFailures(t *testing.T) {
	t.Parallel()

	t.Run("origin missing", func(t *testing.T) {
		t.Parallel()
		assertPreviewTripError(t, `{"origin":" ","destination":"Alexandria"}`, http.StatusBadRequest, "VALIDATION_ERROR", "origin is required")
	})

	t.Run("same places", func(t *testing.T) {
		t.Parallel()
		assertPreviewTripError(t, `{"origin":"Cairo","destination":" cairo "}`, http.StatusBadRequest, "VALIDATION_ERROR", "origin and destination must be different")
	})
}

func executePreviewTripRequest(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	handler := NewTripHandler()
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/trips/preview",
		strings.NewReader(body),
	)
	recorder := httptest.NewRecorder()

	wrapped := middleware.RequestID(http.HandlerFunc(handler.PreviewTrip))
	wrapped.ServeHTTP(recorder, req)

	return recorder
}

func assertPreviewTripError(t *testing.T, body string, wantStatus int, wantErrorCode string, wantErrorMsg string) {
	t.Helper()

	recorder := executePreviewTripRequest(t, body)
	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d", recorder.Code, wantStatus)
	}

	var response struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Error.Code != wantErrorCode {
		t.Fatalf("error.code = %q, want %q", response.Error.Code, wantErrorCode)
	}

	if response.Error.Message != wantErrorMsg {
		t.Fatalf("error.message = %q, want %q", response.Error.Message, wantErrorMsg)
	}

	if response.Error.RequestID == "" {
		t.Fatal("error.request_id = empty, want non-empty")
	}
}
