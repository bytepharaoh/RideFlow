package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	http         *http.Server
	logger       *slog.Logger
	servinceName string
}

func New(port int, serviceName string, logger *slog.Logger) *Server {

	mux := http.NewServeMux()

	s := &Server{
		logger:       logger,
		servinceName: serviceName,
	}
	mux.HandleFunc("/health", healthHandler(serviceName))
	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s

}
func (s *Server) Start() error {
	s.logger.Info("http server listening", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("stopping http server")
	return s.http.Shutdown(ctx)
}
