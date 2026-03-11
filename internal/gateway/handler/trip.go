package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bytepharoh/rideflow/internal/gateway/middleware"
	"github.com/bytepharoh/rideflow/internal/gateway/response"
)

type TripHandler struct {
}

func NewTripHandler() *TripHandler {
	return &TripHandler{}
}

// previewTripRequest is the expected JSON body for POST /api/v1/trips/preview.
type previewTripRequest struct {
	// Origin is the pickup location.
	Origin string `json:"origin"`

	// Destination is the drop-off location.
	Destination string `json:"destination"`
}

// Validate checks that the request contains valid data.
func (r previewTripRequest) Validate() error {
	if strings.TrimSpace(r.Origin) == "" {
		return errors.New("origin is required")

	}
	if strings.TrimSpace(r.Destination) == "" {
		return errors.New("destination is required")

	}
	if strings.EqualFold(
		strings.TrimSpace(r.Origin),
		strings.TrimSpace(r.Destination),
	) {
		return errors.New("origin and destination must be different")
	}
	return nil

}

// previewTripResponse is the JSON body returned for a successful preview.
type previewTripResponse struct {
	DistanceKM   float64 `json:"distance_km"`
	FareEstimate float64 `json:"fare_estimate"`
	ETAMinutes   int     `json:"eta_minutes"`
}

// PreviewTrip handles POST /api/v1/trips/preview.
func (h *TripHandler) PreviewTrip(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"request body must be valid JSON",
			reqID,
		)
		return

	}
	if err := req.Validate(); err != nil {
		response.Error(w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			err.Error(),
			reqID,
		)
		return
	}
	// TODO Phase 5: replace this with a real gRPC call to the Trip Service.
	resp := previewTripResponse{
		DistanceKM:   25.0,
		FareEstimate: 45.00,
		ETAMinutes:   35,
	}
	response.JSON(w, http.StatusOK, resp)

}
