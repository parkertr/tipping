package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/pkg/events"
	"github.com/parkertr/tipping/pkg/utils"
)

type PredictionHandler struct {
	eventStore EventStore
}

func NewPredictionHandler(eventStore EventStore) *PredictionHandler {
	return &PredictionHandler{
		eventStore: eventStore,
	}
}

// CreatePredictionRequest represents the request body for creating a prediction.
type CreatePredictionRequest struct {
	UserID    string `json:"userId"`
	MatchID   string `json:"matchId"`
	HomeGoals int    `json:"homeGoals"`
	AwayGoals int    `json:"awayGoals"`
}

// PredictionResponse represents the response body for prediction operations.
type PredictionResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	MatchID   string    `json:"matchId"`
	HomeGoals int       `json:"homeGoals"`
	AwayGoals int       `json:"awayGoals"`
	CreatedAt time.Time `json:"createdAt"`
	Points    int       `json:"points"`
}

// rebuildMatchFromEvents rebuilds a match from its event history.
func (h *PredictionHandler) rebuildMatchFromEvents(ctx context.Context, matchID string) (*domain.Match, error) {
	events, err := h.eventStore.GetEvents(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	if len(events) == 0 {
		return nil, errors.New("match not found")
	}

	match := &domain.Match{}

	fmt.Printf("Processing %d events\n", len(events))

	for _, event := range events {
		// Debug logging
		data, _ := json.Marshal(event.Data)
		fmt.Printf("Event type: %s, Data: %s\n", event.Type, string(data))

		if err := h.processMatchEvent(match, event, data); err != nil {
			return nil, fmt.Errorf("failed to process event: %w", err)
		}
	}

	return match, nil
}

// processMatchEvent processes a single match event.
func (h *PredictionHandler) processMatchEvent(match *domain.Match, event *events.Event, data []byte) error {
	switch event.Type {
	case "MatchCreated":
		var matchCreated struct {
			ID          string    `json:"id"`
			HomeTeam    string    `json:"homeTeam"`
			AwayTeam    string    `json:"awayTeam"`
			Date        time.Time `json:"date"`
			Competition string    `json:"competition"`
		}

		if err := json.Unmarshal(data, &matchCreated); err != nil {
			return fmt.Errorf("failed to unmarshal MatchCreated: %w", err)
		}

		fmt.Printf("Successfully processed MatchCreated event\n")

		match.ID = matchCreated.ID
		match.HomeTeam = matchCreated.HomeTeam
		match.AwayTeam = matchCreated.AwayTeam
		match.Date = matchCreated.Date
		match.Competition = matchCreated.Competition
		match.Status = domain.MatchStatusScheduled

	case "MatchScoreUpdated":
		var scoreUpdated struct {
			MatchID   string    `json:"matchId"`
			HomeGoals int       `json:"homeGoals"`
			AwayGoals int       `json:"awayGoals"`
			UpdatedAt time.Time `json:"updatedAt"`
		}

		if err := json.Unmarshal(data, &scoreUpdated); err != nil {
			return fmt.Errorf("failed to unmarshal MatchScoreUpdated: %w", err)
		}

		fmt.Printf("Successfully processed MatchScoreUpdated event\n")
		match.UpdateScore(scoreUpdated.HomeGoals, scoreUpdated.AwayGoals)

	case "MatchStatusChanged":
		var statusChanged struct {
			MatchID string    `json:"matchId"`
			Status  string    `json:"status"`
			Date    time.Time `json:"date"`
		}

		if err := json.Unmarshal(data, &statusChanged); err != nil {
			return fmt.Errorf("failed to unmarshal MatchStatusChanged: %w", err)
		}

		fmt.Printf("Successfully processed MatchStatusChanged event, new status: %s\n", statusChanged.Status)
		match.Status = domain.MatchStatus(statusChanged.Status)
	}

	return nil
}

// CreatePrediction handles the creation of a new prediction.
func (h *PredictionHandler) CreatePrediction(w http.ResponseWriter, r *http.Request) {
	var request CreatePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)

		return
	}

	// Check if match exists and is not finished
	match, err := h.rebuildMatchFromEvents(r.Context(), request.MatchID)
	if err != nil {
		if err.Error() == "match not found" {
			http.Error(w, "Match not found with ID "+request.MatchID, http.StatusNotFound)

			return
		}

		http.Error(w, fmt.Sprintf("Failed to retrieve match with ID %s: %v", request.MatchID, err), http.StatusInternalServerError)

		return
	}

	// Debug logging
	fmt.Printf("Match status after events: %s\n", match.Status)

	if match.Status == domain.MatchStatusFinished {
		fmt.Printf("Match is finished, returning 400\n")
		http.Error(w, "cannot create prediction for finished match", http.StatusBadRequest)

		return
	}

	prediction := domain.NewPrediction(
		utils.GenerateID(),
		request.UserID,
		request.MatchID,
		request.HomeGoals,
		request.AwayGoals,
	)

	event := events.NewEvent("PredictionMade", events.PredictionMade{
		ID:        prediction.ID,
		UserID:    prediction.UserID,
		MatchID:   prediction.MatchID,
		HomeGoals: prediction.HomeGoals,
		AwayGoals: prediction.AwayGoals,
		CreatedAt: prediction.CreatedAt,
	})

	if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create prediction for match %s: %v", request.MatchID, err), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(prediction); err != nil {
		fmt.Printf("error encoding prediction for match %s: %v\n", request.MatchID, err)
	}
}

