package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/api/middleware"
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
// UserID is no longer included as it comes from authentication context
type CreatePredictionRequest struct {
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
// Now uses authenticated user from context instead of accepting userID in request
func (h *PredictionHandler) CreatePrediction(writer http.ResponseWriter, request *http.Request) {
	// Get authenticated user from context
	user := middleware.GetUserFromContext(request.Context())
	if user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)
		return
	}

	var createPredictionRequest CreatePredictionRequest
	if err := json.NewDecoder(request.Body).Decode(&createPredictionRequest); err != nil {
		http.Error(writer, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if createPredictionRequest.MatchID == "" {
		http.Error(writer, "missing required field: matchId", http.StatusBadRequest)
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

	// Check if user already has a prediction for this match
	existingPrediction, err := h.getUserPredictionForMatch(request.Context(), user.ID, createPredictionRequest.MatchID)
	if err != nil && !errors.Is(err, ErrPredictionNotFound) {
		http.Error(writer, "failed to check existing prediction", http.StatusInternalServerError)
		return
	}
	if existingPrediction != nil {
		http.Error(writer, "user already has a prediction for this match", http.StatusConflict)
		return
	}

	prediction := domain.NewPrediction(
		utils.GenerateID(),
		user.ID, // Use authenticated user ID
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

var ErrPredictionNotFound = errors.New("prediction not found")

// getUserPredictionForMatch helper method to get a user's prediction for a specific match
func (h *PredictionHandler) getUserPredictionForMatch(
	ctx context.Context,
	userID, matchID string,
) (*domain.Prediction, error) {
	events, err := h.eventStore.GetEventsByType(ctx, "PredictionMade")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve predictions: %w", err)
	}

	for _, event := range events {
		prediction, err := h.unmarshalPredictionEvent(event)
		if err != nil {
			return nil, fmt.Errorf("failed to process prediction data: %w", err)
		}

		if prediction.MatchID == matchID && prediction.UserID == userID {
			return prediction, nil
		}
	}

	return nil, ErrPredictionNotFound
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

// GetUserPredictions retrieves all predictions for the authenticated user.
// No longer takes userID from URL path - uses authenticated user from context
func (h *PredictionHandler) GetUserPredictions(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	h.getPredictions(w, r, func(prediction *domain.Prediction) bool {
		return prediction.UserID == user.ID
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

// GetUserPredictionForMatch retrieves the authenticated user's prediction for a specific match.
func (h *PredictionHandler) GetUserPredictionForMatch(writer http.ResponseWriter, request *http.Request) {
	// Get authenticated user from context
	user := middleware.GetUserFromContext(request.Context())
	if user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(request)
	matchID := vars["matchId"]

	prediction, err := h.getUserPredictionForMatch(request.Context(), user.ID, matchID)
	if err != nil {
		if errors.Is(err, ErrPredictionNotFound) {
			http.Error(writer, "prediction not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "failed to retrieve prediction", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(prediction); err != nil {
		fmt.Printf("error encoding prediction: %v\n", err)
	}
}

// RegisterRoutes registers the prediction handler routes.
// All routes now require authentication
func (h *PredictionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/predictions", h.CreatePrediction).Methods("POST")
	r.HandleFunc("/predictions/me", h.GetUserPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions", h.GetMatchPredictions).Methods("GET")
	r.HandleFunc("/matches/{matchId}/predictions/me", h.GetUserPredictionForMatch).Methods("GET")
}
