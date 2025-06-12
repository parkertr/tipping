package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/eventhandlers"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository"
	"github.com/parkertr2/footy-tipping/pkg/events"
	"github.com/parkertr2/footy-tipping/pkg/utils"
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

// CreateMatch handles the creation of a new match
func (h *MatchHandler) CreateMatch(w http.ResponseWriter, r *http.Request) {
	var request struct {
		HomeTeam    string    `json:"homeTeam"`
		AwayTeam    string    `json:"awayTeam"`
		Date        time.Time `json:"date"`
		Competition string    `json:"competition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	match := domain.NewMatch(
		utils.GenerateID(),
		request.HomeTeam,
		request.AwayTeam,
		request.Date,
		request.Competition,
	)

	event := events.NewEvent("MatchCreated", events.MatchCreated{
		ID:          match.ID,
		HomeTeam:    match.HomeTeam,
		AwayTeam:    match.AwayTeam,
		Date:        match.Date,
		Competition: match.Competition,
	})

	if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
		http.Error(w, "Failed to create match", http.StatusInternalServerError)
		return
	}

	// Process event to update read model
	if err := h.eventHandler.HandleEvent(r.Context(), event); err != nil {
		fmt.Printf("Failed to process event for match creation: %v\n", err)
		// Continue anyway since the event is saved
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(match); err != nil {
		fmt.Printf("error encoding match: %v\n", err)
	}
}

// UpdateMatchScore handles updating a match's score
func (h *MatchHandler) UpdateMatchScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

	var request struct {
		HomeGoals int `json:"homeGoals"`
		AwayGoals int `json:"awayGoals"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event := events.NewEvent("MatchScoreUpdated", events.MatchScoreUpdated{
		MatchID:   matchID,
		HomeGoals: request.HomeGoals,
		AwayGoals: request.AwayGoals,
		UpdatedAt: time.Now(),
	})

	if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
		http.Error(w, "Failed to update match score", http.StatusInternalServerError)
		return
	}

	// Process event to update read model
	if err := h.eventHandler.HandleEvent(r.Context(), event); err != nil {
		fmt.Printf("Failed to process event for score update: %v\n", err)
		// Continue anyway since the event is saved
	}

	w.WriteHeader(http.StatusOK)
}

// GetMatch retrieves a match by ID
func (h *MatchHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

	// Try to get from read model first
	match, err := h.matchRepo.GetByID(r.Context(), matchID)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(match); err != nil {
			fmt.Printf("error encoding match: %v\n", err)
		}
		return
	}

	// Fallback to rebuilding from events if not in read model
	events, err := h.eventStore.GetEvents(r.Context(), matchID)
	if err != nil {
		http.Error(w, "Failed to retrieve match", http.StatusInternalServerError)
		return
	}

	if len(events) == 0 {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}

	// Rebuild match from events
	match = &domain.Match{}
	for _, event := range events {
		switch event.Type {
		case "MatchCreated":
			data, err := json.Marshal(event.Data)
			if err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			var matchCreated struct {
				ID          string    `json:"id"`
				HomeTeam    string    `json:"homeTeam"`
				AwayTeam    string    `json:"awayTeam"`
				Date        time.Time `json:"date"`
				Competition string    `json:"competition"`
			}
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
			var scoreUpdated struct {
				MatchID   string    `json:"matchId"`
				HomeGoals int       `json:"homeGoals"`
				AwayGoals int       `json:"awayGoals"`
				UpdatedAt time.Time `json:"updatedAt"`
			}
			if err := json.Unmarshal(data, &scoreUpdated); err != nil {
				http.Error(w, "Failed to process match data", http.StatusInternalServerError)
				return
			}
			match.UpdateScore(scoreUpdated.HomeGoals, scoreUpdated.AwayGoals)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(match); err != nil {
		fmt.Printf("error encoding match: %v\n", err)
	}
}

// ListMatches retrieves all matches from the read model
func (h *MatchHandler) ListMatches(w http.ResponseWriter, r *http.Request) {
	// Use read model for better performance and consistent date formatting
	matches, err := h.matchRepo.List(r.Context(), repository.MatchFilters{})
	if err != nil {
		http.Error(w, "Failed to retrieve matches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		fmt.Printf("error encoding matches: %v\n", err)
	}
}
