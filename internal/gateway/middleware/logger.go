//example : GET /trips/preview
//logger : = method=GET path=/trips/preview status=200 duration_ms=12

package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func NewResponseHandler(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK}
}
func (rw *responseWriter) WriteHeader(status int) {
	if rw.wrote {
		return
	}
	rw.status = status
	rw.wrote = true
	rw.ResponseWriter.WriteHeader(status)
}

// Logger returns middleware that logs every HTTP request.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := NewResponseHandler(w)
			next.ServeHTTP(wrapped, r)
			log.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", GetRequestID(r.Context()),
				"remote_addr", r.RemoteAddr,
			)

		})
	}
}
