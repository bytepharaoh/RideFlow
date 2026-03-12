package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/bytepharoh/rideflow/internal/gateway/client"
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
		{name: "origin blank", request: previewTripRequest{Origin: "   ", Destination: "Alexandria"}, wantErr: "origin is required"},
		{name: "destination blank", request: previewTripRequest{Origin: "Cairo", Destination: "   "}, wantErr: "destination is required"},
		{name: "same places", request: previewTripRequest{Origin: " Cairo ", Destination: "cairo"}, wantErr: "origin and destination must be different"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.request.Validate()
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestCreateTripRequestValidateSuccess(t *testing.T) {
	t.Parallel()

	req := createTripRequest{
		RiderID:     "rider-1",
		Origin:      "Cairo",
		Destination: "Giza",
		OriginLat:   30.04,
		OriginLng:   31.23,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestCreateTripRequestValidateFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request createTripRequest
		wantErr string
	}{
		{name: "missing rider id", request: createTripRequest{Origin: "Cairo", Destination: "Giza"}, wantErr: "rider_id is required"},
		{name: "missing origin", request: createTripRequest{RiderID: "rider-1", Destination: "Giza"}, wantErr: "origin is required"},
		{name: "missing destination", request: createTripRequest{RiderID: "rider-1", Origin: "Cairo"}, wantErr: "destination is required"},
		{name: "same places", request: createTripRequest{RiderID: "rider-1", Origin: "Cairo", Destination: " cairo "}, wantErr: "origin and destination must be different"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.request.Validate()
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestTripHandlerPreviewTripSuccess(t *testing.T) {
	t.Parallel()

	recorder := executeRequest(t, http.MethodPost, "/api/v1/trips/preview", &fakeTripClient{
		previewResult: &client.PreviewTripResult{DistanceKM: 25.0, FareEstimate: 45.0, ETAMinutes: 35},
	}, `{"origin":"Cairo","destination":"Alexandria"}`, "")
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

	if response.Data.DistanceKM != 25.0 || response.Data.FareEstimate != 45.0 || response.Data.ETAMinutes != 35 {
		t.Fatalf("unexpected preview response: %+v", response.Data)
	}
}

func TestTripHandlerPreviewTripFailures(t *testing.T) {
	t.Parallel()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips/preview", nil, `{"origin":`, ""), http.StatusBadRequest, "INVALID_REQUEST", "request body must be valid JSON")
	})

	t.Run("validation failure", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips/preview", nil, `{"origin":" ","destination":"Alexandria"}`, ""), http.StatusBadRequest, "VALIDATION_ERROR", "origin is required")
	})

	t.Run("service unavailable", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips/preview", &fakeTripClient{
			previewErr: errors.New("service unavailable: dial tcp timeout"),
		}, `{"origin":"Cairo","destination":"Giza"}`, ""), http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "trip service is currently unavailable, please try again")
	})
}

func TestTripHandlerCreateTripSuccess(t *testing.T) {
	t.Parallel()

	recorder := executeRequest(t, http.MethodPost, "/api/v1/trips", &fakeTripClient{
		createResult: &client.CreateTripResult{TripID: "trip-1", Status: "requested", FareEstimate: 30.0},
	}, `{"rider_id":"rider-1","origin":"Cairo","destination":"Giza","origin_lat":30.04,"origin_lng":31.23}`, "")
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	var response struct {
		Data client.CreateTripResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Data.TripID != "trip-1" || response.Data.Status != "requested" || response.Data.FareEstimate != 30.0 {
		t.Fatalf("unexpected create response: %+v", response.Data)
	}
}

func TestTripHandlerCreateTripFailures(t *testing.T) {
	t.Parallel()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips", nil, `{"rider_id":`, ""), http.StatusBadRequest, "INVALID_REQUEST", "request body must be valid JSON")
	})

	t.Run("validation failure", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips", nil, `{"origin":"Cairo","destination":"Giza"}`, ""), http.StatusBadRequest, "VALIDATION_ERROR", "rider_id is required")
	})

	t.Run("service unavailable", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodPost, "/api/v1/trips", &fakeTripClient{
			createErr: errors.New("service unavailable: trip service unavailable"),
		}, `{"rider_id":"rider-1","origin":"Cairo","destination":"Giza","origin_lat":30.04,"origin_lng":31.23}`, ""), http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "trip service is currently unavailable")
	})
}

