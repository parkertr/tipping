package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/parkertr/tipping/internal/domain"
)

type PredictionRepository struct {
	db *sql.DB
}

func NewPredictionRepository(db *sql.DB) *PredictionRepository {
	return &PredictionRepository{db: db}
}

func (r *PredictionRepository) Create(ctx context.Context, prediction *domain.Prediction) error {
	query := `
		INSERT INTO predictions_view (
			id, user_id, match_id, home_goals, away_goals
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		prediction.ID,
		prediction.UserID,
		prediction.MatchID,
		prediction.HomeGoals,
		prediction.AwayGoals,
	)

	if err != nil {
		return fmt.Errorf("failed to create prediction in read model: %w", err)
	}

	return nil
}

func (r *PredictionRepository) Update(ctx context.Context, prediction *domain.Prediction) error {
	query := `
		UPDATE predictions_view
		SET home_goals = $1,
			away_goals = $2,
			points = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		prediction.HomeGoals,
		prediction.AwayGoals,
		prediction.Points,
		prediction.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update prediction in read model: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("prediction not found: %s", prediction.ID)
	}

	return nil
}

func (r *PredictionRepository) GetByID(ctx context.Context, id string) (*domain.Prediction, error) {
	query := `
		SELECT id, user_id, match_id, home_goals, away_goals, points
		FROM predictions_view
		WHERE id = $1
	`

	prediction := &domain.Prediction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&prediction.ID,
		&prediction.UserID,
		&prediction.MatchID,
		&prediction.HomeGoals,
		&prediction.AwayGoals,
		&prediction.Points,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prediction not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	return prediction, nil
}

func (r *PredictionRepository) GetByUserAndMatch(ctx context.Context, userID, matchID string) (*domain.Prediction, error) {
	query := `
		SELECT id, user_id, match_id, home_goals, away_goals, points
		FROM predictions_view
		WHERE user_id = $1 AND match_id = $2
	`

	prediction := &domain.Prediction{}
	err := r.db.QueryRowContext(ctx, query, userID, matchID).Scan(
		&prediction.ID,
		&prediction.UserID,
		&prediction.MatchID,
		&prediction.HomeGoals,
		&prediction.AwayGoals,
		&prediction.Points,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prediction not found for user %s and match %s", userID, matchID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	return prediction, nil
}

func (r *PredictionRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Prediction, error) {
	query := `
		SELECT id, user_id, match_id, home_goals, away_goals, points
		FROM predictions_view
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list predictions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var predictions []*domain.Prediction
	for rows.Next() {
		prediction := &domain.Prediction{}
		err := rows.Scan(
			&prediction.ID,
			&prediction.UserID,
			&prediction.MatchID,
			&prediction.HomeGoals,
			&prediction.AwayGoals,
			&prediction.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prediction: %w", err)
		}
		predictions = append(predictions, prediction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating predictions: %w", err)
	}

	return predictions, nil
}

func (r *PredictionRepository) ListByMatch(ctx context.Context, matchID string) ([]*domain.Prediction, error) {
	query := `
		SELECT id, user_id, match_id, home_goals, away_goals, points
		FROM predictions_view
		WHERE match_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to list predictions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var predictions []*domain.Prediction
	for rows.Next() {
		prediction := &domain.Prediction{}
		err := rows.Scan(
			&prediction.ID,
			&prediction.UserID,
			&prediction.MatchID,
			&prediction.HomeGoals,
			&prediction.AwayGoals,
			&prediction.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prediction: %w", err)
		}
		predictions = append(predictions, prediction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating predictions: %w", err)
	}

	return predictions, nil
}
