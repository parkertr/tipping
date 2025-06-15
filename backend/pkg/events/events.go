package events

import (
	"time"
)

// Event represents a domain event
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Version   int         `json:"version"`
}

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

// MatchStatusChanged represents a match status change event
type MatchStatusChanged struct {
	MatchID   string    `json:"matchId"`
	Status    string    `json:"status"`
	ChangedAt time.Time `json:"changedAt"`
}

// PredictionMade represents a prediction creation event
type PredictionMade struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	MatchID   string    `json:"matchId"`
	HomeGoals int       `json:"homeGoals"`
	AwayGoals int       `json:"awayGoals"`
	CreatedAt time.Time `json:"createdAt"`
}

// PointsAwarded represents points being awarded for a prediction
type PointsAwarded struct {
	UserID    string    `json:"userId"`
	MatchID   string    `json:"matchId"`
	Points    int       `json:"points"`
	AwardedAt time.Time `json:"awardedAt"`
}

// UserRegistered represents a new user registration event
type UserRegistered struct {
	ID        string    `json:"id"`
	GoogleID  string    `json:"googleId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserProfileUpdated represents a user profile update event
type UserProfileUpdated struct {
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserDeactivated represents a user deactivation event
type UserDeactivated struct {
	UserID    string    `json:"userId"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewEvent creates a new event instance
func NewEvent(eventType string, data interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Version:   1,
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000000")
}
