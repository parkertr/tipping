package server

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/auth"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr/tipping/internal/infrastructure/api/middleware"
	"github.com/parkertr/tipping/internal/infrastructure/eventstore"
	"github.com/parkertr/tipping/internal/infrastructure/repository/postgres"
)

// Server represents the HTTP server
type Server struct {
	router     *mux.Router
	eventStore eventstore.EventStore
	matchRepo  *postgres.MatchRepository
	predRepo   *postgres.PredictionRepository
	userRepo   *postgres.UserRepository
}

// NewServer creates a new server instance
func NewServer(db *sql.DB) (*Server, error) {
	// Create repositories
	matchRepo := postgres.NewMatchRepository(db)
	predRepo := postgres.NewPredictionRepository(db)
	userRepo := postgres.NewUserRepository(db)

	// Create event store
	eventStore, err := eventstore.NewPostgresEventStore(db)
	if err != nil {
		return nil, err
	}

	// Create server
	s := &Server{
		router:     mux.NewRouter(),
		eventStore: eventStore,
		matchRepo:  matchRepo,
		predRepo:   predRepo,
		userRepo:   userRepo,
	}

	// Add middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(corsMiddleware)

	// Set up routes
	s.setupRoutes()

	return s, nil
}

// setupRoutes configures the server routes
func (s *Server) setupRoutes() {
	// Create token manager
	tokenManager := auth.NewTokenManager()

	// Create handlers
	matchHandler := handlers.NewMatchHandler(s.eventStore, s.matchRepo)
	predHandler := handlers.NewPredictionHandler(s.eventStore)
	authHandler := handlers.NewAuthHandler(tokenManager, s.userRepo, s.eventStore)

	// Public routes
	s.router.HandleFunc("/api/auth/google", authHandler.GoogleLogin).Methods("GET")
	s.router.HandleFunc("/api/auth/google/callback", authHandler.GoogleCallback).Methods("GET")

	// Protected routes
	protected := s.router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(tokenManager, s.userRepo))

	// Auth routes
	protected.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")
	protected.HandleFunc("/auth/me", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/me", authHandler.UpdateProfile).Methods("PUT")

	// Match routes
	protected.HandleFunc("/matches", matchHandler.CreateMatch).Methods("POST")
	protected.HandleFunc("/matches", matchHandler.ListMatches).Methods("GET")
	protected.HandleFunc("/matches/{id}", matchHandler.GetMatch).Methods("GET")
	protected.HandleFunc("/matches/{id}/score", matchHandler.UpdateMatchScore).Methods("PUT")

	// Prediction routes
	protected.HandleFunc("/predictions", predHandler.CreatePrediction).Methods("POST")
	protected.HandleFunc("/users/{userId}/predictions", predHandler.GetUserPredictions).Methods("GET")
	protected.HandleFunc("/matches/{matchId}/predictions", predHandler.GetMatchPredictions).Methods("GET")
	protected.HandleFunc("/matches/{matchId}/predictions/{userId}", predHandler.GetUserPredictionForMatch).Methods("GET")
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

// Close cleans up any resources used by the server
func (s *Server) Close() error {
	// Add any cleanup logic here if needed
	return nil
}
