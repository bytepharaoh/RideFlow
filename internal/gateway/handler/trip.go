package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bytepharoh/rideflow/internal/gateway/client"
	"github.com/bytepharoh/rideflow/internal/gateway/middleware"
	"github.com/bytepharoh/rideflow/internal/gateway/response"
)

type TripHandler struct {
	tripClient tripPreviewer
}

type tripPreviewer interface {
	PreviewTrip(ctx context.Context, origin string, destination string) (*client.PreviewTripResult, error)
}

func NewTripHandler(tripClient tripPreviewer) *TripHandler {
	return &TripHandler{
		tripClient: tripClient,
	}
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
	ETAMinutes   int32   `json:"eta_minutes"`
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
		response.Error(w, http.StatusBadRequest,
			"VALIDATION_ERROR", err.Error(), reqID)
		return
	}

	result, err := h.tripClient.PreviewTrip(r.Context(), req.Origin, req.Destination)
	if err != nil {
		if strings.HasPrefix(err.Error(), "service unavailable") {

			response.Error(w, http.StatusServiceUnavailable,
				"SERVICE_UNAVAILABLE",
				"trip service is currently unavailable, please try again",
				reqID)
			return
		}
		if strings.HasPrefix(err.Error(), "validation") {
			response.Error(w, http.StatusBadRequest,
				"VALIDATION_ERROR", err.Error(), reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"an unexpected error occurred",
			reqID)
		return
	}

	response.JSON(w, http.StatusOK, previewTripResponse{
		DistanceKM:   result.DistanceKM,
		FareEstimate: result.FareEstimate,
		ETAMinutes:   result.ETAMinutes,
	})

}
