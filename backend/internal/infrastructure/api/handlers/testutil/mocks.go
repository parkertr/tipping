package testutil

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
)

// MockMatchRepo implements repository.MatchRepository for testing
type MockMatchRepo struct {
	mu      sync.RWMutex
	matches map[string]*domain.Match
}

func NewMockMatchRepo() *MockMatchRepo {
	return &MockMatchRepo{
		matches: make(map[string]*domain.Match),
	}
}

func (m *MockMatchRepo) Create(ctx context.Context, match *domain.Match) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matches[match.ID] = match
	return nil
}

func (m *MockMatchRepo) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if match, ok := m.matches[id]; ok {
		return match, nil
	}
	return nil, repository.ErrNotFound
}

func (m *MockMatchRepo) List(ctx context.Context, filters repository.MatchFilters) ([]*domain.Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	matches := make([]*domain.Match, 0, len(m.matches))
	for _, match := range m.matches {
		// Apply filters
		if filters.Competition != nil && match.Competition != *filters.Competition {
			continue
		}
		if filters.StartDate != nil && match.Date.Before(*filters.StartDate) {
			continue
		}
		if filters.EndDate != nil && match.Date.After(*filters.EndDate) {
			continue
		}
		if filters.Status != nil && string(match.Status) != *filters.Status {
			continue
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func (m *MockMatchRepo) Update(ctx context.Context, match *domain.Match) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matches[match.ID] = match
	return nil
}

func (m *MockMatchRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.matches, id)
	return nil
}

// MockEventStore implements eventstore.EventStore for testing
type MockEventStore struct {
	mu     sync.RWMutex
	events []*events.Event
}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		events: make([]*events.Event, 0),
	}
}

func (m *MockEventStore) SaveEvent(ctx context.Context, event *events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*events.Event
	for _, event := range m.events {
		if m.eventMatchesAggregateID(event, aggregateID) {
			result = append(result, event)
		}
	}
	return result, nil
}

// eventMatchesAggregateID checks if an event belongs to the given aggregate ID
// This mimics the PostgreSQL query: WHERE (data->>'ID' = $1) OR (data->>'MatchID' = $1) OR (data->>'UserID' = $1)
func (m *MockEventStore) eventMatchesAggregateID(event *events.Event, aggregateID string) bool {
	// Convert event data to JSON and back to extract fields
	data, err := json.Marshal(event.Data)
	if err != nil {
		return false
	}

	var eventData map[string]interface{}
	if err := json.Unmarshal(data, &eventData); err != nil {
		return false
	}

	// Check if any of the key fields match the aggregate ID
	if id, ok := eventData["id"].(string); ok && id == aggregateID {
		return true
	}
	if matchID, ok := eventData["matchId"].(string); ok && matchID == aggregateID {
		return true
	}
	if userID, ok := eventData["userId"].(string); ok && userID == aggregateID {
		return true
	}

	return false
}

func (m *MockEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*events.Event
	for _, event := range m.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*events.Event
	for _, event := range m.events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}
	return result, nil
}

// MockPredictionRepo implements repository.PredictionRepository for testing
type MockPredictionRepo struct {
	mu          sync.RWMutex
	predictions map[string]*domain.Prediction
}

func NewMockPredictionRepo() *MockPredictionRepo {
	return &MockPredictionRepo{
		predictions: make(map[string]*domain.Prediction),
	}
}

func (m *MockPredictionRepo) Create(ctx context.Context, prediction *domain.Prediction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.predictions[prediction.ID] = prediction
	return nil
}

func (m *MockPredictionRepo) GetByID(ctx context.Context, id string) (*domain.Prediction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if prediction, ok := m.predictions[id]; ok {
		return prediction, nil
	}
	return nil, repository.ErrNotFound
}

func (m *MockPredictionRepo) GetByUserAndMatch(
	ctx context.Context, userID, matchID string,
) (*domain.Prediction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, prediction := range m.predictions {
		if prediction.UserID == userID && prediction.MatchID == matchID {
			return prediction, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *MockPredictionRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Prediction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var predictions []*domain.Prediction
	for _, prediction := range m.predictions {
		if prediction.UserID == userID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *MockPredictionRepo) ListByMatch(ctx context.Context, matchID string) ([]*domain.Prediction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var predictions []*domain.Prediction
	for _, prediction := range m.predictions {
		if prediction.MatchID == matchID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *MockPredictionRepo) Update(ctx context.Context, prediction *domain.Prediction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.predictions[prediction.ID] = prediction
	return nil
}

func (m *MockPredictionRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.predictions, id)
	return nil
}
