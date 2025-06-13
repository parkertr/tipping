package domain

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	// Test case: Create new user
	googleID := "google-123"
	email := "test@example.com"
	name := "Test User"
	picture := "https://example.com/pic.jpg"

	user := NewUser(googleID, email, name, picture)

	if user.GoogleID != googleID {
		t.Errorf("expected GoogleID %v, got %v", googleID, user.GoogleID)
	}
	if user.Email != email {
		t.Errorf("expected Email %v, got %v", email, user.Email)
	}
	if user.Name != name {
		t.Errorf("expected Name %v, got %v", name, user.Name)
	}
	if user.Picture != picture {
		t.Errorf("expected Picture %v, got %v", picture, user.Picture)
	}
	if user.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
	if user.Stats.TotalPoints != 0 {
		t.Errorf("expected TotalPoints 0, got %v", user.Stats.TotalPoints)
	}
	if user.Stats.CorrectPredictions != 0 {
		t.Errorf("expected CorrectPredictions 0, got %v", user.Stats.CorrectPredictions)
	}
	if user.Stats.TotalPredictions != 0 {
		t.Errorf("expected TotalPredictions 0, got %v", user.Stats.TotalPredictions)
	}
	if user.Stats.CurrentRank != 0 {
		t.Errorf("expected CurrentRank 0, got %v", user.Stats.CurrentRank)
	}
}

func TestUpdateStats(t *testing.T) {
	user := NewUser("google-123", "test@example.com", "Test User", "https://example.com/pic.jpg")

	// Test case 1: Correct prediction
	user.UpdateStats(3, true)
	if user.Stats.TotalPoints != 3 {
		t.Errorf("expected TotalPoints 3, got %v", user.Stats.TotalPoints)
	}
	if user.Stats.CorrectPredictions != 1 {
		t.Errorf("expected CorrectPredictions 1, got %v", user.Stats.CorrectPredictions)
	}
	if user.Stats.TotalPredictions != 1 {
		t.Errorf("expected TotalPredictions 1, got %v", user.Stats.TotalPredictions)
	}

	// Test case 2: Incorrect prediction
	user.UpdateStats(0, false)
	if user.Stats.TotalPoints != 3 {
		t.Errorf("expected TotalPoints 3, got %v", user.Stats.TotalPoints)
	}
	if user.Stats.CorrectPredictions != 1 {
		t.Errorf("expected CorrectPredictions 1, got %v", user.Stats.CorrectPredictions)
	}
	if user.Stats.TotalPredictions != 2 {
		t.Errorf("expected TotalPredictions 2, got %v", user.Stats.TotalPredictions)
	}
}

func TestGetSuccessRate(t *testing.T) {
	user := NewUser("google-123", "test@example.com", "Test User", "https://example.com/pic.jpg")

	// Test case 1: No predictions
	if user.GetSuccessRate() != float64(0) {
		t.Errorf("expected success rate 0, got %v", user.GetSuccessRate())
	}

	// Test case 2: One correct prediction
	user.UpdateStats(3, true)
	if user.GetSuccessRate() != float64(100) {
		t.Errorf("expected success rate 100, got %v", user.GetSuccessRate())
	}

	// Test case 3: One correct, one incorrect prediction
	user.UpdateStats(0, false)
	if user.GetSuccessRate() != float64(50) {
		t.Errorf("expected success rate 50, got %v", user.GetSuccessRate())
	}
}
