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
	tripconsumer "github.com/bytepharoh/rideflow/internal/trip/consumer"
	"github.com/bytepharoh/rideflow/internal/trip/domain"
	"github.com/bytepharoh/rideflow/internal/trip/repository"
	"github.com/bytepharoh/rideflow/internal/trip/server"
	"github.com/bytepharoh/rideflow/internal/trip/service"
	pkgconfig "github.com/bytepharoh/rideflow/pkg/config"
	"github.com/bytepharoh/rideflow/pkg/logger"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

func main() {
	cfg, log := mustLoadTripBootstrap()
	log.Info("starting trip service", "http_port", cfg.HTTPPort, "grpc_port", cfg.GRPCPort)

	rabbitConn, publisher, consumerConn := mustSetupTripMessaging(cfg, log)
	defer closeWithLog(log, "rabbitmq connection", rabbitConn.Close)
	defer closeWithLog(log, "rabbitmq publisher", publisher.Close)
	defer closeWithLog(log, "rabbitmq consumer", consumerConn.Close)

	svc := buildTripService(publisher, log)
	driverConsumer := tripconsumer.NewDriverConsumer(svc, consumerConn, log)
	srv := server.New(cfg.HTTPPort, cfg.GRPCPort, cfg.ServiceName, log, svc)

	runTripProcess(log, cfg.ShutdownTimeout, srv.Start(), srv.Shutdown, driverConsumer)
}

func mustLoadTripBootstrap() (*tripconfig.Config, *slog.Logger) {
	if err := pkgconfig.LoadEnv(".env"); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	cfg, err := tripconfig.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	return cfg, logger.New(cfg.ServiceName, cfg.LogLevel)
}

func mustSetupTripMessaging(cfg *tripconfig.Config, log *slog.Logger) (*rabbitmq.Connection, *rabbitmq.Publisher, *rabbitmq.Consumer) {
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

func buildTripService(publisher *rabbitmq.Publisher, log *slog.Logger) *service.TripService {
	repo := repository.NewInMemoryTripRepository()
	fareCalc := domain.NewCalculator(domain.DefaultFareConfig())

	return service.New(repo, fareCalc, uuid.NewString, publisher, log)
}

func runTripProcess(
	log *slog.Logger,
	shutdownTimeout time.Duration,
	errCh <-chan error,
	shutdown func(context.Context),
	driverConsumer *tripconsumer.DriverConsumer,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := driverConsumer.Start(ctx); err != nil {
			log.Error("driver consumer error", "error", err)
		}
	}()

	waitForSignal(log, errCh)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	shutdown(shutdownCtx)
	log.Info("trip service stopped")
}

func waitForSignal(log *slog.Logger, errCh <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		log.Error("server error", "error", err)
	}
}

func closeWithLog(log *slog.Logger, name string, closeFn func() error) {
	if err := closeFn(); err != nil {
		log.Error("failed to close resource", "resource", name, "error", err)
	}
}
