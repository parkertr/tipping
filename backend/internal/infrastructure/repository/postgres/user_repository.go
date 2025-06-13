package postgres

import (
	"context"
	"database/sql"

	"github.com/parkertr2/footy-tipping/internal/domain"
)

// UserRepository implements repository.UserRepository for PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create implements repository.UserRepository
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users_view (
			id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
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
	return err
}

// GetByID implements repository.UserRepository
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE id = $1
	`
	return r.queryUser(ctx, query, id)
}

// GetByGoogleID implements repository.UserRepository
func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE google_id = $1
	`
	return r.queryUser(ctx, query, googleID)
}

// GetByEmail implements repository.UserRepository
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, google_id, email, name, picture_url,
			created_at, updated_at, is_active,
			total_points, correct_predictions, total_predictions, current_rank
		FROM users_view
		WHERE email = $1
	`
	return r.queryUser(ctx, query, email)
}

// Update implements repository.UserRepository
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users_view
		SET name = $1, picture_url = $2, updated_at = $3, is_active = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Name,
		user.Picture,
		user.UpdatedAt,
		user.IsActive,
		user.ID,
	)
	return err
}

// List implements repository.UserRepository
func (r *UserRepository) List(ctx context.Context, activeOnly bool) ([]*domain.User, error) {
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

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// UpdateStats implements repository.UserRepository
func (r *UserRepository) UpdateStats(ctx context.Context, userID string, points int, isCorrect bool) error {
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
	_, err := r.db.ExecContext(ctx, query, points, correct, userID)
	return err
}

// UpdateRank implements repository.UserRepository
func (r *UserRepository) UpdateRank(ctx context.Context, userID string, rank int) error {
	query := `
		UPDATE users_view
		SET current_rank = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, rank, userID)
	return err
}

// Helper function to scan a user from a row
func (r *UserRepository) scanUser(rows *sql.Rows) (*domain.User, error) {
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
		return nil, err
	}
	user.Stats = stats
	return &user, nil
}

// Helper function to query a single user
func (r *UserRepository) queryUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user.Stats = stats
	return &user, nil
}
