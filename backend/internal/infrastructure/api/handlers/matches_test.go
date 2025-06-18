package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers/testutil"
)

func TestCreateMatch(t *testing.T) {
	t.Parallel()

	// Create mocks
	matchRepo := testutil.NewMockMatchRepo()
	eventStore := testutil.NewMockEventStore()

	// Create handler
	handler := handlers.NewMatchHandler(eventStore, matchRepo)

	t.Run("Valid match creation", func(t *testing.T) {
		t.Parallel()

		// Create request
		createMatchRequest := handlers.CreateMatchRequest{
			HomeTeam:    "Home",
			AwayTeam:    "Away",
			Date:        time.Now().Add(24 * time.Hour),
			Competition: "Test League",
		}
		reqBody, err := json.Marshal(createMatchRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreateMatch(rr, req)

		// Check response
		if rr.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, rr.Code)
		}

		// Check that match was created
		var response domain.Match
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify match exists in repository
		match, err := matchRepo.GetByID(context.Background(), response.ID)
		if err != nil {
			t.Fatalf("Failed to get match from repository: %v", err)
		}
		if match == nil {
			t.Fatal("Match not found in repository")
		}

		// Check that event was saved
		events, err := eventStore.GetEvents(context.Background(), response.ID)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event to be saved, got %d", len(events))
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		t.Parallel()

		// Create request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreateMatch(rr, req)

		// Check response
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		t.Parallel()

		// Create request with missing fields
		createMatchRequest := handlers.CreateMatchRequest{
			HomeTeam: "Home",
			// Missing AwayTeam, Date, and Competition
		}
		reqBody, err := json.Marshal(createMatchRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreateMatch(rr, req)

		// Check response
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}