func TestTripHandlerGetTripSuccess(t *testing.T) {
	t.Parallel()

	recorder := executeRequest(t, http.MethodGet, "/api/v1/trips/trip-1", &fakeTripClient{
		getResult: &client.GetTripResult{
			TripID:       "trip-1",
			RiderID:      "rider-1",
			DriverID:     "driver-1",
			Origin:       "Cairo",
			Destination:  "Giza",
			Status:       "accepted",
			FareEstimate: 30.0,
			FinalFare:    0,
		},
	}, "", "trip-1")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Data client.GetTripResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Data.TripID != "trip-1" {
		t.Fatalf("trip_id = %q, want %q", response.Data.TripID, "trip-1")
	}
}

func TestTripHandlerGetTripFailures(t *testing.T) {
	t.Parallel()

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodGet, "/api/v1/trips/", nil, "", ""), http.StatusBadRequest, "VALIDATION_ERROR", "trip id is required")
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodGet, "/api/v1/trips/trip-404", &fakeTripClient{
			getErr: errors.New("not found: trip trip-404 not found"),
		}, "", "trip-404"), http.StatusNotFound, "NOT_FOUND", "trip not found")
	})

	t.Run("service unavailable", func(t *testing.T) {
		t.Parallel()
		assertErrorResponse(t, executeRequest(t, http.MethodGet, "/api/v1/trips/trip-1", &fakeTripClient{
			getErr: errors.New("service unavailable: trip service unavailable"),
		}, "", "trip-1"), http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "trip service is currently unavailable")
	})
}

func executeRequest(t *testing.T, method string, path string, tripClient tripClient, body string, tripID string) *httptest.ResponseRecorder {
	t.Helper()

	if tripClient == nil {
		tripClient = &fakeTripClient{}
	}

	handler := NewTripHandler(tripClient)
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if tripID != "" {
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", tripID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	}

	recorder := httptest.NewRecorder()
	wrapped := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case method == http.MethodPost && path == "/api/v1/trips/preview":
			handler.PreviewTrip(w, r)
		case method == http.MethodPost && path == "/api/v1/trips":
			handler.CreateTrip(w, r)
		case method == http.MethodGet:
			handler.GetTrip(w, r)
		default:
			t.Fatalf("unsupported test route %s %s", method, path)
		}
	}))
	wrapped.ServeHTTP(recorder, req)
	return recorder
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantCode string, wantMessage string) {
	t.Helper()

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
	if response.Error.Code != wantCode {
		t.Fatalf("error.code = %q, want %q", response.Error.Code, wantCode)
	}
	if response.Error.Message != wantMessage {
		t.Fatalf("error.message = %q, want %q", response.Error.Message, wantMessage)
	}
	if response.Error.RequestID == "" {
		t.Fatal("error.request_id = empty, want non-empty")
	}
}

type fakeTripClient struct {
	previewResult *client.PreviewTripResult
	previewErr    error
	createResult  *client.CreateTripResult
	createErr     error
	getResult     *client.GetTripResult
	getErr        error
}

func (f *fakeTripClient) PreviewTrip(_ context.Context, _, _ string) (*client.PreviewTripResult, error) {
	if f.previewErr != nil {
		return nil, f.previewErr
	}
	if f.previewResult != nil {
		return f.previewResult, nil
	}
	return &client.PreviewTripResult{}, nil
}

func (f *fakeTripClient) CreateTrip(_ context.Context, _ client.CreateTripInput) (*client.CreateTripResult, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createResult != nil {
		return f.createResult, nil
	}
	return &client.CreateTripResult{}, nil
}

func (f *fakeTripClient) GetTrip(_ context.Context, _ string) (*client.GetTripResult, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.getResult != nil {
		return f.getResult, nil
	}
	return &client.GetTripResult{}, nil
}
