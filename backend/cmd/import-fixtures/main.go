package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/parkertr/tipping/internal/infrastructure/eventhandlers"
	"github.com/parkertr/tipping/internal/infrastructure/eventstore"
	"github.com/parkertr/tipping/internal/infrastructure/repository/postgres"
	"github.com/parkertr/tipping/pkg/events"
)

// MatchFixture represents a match fixture from the JSON file.
type MatchFixture struct {
	ID          string    `json:"id"`
	HomeTeam    string    `json:"homeTeam"`
	AwayTeam    string    `json:"awayTeam"`
	Date        time.Time `json:"date"`
	Competition string    `json:"competition"`
}

func main() {
	var (
		dbURL        = flag.String("db", "", "Database connection URL")
		fixturesFile = flag.String("fixtures", "fixtures/matches.json", "Path to fixtures JSON file")
		dryRun       = flag.Bool("dry-run", false, "Print what would be imported without actually importing")
	)

	flag.Parse()

	if *dbURL == "" {
		log.Fatal("Database URL is required. Use -db flag or set DATABASE_URL environment variable")
	}

	// Read fixtures file
	fixtures, err := readFixtures(*fixturesFile)
	if err != nil {
		log.Fatalf("Failed to read fixtures: %v", err)
	}

	fmt.Printf("Found %d fixtures to import\n", len(fixtures))

	if *dryRun {
		printDryRun(fixtures)
		return
	}

	db, eventStore, eventHandler := setupDatabase(*dbURL)
	defer db.Close()

	importFixtures(context.Background(), fixtures, eventStore, eventHandler)
}

func printDryRun(fixtures []MatchFixture) {
	fmt.Println("\nDry run mode - showing what would be imported:")
	for _, fixture := range fixtures {
		fmt.Printf("- %s vs %s (%s) on %s\n",
			fixture.HomeTeam, fixture.AwayTeam, fixture.Competition, fixture.Date.Format("2006-01-02 15:04"))
	}
}

func setupDatabase(dbURL string) (*sql.DB, *eventstore.PostgresEventStore, *eventhandlers.MatchEventHandler) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	eventStore, err := eventstore.NewPostgresEventStore(db)
	if err != nil {
		log.Fatalf("Failed to create event store: %v", err)
	}

	matchRepo := postgres.NewMatchRepository(db)
	eventHandler := eventhandlers.NewMatchEventHandler(matchRepo)

	return db, eventStore, eventHandler
}

func importFixtures(ctx context.Context, fixtures []MatchFixture, eventStore *eventstore.PostgresEventStore, eventHandler *eventhandlers.MatchEventHandler) {
	imported := 0
	skipped := 0

	for _, fixture := range fixtures {
		if shouldSkipFixture(ctx, fixture, eventStore) {
			fmt.Printf("Skipping %s vs %s - already exists\n", fixture.HomeTeam, fixture.AwayTeam)
			skipped++
			continue
		}

		if err := importFixture(ctx, fixture, eventStore, eventHandler); err != nil {
			log.Printf("Failed to import match %s vs %s: %v", fixture.HomeTeam, fixture.AwayTeam, err)
			continue
		}

		fmt.Printf("Imported: %s vs %s (%s) on %s\n",
			fixture.HomeTeam, fixture.AwayTeam, fixture.Competition, fixture.Date.Format("2006-01-02 15:04"))
		imported++
	}

	fmt.Printf("\nImport complete: %d imported, %d skipped\n", imported, skipped)
}

func shouldSkipFixture(ctx context.Context, fixture MatchFixture, eventStore *eventstore.PostgresEventStore) bool {
	existingEvents, err := eventStore.GetEvents(ctx, fixture.ID)
	if err != nil {
		log.Printf("Error checking for existing match %s: %v", fixture.ID, err)
		return true
	}
	return len(existingEvents) > 0
}

func importFixture(ctx context.Context, fixture MatchFixture, eventStore *eventstore.PostgresEventStore, eventHandler *eventhandlers.MatchEventHandler) error {
	matchCreated := events.MatchCreated{
		ID:          fixture.ID,
		HomeTeam:    fixture.HomeTeam,
		AwayTeam:    fixture.AwayTeam,
		Date:        fixture.Date,
		Competition: fixture.Competition,
	}

	event := events.NewEvent("MatchCreated", matchCreated)

	if err := eventStore.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	if err := eventHandler.HandleEvent(ctx, event); err != nil {
		log.Printf("Failed to process event for match %s vs %s: %v", fixture.HomeTeam, fixture.AwayTeam, err)
		// Continue anyway since the event is saved
	}

	return nil
}

// readFixtures reads and parses the fixtures JSON file.
func readFixtures(filename string) ([]MatchFixture, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var fixtures []MatchFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return fixtures, nil
}
