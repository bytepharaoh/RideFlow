// Package response provides consistent JSON response helpers for the gateway.
//
// Every HTTP handler in the gateway uses these helpers instead of
// writing json.NewEncoder directly. This guarantees that every
// success and error response has exactly the same shape, which
// makes client-side error handling predictable and simple.

package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type envelope struct {
	Data any `json:"data"`
}
type errorBody struct {
	Error errorDetail `json:"error"`
}
type errorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope{Data: data}); err != nil {
		slog.Error("failed to write json response", "error", err)
	}

}
func Error(w http.ResponseWriter, status int, code, message, reqID string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	body := errorBody{
		Error: errorDetail{
			Code:      code,
			Message:   message,
			RequestID: reqID,
		},
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to write error response", "error", err)

	}
}
