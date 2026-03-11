package config_test

import (
	"testing"
	"time"

	tripconfig "github.com/bytepharoh/rideflow/internal/trip/config"
)

func TestLoadReadsTripPortsFromEnv(t *testing.T) {
	t.Setenv("TRIP_HTTP_PORT", "9090")
	t.Setenv("TRIP_GRPC_PORT", "6000")
	t.Setenv("TRIP_SHUTDOWN_TIMEOUT_SEC", "45")

	cfg, err := tripconfig.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != 9090 {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, 9090)
	}

	if cfg.GRPCPort != 6000 {
		t.Fatalf("GRPCPort = %d, want %d", cfg.GRPCPort, 6000)
	}

	if cfg.ShutdownTimeout != 45*time.Second {
		t.Fatalf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 45*time.Second)
	}
}

func TestLoadUsesDefaultsWhenEnvVarsNotSet(t *testing.T) {
	t.Setenv("TRIP_HTTP_PORT", "")
	t.Setenv("TRIP_GRPC_PORT", "")
	t.Setenv("TRIP_SHUTDOWN_TIMEOUT_SEC", "")

	cfg, err := tripconfig.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != 8082 {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, 8080)
	}

	if cfg.GRPCPort != 50051 {
		t.Fatalf("GRPCPort = %d, want %d", cfg.GRPCPort, 50051)
	}

	if cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 30*time.Second)
	}
}

func TestLoadReturnsErrorForInvalidHTTPPort(t *testing.T) {
	t.Setenv("TRIP_HTTP_PORT", "not-a-number")

	_, err := tripconfig.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}

func TestLoadReturnsErrorForInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("TRIP_SHUTDOWN_TIMEOUT_SEC", "not-a-number")

	_, err := tripconfig.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}
