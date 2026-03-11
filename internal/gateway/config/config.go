// Package config defines and loads all configuration for the API Gateway.
// The gateway sits at the edge of the system — it talks to clients
// over HTTP and to internal services over gRPC. Its config reflects
// both concerns: what port to expose externally, and where to find
// each internal service.

package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
)

type Config struct {
	ServiceName     string
	HTTPPort        int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	TripServiceAddr string
	LogLevel        string
}

func Load() (*Config, error) {
	port, err := pkgconfig.GetInt("GATEWAY_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("gateway config: %w", err)
	}
	readTimeout, err := pkgconfig.GetInt("GATEWAY_READ_TIMEOUT_SECONDS", 10)
	if err != nil {
		return nil, fmt.Errorf("gateway config: %w", err)
	}
	writeTimeout, err := pkgconfig.GetInt("GATEWAY_WRITE_TIMEOUT_SECONDS", 10)
	if err != nil {
		return nil, fmt.Errorf("gateway config: %w", err)
	}

	idleTimeout, err := pkgconfig.GetInt("GATEWAY_IDLE_TIMEOUT_SECONDS", 60)
	if err != nil {
		return nil, fmt.Errorf("gateway config: %w", err)
	}
	return &Config{
		ServiceName:     "gateway",
		HTTPPort:        port,
		ReadTimeout:     time.Duration(readTimeout) * time.Second,
		WriteTimeout:    time.Duration(writeTimeout) * time.Second,
		IdleTimeout:     time.Duration(idleTimeout) * time.Second,
		TripServiceAddr: pkgconfig.GetString("TRIP_SERVICE_ADDR", "localhost:50051"),
		LogLevel:        pkgconfig.GetString("GATEWAY_LOG_LEVEL", "info"),
	}, nil
}
