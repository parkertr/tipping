package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
	"github.com/parkertr/tipping/pkg/utils"
)

var (
	ErrMatchNotFound = errors.New("match not found")
)

type PredictionHandler struct {
	eventStore EventStore
	matchRepo  repository.MatchRepository
}

func NewPredictionHandler(eventStore EventStore, matchRepo repository.MatchRepository) *PredictionHandler {
	return &PredictionHandler{
		eventStore: eventStore,
		matchRepo:  matchRepo,
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

// CreatePrediction handles the creation of a new prediction.
func (h *PredictionHandler) CreatePrediction(writer http.ResponseWriter, request *http.Request) {
	var createPredictionRequest CreatePredictionRequest
	if err := json.NewDecoder(request.Body).Decode(&createPredictionRequest); err != nil {
		http.Error(writer, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if createPredictionRequest.UserID == "" || createPredictionRequest.MatchID == "" {
		http.Error(writer, "missing required fields", http.StatusBadRequest)
		return
	}

	// Get match
	match, err := h.matchRepo.GetByID(request.Context(), createPredictionRequest.MatchID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(writer, "match not found", http.StatusNotFound)
			return
		}
		errMsg := fmt.Sprintf(
			"Failed to retrieve match with ID %s: %v",
			createPredictionRequest.MatchID,
			err,
		)
		http.Error(writer, errMsg, http.StatusInternalServerError)
		return
	}

	// Check if match exists and is not finished
	if match.Status == domain.MatchStatusFinished {
		http.Error(writer, "cannot create prediction for finished match", http.StatusBadRequest)
		return
	}

	prediction := domain.NewPrediction(
		utils.GenerateID(),
		createPredictionRequest.UserID,
		createPredictionRequest.MatchID,
		createPredictionRequest.HomeGoals,
		createPredictionRequest.AwayGoals,
	)

	event := events.NewEvent("PredictionMade", events.PredictionMade{
		ID:        prediction.ID,
		UserID:    prediction.UserID,
		MatchID:   prediction.MatchID,
		HomeGoals: prediction.HomeGoals,
		AwayGoals: prediction.AwayGoals,
		CreatedAt: prediction.CreatedAt,
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		errMsg := fmt.Sprintf(
			"Failed to create prediction for match %s: %v",
			createPredictionRequest.MatchID,
			err,
		)
		http.Error(writer, errMsg, http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(writer).Encode(prediction); err != nil {
		fmt.Printf("error encoding prediction for match %s: %v\n", createPredictionRequest.MatchID, err)
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
func (h *PredictionHandler) getPredictions(
	w http.ResponseWriter,
	r *http.Request,
	filter func(*domain.Prediction) bool,
) {
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
func (h *PredictionHandler) GetUserPredictionForMatch(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	matchID := vars["matchId"]
	userID := vars["userId"]

	events, err := h.eventStore.GetEventsByType(request.Context(), "PredictionMade")
	if err != nil {
		http.Error(writer, "failed to retrieve predictions", http.StatusInternalServerError)

		return
	}

	for _, event := range events {
		prediction, err := h.unmarshalPredictionEvent(event)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Failed to process prediction data: %v", err), http.StatusInternalServerError)

			return
		}

		if prediction.MatchID == matchID && prediction.UserID == userID {
			writer.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(writer).Encode(prediction); err != nil {
				fmt.Printf("error encoding prediction: %v\n", err)
			}

			return
		}
	}

	// No prediction found
	http.Error(writer, "prediction not found", http.StatusNotFound)
}

// RegisterRoutes registers the prediction handler routes.
func (h *PredictionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/predictions", h.CreatePrediction).Methods("POST")
	r.HandleFunc("/users/{userId}/predictions", h.GetUserPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions", h.GetMatchPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions/{userId}", h.GetUserPredictionForMatch).Methods("GET")
}
