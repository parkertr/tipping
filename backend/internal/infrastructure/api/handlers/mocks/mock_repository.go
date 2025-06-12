package mocks

import (
	"context"

	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository"
	"github.com/stretchr/testify/mock"
)

// MockMatchRepository is a mock implementation of repository.MatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(ctx context.Context, match *domain.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) Update(ctx context.Context, match *domain.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Match), args.Error(1)
}

func (m *MockMatchRepository) List(ctx context.Context, filters repository.MatchFilters) ([]*domain.Match, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Match), args.Error(1)
}
