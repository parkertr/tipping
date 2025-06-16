package domain

import (
	"time"

	"github.com/parkertr/tipping/internal/constants"
)

// Prediction represents a user's prediction for a match
type Prediction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	MatchID   string    `json:"matchId"`
	HomeGoals int       `json:"homeGoals"`
	AwayGoals int       `json:"awayGoals"`
	CreatedAt time.Time `json:"createdAt"`
	Points    int       `json:"points"`
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
		Points:    0,
	}
}

// CalculatePoints calculates the points awarded for this prediction
func (p *Prediction) CalculatePoints(match *Match) int {
	if match.Score == nil {
		return constants.NoPoints
	}

	// Exact score prediction
	if p.HomeGoals == match.Score.HomeGoals && p.AwayGoals == match.Score.AwayGoals {
		return constants.ExactScorePoints
	}

	// Correct result prediction
	predictionResult := p.GetResult()
	matchResult := match.GetResult()
	if predictionResult == matchResult {
		return constants.CorrectResultPoints
	}

	return constants.NoPoints
}

// GetResult returns the predicted result (home win, away win, or draw)
func (p *Prediction) GetResult() string {
	if p.HomeGoals > p.AwayGoals {
		return "home"
	}
	if p.AwayGoals > p.HomeGoals {
		return "away"
	}
	return "draw"
}
