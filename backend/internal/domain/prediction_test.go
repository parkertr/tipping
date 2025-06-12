package domain

import (
	"testing"
	"time"
)

func TestNewPrediction(t *testing.T) {
	// Test case: Create new prediction
	id := "pred123"
	userID := "user123"
	matchID := "match123"
	homeGoals := 2
	awayGoals := 1

	prediction := NewPrediction(id, userID, matchID, homeGoals, awayGoals)

	if prediction.ID != id {
		t.Errorf("expected ID %v, got %v", id, prediction.ID)
	}
	if prediction.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, prediction.UserID)
	}
	if prediction.MatchID != matchID {
		t.Errorf("expected MatchID %v, got %v", matchID, prediction.MatchID)
	}
	if prediction.HomeGoals != homeGoals {
		t.Errorf("expected HomeGoals %v, got %v", homeGoals, prediction.HomeGoals)
	}
	if prediction.AwayGoals != awayGoals {
		t.Errorf("expected AwayGoals %v, got %v", awayGoals, prediction.AwayGoals)
	}
	if prediction.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
	if prediction.Points != 0 {
		t.Errorf("expected Points to be 0, got %v", prediction.Points)
	}
}

func TestCalculatePoints(t *testing.T) {
	prediction := NewPrediction("pred123", "user123", "match123", 2, 1)

	// Test case 1: Match has no score
	match := NewMatch("match123", "Team A", "Team B", time.Now(), "Premier League")
	points := prediction.CalculatePoints(match)
	if points != 0 {
		t.Errorf("expected points 0, got %v", points)
	}

	// Test case 2: Exact score prediction
	match.UpdateScore(2, 1)
	points = prediction.CalculatePoints(match)
	if points != 3 {
		t.Errorf("expected points 3, got %v", points)
	}

	// Test case 3: Correct result (home win) but wrong score
	prediction = NewPrediction("pred123", "user123", "match123", 3, 1)
	points = prediction.CalculatePoints(match)
	if points != 1 {
		t.Errorf("expected points 1, got %v", points)
	}

	// Test case 4: Wrong result
	prediction = NewPrediction("pred123", "user123", "match123", 1, 2)
	points = prediction.CalculatePoints(match)
	if points != 0 {
		t.Errorf("expected points 0, got %v", points)
	}
}

func TestGetResult(t *testing.T) {
	// Test case 1: Home win
	result := getResult(2, 1)
	if result != "HOME_WIN" {
		t.Errorf("expected result HOME_WIN, got %v", result)
	}

	// Test case 2: Away win
	result = getResult(1, 2)
	if result != "AWAY_WIN" {
		t.Errorf("expected result AWAY_WIN, got %v", result)
	}

	// Test case 3: Draw
	result = getResult(1, 1)
	if result != "DRAW" {
		t.Errorf("expected result DRAW, got %v", result)
	}
}
