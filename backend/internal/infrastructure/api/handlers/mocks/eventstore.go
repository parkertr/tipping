package mocks

import (
	"context"
	"time"

	"github.com/parkertr/tipping/pkg/events"
	"github.com/stretchr/testify/mock"
)

// MockEventStore is a mock implementation of the EventStore interface
type MockEventStore struct {
	mock.Mock
}

func (m *MockEventStore) SaveEvent(ctx context.Context, event *events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	args := m.Called(ctx, aggregateID)
	return args.Get(0).([]*events.Event), args.Error(1)
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error) {
	args := m.Called(ctx, eventType)
	return args.Get(0).([]*events.Event), args.Error(1)
}

func (m *MockEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error) {
	args := m.Called(ctx, start, end)
	return args.Get(0).([]*events.Event), args.Error(1)
}
