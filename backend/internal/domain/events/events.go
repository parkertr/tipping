package events

import "time"

// MatchCreated represents a match creation event
type MatchCreated struct {
	ID          string    `json:"id"`
	HomeTeam    string    `json:"homeTeam"`
	AwayTeam    string    `json:"awayTeam"`
	Date        time.Time `json:"date"`
	Competition string    `json:"competition"`
}

// MatchScoreUpdated represents a match score update event
type MatchScoreUpdated struct {
	MatchID   string    `json:"matchId"`
	HomeGoals int       `json:"homeGoals"`
	AwayGoals int       `json:"awayGoals"`
	UpdatedAt time.Time `json:"updatedAt"`
}
