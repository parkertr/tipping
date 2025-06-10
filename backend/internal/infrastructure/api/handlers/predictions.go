package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/pkg/events"
	"github.com/parkertr2/footy-tipping/pkg/utils"
)

type PredictionHandler struct {
	eventStore EventStore
}

func NewPredictionHandler(eventStore EventStore) *PredictionHandler {
	return &PredictionHandler{
		eventStore: eventStore,
	}
}

// CreatePrediction handles the creation of a new prediction
func (h *PredictionHandler) CreatePrediction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID    string `json:"userId"`
		MatchID   string `json:"matchId"`
		HomeGoals int    `json:"homeGoals"`
		AwayGoals int    `json:"awayGoals"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if match exists and is not finished
	matchEvents, err := h.eventStore.GetEvents(r.Context(), request.MatchID)
	if err != nil {
		http.Error(w, "Failed to retrieve match", http.StatusInternalServerError)
		return
	}

	if len(matchEvents) == 0 {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}

	match := &domain.Match{}
	for _, event := range matchEvents {
		switch event.Type {
		case "MatchCreated":
			data, err := json.Marshal(event.Data)
			if err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			var matchCreated events.MatchCreated
			if err := json.Unmarshal(data, &matchCreated); err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			match.ID = matchCreated.ID
			match.HomeTeam = matchCreated.HomeTeam
			match.AwayTeam = matchCreated.AwayTeam
			match.Date = matchCreated.Date
			match.Competition = matchCreated.Competition
			match.Status = domain.MatchStatusScheduled
		case "MatchScoreUpdated":
			data, err := json.Marshal(event.Data)
			if err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			var scoreUpdated events.MatchScoreUpdated
			if err := json.Unmarshal(data, &scoreUpdated); err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			match.UpdateScore(scoreUpdated.HomeGoals, scoreUpdated.AwayGoals)
		}
	}

	if match.IsFinished() {
		http.Error(w, "Cannot create prediction for finished match", http.StatusBadRequest)
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
		http.Error(w, "Failed to create prediction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prediction)
}

// GetUserPredictions retrieves all predictions for a user
func (h *PredictionHandler) GetUserPredictions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	events, err := h.eventStore.GetEventsByType(r.Context(), "PredictionMade")
	if err != nil {
		http.Error(w, "Failed to retrieve predictions", http.StatusInternalServerError)
		return
	}

	predictions := make([]*domain.Prediction, 0)
	for _, event := range events {
		data, err := json.Marshal(event.Data)
		if err != nil {
			http.Error(w, "Failed to process prediction data", http.StatusInternalServerError)
			return
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
			http.Error(w, "Failed to process prediction data", http.StatusInternalServerError)
			return
		}
		if predictionMade.UserID == userID {
			prediction := domain.NewPrediction(
				predictionMade.ID,
				predictionMade.UserID,
				predictionMade.MatchID,
				predictionMade.HomeGoals,
				predictionMade.AwayGoals,
			)
			predictions = append(predictions, prediction)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(predictions)
}

// GetMatchPredictions retrieves all predictions for a match
func (h *PredictionHandler) GetMatchPredictions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["matchId"]

	events, err := h.eventStore.GetEventsByType(r.Context(), "PredictionMade")
	if err != nil {
		http.Error(w, "Failed to retrieve predictions", http.StatusInternalServerError)
		return
	}

	predictions := make([]*domain.Prediction, 0)
	for _, event := range events {
		data, err := json.Marshal(event.Data)
		if err != nil {
			http.Error(w, "Failed to process prediction data", http.StatusInternalServerError)
			return
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
			http.Error(w, "Failed to process prediction data", http.StatusInternalServerError)
			return
		}
		if predictionMade.MatchID == matchID {
			prediction := domain.NewPrediction(
				predictionMade.ID,
				predictionMade.UserID,
				predictionMade.MatchID,
				predictionMade.HomeGoals,
				predictionMade.AwayGoals,
			)
			predictions = append(predictions, prediction)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(predictions)
}
