// Package logger provides a shared structured logger for all RideFlow services.
// We use slog from the Go standard library. Every log entry is written
// as JSON to stdout. In production, stdout is captured by Kubernetes
// and forwarded to your log aggregation system (Datadog, Loki, etc).

package logger

import (
	"log/slog"
	"os"
)

func New(serviceName, level string) *slog.Logger {
	var loglevel slog.Level
	if err := loglevel.UnmarshalText([]byte(level)); err != nil {
		loglevel = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     loglevel,
		AddSource: false,
	})
	return slog.New(handler).With("service", serviceName)

}
