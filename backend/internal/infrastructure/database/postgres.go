package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB() (*sql.DB, error) {
	// Get database connection string from environment variable or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/footy_tipping?sslmode=disable"
	}

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CloseDB closes the database connection
func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("error closing db: %v", err)
	}
}
