package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/api/server"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
)

// Mock implementations
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (mockUserRepo *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	mockUserRepo.users[user.ID] = user
	return nil
}

func (mockUserRepo *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if user, ok := mockUserRepo.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (mockUserRepo *mockUserRepo) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	for _, user := range mockUserRepo.users {
		if user.GoogleID == googleID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (mockUserRepo *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range mockUserRepo.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (mockUserRepo *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	mockUserRepo.users[user.ID] = user
	return nil
}

func (mockUserRepo *mockUserRepo) List(ctx context.Context, activeOnly bool) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(mockUserRepo.users))
	for _, user := range mockUserRepo.users {
		users = append(users, user)
	}
	return users, nil
}

func (mockUserRepo *mockUserRepo) UpdateStats(ctx context.Context, userID string, points int, isCorrect bool) error {
	if user, ok := mockUserRepo.users[userID]; ok {
		user.Stats.TotalPoints += points
		return nil
	}
	return errors.New("user not found")
}

func (mockUserRepo *mockUserRepo) UpdateRank(ctx context.Context, userID string, rank int) error {
	if user, ok := mockUserRepo.users[userID]; ok {
		user.Stats.CurrentRank = rank
		return nil
	}
	return errors.New("user not found")
}

type mockMatchRepo struct {
	matches map[string]*domain.Match
}

func newMockMatchRepo() *mockMatchRepo {
	return &mockMatchRepo{
		matches: make(map[string]*domain.Match),
	}
}

func (mockMatchRepo *mockMatchRepo) Create(ctx context.Context, match *domain.Match) error {
	mockMatchRepo.matches[match.ID] = match
	return nil
}

func (mockMatchRepo *mockMatchRepo) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	if match, ok := mockMatchRepo.matches[id]; ok {
		return match, nil
	}
	return nil, errors.New("match not found")
}

func (mockMatchRepo *mockMatchRepo) List(ctx context.Context, filters repository.MatchFilters) ([]*domain.Match, error) {
	matches := make([]*domain.Match, 0, len(mockMatchRepo.matches))
	for _, match := range mockMatchRepo.matches {
		matches = append(matches, match)
	}
	return matches, nil
}

func (mockMatchRepo *mockMatchRepo) Update(ctx context.Context, match *domain.Match) error {
	mockMatchRepo.matches[match.ID] = match
	return nil
}

type mockPredictionRepo struct {
	predictions map[string]*domain.Prediction
}

func newMockPredictionRepo() *mockPredictionRepo {
	return &mockPredictionRepo{
		predictions: make(map[string]*domain.Prediction),
	}
}

func (m *mockPredictionRepo) Create(ctx context.Context, prediction *domain.Prediction) error {
	m.predictions[prediction.ID] = prediction
	return nil
}

func (m *mockPredictionRepo) GetByID(ctx context.Context, id string) (*domain.Prediction, error) {
	if prediction, ok := m.predictions[id]; ok {
		return prediction, nil
	}
	return nil, nil
}

func (m *mockPredictionRepo) GetByUserAndMatch(ctx context.Context, userID, matchID string) (*domain.Prediction, error) {
	for _, prediction := range m.predictions {
		if prediction.UserID == userID && prediction.MatchID == matchID {
			return prediction, nil
		}
	}
	return nil, nil
}

