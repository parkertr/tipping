package handlers

import (
	"context"
	"time"

	"github.com/parkertr2/footy-tipping/pkg/events"
)

// EventStore defines the interface for event storage operations needed by handlers
type EventStore interface {
	SaveEvent(ctx context.Context, event *events.Event) error
	GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error)
	GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error)
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error)
}
