package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/parkertr/tipping/internal/constants"
	"github.com/parkertr/tipping/internal/infrastructure/api/server"
	"github.com/parkertr/tipping/internal/infrastructure/database"
	"github.com/parkertr/tipping/internal/infrastructure/eventstore"
	"github.com/parkertr/tipping/internal/infrastructure/repository/postgres"
)

func main() {
	// Connect to the database
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(db)

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	matchRepo := postgres.NewMatchRepository(db)
	predictionRepo := postgres.NewPredictionRepository(db)

	// Create event store
	eventStore, err := eventstore.NewPostgresEventStore(db)
	if err != nil {
		log.Fatalf("Failed to create event store: %v", err)
	}

	// Create server
	srv := server.NewServer(userRepo, matchRepo, predictionRepo, eventStore)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           srv.Handler(),
		ReadTimeout:       constants.DefaultReadTimeout,
		WriteTimeout:      constants.DefaultWriteTimeout,
		IdleTimeout:       constants.DefaultIdleTimeout,
		MaxHeaderBytes:    constants.DefaultMaxHeaderBytes,
		ReadHeaderTimeout: constants.DefaultHeaderTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server is running on http://localhost%s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultShutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
