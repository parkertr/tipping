package repository

import (
	"context"
	"time"

	"github.com/parkertr2/footy-tipping/internal/domain"
)

// MatchRepository defines the interface for match read model operations
type MatchRepository interface {
	// Create creates a new match in the read model
	Create(ctx context.Context, match *domain.Match) error

	// Update updates an existing match in the read model
	Update(ctx context.Context, match *domain.Match) error

	// GetByID retrieves a match by its ID
	GetByID(ctx context.Context, id string) (*domain.Match, error)

	// List retrieves all matches with optional filters
	List(ctx context.Context, filters MatchFilters) ([]*domain.Match, error)
}

// PredictionRepository defines the interface for prediction read model operations
type PredictionRepository interface {
	// Create creates a new prediction in the read model
	Create(ctx context.Context, prediction *domain.Prediction) error

	// Update updates an existing prediction in the read model
	Update(ctx context.Context, prediction *domain.Prediction) error

	// GetByID retrieves a prediction by its ID
	GetByID(ctx context.Context, id string) (*domain.Prediction, error)

	// GetByUserAndMatch retrieves a prediction by user ID and match ID
	GetByUserAndMatch(ctx context.Context, userID, matchID string) (*domain.Prediction, error)

	// ListByUser retrieves all predictions for a user
	ListByUser(ctx context.Context, userID string) ([]*domain.Prediction, error)

	// ListByMatch retrieves all predictions for a match
	ListByMatch(ctx context.Context, matchID string) ([]*domain.Prediction, error)
}

// MatchFilters defines the available filters for listing matches
type MatchFilters struct {
	Competition *string    // Filter by competition
	StartDate   *time.Time // Filter by start date
	EndDate     *time.Time // Filter by end date
	Status      *string    // Filter by match status
}
