// Package server contains the HTTP server for the API Gateway.
//
// The server owns the router and middleware chain.
// It is the single place where you can see:
//   - What middleware runs on every request
//   - What routes are registered
//   - What timeouts are configured

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/bytepharoh/rideflow/internal/gateway/client"
	"github.com/bytepharoh/rideflow/internal/gateway/config"
	"github.com/bytepharoh/rideflow/internal/gateway/handler"
	"github.com/bytepharoh/rideflow/internal/gateway/middleware"
	"github.com/bytepharoh/rideflow/internal/gateway/response"
	"github.com/bytepharoh/rideflow/internal/gateway/ws"
)

type Server struct {
	http   *http.Server
	logger *slog.Logger
}

// New builds the server with all routes and middleware registered.
// It does not start listening — call Start() for that.
func New(cfg *config.Config, logger *slog.Logger, tripClient *client.TripClient, wsHandler *ws.Handler) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", healthHandler(cfg.ServiceName))
	r.Get("/ws", wsHandler.ServeWS)

	tripHandler := handler.NewTripHandler(tripClient)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/trips", func(r chi.Router) {
			r.Post("/preview", tripHandler.PreviewTrip)
			r.Post("/", tripHandler.CreateTrip)
			r.Get("/{id}", tripHandler.GetTrip)
		})
	})
	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w,
			http.StatusNotFound,
			"NOT_FOUND",
			fmt.Sprintf("route %s %s does not exist", r.Method, r.URL.Path),
			middleware.GetRequestID(r.Context()),
		)
	})
	// Method not allowed handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			fmt.Sprintf("method %s is not allowed on %s", r.Method, r.URL.Path),
			middleware.GetRequestID(r.Context()),
		)
	})
	return &Server{
		logger: logger,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler:      r,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("http server listening", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("stopping http server")
	return s.http.Shutdown(ctx)
}
