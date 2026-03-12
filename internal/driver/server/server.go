package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	driverv1 "github.com/bytepharoh/rideflow/internal/driver/gen/proto/driver"
	drivergrpc "github.com/bytepharoh/rideflow/internal/driver/grpc"
	"github.com/bytepharoh/rideflow/internal/driver/service"
)

type Server struct {
	http     *http.Server
	grpc     *grpc.Server
	grpcAddr string
	logger   *slog.Logger
}

func New(httpPort, grpcPort int, serviceName string, logger *slog.Logger, svc *service.DriverService) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": serviceName,
		}); err != nil {
			logger.Error("failed to encode driver health response", "error", err)
		}
	})

	grpcSrv := grpc.NewServer()
	driverv1.RegisterDriverServiceServer(grpcSrv, drivergrpc.New(svc, logger))
	reflection.Register(grpcSrv)

	return &Server{
		logger:   logger,
		grpcAddr: fmt.Sprintf(":%d", grpcPort),
		grpc:     grpcSrv,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", httpPort),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() <-chan error {
	errCh := make(chan error, 2)

	go func() {
		s.logger.Info("http server listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http: %w", err)
		}
	}()

	go func() {
		var listenConfig net.ListenConfig
		lis, err := listenConfig.Listen(context.Background(), "tcp", s.grpcAddr)
		if err != nil {
			errCh <- fmt.Errorf("grpc listen: %w", err)
			return
		}
		s.logger.Info("grpc server listening", "addr", s.grpcAddr)
		if err := s.grpc.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
	}()

	return errCh
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("shutting down driver service")
	s.grpc.GracefulStop()
	_ = s.http.Shutdown(ctx)
}
