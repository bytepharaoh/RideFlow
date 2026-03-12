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
	driverconsumer "github.com/bytepharoh/rideflow/internal/driver/consumer"
	"github.com/bytepharoh/rideflow/internal/driver/repository"
	"github.com/bytepharoh/rideflow/internal/driver/server"
	"github.com/bytepharoh/rideflow/internal/driver/service"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

func main() {
	cfg, log := mustLoadDriverBootstrap()
	log.Info("starting driver service", "http_port", cfg.HTTPPort, "grpc_port", cfg.GRPCPort)

	rabbitConn, publisher, consumerConn := mustSetupDriverMessaging(cfg, log)
	defer closeDriverResource(log, "rabbitmq connection", rabbitConn.Close)
	defer closeDriverResource(log, "rabbitmq publisher", publisher.Close)
	defer closeDriverResource(log, "rabbitmq consumer", consumerConn.Close)

	repo := repository.NewInMemoryDriverRepository()
	svc := service.New(repo, uuid.NewString, log)
	tripConsumer := driverconsumer.NewTripConsumer(svc, consumerConn, publisher, log)
	srv := server.New(cfg.HTTPPort, cfg.GRPCPort, cfg.ServiceName, log, svc)

	runDriverProcess(log, 30*time.Second, srv.Start(), srv.Shutdown, tripConsumer)
}

func mustLoadDriverBootstrap() (*driverconfig.Config, *slog.Logger) {
	if err := pkgconfig.LoadEnv(".env"); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	cfg, err := driverconfig.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	return cfg, logger.New(cfg.ServiceName, cfg.LogLevel)
}

func mustSetupDriverMessaging(cfg *driverconfig.Config, log *slog.Logger) (*rabbitmq.Connection, *rabbitmq.Publisher, *rabbitmq.Consumer) {
	rabbitConn, err := rabbitmq.NewConnection(cfg.RabbitMQURL, log)
	if err != nil {
		log.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}

	publisher, err := rabbitmq.NewPublisher(rabbitConn, events.Exchange)
	if err != nil {
		log.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}

	consumerConn, err := rabbitmq.NewConsumer(rabbitConn, events.Exchange, log)
	if err != nil {
		log.Error("failed to create consumer", "error", err)
		os.Exit(1)
	}

	return rabbitConn, publisher, consumerConn
}

func runDriverProcess(
	log *slog.Logger,
	shutdownTimeout time.Duration,
	errCh <-chan error,
	shutdown func(context.Context),
	tripConsumer *driverconsumer.TripConsumer,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := tripConsumer.Start(ctx); err != nil {
			log.Error("trip consumer error", "error", err)
		}
	}()

	waitForDriverSignal(log, errCh)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	shutdown(shutdownCtx)
	log.Info("driver service stopped")
}

func waitForDriverSignal(log *slog.Logger, errCh <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		log.Error("server error", "error", err)
	}
}

func closeDriverResource(log *slog.Logger, name string, closeFn func() error) {
	if err := closeFn(); err != nil {
		log.Error("failed to close resource", "resource", name, "error", err)
	}
}
