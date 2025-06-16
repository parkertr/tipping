package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/auth"
	"github.com/parkertr/tipping/pkg/events"
	"github.com/parkertr/tipping/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo     repository.UserRepository
	tokenManager *auth.TokenManager
	oauthConfig  *oauth2.Config
	eventStore   EventStore
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepository, tokenManager *auth.TokenManager, eventStore EventStore) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		oauthConfig: &oauth2.Config{
			ClientID:     utils.GetEnvOrDefault("GOOGLE_CLIENT_ID", ""),
			ClientSecret: utils.GetEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  utils.GetEnvOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		eventStore: eventStore,
	}
}

// RegisterRoutes registers the auth handler routes
func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/auth/google", h.GoogleLogin).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")
	r.HandleFunc("/auth/refresh", h.RefreshToken).Methods("POST")
}

// GoogleLogin initiates the Google OAuth flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Get or create user
	user, err := h.userRepo.GetByGoogleID(r.Context(), userInfo.ID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Create new user
		user = &domain.User{
			GoogleID: userInfo.ID,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			Picture:  userInfo.Picture,
		}

		if err := h.userRepo.Create(r.Context(), user); err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		event := events.NewEvent("UserRegistered", events.UserRegistered{
			ID:        user.ID,
			GoogleID:  userInfo.ID,
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			Picture:   userInfo.Picture,
			CreatedAt: time.Now(),
		})

		if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
			http.Error(w, "Failed to save user registration event", http.StatusInternalServerError)
			return
		}
	} else {
		user.UpdateProfile(userInfo.Name, userInfo.Picture)

		event := events.NewEvent("UserProfileUpdated", events.UserProfileUpdated{
			UserID:    user.ID,
			Name:      userInfo.Name,
			Picture:   userInfo.Picture,
			UpdatedAt: time.Now(),
		})

		if err := h.eventStore.SaveEvent(r.Context(), event); err != nil {
			http.Error(w, "Failed to save profile update event", http.StatusInternalServerError)
			return
		}
	}

	// Generate JWT token
	tokenString, err := h.tokenManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

// RefreshToken refreshes a JWT token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenString, err := h.tokenManager.RefreshToken(req.Token)
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
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
