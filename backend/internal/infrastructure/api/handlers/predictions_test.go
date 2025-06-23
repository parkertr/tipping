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
	"github.com/parkertr/tipping/internal/infrastructure/api/middleware"
)

func TestCreatePrediction(t *testing.T) {
	t.Parallel()

	// Create mocks
	matchRepo := testutil.NewMockMatchRepo()
	eventStore := testutil.NewMockEventStore()

	// Create handler
	handler := handlers.NewPredictionHandler(eventStore, matchRepo)

	// Create test user
	testUser := &domain.User{
		ID:       "user1",
		GoogleID: "google123",
		Email:    "test@example.com",
		Name:     "Test User",
		Picture:  "https://example.com/avatar.jpg",
	}

	t.Run("Valid prediction creation", func(t *testing.T) {
		t.Parallel()

		// Create a match first
		match := &domain.Match{
			ID:          "match1",
			HomeTeam:    "Home",
			AwayTeam:    "Away",
			Date:        time.Now().Add(24 * time.Hour),
			Competition: "Test League",
			Status:      domain.MatchStatusScheduled,
			Score:       &domain.Score{HomeGoals: 0, AwayGoals: 0},
		}
		if err := matchRepo.Create(context.Background(), match); err != nil {
			t.Fatalf("Failed to create test match: %v", err)
		}

		// Create request (no UserID in body anymore)
		createPredictionRequest := handlers.CreatePredictionRequest{
			MatchID:   match.ID,
			HomeGoals: 2,
			AwayGoals: 1,
		}
		reqBody, err := json.Marshal(createPredictionRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/predictions", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Add user to context (simulate authentication middleware)
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
		req = req.WithContext(ctx)

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreatePrediction(rr, req)

		// Check response
		if rr.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, rr.Code)
		}

		// Decode response to get prediction ID
		var response domain.Prediction
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
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

	t.Run("Match not found", func(t *testing.T) {
		t.Parallel()

		// Create request
		createPredictionRequest := handlers.CreatePredictionRequest{
			MatchID:   "nonexistent",
			HomeGoals: 2,
			AwayGoals: 1,
		}
		reqBody, err := json.Marshal(createPredictionRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/predictions", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Add user to context
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
		req = req.WithContext(ctx)

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreatePrediction(rr, req)

		// Check response
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		t.Parallel()

		// Create request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/predictions", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		// Add user to context
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
		req = req.WithContext(ctx)

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreatePrediction(rr, req)

		// Check response
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		t.Parallel()

		// Create request with missing MatchID
		createPredictionRequest := handlers.CreatePredictionRequest{
			HomeGoals: 2,
			AwayGoals: 1,
			// Missing MatchID
		}
		reqBody, err := json.Marshal(createPredictionRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/predictions", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Add user to context
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
		req = req.WithContext(ctx)

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreatePrediction(rr, req)

		// Check response
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("User not authenticated", func(t *testing.T) {
		t.Parallel()

		// Create request
		createPredictionRequest := handlers.CreatePredictionRequest{
			MatchID:   "match1",
			HomeGoals: 2,
			AwayGoals: 1,
		}
		reqBody, err := json.Marshal(createPredictionRequest)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/predictions", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Don't add user to context (simulate unauthenticated request)

		// Handle request
		rr := httptest.NewRecorder()
		handler.CreatePrediction(rr, req)

		// Check response
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})
}
