package events

import (
	"time"
)

// Event represents a domain event
type Event struct {
	ID        string
	Type      string
	Data      interface{}
	Timestamp time.Time
	Version   int
}

// MatchCreated represents a match creation event
type MatchCreated struct {
	ID          string
	HomeTeam    string
	AwayTeam    string
	Date        time.Time
	Competition string
}

// MatchScoreUpdated represents a match score update event
type MatchScoreUpdated struct {
	MatchID   string
	HomeGoals int
	AwayGoals int
	UpdatedAt time.Time
}

// MatchStatusChanged represents a match status change event
type MatchStatusChanged struct {
	MatchID   string
	Status    string
	ChangedAt time.Time
}

// PredictionMade represents a prediction creation event
type PredictionMade struct {
	ID        string
	UserID    string
	MatchID   string
	HomeGoals int
	AwayGoals int
	CreatedAt time.Time
}

// PointsAwarded represents points being awarded for a prediction
type PointsAwarded struct {
	UserID    string
	MatchID   string
	Points    int
	AwardedAt time.Time
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
