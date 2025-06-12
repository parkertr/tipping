package domain

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	// Test case: Create new user
	id := "user123"
	username := "testuser"
	email := "test@example.com"

	user := NewUser(id, username, email)

	if user.ID != id {
		t.Errorf("expected ID %v, got %v", id, user.ID)
	}
	if user.Username != username {
		t.Errorf("expected Username %v, got %v", username, user.Username)
	}
	if user.Email != email {
		t.Errorf("expected Email %v, got %v", email, user.Email)
	}
	if user.JoinDate.IsZero() {
		t.Errorf("expected JoinDate to be set")
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
	user := NewUser("user123", "testuser", "test@example.com")

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
	user := NewUser("user123", "testuser", "test@example.com")

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
