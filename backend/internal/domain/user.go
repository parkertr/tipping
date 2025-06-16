package domain

import (
	"strconv"
	"time"

	"github.com/parkertr/tipping/internal/constants"
)

// User represents a user in the system.
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

// UserStats represents a user's statistics.
type UserStats struct {
	TotalPoints        int `json:"totalPoints"`
	CorrectPredictions int `json:"correctPredictions"`
	TotalPredictions   int `json:"totalPredictions"`
	CurrentRank        int `json:"currentRank"`
}

// NewUser creates a new user instance.
func NewUser(googleID, email, name, picture string) *User {
	now := time.Now()

	return &User{
		ID:        strconv.FormatInt(now.UnixNano(), 10),
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

// UpdateProfile updates the user's profile information.
func (user *User) UpdateProfile(name, picture string) {
	user.Name = name
	user.Picture = picture
	user.UpdatedAt = time.Now()
}

// Deactivate marks the user as inactive.
func (user *User) Deactivate() {
	user.IsActive = false
	user.UpdatedAt = time.Now()
}

// Activate marks the user as active.
func (user *User) Activate() {
	user.IsActive = true
	user.UpdatedAt = time.Now()
}

// UpdateStats updates the user's statistics.
func (user *User) UpdateStats(points int, isCorrect bool) {
	user.Stats.TotalPoints += points
	user.Stats.TotalPredictions++

	if isCorrect {
		user.Stats.CorrectPredictions++
	}
}

// GetSuccessRate returns the user's prediction success rate.
func (user *User) GetSuccessRate() float64 {
	if user.Stats.TotalPredictions == 0 {
		return 0
	}

	return float64(user.Stats.CorrectPredictions) / float64(user.Stats.TotalPredictions) * 100
}
