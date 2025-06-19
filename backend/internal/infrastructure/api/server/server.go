package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr/tipping/internal/infrastructure/api/middleware"
	"github.com/parkertr/tipping/internal/infrastructure/eventstore"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
)

// Server represents the HTTP server.
type Server struct {
	Router     *mux.Router
	server     *http.Server
	userRepo   repository.UserRepository
	matchRepo  repository.MatchRepository
	tokenMgr   *auth.TokenManager
	eventStore eventstore.EventStore
}

// New creates a new server instance.
func New(
	userRepo repository.UserRepository,
	matchRepo repository.MatchRepository,
	tokenMgr *auth.TokenManager,
	eventStore eventstore.EventStore,
) *Server {
	router := mux.NewRouter()
	s := &Server{
		Router:     router,
		userRepo:   userRepo,
		matchRepo:  matchRepo,
		tokenMgr:   tokenMgr,
		eventStore: eventStore,
	}

	s.registerRoutes()

	return s
}

// Start starts the server on the specified port.
func (s *Server) Start(port int) error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// registerRoutes registers all the routes for the server.
func (s *Server) registerRoutes() {
	// Create handlers
	authHandler := handlers.NewAuthHandler(s.userRepo, s.tokenMgr, s.eventStore)
	matchHandler := handlers.NewMatchHandler(s.eventStore, s.matchRepo)
	predictionHandler := handlers.NewPredictionHandler(s.eventStore, s.matchRepo)

	// Public routes (no authentication required)
	api := s.Router.PathPrefix("/api").Subrouter()

	// Health check
	api.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	// Auth routes
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authHandler.RegisterRoutes(authRoutes)

	// Protected routes (authentication required)
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(s.tokenMgr, s.userRepo))

	// Match routes
	matchHandler.RegisterRoutes(protected)

	// Prediction routes (these will be updated to use authenticated user)
	predictionHandler.RegisterRoutes(protected)
}

// healthCheck handles the health check endpoint.
func (s *Server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
