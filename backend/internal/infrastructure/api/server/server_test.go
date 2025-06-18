package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/api/server"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrMatchNotFound = errors.New("match not found")
)

// Mock implementations.
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}

	return nil, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, nil
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.users[user.ID] = user

	return nil
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.users[user.ID] = user

	return nil
}

func (m *mockUserRepo) List(_ context.Context, activeOnly bool) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(m.users))
	for _, user := range m.users {
		if !activeOnly || user.IsActive {
			users = append(users, user)
		}
	}

	return users, nil
}

func (m *mockUserRepo) GetByGoogleID(_ context.Context, googleID string) (*domain.User, error) {
	for _, user := range m.users {
		if user.GoogleID == googleID {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

func (m *mockUserRepo) UpdateStats(_ context.Context, userID string, points int, isCorrect bool) error {
	if user, ok := m.users[userID]; ok {
		user.Stats.TotalPoints += points
		user.Stats.TotalPredictions++
		if isCorrect {
			user.Stats.CorrectPredictions++
		}

		return nil
	}

	return ErrUserNotFound
}

func (m *mockUserRepo) UpdateRank(_ context.Context, userID string, rank int) error {
	if user, ok := m.users[userID]; ok {
		user.Stats.CurrentRank = rank

		return nil
	}

	return ErrUserNotFound
}

type mockMatchRepo struct {
	matches map[string]*domain.Match
}

func newMockMatchRepo() *mockMatchRepo {
	return &mockMatchRepo{
		matches: make(map[string]*domain.Match),
	}
}

func (m *mockMatchRepo) GetByID(_ context.Context, id string) (*domain.Match, error) {
	if match, ok := m.matches[id]; ok {
		return match, nil
	}

	return nil, nil
}

func (m *mockMatchRepo) List(_ context.Context, filters repository.MatchFilters) ([]*domain.Match, error) {
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

func (m *mockMatchRepo) Create(_ context.Context, match *domain.Match) error {
	m.matches[match.ID] = match

	return nil
}

func (m *mockMatchRepo) Update(_ context.Context, match *domain.Match) error {
	m.matches[match.ID] = match

	return nil
}

func (m *mockMatchRepo) Delete(_ context.Context, id string) error {
	delete(m.matches, id)

	return nil
}

func TestServer_HealthCheck(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	tokenMgr := auth.NewTokenManager()
	srv := server.New(userRepo, matchRepo, tokenMgr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content type %s, got %s", "application/json", w.Header().Get("Content-Type"))
	}
}

func TestServer_Register(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	tokenMgr := auth.NewTokenManager()
	srv := server.New(userRepo, matchRepo, tokenMgr)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status code %d, got %d", http.StatusNotImplemented, w.Code)
	}
}

func TestServer_Login(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	tokenMgr := auth.NewTokenManager()
	srv := server.New(userRepo, matchRepo, tokenMgr)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", nil)
	w := httptest.NewRecorder()

	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status code %d, got %d", http.StatusNotImplemented, w.Code)
	}
}

func TestServerRoutes(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	tokenMgr := auth.NewTokenManager()
	srv := server.New(userRepo, matchRepo, tokenMgr)

	// Test routes
	testCases := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/api/v1/health", http.StatusOK},
		{"POST", "/api/v1/users", http.StatusNotImplemented},
		{"POST", "/api/v1/users/login", http.StatusNotImplemented},
		{"GET", "/api/v1/users/me", http.StatusUnauthorized},
		{"PUT", "/api/v1/users/me", http.StatusUnauthorized},
		{"GET", "/api/v1/matches", http.StatusUnauthorized},
		{"POST", "/api/v1/matches", http.StatusUnauthorized},
		{"GET", "/api/v1/matches/123", http.StatusUnauthorized},
		{"PUT", "/api/v1/matches/123", http.StatusUnauthorized},
		{"DELETE", "/api/v1/matches/123", http.StatusUnauthorized},
	}

	for _, testCase := range testCases {
		t.Run(testCase.method+" "+testCase.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(testCase.method, testCase.path, nil)
			rr := httptest.NewRecorder()
			srv.Router.ServeHTTP(rr, req)

			if rr.Code != testCase.code {
				t.Errorf("Expected status code %d, got %d", testCase.code, rr.Code)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	matchRepo := newMockMatchRepo()
	tokenMgr := auth.NewTokenManager()
	srv := server.New(userRepo, matchRepo, tokenMgr)

	// Test cases
	testCases := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			path:           "/api/v1/users/me",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid auth header format",
			path:           "/api/v1/users/me",
			authHeader:     "Invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			path:           "/api/v1/users/me",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid token but user not found",
			path:           "/api/v1/users/me",
			authHeader:     "Bearer valid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			if testCase.authHeader != "" {
				req.Header.Set("Authorization", testCase.authHeader)
			}

			rr := httptest.NewRecorder()
			srv.Router.ServeHTTP(rr, req)

			if rr.Code != testCase.expectedStatus {
				t.Errorf("Expected status code %d, got %d", testCase.expectedStatus, rr.Code)
			}
		})
	}
}
