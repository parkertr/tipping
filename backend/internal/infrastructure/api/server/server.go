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

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow requests from frontend origin
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}

		// Check if origin is allowed
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// New creates a new server instance.
func New(
	userRepo repository.UserRepository,
	matchRepo repository.MatchRepository,
	tokenMgr *auth.TokenManager,
	eventStore eventstore.EventStore,
) *Server {
	router := mux.NewRouter()

	// Add CORS middleware to all routes
	router.Use(corsMiddleware)

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

	// Global OPTIONS handler for CORS preflight requests
	s.Router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers are already set by corsMiddleware
		w.WriteHeader(http.StatusOK)
	})

	// Public routes (no authentication required)
	s.Router.HandleFunc("/api/health", s.healthCheck).Methods(http.MethodGet)

	// Public match routes (viewing matches doesn't require auth)
	s.Router.HandleFunc("/api/matches", matchHandler.ListMatches).Methods("GET")
	s.Router.HandleFunc("/api/matches/{id}", matchHandler.GetMatch).Methods("GET")

	// Auth routes (public)
	authRoutes := s.Router.PathPrefix("/api/auth").Subrouter()
	authHandler.RegisterRoutes(authRoutes)

	// Protected routes use a different prefix to avoid conflicts
	protectedAPI := s.Router.PathPrefix("/api/protected").Subrouter()
	protectedAPI.Use(middleware.AuthMiddleware(s.tokenMgr, s.userRepo))

	// Protected match routes (creating/updating matches requires auth)
	protectedAPI.HandleFunc("/matches", matchHandler.CreateMatch).Methods("POST")
	protectedAPI.HandleFunc("/matches/{id}/score", matchHandler.UpdateMatchScore).Methods("PUT")

	// Prediction routes (all require authentication) - also need to update the handler to use /api/protected
	s.registerPredictionRoutes(protectedAPI, predictionHandler)
}

// registerPredictionRoutes registers prediction routes with authentication
func (s *Server) registerPredictionRoutes(r *mux.Router, handler *handlers.PredictionHandler) {
	r.HandleFunc("/predictions", handler.CreatePrediction).Methods("POST")
	r.HandleFunc("/predictions/me", handler.GetUserPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions", handler.GetMatchPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions/me", handler.GetUserPredictionForMatch).Methods("GET")
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
