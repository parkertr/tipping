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

	_ "github.com/lib/pq"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/eventstore"
	"github.com/parkertr2/footy-tipping/pkg/events"
)

// MatchFixture represents a match fixture from the JSON file
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
		fmt.Println("\nDry run mode - showing what would be imported:")
		for _, fixture := range fixtures {
			fmt.Printf("- %s vs %s (%s) on %s\n",
				fixture.HomeTeam, fixture.AwayTeam, fixture.Competition, fixture.Date.Format("2006-01-02 15:04"))
		}
		return
	}

	// Connect to database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create event store
	eventStore, err := eventstore.NewPostgresEventStore(db)
	if err != nil {
		log.Fatalf("Failed to create event store: %v", err)
	}

	// Import fixtures
	ctx := context.Background()
	imported := 0
	skipped := 0

	for _, fixture := range fixtures {
		// Check if match already exists
		existingEvents, err := eventStore.GetEvents(ctx, fixture.ID)
		if err != nil {
			log.Printf("Error checking for existing match %s: %v", fixture.ID, err)
			continue
		}

		if len(existingEvents) > 0 {
			fmt.Printf("Skipping %s vs %s - already exists\n", fixture.HomeTeam, fixture.AwayTeam)
			skipped++
			continue
		}

		// Create MatchCreated event
		matchCreated := events.MatchCreated{
			ID:          fixture.ID,
			HomeTeam:    fixture.HomeTeam,
			AwayTeam:    fixture.AwayTeam,
			Date:        fixture.Date,
			Competition: fixture.Competition,
		}

		event := events.NewEvent("MatchCreated", matchCreated)

		// Save event
		if err := eventStore.SaveEvent(ctx, event); err != nil {
			log.Printf("Failed to import match %s vs %s: %v", fixture.HomeTeam, fixture.AwayTeam, err)
			continue
		}

		fmt.Printf("Imported: %s vs %s (%s) on %s\n",
			fixture.HomeTeam, fixture.AwayTeam, fixture.Competition, fixture.Date.Format("2006-01-02 15:04"))
		imported++
	}

	fmt.Printf("\nImport complete: %d imported, %d skipped\n", imported, skipped)
}

// readFixtures reads and parses the fixtures JSON file
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
