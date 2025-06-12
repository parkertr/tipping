package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/api/handlers/mocks"
	"github.com/parkertr2/footy-tipping/pkg/events"
	"github.com/stretchr/testify/mock"
)

// Test cases for match-related handlers
func TestCreateMatch(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewMatchHandler(mockStore)

	// Test case 1: Valid match creation
	t.Run("Valid match creation", func(t *testing.T) {
		matchDate := time.Now().Add(24 * time.Hour)
		match := domain.Match{
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        matchDate,
			Competition: "Premier League",
			Status:      domain.MatchStatusScheduled,
		}

		// Create request
		body, _ := json.Marshal(match)
		req := httptest.NewRequest("POST", "/api/matches", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		// Set up mock expectation - use mock.MatchedBy for dynamic fields
		mockStore.On("SaveEvent", req.Context(), mock.MatchedBy(func(event *events.Event) bool {
			// Type check
			if event.Type != "MatchCreated" {
				return false
			}

			// Cast data to MatchCreated
			matchCreated, ok := event.Data.(events.MatchCreated)
			if !ok {
				return false
			}

			// Check static fields
			return matchCreated.HomeTeam == match.HomeTeam &&
				matchCreated.AwayTeam == match.AwayTeam &&
				matchCreated.Competition == match.Competition
		})).Return(nil)

		// Handle request
		handler.CreateMatch(rr, req)

		// Assert response
		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/matches", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		handler.CreateMatch(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

func TestGetMatch(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewMatchHandler(mockStore)

	// Test case 1: Match found
	t.Run("Match found", func(t *testing.T) {
		matchID := "123"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID, nil)
		rr := httptest.NewRecorder()

		// Add route parameters to request context
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		// Create event with mock data
		matchCreated := events.MatchCreated{
			ID:          matchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now(),
			Competition: "Premier League",
		}
		event := &events.Event{
			ID:        "event123",
			Type:      "MatchCreated",
			Data:      matchCreated,
			Timestamp: time.Now(),
			Version:   1,
		}

		mockStore.On("GetEvents", req.Context(), matchID).Return([]*events.Event{event}, nil)

		// Handle request
		handler.GetMatch(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Match not found
	t.Run("Match not found", func(t *testing.T) {
		matchID := "456"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID, nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		mockStore.On("GetEvents", req.Context(), matchID).Return([]*events.Event{}, nil)

		handler.GetMatch(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})
}

func TestListMatches(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewMatchHandler(mockStore)

	// Test case: List all matches
	t.Run("List all matches", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/matches", nil)
		rr := httptest.NewRecorder()

		// Create events with mock data
		matchCreated1 := events.MatchCreated{
			ID:          "123",
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now(),
			Competition: "Premier League",
		}
		event1 := &events.Event{
			ID:        "event123",
			Type:      "MatchCreated",
			Data:      matchCreated1,
			Timestamp: time.Now(),
			Version:   1,
		}

		matchCreated2 := events.MatchCreated{
			ID:          "456",
			HomeTeam:    "Team C",
			AwayTeam:    "Team D",
			Date:        time.Now(),
			Competition: "Premier League",
		}
		event2 := &events.Event{
			ID:        "event456",
			Type:      "MatchCreated",
			Data:      matchCreated2,
			Timestamp: time.Now(),
			Version:   1,
		}

		mockStore.On("GetEventsByType", req.Context(), "MatchCreated").Return([]*events.Event{event1, event2}, nil)

		handler.ListMatches(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})
}

func TestUpdateMatchScore(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewMatchHandler(mockStore)

	// Test case 1: Valid score update
	t.Run("Valid score update", func(t *testing.T) {
		matchID := "123"
		score := struct {
			HomeGoals int `json:"homeGoals"`
			AwayGoals int `json:"awayGoals"`
		}{
			HomeGoals: 2,
			AwayGoals: 1,
		}

		// Create request
		body, _ := json.Marshal(score)
		req := httptest.NewRequest("PUT", "/api/matches/"+matchID+"/score", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		// Set up mock expectation - use mock.MatchedBy for dynamic fields
		mockStore.On("SaveEvent", req.Context(), mock.MatchedBy(func(event *events.Event) bool {
			// Type check
			if event.Type != "MatchScoreUpdated" {
				return false
			}

			// Cast data to MatchScoreUpdated
			scoreUpdated, ok := event.Data.(events.MatchScoreUpdated)
			if !ok {
				return false
			}

			// Check static fields
			return scoreUpdated.MatchID == matchID &&
				scoreUpdated.HomeGoals == score.HomeGoals &&
				scoreUpdated.AwayGoals == score.AwayGoals
		})).Return(nil)

		handler.UpdateMatchScore(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		matchID := "123"
		req := httptest.NewRequest("PUT", "/api/matches/"+matchID+"/score", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		handler.UpdateMatchScore(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}
