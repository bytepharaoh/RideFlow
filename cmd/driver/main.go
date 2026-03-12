package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	driverconfig "github.com/bytepharoh/rideflow/internal/driver/config"
	"github.com/bytepharoh/rideflow/internal/driver/repository"
	"github.com/bytepharoh/rideflow/internal/driver/server"
	"github.com/bytepharoh/rideflow/internal/driver/service"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
)

func main() {
	if err := pkgconfig.LoadEnv(".env"); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	cfg, err := driverconfig.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting driver service",
		"http_port", cfg.HTTPPort,
		"grpc_port", cfg.GRPCPort,
	)

	repo := repository.NewInMemoryDriverRepository()
	svc := service.New(repo, uuid.NewString, log)
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
