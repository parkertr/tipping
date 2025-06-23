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
	"github.com/parkertr/tipping/internal/domain/constants"
	"github.com/parkertr/tipping/internal/domain/events"
	"github.com/parkertr/tipping/internal/infrastructure/eventhandlers"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	pkgEvents "github.com/parkertr/tipping/pkg/events"
	"github.com/parkertr/tipping/pkg/utils"
)

type MatchHandler struct {
	eventStore   EventStore
	matchRepo    repository.MatchRepository
	eventHandler *eventhandlers.MatchEventHandler
}

func NewMatchHandler(eventStore EventStore, matchRepo repository.MatchRepository) *MatchHandler {
	return &MatchHandler{
		eventStore:   eventStore,
		matchRepo:    matchRepo,
		eventHandler: eventhandlers.NewMatchEventHandler(matchRepo),
	}
}

// CreateMatchRequest represents the request body for creating a match.
type CreateMatchRequest struct {
	HomeTeam    string    `json:"homeTeam"`
	AwayTeam    string    `json:"awayTeam"`
	Date        time.Time `json:"date"`
	Competition string    `json:"competition"`
}

// UpdateScoreRequest represents the request body for updating a match score.
type UpdateScoreRequest struct {
	HomeGoals int `json:"homeGoals"`
	AwayGoals int `json:"awayGoals"`
}

// MatchResponse represents the response body for match operations.
type MatchResponse struct {
	ID          string    `json:"id"`
	HomeTeam    string    `json:"homeTeam"`
	AwayTeam    string    `json:"awayTeam"`
	Date        time.Time `json:"date"`
	Competition string    `json:"competition"`
	Status      string    `json:"status"`
	Score       *Score    `json:"score"`
}