func (m *mockPredictionRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Prediction, error) {
	predictions := make([]*domain.Prediction, 0)
	for _, prediction := range m.predictions {
		if prediction.UserID == userID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *mockPredictionRepo) ListByMatch(ctx context.Context, matchID string) ([]*domain.Prediction, error) {
	predictions := make([]*domain.Prediction, 0)
	for _, prediction := range m.predictions {
		if prediction.MatchID == matchID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *mockPredictionRepo) Update(ctx context.Context, prediction *domain.Prediction) error {
	m.predictions[prediction.ID] = prediction
	return nil
}

type mockEventStore struct {
	events map[string][]*events.Event
}

func newMockEventStore() *mockEventStore {
	return &mockEventStore{
		events: make(map[string][]*events.Event),
	}
}

func (m *mockEventStore) SaveEvent(ctx context.Context, event *events.Event) error {
	m.events[event.ID] = append(m.events[event.ID], event)
	return nil
}

func (m *mockEventStore) GetEvents(ctx context.Context, id string) ([]*events.Event, error) {
	return m.events[id], nil
}

func (m *mockEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error) {
	var result []*events.Event
	for _, events := range m.events {
		for _, event := range events {
			if event.Type == eventType {
				result = append(result, event)
			}
		}
	}
	return result, nil
}

func (m *mockEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*events.Event, error) {
	var result []*events.Event
	for _, events := range m.events {
		for _, event := range events {
			if event.Timestamp.After(start) && event.Timestamp.Before(end) {
				result = append(result, event)
			}
		}
	}
	return result, nil
}

func TestNewServer(t *testing.T) {
	t.Parallel()
	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	predictionRepo := newMockPredictionRepo()
	eventStore := newMockEventStore()

	srv := server.NewServer(userRepo, matchRepo, predictionRepo, eventStore)
	if srv == nil {
		t.Fatal("Expected server to be created")
	}
}

func TestServerRoutes(t *testing.T) {
	t.Parallel()
	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	predictionRepo := newMockPredictionRepo()
	eventStore := newMockEventStore()

	// Create test data
	user := &domain.User{
		ID:       "123",
		GoogleID: "google123",
		Email:    "test@example.com",
		Name:     "Test User",
	}
	userRepo.Create(context.Background(), user)

	match := &domain.Match{
		ID:          "123",
		HomeTeam:    "Home",
		AwayTeam:    "Away",
		Date:        time.Now(),
		Competition: "Test League",
		Status:      domain.MatchStatusScheduled,
		Score:       &domain.Score{HomeGoals: 0, AwayGoals: 0},
	}
	matchRepo.Create(context.Background(), match)

	prediction := &domain.Prediction{
		ID:        "123",
		UserID:    "123",
		MatchID:   "123",
		HomeGoals: 2,
		AwayGoals: 1,
		CreatedAt: time.Now(),
	}
	predictionRepo.Create(context.Background(), prediction)

	srv := server.NewServer(userRepo, matchRepo, predictionRepo, eventStore)

	// Test routes
	testCases := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/api/auth/google", http.StatusTemporaryRedirect},
		{"GET", "/api/auth/google/callback", http.StatusBadRequest},
		{"POST", "/api/auth/refresh", http.StatusUnauthorized},
		{"GET", "/api/auth/me", http.StatusUnauthorized},
		{"PUT", "/api/auth/me", http.StatusUnauthorized},
		{"POST", "/api/auth/me/deactivate", http.StatusUnauthorized},
		{"GET", "/api/auth/me/stats", http.StatusUnauthorized},
		{"GET", "/api/auth/me/ranking", http.StatusUnauthorized},
		{"POST", "/api/matches", http.StatusUnauthorized},
		{"GET", "/api/matches", http.StatusUnauthorized},
		{"GET", "/api/matches/123", http.StatusUnauthorized},
		{"PUT", "/api/matches/123/score", http.StatusUnauthorized},
		{"POST", "/api/predictions", http.StatusUnauthorized},
		{"GET", "/api/users/123/predictions", http.StatusUnauthorized},
		{"GET", "/api/matches/123/predictions", http.StatusUnauthorized},
		{"GET", "/api/matches/123/predictions/456", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rr, req)
			if rr.Code != tc.code {
				t.Errorf("Expected status code %d, got %d", tc.code, rr.Code)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()
	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	predictionRepo := newMockPredictionRepo()
	eventStore := newMockEventStore()

	srv := server.NewServer(userRepo, matchRepo, predictionRepo, eventStore)

	// Test cases
	testCases := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			path:           "/api/auth/me",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid auth header format",
			path:           "/api/auth/me",
			authHeader:     "Invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			path:           "/api/auth/me",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid token but user not found",
			path:           "/api/auth/me",
			authHeader:     "Bearer valid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}
