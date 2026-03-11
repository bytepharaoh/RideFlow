package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tripconfig "github.com/bytepharoh/rideflow/internal/trip/config"
	"github.com/bytepharoh/rideflow/internal/trip/server"
	"github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
)

func main() {
	if err := config.LoadEnv(".env"); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	// step 1 : load the configurations
	cfg, err := tripconfig.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting trip service",
		"port", cfg.HTTPPort,
		"grpc_port", cfg.GRPCPort,
		"log_level", cfg.LogLevel,
	)
	srv := server.New(cfg.HTTPPort, cfg.ServiceName, log)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	// signal.Notify registers our channel to receive:
	// - SIGINT  → Ctrl+C from terminal
	// - SIGTERM → Kubernetes stopping the pod

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	log.Info("trip service stopped gracefully")

}