// Score represents a match score.
type Score struct {
	HomeGoals int       `json:"homeGoals"`
	AwayGoals int       `json:"awayGoals"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateMatch handles the creation of a new match.
func (h *MatchHandler) CreateMatch(writer http.ResponseWriter, request *http.Request) {
	var matchRequest struct {
		HomeTeam    string    `json:"homeTeam"`
		AwayTeam    string    `json:"awayTeam"`
		Date        time.Time `json:"date"`
		Competition string    `json:"competition"`
	}

	if err := json.NewDecoder(request.Body).Decode(&matchRequest); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if matchRequest.HomeTeam == "" || matchRequest.AwayTeam == "" || matchRequest.Competition == "" {
		http.Error(writer, "missing required fields", http.StatusBadRequest)
		return
	}

	match := &domain.Match{
		ID:          utils.GenerateID(),
		HomeTeam:    matchRequest.HomeTeam,
		AwayTeam:    matchRequest.AwayTeam,
		Date:        matchRequest.Date,
		Competition: matchRequest.Competition,
		Status:      domain.MatchStatusScheduled,
		Score:       &domain.Score{HomeGoals: 0, AwayGoals: 0},
	}

	if err := h.matchRepo.Create(request.Context(), match); err != nil {
		http.Error(writer, "failed to create match", http.StatusInternalServerError)
		return
	}

	event := pkgEvents.NewEvent("MatchCreated", events.MatchCreated{
		ID:          match.ID,
		HomeTeam:    match.HomeTeam,
		AwayTeam:    match.AwayTeam,
		Date:        match.Date,
		Competition: match.Competition,
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		http.Error(writer, "failed to create match", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(writer).Encode(match); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateMatchScore handles updating a match's score.
func (h *MatchHandler) UpdateMatchScore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	matchID := vars["id"]

	var scoreRequest struct {
		HomeGoals int `json:"homeGoals"`
		AwayGoals int `json:"awayGoals"`
	}

	if err := json.NewDecoder(request.Body).Decode(&scoreRequest); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)

		return
	}

	event := pkgEvents.NewEvent("MatchScoreUpdated", events.MatchScoreUpdated{
		MatchID:   matchID,
		HomeGoals: scoreRequest.HomeGoals,
		AwayGoals: scoreRequest.AwayGoals,
		UpdatedAt: time.Now(),
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		http.Error(writer, "failed to update match score", http.StatusInternalServerError)

		return
	}

	// Process event to update read model
	if err := h.eventHandler.HandleEvent(request.Context(), event); err != nil {
		fmt.Printf("Failed to process event for score update: %v\n", err)
		// Continue anyway since the event is saved
	}

	writer.WriteHeader(http.StatusOK)
}

// rebuildMatchFromEvents rebuilds a match from its event history.
func (h *MatchHandler) rebuildMatchFromEvents(ctx context.Context, matchID string) (*domain.Match, error) {
	events, err := h.eventStore.GetEvents(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	if len(events) == 0 {
		return nil, repository.ErrNotFound
	}

	match := &domain.Match{
		ID:          "",
		HomeTeam:    "",
		AwayTeam:    "",
		Date:        time.Time{},
		Competition: "",
		Status:      domain.MatchStatusScheduled,
		Score:       &domain.Score{HomeGoals: constants.InitialScore, AwayGoals: constants.InitialScore},
	}

	for _, event := range events {
		if err := h.processMatchEvent(match, *event); err != nil {
			return nil, fmt.Errorf("failed to process event: %w", err)
		}
	}

	return match, nil
}

// processMatchEvent processes a single match event.
func (h *MatchHandler) processMatchEvent(match *domain.Match, event pkgEvents.Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	switch event.Type {
	case "MatchCreated":
		var matchCreated events.MatchCreated
		if err := json.Unmarshal(data, &matchCreated); err != nil {
			return fmt.Errorf("failed to unmarshal match created data: %w", err)
		}

		match.ID = matchCreated.ID
		match.HomeTeam = matchCreated.HomeTeam
		match.AwayTeam = matchCreated.AwayTeam
		match.Date = matchCreated.Date
		match.Competition = matchCreated.Competition
		match.Status = domain.MatchStatusScheduled

	case "MatchScoreUpdated":
		var scoreUpdated events.MatchScoreUpdated
		if err := json.Unmarshal(data, &scoreUpdated); err != nil {
			return fmt.Errorf("failed to unmarshal score updated data: %w", err)
		}

		match.UpdateScore(scoreUpdated.HomeGoals, scoreUpdated.AwayGoals)
	}

	return nil
}

// GetMatch retrieves a match by ID.
func (h *MatchHandler) GetMatch(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	matchID := vars["id"]

	// Try to get from read model first
	match, err := h.matchRepo.GetByID(request.Context(), matchID)
	if err == nil {
		h.writeMatchResponse(writer, match)

		return
	}

	// Fallback to rebuilding from events if not in read model
	match, err = h.rebuildMatchFromEvents(request.Context(), matchID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(writer, "match not found", http.StatusNotFound)

			return
		}

		http.Error(writer, "failed to retrieve match", http.StatusInternalServerError)

		return
	}

	h.writeMatchResponse(writer, match)
}

// writeMatchResponse writes the match response to the HTTP response.
func (h *MatchHandler) writeMatchResponse(writer http.ResponseWriter, match *domain.Match) {
	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(match); err != nil {
		fmt.Printf("error encoding match: %v\n", err)
	}
}

// ListMatches retrieves all matches from the read model.
func (h *MatchHandler) ListMatches(writer http.ResponseWriter, request *http.Request) {
	// Use read model for better performance and consistent date formatting
	matches, err := h.matchRepo.List(request.Context(), repository.MatchFilters{
		Competition: nil,
		StartDate:   nil,
		EndDate:     nil,
		Status:      nil,
	})
	if err != nil {
		http.Error(writer, "failed to retrieve matches", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(matches); err != nil {
		fmt.Printf("error encoding matches: %v\n", err)
	}
}

// ListUpcomingMatches retrieves upcoming matches (scheduled matches).
func (h *MatchHandler) ListUpcomingMatches(writer http.ResponseWriter, request *http.Request) {
	status := string(domain.MatchStatusScheduled)

	// Use read model with filters for upcoming matches
	// For demo purposes, we'll show all scheduled matches regardless of date
	matches, err := h.matchRepo.List(request.Context(), repository.MatchFilters{
		Status:      &status,
		Competition: nil,
		StartDate:   nil,
		EndDate:     nil,
	})
	if err != nil {
		http.Error(writer, "failed to retrieve upcoming matches", http.StatusInternalServerError)

		return
	}

	// Limit to next 5 matches for the home page
	if len(matches) > constants.MaxUpcomingMatches {
		matches = matches[:constants.MaxUpcomingMatches]
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(matches); err != nil {
		fmt.Printf("error encoding upcoming matches: %v\n", err)
	}
}

// RegisterRoutes registers the match handler routes.
func (h *MatchHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/matches", h.CreateMatch).Methods("POST")
	r.HandleFunc("/matches", h.ListMatches).Methods("GET")
	r.HandleFunc("/matches/{id}", h.GetMatch).Methods("GET")
	r.HandleFunc("/matches/{id}/score", h.UpdateMatchScore).Methods("PUT")
}
