package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository"
	"github.com/parkertr2/footy-tipping/pkg/events"
)

// MatchEventHandler handles match-related events and updates the read model
type MatchEventHandler struct {
	matchRepo repository.MatchRepository
}

// NewMatchEventHandler creates a new match event handler
func NewMatchEventHandler(matchRepo repository.MatchRepository) *MatchEventHandler {
	return &MatchEventHandler{
		matchRepo: matchRepo,
	}
}

// HandleEvent processes events and updates the read model accordingly
func (h *MatchEventHandler) HandleEvent(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case "MatchCreated":
		return h.handleMatchCreated(ctx, event)
	case "MatchScoreUpdated":
		return h.handleMatchScoreUpdated(ctx, event)
	case "MatchStatusChanged":
		return h.handleMatchStatusChanged(ctx, event)
	default:
		// Ignore unknown event types
		return nil
	}
}

// handleMatchCreated processes MatchCreated events
func (h *MatchEventHandler) handleMatchCreated(ctx context.Context, event *events.Event) error {
	// Extract event data
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var matchCreated events.MatchCreated
	if err := json.Unmarshal(data, &matchCreated); err != nil {
		return fmt.Errorf("failed to unmarshal MatchCreated event: %w", err)
	}

	// Create domain match
	match := domain.NewMatch(
		matchCreated.ID,
		matchCreated.HomeTeam,
		matchCreated.AwayTeam,
		matchCreated.Date,
		matchCreated.Competition,
	)

	// Save to read model
	if err := h.matchRepo.Create(ctx, match); err != nil {
		log.Printf("Failed to create match in read model: %v", err)
		return fmt.Errorf("failed to create match in read model: %w", err)
	}

	log.Printf("Created match in read model: %s vs %s", match.HomeTeam, match.AwayTeam)
	return nil
}

// handleMatchScoreUpdated processes MatchScoreUpdated events
func (h *MatchEventHandler) handleMatchScoreUpdated(ctx context.Context, event *events.Event) error {
	// Extract event data
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var scoreUpdated events.MatchScoreUpdated
	if err := json.Unmarshal(data, &scoreUpdated); err != nil {
		return fmt.Errorf("failed to unmarshal MatchScoreUpdated event: %w", err)
	}

	// Get existing match from read model
	match, err := h.matchRepo.GetByID(ctx, scoreUpdated.MatchID)
	if err != nil {
		return fmt.Errorf("failed to get match from read model: %w", err)
	}

	// Update score
	match.UpdateScore(scoreUpdated.HomeGoals, scoreUpdated.AwayGoals)

	// Save updated match to read model
	if err := h.matchRepo.Update(ctx, match); err != nil {
		return fmt.Errorf("failed to update match in read model: %w", err)
	}

	log.Printf("Updated match score in read model: %s %d-%d %s",
		match.HomeTeam, scoreUpdated.HomeGoals, scoreUpdated.AwayGoals, match.AwayTeam)
	return nil
}

// handleMatchStatusChanged processes MatchStatusChanged events
func (h *MatchEventHandler) handleMatchStatusChanged(ctx context.Context, event *events.Event) error {
	// Extract event data
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var statusChanged events.MatchStatusChanged
	if err := json.Unmarshal(data, &statusChanged); err != nil {
		return fmt.Errorf("failed to unmarshal MatchStatusChanged event: %w", err)
	}

	// Get existing match from read model
	match, err := h.matchRepo.GetByID(ctx, statusChanged.MatchID)
	if err != nil {
		return fmt.Errorf("failed to get match from read model: %w", err)
	}

	// Update status
	match.Status = domain.MatchStatus(statusChanged.Status)

	// Save updated match to read model
	if err := h.matchRepo.Update(ctx, match); err != nil {
		return fmt.Errorf("failed to update match in read model: %w", err)
	}

	log.Printf("Updated match status in read model: %s vs %s -> %s",
		match.HomeTeam, match.AwayTeam, match.Status)
	return nil
}
