package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parkertr/tipping/internal/infrastructure/api/server"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	s, err := server.NewServer(db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	}()

	httpServer := &http.Server{
		Addr:                         ":8080",
		Handler:                      s.router,
		ReadTimeout:                  15 * time.Second,
		WriteTimeout:                 15 * time.Second,
		IdleTimeout:                  60 * time.Second,
		MaxHeaderBytes:               1 << 20,
		ReadHeaderTimeout:            5 * time.Second,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        &http.HTTP2Config{},
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

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
