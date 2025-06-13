package repository

import (
	"context"

	"github.com/parkertr2/footy-tipping/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByGoogleID retrieves a user by their Google ID
	GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error)

	// GetByEmail retrieves a user by their email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update updates a user's information
	Update(ctx context.Context, user *domain.User) error

	// List retrieves all users with optional filters
	List(ctx context.Context, activeOnly bool) ([]*domain.User, error)

	// UpdateStats updates a user's statistics
	UpdateStats(ctx context.Context, userID string, points int, isCorrect bool) error

	// UpdateRank updates a user's rank
	UpdateRank(ctx context.Context, userID string, rank int) error
}
