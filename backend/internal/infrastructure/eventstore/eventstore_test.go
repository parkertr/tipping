package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/parkertr2/footy-tipping/pkg/events"
)

func TestNewPostgresEventStore(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		_ = db.Close() // Ignore close errors for mock database
	}()

	// Set up expectations for database ping
	mock.ExpectPing()

	// Create event store
	store, err := NewPostgresEventStore(db)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if store == nil {
		t.Errorf("expected store to be non-nil")
	}
}

func TestSaveEvent(t *testing.T) {
	// Test case 1: Save match created event
	t.Run("Save match created event", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer func() {
			_ = db.Close() // Ignore close errors for mock database
		}()

		// Set up expectations for database ping
		mock.ExpectPing()

		// Create event store
		store, err := NewPostgresEventStore(db)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.Background()
		matchCreated := &events.MatchCreated{
			ID:          "match123",
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        time.Now(),
			Competition: "Premier League",
		}

		event := events.NewEvent("MatchCreated", matchCreated)

		// Set up expectations for event insertion
		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Type, sqlmock.AnyArg(), event.Timestamp, event.Version).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Save event
		err = store.SaveEvent(ctx, event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// Test case 2: Save prediction made event
	t.Run("Save prediction made event", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer func() {
			_ = db.Close() // Ignore close errors for mock database
		}()

		// Set up expectations for database ping
		mock.ExpectPing()

		// Create event store
		store, err := NewPostgresEventStore(db)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.Background()
		predictionMade := &events.PredictionMade{
			ID:        "pred123",
			UserID:    "user123",
			MatchID:   "match123",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: time.Now(),
		}

		event := events.NewEvent("PredictionMade", predictionMade)

		// Set up expectations for event insertion
		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Type, sqlmock.AnyArg(), event.Timestamp, event.Version).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Save event
		err = store.SaveEvent(ctx, event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestGetEvents(t *testing.T) {
	// Test case 1: Get match events
	t.Run("Get match events", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer func() {
			_ = db.Close() // Ignore close errors for mock database
		}()

		// Set up expectations for database ping
		mock.ExpectPing()

		// Create event store
		store, err := NewPostgresEventStore(db)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.Background()
		matchID := "match123"
		now := time.Now()
		matchCreated := events.MatchCreated{
			ID:          matchID,
			HomeTeam:    "Team A",
			AwayTeam:    "Team B",
			Date:        now,
			Competition: "Premier League",
		}

		event := events.NewEvent("MatchCreated", matchCreated)
		eventData, err := json.Marshal(matchCreated)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Set up expectations for event retrieval
		rows := sqlmock.NewRows([]string{"id", "type", "data", "timestamp", "version"}).
			AddRow(event.ID, event.Type, eventData, event.Timestamp, event.Version)

		mock.ExpectQuery(`SELECT id, type, data, timestamp, version
                        FROM events
                        WHERE \(data->>'ID' = \$1\) OR \(data->>'MatchID' = \$1\) OR \(data->>'UserID' = \$1\)
                        ORDER BY timestamp ASC`).
			WithArgs(matchID).
			WillReturnRows(rows)

		// Get events
		events, err := store.GetEvents(ctx, matchID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}

		// Verify event data
		if events[0].Type != event.Type {
			t.Errorf("expected event type %v, got %v", event.Type, events[0].Type)
		}

		// Compare individual fields of MatchCreated
		data, err := json.Marshal(events[0].Data)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		var resultData map[string]interface{}
		err = json.Unmarshal(data, &resultData)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if matchCreated.ID != resultData["ID"] {
			t.Errorf("expected ID %v, got %v", matchCreated.ID, resultData["ID"])
		}
		if matchCreated.HomeTeam != resultData["HomeTeam"] {
			t.Errorf("expected HomeTeam %v, got %v", matchCreated.HomeTeam, resultData["HomeTeam"])
		}
		if matchCreated.AwayTeam != resultData["AwayTeam"] {
			t.Errorf("expected AwayTeam %v, got %v", matchCreated.AwayTeam, resultData["AwayTeam"])
		}
		if matchCreated.Competition != resultData["Competition"] {
			t.Errorf("expected Competition %v, got %v", matchCreated.Competition, resultData["Competition"])
		}
	})

	// Test case 2: Get prediction events
	t.Run("Get prediction events", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer func() {
			_ = db.Close() // Ignore close errors for mock database
		}()

		// Set up expectations for database ping
		mock.ExpectPing()

		// Create event store
		store, err := NewPostgresEventStore(db)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.Background()
		userID := "user123"
		now := time.Now()
		predictionMade := events.PredictionMade{
			ID:        "pred123",
			UserID:    userID,
			MatchID:   "match123",
			HomeGoals: 2,
			AwayGoals: 1,
			CreatedAt: now,
		}

		event := events.NewEvent("PredictionMade", predictionMade)
		eventData, err := json.Marshal(predictionMade)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Set up expectations for event retrieval
		rows := sqlmock.NewRows([]string{"id", "type", "data", "timestamp", "version"}).
			AddRow(event.ID, event.Type, eventData, event.Timestamp, event.Version)

		mock.ExpectQuery(`SELECT id, type, data, timestamp, version
                        FROM events
                        WHERE \(data->>'ID' = \$1\) OR \(data->>'MatchID' = \$1\) OR \(data->>'UserID' = \$1\)
                        ORDER BY timestamp ASC`).
			WithArgs(userID).
			WillReturnRows(rows)

		// Get events
		events, err := store.GetEvents(ctx, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}

		// Verify event data
		if events[0].Type != event.Type {
			t.Errorf("expected event type %v, got %v", event.Type, events[0].Type)
		}

		// Compare individual fields of PredictionMade
		data, err := json.Marshal(events[0].Data)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		var resultData map[string]interface{}
		err = json.Unmarshal(data, &resultData)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if predictionMade.ID != resultData["ID"] {
			t.Errorf("expected ID %v, got %v", predictionMade.ID, resultData["ID"])
		}
		if predictionMade.UserID != resultData["UserID"] {
			t.Errorf("expected UserID %v, got %v", predictionMade.UserID, resultData["UserID"])
		}
		if predictionMade.MatchID != resultData["MatchID"] {
			t.Errorf("expected MatchID %v, got %v", predictionMade.MatchID, resultData["MatchID"])
		}
		if float64(predictionMade.HomeGoals) != resultData["HomeGoals"] {
			t.Errorf("expected HomeGoals %v, got %v", predictionMade.HomeGoals, resultData["HomeGoals"])
		}
		if float64(predictionMade.AwayGoals) != resultData["AwayGoals"] {
			t.Errorf("expected AwayGoals %v, got %v", predictionMade.AwayGoals, resultData["AwayGoals"])
		}
	})
}
