package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerRejectsNonGet(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	handler := healthHundler("trip")
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
