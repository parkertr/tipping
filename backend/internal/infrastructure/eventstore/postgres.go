package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// Blank import for PostgreSQL driver.
	_ "github.com/lib/pq"
	"github.com/parkertr/tipping/pkg/events"
)

// PostgresEventStore implements EventStore using PostgreSQL.
type PostgresEventStore struct {
	db *sql.DB
}

// NewPostgresEventStore creates a new PostgreSQL event store.
func NewPostgresEventStore(db *sql.DB) (*PostgresEventStore, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresEventStore{db: db}, nil
}

// SaveEvent persists an event to PostgreSQL.
func (store *PostgresEventStore) SaveEvent(ctx context.Context, event *events.Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		INSERT INTO events (id, type, data, timestamp, version)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = store.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		data,
		event.Timestamp,
		event.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for a given aggregate ID.
func (store *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	query := `
		SELECT id, type, data, timestamp, version
		FROM events
		WHERE (data->>'ID' = $1) OR (data->>'MatchID' = $1) OR (data->>'UserID' = $1)
		ORDER BY timestamp ASC
	`

	rows, err := store.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for aggregateID %s: %w", aggregateID, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var result []*events.Event

	for rows.Next() {
		var event events.Event

		var data []byte
		if err := rows.Scan(&event.ID, &event.Type, &data, &event.Timestamp, &event.Version); err != nil {
			return nil, fmt.Errorf("failed to scan event for aggregateID %s: %w", aggregateID, err)
		}

		// Unmarshal the event data based on the event type
		switch event.Type {
		case "MatchCreated":
			var matchCreated events.MatchCreated
			if err := json.Unmarshal(data, &matchCreated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchCreated for event ID %s: %w", event.ID, err)
			}

			event.Data = matchCreated
		case "MatchScoreUpdated":
			var scoreUpdated events.MatchScoreUpdated
			if err := json.Unmarshal(data, &scoreUpdated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchScoreUpdated for event ID %s: %w", event.ID, err)
			}

			event.Data = scoreUpdated
		case "MatchStatusChanged":
			var statusChanged events.MatchStatusChanged
			if err := json.Unmarshal(data, &statusChanged); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchStatusChanged for event ID %s: %w", event.ID, err)
			}

			event.Data = statusChanged
		case "PredictionMade":
			var predictionMade events.PredictionMade
			if err := json.Unmarshal(data, &predictionMade); err != nil {
				return nil, fmt.Errorf("failed to unmarshal PredictionMade for event ID %s: %w", event.ID, err)
			}

			event.Data = predictionMade
		}

		result = append(result, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events for aggregateID %s: %w", aggregateID, err)
	}

	return result, nil
}

// GetEventsByType retrieves all events of a specific type.
func (store *PostgresEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error) {
	query := `
		SELECT id, type, data, timestamp, version
		FROM events
		WHERE type = $1
		ORDER BY timestamp ASC
	`

	rows, err := store.db.QueryContext(ctx, query, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for eventType %s: %w", eventType, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var result []*events.Event

	for rows.Next() {
		var event events.Event

		var data []byte
		if err := rows.Scan(&event.ID, &event.Type, &data, &event.Timestamp, &event.Version); err != nil {
			return nil, fmt.Errorf("failed to scan event for eventType %s: %w", eventType, err)
		}

		// Unmarshal the event data based on the event type
		switch event.Type {
		case "MatchCreated":
			var matchCreated events.MatchCreated
			if err := json.Unmarshal(data, &matchCreated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchCreated for event ID %s: %w", event.ID, err)
			}

			event.Data = matchCreated
		case "MatchScoreUpdated":
			var scoreUpdated events.MatchScoreUpdated
			if err := json.Unmarshal(data, &scoreUpdated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchScoreUpdated for event ID %s: %w", event.ID, err)
			}

			event.Data = scoreUpdated
		case "MatchStatusChanged":
			var statusChanged events.MatchStatusChanged
			if err := json.Unmarshal(data, &statusChanged); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchStatusChanged for event ID %s: %w", event.ID, err)
			}

			event.Data = statusChanged
		case "PredictionMade":
			var predictionMade events.PredictionMade
			if err := json.Unmarshal(data, &predictionMade); err != nil {
				return nil, fmt.Errorf("failed to unmarshal PredictionMade for event ID %s: %w", event.ID, err)
			}

			event.Data = predictionMade
		}

		result = append(result, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events for eventType %s: %w", eventType, err)
	}

	return result, nil
}

// GetEventsByTimeRange retrieves events within a time range.
func (store *PostgresEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error) {
	query := `
		SELECT id, type, data, timestamp, version
		FROM events
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp ASC
	`

	rows, err := store.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for time range %s to %s: %w", start.Format(time.RFC3339), end.Format(time.RFC3339), err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var result []*events.Event

	for rows.Next() {
		var event events.Event

		var data []byte
		if err := rows.Scan(&event.ID, &event.Type, &data, &event.Timestamp, &event.Version); err != nil {
			return nil, fmt.Errorf("failed to scan event for time range %s to %s: %w", start.Format(time.RFC3339), end.Format(time.RFC3339), err)
		}

		// Unmarshal the event data based on the event type
		switch event.Type {
		case "MatchCreated":
			var matchCreated events.MatchCreated
			if err := json.Unmarshal(data, &matchCreated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchCreated for event ID %s: %w", event.ID, err)
			}

			event.Data = matchCreated
		case "MatchScoreUpdated":
			var scoreUpdated events.MatchScoreUpdated
			if err := json.Unmarshal(data, &scoreUpdated); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchScoreUpdated for event ID %s: %w", event.ID, err)
			}

			event.Data = scoreUpdated
		case "MatchStatusChanged":
			var statusChanged events.MatchStatusChanged
			if err := json.Unmarshal(data, &statusChanged); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MatchStatusChanged for event ID %s: %w", event.ID, err)
			}

			event.Data = statusChanged
		case "PredictionMade":
			var predictionMade events.PredictionMade
			if err := json.Unmarshal(data, &predictionMade); err != nil {
				return nil, fmt.Errorf("failed to unmarshal PredictionMade for event ID %s: %w", event.ID, err)
			}

			event.Data = predictionMade
		}

		result = append(result, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events for time range %s to %s: %w", start.Format(time.RFC3339), end.Format(time.RFC3339), err)
	}

	return result, nil
}
