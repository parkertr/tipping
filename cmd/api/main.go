package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/parkertr/tipping/internal/infrastructure/api/server"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
)

func main() {
	// Initialize repositories
	userRepo := repository.NewUserRepository()
	matchRepo := repository.NewMatchRepository()

	// Initialize token manager
	tokenMgr := auth.NewTokenManager()

	// Create and start server
	srv := server.New(userRepo, matchRepo, tokenMgr)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(8080); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown server
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
