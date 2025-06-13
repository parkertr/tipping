package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/api/handlers/mocks"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
	"github.com/stretchr/testify/mock"
)

// Test cases for match-related handlers
func TestCreateMatch(t *testing.T) {
	// Create mock event store and repository
	mockStore := new(mocks.MockEventStore)
	mockRepo := new(mocks.MockMatchRepository)
	handler := NewMatchHandler(mockStore, mockRepo)

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

		// Mock repository create call (from event handler)
		mockRepo.On("Create", req.Context(), mock.AnythingOfType("*domain.Match")).Return(nil)

		// Handle request
		handler.CreateMatch(rr, req)

		// Assert response
		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
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
	// Create mock event store and repository
	mockStore := new(mocks.MockEventStore)
	mockRepo := new(mocks.MockMatchRepository)
	handler := NewMatchHandler(mockStore, mockRepo)

	// Test case 1: Match found in repository
	t.Run("Match found in repository", func(t *testing.T) {
		matchID := "123"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID, nil)
		rr := httptest.NewRecorder()

		// Add route parameters to request context
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		// Create mock match
		match := &domain.Match{
			ID:          matchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now(),
			Competition: "Premier League",
			Status:      domain.MatchStatusScheduled,
		}

		mockRepo.On("GetByID", req.Context(), matchID).Return(match, nil)

		// Handle request
		handler.GetMatch(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockRepo.AssertExpectations(t)
	})

	// Test case 2: Match not found, fallback to events
	t.Run("Match not found, fallback to events", func(t *testing.T) {
		matchID := "456"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID, nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		// Mock repository returns error (not found)
		mockRepo.On("GetByID", req.Context(), matchID).Return(nil, repository.ErrNotFound)

		// Mock event store returns empty events (not found)
		mockStore.On("GetEvents", req.Context(), matchID).Return([]*events.Event{}, nil)

		handler.GetMatch(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})
}

func TestListMatches(t *testing.T) {
	// Create mock event store and repository
	mockStore := new(mocks.MockEventStore)
	mockRepo := new(mocks.MockMatchRepository)
	handler := NewMatchHandler(mockStore, mockRepo)

	// Test case: List all matches from repository
	t.Run("List all matches from repository", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/api/matches", nil)
		rr := httptest.NewRecorder()

		// Create mock matches
		matches := []*domain.Match{
			{
				ID:          "123",
				HomeTeam:    "Team A",
				AwayTeam:    "Team B",
				Date:        time.Now(),
				Competition: "Premier League",
				Status:      domain.MatchStatusScheduled,
			},
			{
				ID:          "456",
				HomeTeam:    "Team C",
				AwayTeam:    "Team D",
				Date:        time.Now(),
				Competition: "Premier League",
				Status:      domain.MatchStatusScheduled,
			},
		}

		mockRepo.On("List", req.Context(), repository.MatchFilters{}).Return(matches, nil)

		handler.ListMatches(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateMatchScore(t *testing.T) {
	// Create mock event store and repository
	mockStore := new(mocks.MockEventStore)
	mockRepo := new(mocks.MockMatchRepository)
	handler := NewMatchHandler(mockStore, mockRepo)

	// Test case: Valid score update
	t.Run("Valid score update", func(t *testing.T) {
		matchID := "123"

		// Create request with mux vars
		req := httptest.NewRequest("PUT", "/api/matches/"+matchID+"/score", bytes.NewBufferString(`{"homeGoals": 2, "awayGoals": 1}`))
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": matchID})

		// Set up mock expectation for SaveEvent
		mockStore.On("SaveEvent", req.Context(), mock.MatchedBy(func(event *events.Event) bool {
			return event.Type == "MatchScoreUpdated"
		})).Return(nil)

		// Mock repository calls for event handler
		match := &domain.Match{
			ID:          matchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now(),
			Competition: "Premier League",
			Status:      domain.MatchStatusScheduled,
		}
		mockRepo.On("GetByID", req.Context(), matchID).Return(match, nil)
		mockRepo.On("Update", req.Context(), mock.AnythingOfType("*domain.Match")).Return(nil)

		handler.UpdateMatchScore(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
