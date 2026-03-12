package config

import (
	"fmt"

	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
)

type Config struct {
	ServiceName string
	HTTPPort    int
	GRPCPort    int
	MongoURI    string
	MongoDB     string
	LogLevel    string
}

func Load() (*Config, error) {
	httpPort, err := pkgconfig.GetInt("DRIVER_HTTP_PORT", 8083)
	if err != nil {
		return nil, fmt.Errorf("driver config: %w", err)
	}

	grpcPort, err := pkgconfig.GetInt("DRIVER_GRPC_PORT", 50052)
	if err != nil {
		return nil, fmt.Errorf("driver config: %w", err)
	}

	return &Config{
		ServiceName: "driver",
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		MongoURI:    pkgconfig.GetString("DRIVER_MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:     pkgconfig.GetString("DRIVER_MONGO_DB", "rideflow_driver"),
		LogLevel:    pkgconfig.GetString("DRIVER_LOG_LEVEL", "info"),
	}, nil
}
