package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
	"github.com/parkertr/tipping/pkg/utils"
)

// Server represents the HTTP server
type Server struct {
	router *mux.Router
	server *http.Server
}

// NewServer creates a new server instance
func NewServer(userRepo repository.UserRepository, matchRepo repository.MatchRepository, predictionRepo repository.PredictionRepository, eventStore handlers.EventStore) *Server {
	router := mux.NewRouter()

	// Create token manager
	tokenManager := auth.NewTokenManager()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenManager, eventStore)
	matchHandler := handlers.NewMatchHandler(eventStore, matchRepo)
	predictionHandler := handlers.NewPredictionHandler(eventStore)

	// Register routes
	authHandler.RegisterRoutes(router)
	matchHandler.RegisterRoutes(router)
	predictionHandler.RegisterRoutes(router)

	return &Server{
		router: router,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", utils.GetEnvOrDefault("PORT", "8080")),
			Handler:      router,
			ReadTimeout:  utils.GetEnvOrDefaultDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: utils.GetEnvOrDefaultDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  utils.GetEnvOrDefaultDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
	}
}

// Start starts the server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Handler returns the main HTTP handler for the server
func (s *Server) Handler() http.Handler {
	return s.router
}