// unmarshalPredictionEvent unmarshals a prediction event into a domain.Prediction.
func (h *PredictionHandler) unmarshalPredictionEvent(event *events.Event) (*domain.Prediction, error) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	var predictionMade struct {
		ID        string    `json:"id"`
		UserID    string    `json:"userId"`
		MatchID   string    `json:"matchId"`
		HomeGoals int       `json:"homeGoals"`
		AwayGoals int       `json:"awayGoals"`
		CreatedAt time.Time `json:"createdAt"`
	}

	if err := json.Unmarshal(data, &predictionMade); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prediction data for event ID %s: %w", event.ID, err)
	}

	return domain.NewPrediction(
		predictionMade.ID,
		predictionMade.UserID,
		predictionMade.MatchID,
		predictionMade.HomeGoals,
		predictionMade.AwayGoals,
	), nil
}

// getPredictions retrieves predictions based on the given filter.
func (h *PredictionHandler) getPredictions(w http.ResponseWriter, r *http.Request, filter func(*domain.Prediction) bool) {
	events, err := h.eventStore.GetEventsByType(r.Context(), "PredictionMade")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve predictions: %v", err), http.StatusInternalServerError)

		return
	}

	predictions := make([]*domain.Prediction, 0)

	for _, event := range events {
		prediction, err := h.unmarshalPredictionEvent(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to process prediction data: %v", err), http.StatusInternalServerError)

			return
		}

		if filter(prediction) {
			predictions = append(predictions, prediction)
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(predictions); err != nil {
		fmt.Printf("error encoding predictions: %v\n", err)
	}
}

// GetUserPredictions retrieves all predictions for a user.
func (h *PredictionHandler) GetUserPredictions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	h.getPredictions(w, r, func(prediction *domain.Prediction) bool {
		return prediction.UserID == userID
	})
}

// GetMatchPredictions retrieves all predictions for a match.
func (h *PredictionHandler) GetMatchPredictions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["matchId"]

	h.getPredictions(w, r, func(prediction *domain.Prediction) bool {
		return prediction.MatchID == matchID
	})
}

// GetUserPredictionForMatch retrieves a specific user's prediction for a specific match.
func (h *PredictionHandler) GetUserPredictionForMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["matchId"]
	userID := vars["userId"]

	events, err := h.eventStore.GetEventsByType(r.Context(), "PredictionMade")
	if err != nil {
		http.Error(w, "failed to retrieve predictions", http.StatusInternalServerError)

		return
	}

	for _, event := range events {
		prediction, err := h.unmarshalPredictionEvent(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to process prediction data: %v", err), http.StatusInternalServerError)

			return
		}

		if prediction.MatchID == matchID && prediction.UserID == userID {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(prediction); err != nil {
				fmt.Printf("error encoding prediction: %v\n", err)
			}

			return
		}
	}

	// No prediction found
	http.Error(w, "prediction not found", http.StatusNotFound)
}

// RegisterRoutes registers the prediction handler routes.
func (h *PredictionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/predictions", h.CreatePrediction).Methods("POST")
	r.HandleFunc("/users/{userId}/predictions", h.GetUserPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions", h.GetMatchPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions/{userId}", h.GetUserPredictionForMatch).Methods("GET")
}
