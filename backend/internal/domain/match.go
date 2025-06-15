package domain

import (
	"time"
)

// Match represents a football match in the system
type Match struct {
	ID          string      `json:"id"`
	HomeTeam    string      `json:"homeTeam"`
	AwayTeam    string      `json:"awayTeam"`
	Date        time.Time   `json:"date"`
	Competition string      `json:"competition"`
	Status      MatchStatus `json:"status"`
	Score       *Score      `json:"score"`
}

// Score represents the match score
type Score struct {
	HomeGoals int `json:"homeGoals"`
	AwayGoals int `json:"awayGoals"`
}

// MatchStatus represents the current status of a match
type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "SCHEDULED"
	MatchStatusLive      MatchStatus = "LIVE"
	MatchStatusFinished  MatchStatus = "FINISHED"
	MatchStatusCancelled MatchStatus = "CANCELLED"
)

// NewMatch creates a new match instance
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

// UpdateScore updates the match score
func (m *Match) UpdateScore(homeGoals, awayGoals int) {
	m.Score = &Score{
		HomeGoals: homeGoals,
		AwayGoals: awayGoals,
	}
}

// IsFinished returns true if the match is finished
func (m *Match) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

// IsLive returns true if the match is currently live
func (m *Match) IsLive() bool {
	return m.Status == MatchStatusLive
}

// GetResult returns the match result (home win, away win, or draw)
func (m *Match) GetResult() string {
	if m.Score == nil {
		return ""
	}
	if m.Score.HomeGoals > m.Score.AwayGoals {
		return "home"
	}
	if m.Score.AwayGoals > m.Score.HomeGoals {
		return "away"
	}
	return "draw"
}
