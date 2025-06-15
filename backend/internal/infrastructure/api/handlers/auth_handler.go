package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/parkertr/tipping/internal/auth"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
	"github.com/parkertr/tipping/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	config       *oauth2.Config
	tokenManager *auth.TokenManager
	userRepo     repository.UserRepository
	eventStore   EventStore
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(tokenManager *auth.TokenManager, userRepo repository.UserRepository, eventStore EventStore) *AuthHandler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		config:       config,
		tokenManager: tokenManager,
		userRepo:     userRepo,
		eventStore:   eventStore,
	}
}

// GoogleLogin initiates the Google OAuth login flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	userInfo, err := h.getGoogleUserInfo(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.getOrCreateUser(r.Context(), userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.tokenManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getGoogleUserInfo retrieves user information from Google
func (h *AuthHandler) getGoogleUserInfo(ctx context.Context, code string) (*struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}, error) {
	token, err := h.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := h.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verifiedEmail"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// getOrCreateUser retrieves an existing user or creates a new one
func (h *AuthHandler) getOrCreateUser(ctx context.Context, userInfo *struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}) (*domain.User, error) {
	existingUser, err := h.userRepo.GetByGoogleID(ctx, userInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	var user *domain.User
	if existingUser == nil {
		user = domain.NewUser(userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
		user.ID = utils.GenerateID()

		event := events.NewEvent("UserRegistered", events.UserRegistered{
			ID:        user.ID,
			GoogleID:  user.GoogleID,
			Email:     user.Email,
			Name:      user.Name,
			Picture:   user.Picture,
			CreatedAt: user.CreatedAt,
		})

		if err := h.eventStore.SaveEvent(ctx, event); err != nil {
			return nil, fmt.Errorf("failed to save user registration event: %w", err)
		}
	} else {
		user = existingUser
		user.UpdateProfile(userInfo.Name, userInfo.Picture)

		event := events.NewEvent("UserProfileUpdated", events.UserProfileUpdated{
			UserID:    user.ID,
			Name:      user.Name,
			Picture:   user.Picture,
			UpdatedAt: user.UpdatedAt,
		})

		if err := h.eventStore.SaveEvent(ctx, event); err != nil {
			return nil, fmt.Errorf("failed to save profile update event: %w", err)
		}
	}

	return user, nil
}

// RefreshToken generates a new JWT token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	token, err := h.tokenManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*domain.User)
	if user == nil {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*domain.User)
	if user == nil {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var update struct {
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.UpdateProfile(update.Name, update.Picture)

	event := events.NewEvent("UserProfileUpdated", events.UserProfileUpdated{
		UserID:    user.ID,
		Name:      user.Name,
		Picture:   user.Picture,
		UpdatedAt: user.UpdatedAt,
	})

	if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
		http.Error(w, "Failed to save profile update event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeactivateProfile deactivates the current user's account
func (h *AuthHandler) DeactivateProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*domain.User)
	if user == nil {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	user.Deactivate()

	event := events.NewEvent("UserDeactivated", events.UserDeactivated{
		UserID:    user.ID,
		UpdatedAt: user.UpdatedAt,
	})

	if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
		http.Error(w, "Failed to save deactivation event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUserStats returns the current user's statistics
func (h *AuthHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*domain.User)
	if user == nil {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"totalPoints":        user.Stats.TotalPoints,
		"correctPredictions": user.Stats.CorrectPredictions,
		"totalPredictions":   user.Stats.TotalPredictions,
		"currentRank":        user.Stats.CurrentRank,
		"successRate":        user.GetSuccessRate(),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetUserRanking returns the user's ranking and leaderboard position
func (h *AuthHandler) GetUserRanking(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*domain.User)
	if user == nil {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	// Get all active users sorted by points
	users, err := h.userRepo.List(r.Context(), true)
	if err != nil {
		http.Error(w, "Failed to get user rankings", http.StatusInternalServerError)
		return
	}

	// Find user's position in the ranking
	var position int
	for i, u := range users {
		if u.ID == user.ID {
			position = i + 1
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"position":    position,
		"totalUsers":  len(users),
		"currentRank": user.Stats.CurrentRank,
		"totalPoints": user.Stats.TotalPoints,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
