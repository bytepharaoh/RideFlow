package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytepharoh/rideflow/internal/gateway/config"
	"github.com/bytepharoh/rideflow/internal/gateway/server"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
)

func main() {
	if err := loadEnv(".env"); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)

	}
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)

	}
	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting api gateway", "port", cfg.HTTPPort)
	srv := server.New(cfg, log)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("api gateway stopped gracefully")
}
func loadEnv(path string) error {
	return pkgconfig.LoadEnv(path)
}
