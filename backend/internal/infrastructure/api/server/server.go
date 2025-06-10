package server

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/eventstore"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository/postgres"
)

// Server represents the HTTP server
type Server struct {
	router     *mux.Router
	eventStore eventstore.EventStore
	matchRepo  *postgres.MatchRepository
	predRepo   *postgres.PredictionRepository
}

// NewServer creates a new server instance
func NewServer(db *sql.DB) (*Server, error) {
	// Create repositories
	matchRepo := postgres.NewMatchRepository(db)
	predRepo := postgres.NewPredictionRepository(db)

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
	// Create handlers
	matchHandler := handlers.NewMatchHandler(s.eventStore)
	predictionHandler := handlers.NewPredictionHandler(s.eventStore)

	// Match routes
	s.router.HandleFunc("/api/matches", matchHandler.CreateMatch).Methods("POST")
	s.router.HandleFunc("/api/matches", matchHandler.ListMatches).Methods("GET")
	s.router.HandleFunc("/api/matches/{id}", matchHandler.GetMatch).Methods("GET")
	s.router.HandleFunc("/api/matches/{id}/score", matchHandler.UpdateMatchScore).Methods("PUT")

	// Prediction routes
	s.router.HandleFunc("/api/predictions", predictionHandler.CreatePrediction).Methods("POST")
	s.router.HandleFunc("/api/users/{userId}/predictions", predictionHandler.GetUserPredictions).Methods("GET")
	s.router.HandleFunc("/api/matches/{matchId}/predictions", predictionHandler.GetMatchPredictions).Methods("GET")
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Close cleans up any resources used by the server
func (s *Server) Close() error {
	// Add any cleanup logic here if needed
	return nil
}
