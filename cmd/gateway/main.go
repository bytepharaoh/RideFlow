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

	"github.com/bytepharoh/rideflow/internal/gateway/client"
	"github.com/bytepharoh/rideflow/internal/gateway/config"
	gatewayconsumer "github.com/bytepharoh/rideflow/internal/gateway/consumer"
	"github.com/bytepharoh/rideflow/internal/gateway/server"
	"github.com/bytepharoh/rideflow/internal/gateway/ws"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

func main() {
	cfg, log := mustLoadGatewayBootstrap()
	log.Info("starting api gateway", "port", cfg.HTTPPort)

	tripClient := mustCreateTripClient(cfg, log)
	log.Info("trip service client created", "addr", cfg.TripServiceAddr)

	rabbitConn, mqConsumer := mustSetupGatewayMessaging(cfg, log)
	defer closeWithLog(log, "rabbitmq connection", rabbitConn.Close)
	defer closeWithLog(log, "rabbitmq consumer", mqConsumer.Close)

	srv, eventConsumer := buildGatewayRuntime(cfg, log, tripClient, mqConsumer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}()
	go eventConsumer.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("shutdown signal received", "signal", sig.String())
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("api gateway stopped gracefully")
}

func mustLoadGatewayBootstrap() (*config.Config, *slog.Logger) {
	if err := loadEnv(".env"); err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	return cfg, logger.New(cfg.ServiceName, cfg.LogLevel)
}

func mustCreateTripClient(cfg *config.Config, log *slog.Logger) *client.TripClient {
	tripClient, err := client.NewTripClient(cfg.TripServiceAddr)
	if err != nil {
		log.Error("failed to create trip client", "error", err)
		os.Exit(1)
	}

	return tripClient
}

func mustSetupGatewayMessaging(cfg *config.Config, log *slog.Logger) (*rabbitmq.Connection, *rabbitmq.Consumer) {
	rabbitConn, err := rabbitmq.NewConnection(cfg.RabbitMQURL, log)
	if err != nil {
		log.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}

	mqConsumer, err := rabbitmq.NewConsumer(rabbitConn, events.Exchange, log)
	if err != nil {
		log.Error("failed to create rabbitmq consumer", "error", err)
		os.Exit(1)
	}

	return rabbitConn, mqConsumer
}

func buildGatewayRuntime(
	cfg *config.Config,
	log *slog.Logger,
	tripClient *client.TripClient,
	mqConsumer *rabbitmq.Consumer,
) (*server.Server, *gatewayconsumer.EventConsumer) {
	wsManager := ws.NewManager(log)
	wsHandler := ws.NewHandler(wsManager, log)
	eventConsumer := gatewayconsumer.NewEventConsumer(wsManager, mqConsumer, log)
	srv := server.New(cfg, log, tripClient, wsHandler)

	return srv, eventConsumer
}

func loadEnv(path string) error {
	return pkgconfig.LoadEnv(path)
}

func closeWithLog(log *slog.Logger, name string, closeFn func() error) {
	if err := closeFn(); err != nil {
		log.Error("failed to close resource", "resource", name, "error", err)
	}
}
