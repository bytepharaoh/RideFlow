// Package config defines and loads all configuration for the Trip Service.
// Configuration is loaded once at startup in main.go.
// If any required value is missing or invalid, Load() returns an error
// and the service refuses to start. This is called "fail fast" —
// better to crash clearly at boot than fail mysteriously under load.

package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
)

type Config struct {
	ServiceName     string
	HTTPPort        int
	GRPCPort        int
	MongoURI        string
	MongoDB         string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	port, err := pkgconfig.GetInt("TRIP_HTTP_PORT", 8082)
	if err != nil {
		return nil, fmt.Errorf("trip config: %w", err)
	}
	grpcPort, err := pkgconfig.GetInt("TRIP_GRPC_PORT", 50051)
	if err != nil {
		return nil, fmt.Errorf("trip config: %w", err)
	}
	shutdownTimeoutSeconds, err := pkgconfig.GetInt("TRIP_SHUTDOWN_TIMEOUT_SEC", 30)
	if err != nil {
		return nil, fmt.Errorf("trip config: %w", err)
	}
	return &Config{
		ServiceName:     "trip",
		HTTPPort:        port,
		GRPCPort:        grpcPort,
		MongoURI:        pkgconfig.GetString("TRIP_MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:         pkgconfig.GetString("TRIP_MONGO_DB", "rideflow_trip"),
		LogLevel:        pkgconfig.GetString("TRIP_LOG_LEVEL", "info"),
		ShutdownTimeout: time.Duration(shutdownTimeoutSeconds) * time.Second,
	}, nil
}
