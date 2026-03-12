package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bytepharoh/rideflow/internal/gateway/client"
	"github.com/bytepharoh/rideflow/internal/gateway/middleware"
	"github.com/bytepharoh/rideflow/internal/gateway/response"
)

type TripHandler struct {
	tripClient tripClient
}

type tripClient interface {
	PreviewTrip(ctx context.Context, origin string, destination string) (*client.PreviewTripResult, error)
	CreateTrip(ctx context.Context, input client.CreateTripInput) (*client.CreateTripResult, error)
	GetTrip(ctx context.Context, tripID string) (*client.GetTripResult, error)
}

func NewTripHandler(tripClient tripClient) *TripHandler {
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

type createTripRequest struct {
	RiderID     string  `json:"rider_id"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	OriginLat   float64 `json:"origin_lat"`
	OriginLng   float64 `json:"origin_lng"`
}

func (r createTripRequest) Validate() error {
	if strings.TrimSpace(r.RiderID) == "" {
		return errors.New("rider_id is required")
	}
	if strings.TrimSpace(r.Origin) == "" {
		return errors.New("origin is required")
	}
	if strings.TrimSpace(r.Destination) == "" {
		return errors.New("destination is required")
	}
	if strings.EqualFold(strings.TrimSpace(r.Origin), strings.TrimSpace(r.Destination)) {
		return errors.New("origin and destination must be different")
	}
	return nil
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

func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req createTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "request body must be valid JSON", reqID)
		return
	}

	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), reqID)
		return
	}

	result, err := h.tripClient.CreateTrip(r.Context(), client.CreateTripInput{
		RiderID:     req.RiderID,
		Origin:      req.Origin,
		Destination: req.Destination,
		OriginLat:   req.OriginLat,
		OriginLng:   req.OriginLng,
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "service unavailable") {
			response.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "trip service is currently unavailable", reqID)
			return
		}
		if strings.HasPrefix(err.Error(), "validation") {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", reqID)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tripID := chi.URLParam(r, "id")
	if tripID == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip id is required", reqID)
		return
	}

	result, err := h.tripClient.GetTrip(r.Context(), tripID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "not found") {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "trip not found", reqID)
			return
		}
		if strings.HasPrefix(err.Error(), "service unavailable") {
			response.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "trip service is currently unavailable", reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", reqID)
		return
	}

	response.JSON(w, http.StatusOK, result)
}
