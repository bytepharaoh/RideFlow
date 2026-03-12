package ws

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServeWSRequiresUserID(t *testing.T) {
	t.Parallel()

	handler := NewHandler(NewManager(slog.New(slog.NewJSONHandler(io.Discard, nil))), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeWS(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	if body := rec.Body.String(); !strings.Contains(body, "user_id is required") {
		t.Fatalf("body = %q, want user_id error", body)
	}
}
