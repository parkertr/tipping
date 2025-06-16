// Package domain provides the core domain models and business logic for the tipping application.
// It includes types and methods for managing matches, predictions, and user data.
package domain

import (
	"time"
)

// Match represents a football match in the system.
type Match struct {
	ID          string      `json:"id"`
	HomeTeam    string      `json:"homeTeam"`
	AwayTeam    string      `json:"awayTeam"`
	Date        time.Time   `json:"date"`
	Competition string      `json:"competition"`
	Status      MatchStatus `json:"status"`
	Score       *Score      `json:"score"`
}

// Score represents the match score.
type Score struct {
	HomeGoals int `json:"homeGoals"`
	AwayGoals int `json:"awayGoals"`
}

// MatchStatus represents the current status of a match.
type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "SCHEDULED"
	MatchStatusLive      MatchStatus = "LIVE"
	MatchStatusFinished  MatchStatus = "FINISHED"
	MatchStatusCancelled MatchStatus = "CANCELLED"
)

// NewMatch creates a new match instance.
func NewMatch(id, homeTeam, awayTeam string, date time.Time, competition string) *Match {
	return &Match{
		ID:          id,
		HomeTeam:    homeTeam,
		AwayTeam:    awayTeam,
		Date:        date,
		Competition: competition,
		Status:      MatchStatusScheduled,
		Score:       &Score{HomeGoals: 0, AwayGoals: 0},
	}
}

// UpdateScore updates the match score.
func (match *Match) UpdateScore(homeGoals, awayGoals int) {
	match.Score = &Score{
		HomeGoals: homeGoals,
		AwayGoals: awayGoals,
	}
}

// IsFinished returns true if the match is finished.
func (match *Match) IsFinished() bool {
	return match.Status == MatchStatusFinished
}

// IsLive returns true if the match is currently live.
func (match *Match) IsLive() bool {
	return match.Status == MatchStatusLive
}

// GetResult returns the match result (home win, away win, or draw).
func (match *Match) GetResult() string {
	if match.Score == nil {
		return ""
	}

	if match.Score.HomeGoals > match.Score.AwayGoals {
		return "home"
	}

	if match.Score.AwayGoals > match.Score.HomeGoals {
		return "away"
	}

	return "draw"
}
