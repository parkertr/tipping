package domain

import (
	"time"
)

// Prediction represents a user's prediction for a match
type Prediction struct {
	ID        string
	UserID    string
	MatchID   string
	HomeGoals int
	AwayGoals int
	CreatedAt time.Time
	Points    int
}

// NewPrediction creates a new prediction instance
func NewPrediction(id, userID, matchID string, homeGoals, awayGoals int) *Prediction {
	return &Prediction{
		ID:        id,
		UserID:    userID,
		MatchID:   matchID,
		HomeGoals: homeGoals,
		AwayGoals: awayGoals,
		CreatedAt: time.Now(),
	}
}

// CalculatePoints calculates the points earned for this prediction
func (p *Prediction) CalculatePoints(match *Match) int {
	if match.Score == nil {
		return 0
	}

	// Exact score prediction
	if p.HomeGoals == match.Score.HomeGoals && p.AwayGoals == match.Score.AwayGoals {
		return 3
	}

	// Correct result (win/draw/loss)
	predictionResult := getResult(p.HomeGoals, p.AwayGoals)
	actualResult := getResult(match.Score.HomeGoals, match.Score.AwayGoals)
	if predictionResult == actualResult {
		return 1
	}

	return 0
}

// getResult determines the result of a match based on goals
func getResult(homeGoals, awayGoals int) string {
	if homeGoals > awayGoals {
		return "HOME_WIN"
	}
	if awayGoals > homeGoals {
		return "AWAY_WIN"
	}
	return "DRAW"
}
