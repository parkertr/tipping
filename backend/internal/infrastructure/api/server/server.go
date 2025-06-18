package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/infrastructure/api/middleware"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
)

// Server represents the HTTP server.
type Server struct {
	Router    *mux.Router
	server    *http.Server
	userRepo  repository.UserRepository
	matchRepo repository.MatchRepository
	tokenMgr  *auth.TokenManager
}

// New creates a new server instance.
func New(
	userRepo repository.UserRepository,
	matchRepo repository.MatchRepository,
	tokenMgr *auth.TokenManager,
) *Server {
	router := mux.NewRouter()
	s := &Server{
		Router:    router,
		userRepo:  userRepo,
		matchRepo: matchRepo,
		tokenMgr:  tokenMgr,
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
	// Public routes
	s.Router.HandleFunc("/api/v1/health", s.healthCheck).Methods(http.MethodGet)
	s.Router.HandleFunc("/api/v1/users", s.handleRegister).Methods(http.MethodPost)
	s.Router.HandleFunc("/api/v1/users/login", s.handleLogin).Methods(http.MethodPost)

	// Protected routes
	protected := s.Router.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(s.tokenMgr, s.userRepo))

	// User routes
	protected.HandleFunc("/users/me", s.handleGetMe).Methods(http.MethodGet)
	protected.HandleFunc("/users/me", s.handleUpdateMe).Methods(http.MethodPut)

	// Match routes
	protected.HandleFunc("/matches", s.handleCreateMatch).Methods(http.MethodPost)
	protected.HandleFunc("/matches", s.handleListMatches).Methods(http.MethodGet)
	protected.HandleFunc("/matches/{id}", s.handleGetMatch).Methods(http.MethodGet)
	protected.HandleFunc("/matches/{id}", s.handleUpdateMatch).Methods(http.MethodPut)
	protected.HandleFunc("/matches/{id}", s.handleDeleteMatch).Methods(http.MethodDelete)
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

// handleRegister handles user registration.
func (s *Server) handleRegister(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleLogin handles user login.
func (s *Server) handleLogin(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleGetMe handles getting the current user.
func (s *Server) handleGetMe(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleUpdateMe handles updating the current user.
func (s *Server) handleUpdateMe(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleCreateMatch handles creating a new match.
func (s *Server) handleCreateMatch(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleListMatches handles listing matches.
func (s *Server) handleListMatches(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleGetMatch handles getting a match.
func (s *Server) handleGetMatch(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleUpdateMatch handles updating a match.
func (s *Server) handleUpdateMatch(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleDeleteMatch handles deleting a match.
func (s *Server) handleDeleteMatch(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement delete match
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
