package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	tripv1 "github.com/bytepharoh/rideflow/internal/trip/gen/proto/trip"
	tripgrpc "github.com/bytepharoh/rideflow/internal/trip/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	http         *http.Server
	grpc         *grpc.Server
	grpcAddr     string
	logger       *slog.Logger
	servinceName string
}

func New(httpPort int, grpcPort int, serviceName string, logger *slog.Logger) *Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(serviceName))

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	tripGRPCServer := tripgrpc.New(logger)
	tripv1.RegisterTripServiceServer(grpcSrv, tripGRPCServer)

	return &Server{
		http:         httpSrv,
		grpc:         grpcSrv,
		grpcAddr:     fmt.Sprintf(":%d", grpcPort),
		logger:       logger,
		servinceName: serviceName,
	}
}

func (s *Server) Start() <-chan error {
	errCh := make(chan error, 2)
	go func() {
		s.logger.Info("http server listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server: %w", err)

		}
	}()
	// Start gRPC server
	go func() {
		var listenConfig net.ListenConfig
		lis, err := listenConfig.Listen(context.Background(), "tcp", s.grpcAddr)
		if err != nil {
			errCh <- fmt.Errorf("grpc listen: %w", err)
			return
		}

		s.logger.Info("grpc server listening", "addr", s.grpcAddr)
		if err := s.grpc.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()
	return errCh

}
func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("stopping http server")
	s.grpc.GracefulStop()
	if err := s.http.Shutdown(ctx); err != nil {
		s.logger.Error("http shutdown error", "error", err)
	}
	s.logger.Info("trip service stopped")
}
