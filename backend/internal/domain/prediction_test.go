package domain_test

import (
	"testing"
	"time"

	"github.com/parkertr/tipping/internal/domain"
)

func TestNewPrediction(t *testing.T) {
	t.Parallel()

	id := "pred1"
	userID := "user1"
	matchID := "match1"
	homeGoals := 2
	awayGoals := 1

	prediction := domain.NewPrediction(
		"pred1",
		"user1",
		"match1",
		2,
		1,
	)

	if prediction.ID != id {
		t.Errorf("Expected ID %s, got %s", id, prediction.ID)
	}

	if prediction.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, prediction.UserID)
	}

	if prediction.MatchID != matchID {
		t.Errorf("Expected MatchID %s, got %s", matchID, prediction.MatchID)
	}

	if prediction.HomeGoals != homeGoals {
		t.Errorf("Expected HomeGoals %d, got %d", homeGoals, prediction.HomeGoals)
	}

	if prediction.AwayGoals != awayGoals {
		t.Errorf("Expected AwayGoals %d, got %d", awayGoals, prediction.AwayGoals)
	}

	if prediction.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}

	if prediction.Points != 0 {
		t.Errorf("expected Points to be 0, got %v", prediction.Points)
	}
}

func TestCalculatePoints(t *testing.T) {
	t.Parallel()

	t.Run("exact score", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 2, 1)
		match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
		match.UpdateScore(2, 1)

		points := prediction.CalculatePoints(match)
		if points != 3 {
			t.Errorf("Expected 3 points for exact score, got %d", points)
		}
	})

	t.Run("correct result", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 3, 1)
		match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
		match.UpdateScore(2, 1)

		points := prediction.CalculatePoints(match)
		if points != 1 {
			t.Errorf("Expected 1 point for correct result, got %d", points)
		}
	})

	t.Run("wrong prediction", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 2, 1)
		match := domain.NewMatch("match1", "Arsenal", "Chelsea", time.Now(), "Premier League")
		match.UpdateScore(1, 2)

		points := prediction.CalculatePoints(match)
		if points != 0 {
			t.Errorf("Expected 0 points for wrong prediction, got %d", points)
		}
	})
}

func TestGetResult(t *testing.T) {
	t.Parallel()

	t.Run("home win", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 2, 1)

		result := prediction.GetResult()
		if result != "home" {
			t.Errorf("Expected result 'home', got '%s'", result)
		}
	})

	t.Run("away win", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 1, 2)

		result := prediction.GetResult()
		if result != "away" {
			t.Errorf("Expected result 'away', got '%s'", result)
		}
	})

	t.Run("draw", func(t *testing.T) {
		t.Parallel()

		prediction := domain.NewPrediction("pred1", "user1", "match1", 1, 1)

		result := prediction.GetResult()
		if result != "draw" {
			t.Errorf("Expected result 'draw', got '%s'", result)
		}
	})
}
