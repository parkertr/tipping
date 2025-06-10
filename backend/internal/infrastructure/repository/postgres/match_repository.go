package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository"
)

type MatchRepository struct {
	db *sql.DB
}

func NewMatchRepository(db *sql.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(ctx context.Context, match *domain.Match) error {
	var homeGoals, awayGoals sql.NullInt32
	if match.Score != nil {
		homeGoals = sql.NullInt32{Int32: int32(match.Score.HomeGoals), Valid: true}
		awayGoals = sql.NullInt32{Int32: int32(match.Score.AwayGoals), Valid: true}
	}

	query := `
		INSERT INTO matches_view (
			id, home_team, away_team, match_date, competition, status, home_goals, away_goals
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		match.ID,
		match.HomeTeam,
		match.AwayTeam,
		match.Date,
		match.Competition,
		match.Status,
		homeGoals,
		awayGoals,
	)

	if err != nil {
		return fmt.Errorf("failed to create match in read model: %w", err)
	}

	return nil
}

func (r *MatchRepository) Update(ctx context.Context, match *domain.Match) error {
	var homeGoals, awayGoals sql.NullInt32
	if match.Score != nil {
		homeGoals = sql.NullInt32{Int32: int32(match.Score.HomeGoals), Valid: true}
		awayGoals = sql.NullInt32{Int32: int32(match.Score.AwayGoals), Valid: true}
	}

	query := `
		UPDATE matches_view
		SET home_team = $1,
			away_team = $2,
			match_date = $3,
			competition = $4,
			status = $5,
			home_goals = $6,
			away_goals = $7,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		match.HomeTeam,
		match.AwayTeam,
		match.Date,
		match.Competition,
		match.Status,
		homeGoals,
		awayGoals,
		match.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update match in read model: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("match not found: %s", match.ID)
	}

	return nil
}

func (r *MatchRepository) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	query := `
		SELECT id, home_team, away_team, match_date, competition, status, home_goals, away_goals
		FROM matches_view
		WHERE id = $1
	`

	var homeGoals, awayGoals sql.NullInt32
	match := &domain.Match{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.HomeTeam,
		&match.AwayTeam,
		&match.Date,
		&match.Competition,
		&match.Status,
		&homeGoals,
		&awayGoals,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if homeGoals.Valid && awayGoals.Valid {
		match.Score = &domain.Score{
			HomeGoals: int(homeGoals.Int32),
			AwayGoals: int(awayGoals.Int32),
		}
	}

	return match, nil
}

func (r *MatchRepository) List(ctx context.Context, filters repository.MatchFilters) ([]*domain.Match, error) {
	var conditions []string
	var args []interface{}
	argPos := 1

	if filters.Competition != nil {
		conditions = append(conditions, fmt.Sprintf("competition = $%d", argPos))
		args = append(args, *filters.Competition)
		argPos++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("match_date >= $%d", argPos))
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("match_date <= $%d", argPos))
		args = append(args, *filters.EndDate)
		argPos++
	}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *filters.Status)
		argPos++
	}

	query := `
		SELECT id, home_team, away_team, match_date, competition, status, home_goals, away_goals
		FROM matches_view
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY match_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list matches: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		var homeGoals, awayGoals sql.NullInt32
		match := &domain.Match{}
		err := rows.Scan(
			&match.ID,
			&match.HomeTeam,
			&match.AwayTeam,
			&match.Date,
			&match.Competition,
			&match.Status,
			&homeGoals,
			&awayGoals,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}

		if homeGoals.Valid && awayGoals.Valid {
			match.Score = &domain.Score{
				HomeGoals: int(homeGoals.Int32),
				AwayGoals: int(awayGoals.Int32),
			}
		}

		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}
