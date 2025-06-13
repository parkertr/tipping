package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string
	GoogleID  string
	Email     string
	Name      string
	Picture   string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
	Stats     UserStats
}

// UserStats represents a user's statistics
type UserStats struct {
	TotalPoints        int
	CorrectPredictions int
	TotalPredictions   int
	CurrentRank        int
}

// NewUser creates a new user instance
func NewUser(googleID, email, name, picture string) *User {
	now := time.Now()
	return &User{
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		Picture:   picture,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
		Stats: UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(name, picture string) {
	u.Name = name
	u.Picture = picture
	u.UpdatedAt = time.Now()
}

// Deactivate marks the user as inactive
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate marks the user as active
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
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
