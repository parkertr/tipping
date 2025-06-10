package eventstore

import (
	"context"
	"time"

	"github.com/parkertr2/footy-tipping/pkg/events"
)

// EventStore defines the interface for storing and retrieving events
type EventStore interface {
	// SaveEvent persists an event to the store
	SaveEvent(ctx context.Context, event *events.Event) error

	// GetEvents retrieves all events for a given aggregate ID
	GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error)

	// GetEventsByType retrieves all events of a specific type
	GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error)

	// GetEventsByTimeRange retrieves events within a time range
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error)
}
