package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Email    string
	JoinDate time.Time
	Stats    UserStats
}

// UserStats represents a user's statistics
type UserStats struct {
	TotalPoints        int
	CorrectPredictions int
	TotalPredictions   int
	CurrentRank        int
}

// NewUser creates a new user instance
func NewUser(id, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		JoinDate: time.Now(),
		Stats: UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}
}

// UpdateStats updates the user's statistics
func (u *User) UpdateStats(points int, isCorrect bool) {
	u.Stats.TotalPoints += points
	u.Stats.TotalPredictions++
	if isCorrect {
		u.Stats.CorrectPredictions++
	}
}

// GetSuccessRate returns the user's prediction success rate
func (u *User) GetSuccessRate() float64 {
	if u.Stats.TotalPredictions == 0 {
		return 0
	}
	return float64(u.Stats.CorrectPredictions) / float64(u.Stats.TotalPredictions) * 100
}
