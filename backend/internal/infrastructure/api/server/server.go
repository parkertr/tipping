package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
	"github.com/parkertr/tipping/pkg/utils"
)

// Server represents the HTTP server.
type Server struct {
	router *mux.Router
	server *http.Server
}

// NewServer creates a new server instance.
func NewServer(userRepo repository.UserRepository, matchRepo repository.MatchRepository, predictionRepo repository.PredictionRepository, eventStore handlers.EventStore) *Server {
	router := mux.NewRouter()

	// Create token manager
	tokenManager := auth.NewTokenManager()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenManager, eventStore)
	matchHandler := handlers.NewMatchHandler(eventStore, matchRepo)
	predictionHandler := handlers.NewPredictionHandler(eventStore, matchRepo)

	// Create subrouters for authenticated and unauthenticated routes
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Add authentication middleware to API routes
	apiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for certain routes
			if strings.HasPrefix(r.URL.Path, "/api/auth/google") || r.URL.Path == "/api/auth/refresh" {
				next.ServeHTTP(w, r)

				return
			}

			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)

				return
			}

			// Extract token
			tokenString := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			// Validate token
			claims, err := tokenManager.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)

				return
			}

			// Get user from repository
			user, err := userRepo.GetByID(r.Context(), claims.Subject)
			if err != nil {
				http.Error(w, "user not found", http.StatusUnauthorized)

				return
			}

			if user == nil {
				http.Error(w, "user not found", http.StatusUnauthorized)

				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Register routes
	authHandler.RegisterRoutes(authRouter)
	matchHandler.RegisterRoutes(apiRouter)
	predictionHandler.RegisterRoutes(apiRouter)

	return &Server{
		router: router,
		server: &http.Server{
			Addr:         ":" + utils.GetEnvOrDefault("PORT", "8080"),
			Handler:      router,
			ReadTimeout:  utils.GetEnvOrDefaultDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: utils.GetEnvOrDefaultDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  utils.GetEnvOrDefaultDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
	}
}

// Start starts the server.
func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// Handler returns the main HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}
