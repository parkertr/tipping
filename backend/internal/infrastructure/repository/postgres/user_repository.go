package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
)

// UserRepository implements repository.UserRepository for PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create implements repository.UserRepository.
func (repo *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users_view (
			id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := repo.db.ExecContext(ctx, query,
		user.ID,
		user.GoogleID,
		user.Email,
		user.Name,
		user.Picture,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
		user.Stats.TotalPoints,
		user.Stats.CorrectPredictions,
		user.Stats.TotalPredictions,
		user.Stats.CurrentRank,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID implements repository.UserRepository.
func (repo *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE id = $1
	`

	return repo.queryUser(ctx, query, userID)
}

// GetByGoogleID implements repository.UserRepository.
func (repo *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE google_id = $1
	`

	return repo.queryUser(ctx, query, googleID)
}

// GetByEmail implements repository.UserRepository.
func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE email = $1
	`

	return repo.queryUser(ctx, query, email)
}

// Update implements repository.UserRepository.
func (repo *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users_view
		SET name = $1, picture_url = $2, updated_at = $3, is_active = $4
		WHERE id = $5
	`

	_, err := repo.db.ExecContext(ctx, query,
		user.Name,
		user.Picture,
		user.UpdatedAt,
		user.IsActive,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// List implements repository.UserRepository.
func (repo *UserRepository) List(ctx context.Context, activeOnly bool) ([]*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
	`
	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY total_points DESC"

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil {
			// Log the error but don't fail the request
			log.Printf("Error closing rows: %v", cerr)
		}
	}()

	var users []*domain.User

	for rows.Next() {
		user, err := repo.scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UpdateStats implements repository.UserRepository.
func (repo *UserRepository) UpdateStats(ctx context.Context, userID string, points int, isCorrect bool) error {
	query := `
		UPDATE users_view
		SET total_points = total_points + $1,
			total_predictions = total_predictions + 1,
			correct_predictions = correct_predictions + $2
		WHERE id = $3
	`

	correct := 0
	if isCorrect {
		correct = 1
	}

	_, err := repo.db.ExecContext(ctx, query, points, correct, userID)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

// UpdateRank implements repository.UserRepository.
func (repo *UserRepository) UpdateRank(ctx context.Context, userID string, rank int) error {
	query := `
		UPDATE users_view
		SET current_rank = $1
		WHERE id = $2
	`

	_, err := repo.db.ExecContext(ctx, query, rank, userID)
	if err != nil {
		return fmt.Errorf("failed to update user rank: %w", err)
	}

	return nil
}

// Helper function to scan a user from a row.
func (repo *UserRepository) scanUser(rows *sql.Rows) (*domain.User, error) {
	var user domain.User

	var stats domain.UserStats

	err := rows.Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&stats.TotalPoints,
		&stats.CorrectPredictions,
		&stats.TotalPredictions,
		&stats.CurrentRank,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}

	user.Stats = stats

	return &user, nil
}

// Helper function to query a single user.
func (repo *UserRepository) queryUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	row := repo.db.QueryRowContext(ctx, query, args...)

	var user domain.User

	var stats domain.UserStats

	err := row.Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&stats.TotalPoints,
		&stats.CorrectPredictions,
		&stats.TotalPredictions,
		&stats.CurrentRank,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}

	user.Stats = stats

	return &user, nil
}
