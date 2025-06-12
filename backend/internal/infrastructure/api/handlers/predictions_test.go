package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Test cases for prediction-related handlers
func TestCreatePrediction(t *testing.T) {
	// Test case 1: Valid prediction creation
	t.Run("Valid prediction creation", func(t *testing.T) {
		mockStore := new(mocks.MockEventStore)
		handler := NewPredictionHandler(mockStore)
		prediction := domain.Prediction{
			UserID:    "user123",
			MatchID:   "match123",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}

		// Create request
		body, _ := json.Marshal(prediction)
		req := httptest.NewRequest("POST", "/api/predictions", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		// Create mock match event
		matchCreated := events.MatchCreated{
			ID:          prediction.MatchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now().Add(24 * time.Hour), // Future match
			Competition: "Premier League",
		}
		matchEvent := &events.Event{
			ID:        "event123",
			Type:      "MatchCreated",
			Data:      matchCreated,
			Timestamp: time.Now(),
			Version:   1,
		}

		// Set up mock expectation for GetEvents (match check)
		mockStore.On("GetEvents", req.Context(), prediction.MatchID).Return([]*events.Event{matchEvent}, nil)

		// Set up mock expectation for SaveEvent - use mock.MatchedBy for dynamic fields
		mockStore.On("SaveEvent", req.Context(), mock.MatchedBy(func(event *events.Event) bool {
			// Type check
			if event.Type != "PredictionMade" {
				return false
			}

			// Cast data to PredictionMade
			predictionMade, ok := event.Data.(events.PredictionMade)
			if !ok {
				return false
			}

			// Check static fields
			return predictionMade.UserID == prediction.UserID &&
				predictionMade.MatchID == prediction.MatchID &&
				predictionMade.HomeGoals == prediction.HomeGoals &&
				predictionMade.AwayGoals == prediction.AwayGoals
		})).Return(nil)

		// Handle request
		handler.CreatePrediction(rr, req)

		// Assert response
		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStore := new(mocks.MockEventStore)
		handler := NewPredictionHandler(mockStore)
		req := httptest.NewRequest("POST", "/api/predictions", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		handler.CreatePrediction(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	// Test case 3: Match not found
	t.Run("Match not found", func(t *testing.T) {
		mockStore := new(mocks.MockEventStore)
		handler := NewPredictionHandler(mockStore)
		prediction := domain.Prediction{
			UserID:    "user123",
			MatchID:   "nonexistent",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}

		// Create request
		body, _ := json.Marshal(prediction)
		req := httptest.NewRequest("POST", "/api/predictions", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		// Set up mock expectation for GetEvents (match not found)
		mockStore.On("GetEvents", req.Context(), prediction.MatchID).Return([]*events.Event{}, nil)

		// Handle request
		handler.CreatePrediction(rr, req)

		// Assert response
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 4: Match already finished
	t.Run("Match already finished", func(t *testing.T) {
		mockStore := new(mocks.MockEventStore)
		handler := NewPredictionHandler(mockStore)
		prediction := domain.Prediction{
			UserID:    "user123",
			MatchID:   "match123",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}

		// Create request
		body, _ := json.Marshal(prediction)
		req := httptest.NewRequest("POST", "/api/predictions", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		// Create mock match events
		matchCreated := events.MatchCreated{
			ID:          prediction.MatchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now().Add(-24 * time.Hour), // Past match
			Competition: "Premier League",
		}
		matchEvent := &events.Event{
			ID:        "event123",
			Type:      "MatchCreated",
			Data:      matchCreated,
			Timestamp: time.Now().Add(-24 * time.Hour),
			Version:   1,
		}

		scoreUpdated := events.MatchScoreUpdated{
			MatchID:   prediction.MatchID,
			HomeGoals: 1,
			AwayGoals: 0,
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		}
		scoreEvent := &events.Event{
			ID:        "event124",
			Type:      "MatchScoreUpdated",
			Data:      scoreUpdated,
			Timestamp: time.Now().Add(-12 * time.Hour),
			Version:   1,
		}

		statusChanged := events.MatchStatusChanged{
			MatchID:   prediction.MatchID,
			Status:    "FINISHED",
			ChangedAt: time.Now().Add(-12 * time.Hour),
		}
		statusEvent := &events.Event{
			ID:        "event125",
			Type:      "MatchStatusChanged",
			Data:      statusChanged,
			Timestamp: time.Now().Add(-12 * time.Hour),
			Version:   1,
		}

		// Set up mock expectation for GetEvents (finished match)
		mockStore.On("GetEvents", req.Context(), prediction.MatchID).Return([]*events.Event{matchEvent, scoreEvent, statusEvent}, nil)

		// Handle request
		handler.CreatePrediction(rr, req)

		// Debug output
		fmt.Printf("TestCreatePrediction/Match_already_finished: response code = %d, body = %s\n", rr.Code, rr.Body.String())

		// Assert response
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})
}

func TestGetUserPredictions(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewPredictionHandler(mockStore)

	// Test case 1: User has predictions
	t.Run("User has predictions", func(t *testing.T) {
		userID := "user123"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/users/"+userID+"/predictions", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"userId": userID})

		// Create events with mock data
		predictionMade1 := events.PredictionMade{
			ID:        "pred123",
			UserID:    userID,
			MatchID:   "match123",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}
		event1 := &events.Event{
			ID:        "event123",
			Type:      "PredictionMade",
			Data:      predictionMade1,
			Timestamp: time.Now(),
			Version:   1,
		}

		predictionMade2 := events.PredictionMade{
			ID:        "pred456",
			UserID:    userID,
			MatchID:   "match456",
			HomeGoals: 0,
			AwayGoals: 0,
			CreatedAt: time.Now(),
		}
		event2 := &events.Event{
			ID:        "event456",
			Type:      "PredictionMade",
			Data:      predictionMade2,
			Timestamp: time.Now(),
			Version:   1,
		}

		mockStore.On("GetEventsByType", req.Context(), "PredictionMade").Return([]*events.Event{event1, event2}, nil)

		// Handle request
		handler.GetUserPredictions(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: User has no predictions
	t.Run("User has no predictions", func(t *testing.T) {
		userID := "user456"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/users/"+userID+"/predictions", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"userId": userID})

		mockStore.On("GetEventsByType", req.Context(), "PredictionMade").Return([]*events.Event{}, nil)

		handler.GetUserPredictions(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})
}

func TestGetMatchPredictions(t *testing.T) {
	// Create mock event store
	mockStore := new(mocks.MockEventStore)
	handler := NewPredictionHandler(mockStore)

	// Test case 1: Match has predictions
	t.Run("Match has predictions", func(t *testing.T) {
		matchID := "match123"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID+"/predictions", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"matchId": matchID})

		// Create events with mock data
		predictionMade1 := events.PredictionMade{
			ID:        "pred123",
			UserID:    "user123",
			MatchID:   matchID,
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}
		event1 := &events.Event{
			ID:        "event123",
			Type:      "PredictionMade",
			Data:      predictionMade1,
			Timestamp: time.Now(),
			Version:   1,
		}

		predictionMade2 := events.PredictionMade{
			ID:        "pred456",
			UserID:    "user456",
			MatchID:   matchID,
			HomeGoals: 1,
			AwayGoals: 2,
			CreatedAt: time.Now(),
		}
		event2 := &events.Event{
			ID:        "event456",
			Type:      "PredictionMade",
			Data:      predictionMade2,
			Timestamp: time.Now(),
			Version:   1,
		}

		mockStore.On("GetEventsByType", req.Context(), "PredictionMade").Return([]*events.Event{event1, event2}, nil)

		// Handle request
		handler.GetMatchPredictions(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Match has no predictions
	t.Run("Match has no predictions", func(t *testing.T) {
		matchID := "match456"

		// Create request with mux vars
		req := httptest.NewRequest("GET", "/api/matches/"+matchID+"/predictions", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"matchId": matchID})

		mockStore.On("GetEventsByType", req.Context(), "PredictionMade").Return([]*events.Event{}, nil)

		handler.GetMatchPredictions(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		mockStore.AssertExpectations(t)
	})
}
