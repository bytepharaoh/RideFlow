package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	tripconfig "github.com/bytepharoh/rideflow/internal/trip/config"
	"github.com/bytepharoh/rideflow/internal/trip/domain"
	"github.com/bytepharoh/rideflow/internal/trip/repository"
	"github.com/bytepharoh/rideflow/internal/trip/server"
	"github.com/bytepharoh/rideflow/internal/trip/service"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
)

func main() {
	if err := pkgconfig.LoadEnv(".env"); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	cfg, err := tripconfig.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting trip service",
		"http_port", cfg.HTTPPort,
		"grpc_port", cfg.GRPCPort,
	)

	// ── Build the dependency chain ────────────────────────
	//
	// We build from the bottom up:
	// repository → fare calculator → service → grpc server
	//
	// Each layer only depends on the layer below it.
	// main.go is the only place that knows about all layers.

	// 1. Repository — in-memory until Phase 11
	repo := repository.NewInMemoryTripRepository()

	// 2. Fare calculator with default config
	fareCalc := domain.NewCalculator(domain.DefaultFareConfig())

	// 3. Service layer — inject repository and calculator
	svc := service.New(
		repo,
		fareCalc,
		uuid.NewString, // ID generator — produces a real UUID per trip
		log,
	)

	// 4. Build and start the server with the service injected
	srv := server.New(cfg.HTTPPort, cfg.GRPCPort, cfg.ServiceName, log, svc)

	errCh := srv.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		log.Error("server error", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
}
