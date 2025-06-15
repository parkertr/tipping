package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/parkertr/tipping/internal/infrastructure/api/server"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	// Create mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Test server creation
	srv, err := server.NewServer(db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if srv == nil {
		t.Fatal("Expected server to be created")
	}
}

func TestServerRoutes(t *testing.T) {
	t.Parallel()
	// Create mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create server
	srv, err := server.NewServer(db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test routes
	testCases := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/api/auth/google", http.StatusTemporaryRedirect},
		{"GET", "/api/auth/google/callback", http.StatusBadRequest},
		{"POST", "/api/auth/refresh", http.StatusUnauthorized},
		{"GET", "/api/auth/me", http.StatusUnauthorized},
		{"PUT", "/api/auth/me", http.StatusUnauthorized},
		{"POST", "/api/auth/me/deactivate", http.StatusUnauthorized},
		{"GET", "/api/auth/me/stats", http.StatusUnauthorized},
		{"GET", "/api/auth/me/ranking", http.StatusUnauthorized},
		{"POST", "/api/matches", http.StatusUnauthorized},
		{"GET", "/api/matches", http.StatusUnauthorized},
		{"GET", "/api/matches/123", http.StatusUnauthorized},
		{"PUT", "/api/matches/123/score", http.StatusUnauthorized},
		{"POST", "/api/predictions", http.StatusUnauthorized},
		{"GET", "/api/users/123/predictions", http.StatusUnauthorized},
		{"GET", "/api/matches/123/predictions", http.StatusUnauthorized},
		{"GET", "/api/matches/123/predictions/456", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)
			if rr.Code != tc.code {
				t.Errorf("Expected status code %d, got %d", tc.code, rr.Code)
			}
		})
	}
}

func TestServerClose(t *testing.T) {
	t.Parallel()
	// Create mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create server
	srv, err := server.NewServer(db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test close
	if err := srv.Close(); err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()
	// Create mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create server
	srv, err := server.NewServer(db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test middleware
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
