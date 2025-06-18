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
	"github.com/parkertr/tipping/internal/infrastructure/repository/postgres"
	"github.com/parkertr/tipping/pkg/auth"
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

	// Create token manager
	tokenMgr := auth.NewTokenManager()

	// Create server
	srv := server.New(userRepo, matchRepo, tokenMgr)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:                         ":8080",
		Handler:                      srv.Router,
		ReadTimeout:                  constants.DefaultReadTimeout,
		WriteTimeout:                 constants.DefaultWriteTimeout,
		IdleTimeout:                  constants.DefaultIdleTimeout,
		MaxHeaderBytes:               constants.DefaultMaxHeaderBytes,
		ReadHeaderTimeout:            constants.DefaultHeaderTimeout,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        nil,
		Protocols:                    nil,
	}

	// Start server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown server
	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
