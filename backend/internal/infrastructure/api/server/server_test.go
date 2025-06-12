package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewServer(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	mock.ExpectClose()
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	// Set up expectations for event store initialization
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Test server creation
	server, err := NewServer(db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if server == nil {
		t.Fatalf("expected server to be non-nil")
	}
	if server.router == nil {
		t.Errorf("expected router to be non-nil")
	}
	if server.eventStore == nil {
		t.Errorf("expected eventStore to be non-nil")
	}
	if server.matchRepo == nil {
		t.Errorf("expected matchRepo to be non-nil")
	}
	if server.predRepo == nil {
		t.Errorf("expected predRepo to be non-nil")
	}
}

func TestServerRoutes(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	mock.ExpectClose()
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	// Set up expectations for event store initialization
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Create server
	server, err := NewServer(db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test cases for different routes
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"List Matches", "GET", "/api/matches", http.StatusOK},
		{"Get Match", "GET", "/api/matches/123", http.StatusOK},
		{"Create Match", "POST", "/api/matches", http.StatusOK},
		{"Update Match Score", "PUT", "/api/matches/123/score", http.StatusOK},
		{"Create Prediction", "POST", "/api/predictions", http.StatusOK},
		{"Get User Predictions", "GET", "/api/users/123/predictions", http.StatusOK},
		{"Get Match Predictions", "GET", "/api/matches/123/predictions", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			server.ServeHTTP(rr, req)
			if rr.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", tc.method, tc.path)
			}
		})
	}
}

func TestServerClose(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	mock.ExpectClose()
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	// Set up expectations for event store initialization
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Create server
	server, err := NewServer(db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test server close
	err = server.Close()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMiddleware(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	mock.ExpectClose()
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	// Set up expectations for event store initialization
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Create server
	server, err := NewServer(db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test middleware
	req := httptest.NewRequest("GET", "/api/matches", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	// Check CORS headers
	if _, ok := rr.Header()["Access-Control-Allow-Origin"]; !ok {
		t.Errorf("expected Access-Control-Allow-Origin header to be set")
	}
	if _, ok := rr.Header()["Access-Control-Allow-Methods"]; !ok {
		t.Errorf("expected Access-Control-Allow-Methods header to be set")
	}
	if _, ok := rr.Header()["Access-Control-Allow-Headers"]; !ok {
		t.Errorf("expected Access-Control-Allow-Headers header to be set")
	}
}
