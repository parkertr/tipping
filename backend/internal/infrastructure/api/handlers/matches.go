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

type MatchHandler struct {
	eventStore EventStore
}

func NewMatchHandler(eventStore EventStore) *MatchHandler {
	return &MatchHandler{
		eventStore: eventStore,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(match)
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

	w.WriteHeader(http.StatusOK)
}

// GetMatch retrieves a match by ID
func (h *MatchHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

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
	match := &domain.Match{}
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
	json.NewEncoder(w).Encode(match)
}

// ListMatches retrieves all matches
func (h *MatchHandler) ListMatches(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventStore.GetEventsByType(r.Context(), "MatchCreated")
	if err != nil {
		http.Error(w, "Failed to retrieve matches", http.StatusInternalServerError)
		return
	}

	matches := make([]*domain.Match, 0)
	for _, event := range events {
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
		match := domain.NewMatch(
			matchCreated.ID,
			matchCreated.HomeTeam,
			matchCreated.AwayTeam,
			matchCreated.Date,
			matchCreated.Competition,
		)
		matches = append(matches, match)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
