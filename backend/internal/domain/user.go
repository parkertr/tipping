package domain

import (
	"fmt"
	"time"

	"github.com/parkertr/tipping/internal/constants"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	GoogleID  string    `json:"googleId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsActive  bool      `json:"isActive"`
	Stats     UserStats `json:"stats"`
}

// UserStats represents a user's statistics
type UserStats struct {
	TotalPoints        int `json:"totalPoints"`
	CorrectPredictions int `json:"correctPredictions"`
	TotalPredictions   int `json:"totalPredictions"`
	CurrentRank        int `json:"currentRank"`
}

// NewUser creates a new user instance
func NewUser(googleID, email, name, picture string) *User {
	now := time.Now()

	return &User{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		Picture:   picture,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
		Stats: UserStats{
			TotalPoints:        constants.InitialStats,
			CorrectPredictions: constants.InitialStats,
			TotalPredictions:   constants.InitialStats,
			CurrentRank:        constants.InitialStats,
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
