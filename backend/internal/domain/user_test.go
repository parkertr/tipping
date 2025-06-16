package domain_test

import (
	"testing"
	"time"

	"github.com/parkertr/tipping/internal/domain"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	id := "user1"
	email := "test@example.com"
	name := "Test User"
	picture := "https://example.com/picture.jpg"

	user := &domain.User{
		ID:        id,
		GoogleID:  "google123",
		Email:     email,
		Name:      name,
		Picture:   picture,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}

	if user.ID != id {
		t.Errorf("Expected ID %s, got %s", id, user.ID)
	}

	if user.Email != email {
		t.Errorf("Expected Email %s, got %s", email, user.Email)
	}

	if user.Name != name {
		t.Errorf("Expected Name %s, got %s", name, user.Name)
	}

	if user.Picture != picture {
		t.Errorf("Expected Picture %s, got %s", picture, user.Picture)
	}

	if user.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}

	if user.IsActive != true {
		t.Errorf("expected IsActive to be true")
	}
}

func TestUpdateProfile(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:        "user1",
		GoogleID:  "google123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}
	newName := "Updated Name"
	newPicture := "https://example.com/new-picture.jpg"

	user.UpdateProfile(newName, newPicture)

	if user.Name != newName {
		t.Errorf("Expected Name %s, got %s", newName, user.Name)
	}

	if user.Picture != newPicture {
		t.Errorf("Expected Picture %s, got %s", newPicture, user.Picture)
	}

	if user.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}
}

func TestDeactivate(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:        "user1",
		GoogleID:  "google123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}
	user.Deactivate()

	if user.IsActive != false {
		t.Errorf("expected IsActive to be false")
	}

	if user.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}
}

func TestUserStats(t *testing.T) {
	t.Parallel()

	stats := domain.UserStats{
		TotalPoints:        9,
		CorrectPredictions: 5,
		TotalPredictions:   10,
		CurrentRank:        1,
	}

	if stats.TotalPoints != 9 {
		t.Errorf("Expected TotalPoints %d, got %d", 9, stats.TotalPoints)
	}

	if stats.CorrectPredictions != 5 {
		t.Errorf("Expected CorrectPredictions %d, got %d", 5, stats.CorrectPredictions)
	}

	if stats.TotalPredictions != 10 {
		t.Errorf("Expected TotalPredictions %d, got %d", 10, stats.TotalPredictions)
	}

	if stats.CurrentRank != 1 {
		t.Errorf("Expected CurrentRank %d, got %d", 1, stats.CurrentRank)
	}
}

func TestUpdateStats(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:        "user1",
		GoogleID:  "google123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}

	// Test case 1: Correct prediction
	user.UpdateStats(3, true)

	if user.Stats.TotalPoints != 3 {
		t.Errorf("Expected TotalPoints 3, got %d", user.Stats.TotalPoints)
	}

	if user.Stats.CorrectPredictions != 1 {
		t.Errorf("Expected CorrectPredictions 1, got %d", user.Stats.CorrectPredictions)
	}

	if user.Stats.TotalPredictions != 1 {
		t.Errorf("Expected TotalPredictions 1, got %d", user.Stats.TotalPredictions)
	}

	// Test case 2: Wrong prediction
	user.UpdateStats(0, false)

	if user.Stats.TotalPoints != 3 {
		t.Errorf("Expected TotalPoints 3, got %d", user.Stats.TotalPoints)
	}

	if user.Stats.CorrectPredictions != 1 {
		t.Errorf("Expected CorrectPredictions 1, got %d", user.Stats.CorrectPredictions)
	}

	if user.Stats.TotalPredictions != 2 {
		t.Errorf("Expected TotalPredictions 2, got %d", user.Stats.TotalPredictions)
	}
}

func TestGetSuccessRate(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:        "user1",
		GoogleID:  "google123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}

	// Test case 1: No predictions
	rate := user.GetSuccessRate()
	if rate != 0 {
		t.Errorf("Expected success rate 0, got %f", rate)
	}

	// Test case 2: 50% success rate
	user.UpdateStats(3, true)
	user.UpdateStats(0, false)

	rate = user.GetSuccessRate()
	if rate != 50 {
		t.Errorf("Expected success rate 50, got %f", rate)
	}

	// Test case 3: 100% success rate
	user = &domain.User{
		ID:        "user1",
		GoogleID:  "google123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}
	user.UpdateStats(3, true)
	user.UpdateStats(1, true)

	rate = user.GetSuccessRate()
	if rate != 100 {
		t.Errorf("Expected success rate 100, got %f", rate)
	}
}
