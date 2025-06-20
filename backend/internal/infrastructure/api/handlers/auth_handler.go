package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	userRepo     repository.UserRepository
	tokenManager *auth.TokenManager
	oauthConfig  *oauth2.Config
	eventStore   EventStore
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	userRepo repository.UserRepository,
	tokenManager *auth.TokenManager,
	eventStore EventStore,
) *AuthHandler {
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

// RegisterRoutes registers the auth handler routes.
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/google", h.GoogleLogin).Methods("GET")
	router.HandleFunc("/google/callback", h.GoogleCallback).Methods("GET")
	router.HandleFunc("/google/callback", h.GoogleCredentialCallback).Methods("POST") // New credential-based flow
	router.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	router.HandleFunc("/me", h.GetProfile).Methods("GET")
	router.HandleFunc("/me", h.UpdateProfile).Methods("PUT")
	router.HandleFunc("/me/deactivate", h.DeactivateProfile).Methods("POST")
	router.HandleFunc("/me/stats", h.GetUserStats).Methods("GET")
	router.HandleFunc("/me/ranking", h.GetUserRanking).Methods("GET")
	router.HandleFunc("/me/predictions", h.GetUserPredictions).Methods("GET")
	router.HandleFunc("/me/preferences", h.UpdateUserPreferences).Methods("PUT")
}

// GoogleLogin initiates the Google OAuth flow.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// getUserInfoFromGoogle retrieves user info from Google OAuth.
func (h *AuthHandler) getUserInfoFromGoogle(ctx context.Context, code string) (*struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}, error) {
	token, err := h.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := h.oauthConfig.Client(ctx, token)

	// Get user info from Google
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
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

// GoogleCredentialCallback handles the Google credential-based authentication (for frontend)
func (h *AuthHandler) GoogleCredentialCallback(writer http.ResponseWriter, request *http.Request) {
	var req struct {
		Credential string `json:"credential"`
	}

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Credential == "" {
		http.Error(writer, "credential is required", http.StatusBadRequest)
		return
	}

	// Parse the JWT credential from Google
	// Note: In production, you should verify the JWT signature
	// For now, we'll decode it to get user info
	userInfo, err := h.parseGoogleCredential(req.Credential)
	if err != nil {
		http.Error(writer, fmt.Sprintf("failed to parse credential: %v", err), http.StatusBadRequest)
		return
	}

	user, err := h.createOrUpdateUser(request.Context(), userInfo)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	tokenString, err := h.tokenManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(writer, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return both user and token
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"user":  user,
		"token": tokenString,
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseGoogleCredential parses the Google ID token credential
func (h *AuthHandler) parseGoogleCredential(credential string) (*struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}, error) {
	// Note: This is a simplified implementation
	// In production, you should verify the JWT signature using Google's public keys

	// For now, we'll decode the JWT payload (base64 encoded)
	parts := strings.Split(credential, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	return &struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verifiedEmail"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}{
		ID:            claims.Sub,
		Email:         claims.Email,
		VerifiedEmail: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

// createOrUpdateUser creates a new user or updates an existing one.
func (h *AuthHandler) createOrUpdateUser(ctx context.Context, userInfo *struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}) (*domain.User, error) {
	user, err := h.userRepo.GetByGoogleID(ctx, userInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user
		user = &domain.User{
			GoogleID: userInfo.ID,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			Picture:  userInfo.Picture,
		}

		if err := h.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		event := events.NewEvent("UserRegistered", events.UserRegistered{
			ID:        user.ID,
			GoogleID:  userInfo.ID,
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			Picture:   userInfo.Picture,
			CreatedAt: time.Now(),
		})

		if err := h.eventStore.SaveEvent(ctx, event); err != nil {
			return nil, fmt.Errorf("failed to save user registration event: %w", err)
		}
	} else {
		user.UpdateProfile(userInfo.Name, userInfo.Picture)

		event := events.NewEvent("UserProfileUpdated", events.UserProfileUpdated{
			UserID:    user.ID,
			Name:      userInfo.Name,
			Picture:   userInfo.Picture,
			UpdatedAt: time.Now(),
		})

		if err := h.eventStore.SaveEvent(ctx, event); err != nil {
			return nil, fmt.Errorf("failed to save profile update event: %w", err)
		}
	}

	return user, nil
}

// GoogleCallback handles the Google OAuth callback.
func (h *AuthHandler) GoogleCallback(writer http.ResponseWriter, request *http.Request) {
	code := request.URL.Query().Get("code")
	if code == "" {
		http.Error(writer, "code not found", http.StatusBadRequest)

		return
	}

	userInfo, err := h.getUserInfoFromGoogle(request.Context(), code)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	user, err := h.createOrUpdateUser(request.Context(), userInfo)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	// Generate JWT token
	tokenString, err := h.tokenManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(writer, "failed to generate token", http.StatusInternalServerError)

		return
	}

	// Return token
	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(map[string]string{
		"token": tokenString,
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// RefreshToken refreshes a JWT token.
func (h *AuthHandler) RefreshToken(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)

		return
	}

	tokenString, err := h.tokenManager.RefreshToken(req.Token)
	if err != nil {
		http.Error(writer, "failed to refresh token", http.StatusUnauthorized)

		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(map[string]string{
		"token": tokenString,
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// GetProfile returns the current user's profile.
func (h *AuthHandler) GetProfile(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(user); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// UpdateProfile updates the current user's profile.
func (h *AuthHandler) UpdateProfile(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	var update struct {
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(request.Body).Decode(&update); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)

		return
	}

	user.UpdateProfile(update.Name, update.Picture)

	event := events.NewEvent("UserProfileUpdated", events.UserProfileUpdated{
		UserID:    user.ID,
		Name:      user.Name,
		Picture:   user.Picture,
		UpdatedAt: user.UpdatedAt,
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		http.Error(writer, "failed to save profile update event", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(user); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// DeactivateProfile deactivates the current user's profile.
func (h *AuthHandler) DeactivateProfile(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	user.Deactivate()

	event := events.NewEvent("UserDeactivated", events.UserDeactivated{
		UserID:    user.ID,
		UpdatedAt: user.UpdatedAt,
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		http.Error(writer, "failed to save deactivation event", http.StatusInternalServerError)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// GetUserStats returns the current user's statistics.
func (h *AuthHandler) GetUserStats(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"totalPoints":        user.Stats.TotalPoints,
		"correctPredictions": user.Stats.CorrectPredictions,
		"totalPredictions":   user.Stats.TotalPredictions,
		"currentRank":        user.Stats.CurrentRank,
		"successRate":        user.GetSuccessRate(),
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// GetUserRanking returns the current user's ranking.
func (h *AuthHandler) GetUserRanking(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)

		return
	}

	// Get all active users sorted by points
	users, err := h.userRepo.List(request.Context(), true)
	if err != nil {
		http.Error(writer, "failed to get user rankings", http.StatusInternalServerError)

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

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"position":    position,
		"totalUsers":  len(users),
		"currentRank": user.Stats.CurrentRank,
		"totalPoints": user.Stats.TotalPoints,
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)

		return
	}
}

// GetUserPredictions returns the current user's prediction history.
func (h *AuthHandler) GetUserPredictions(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Get all prediction events for this user
	events, err := h.eventStore.GetEventsByType(request.Context(), "PredictionMade")
	if err != nil {
		http.Error(writer, "failed to retrieve predictions", http.StatusInternalServerError)
		return
	}

	var predictions []map[string]interface{}
	for _, event := range events {
		// Unmarshal the event data
		data, err := json.Marshal(event.Data)
		if err != nil {
			continue
		}

		var predictionData struct {
			ID        string    `json:"id"`
			UserID    string    `json:"userId"`
			MatchID   string    `json:"matchId"`
			HomeGoals int       `json:"homeGoals"`
			AwayGoals int       `json:"awayGoals"`
			CreatedAt time.Time `json:"createdAt"`
		}

		if err := json.Unmarshal(data, &predictionData); err != nil {
			continue
		}

		// Only include predictions for this user
		if predictionData.UserID == user.ID {
			predictions = append(predictions, map[string]interface{}{
				"id":        predictionData.ID,
				"matchId":   predictionData.MatchID,
				"homeGoals": predictionData.HomeGoals,
				"awayGoals": predictionData.AwayGoals,
				"createdAt": predictionData.CreatedAt,
				"points":    0, // TODO: Calculate points based on match results
			})
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"predictions": predictions,
		"total":       len(predictions),
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateUserPreferences updates the current user's preferences/settings.
func (h *AuthHandler) UpdateUserPreferences(writer http.ResponseWriter, request *http.Request) {
	user, ok := request.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(writer, "user not found in context", http.StatusUnauthorized)
		return
	}

	var preferences struct {
		EmailNotifications bool   `json:"emailNotifications"`
		Timezone           string `json:"timezone"`
		Language           string `json:"language"`
	}

	if err := json.NewDecoder(request.Body).Decode(&preferences); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	// Update user preferences (extend User domain model to include preferences)
	// For now, we'll store this as a simple event
	event := events.NewEvent("UserPreferencesUpdated", events.UserPreferencesUpdated{
		UserID:             user.ID,
		EmailNotifications: preferences.EmailNotifications,
		Timezone:           preferences.Timezone,
		Language:           preferences.Language,
		UpdatedAt:          time.Now(),
	})

	if err := h.eventStore.SaveEvent(request.Context(), event); err != nil {
		http.Error(writer, "failed to save preferences", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"message":            "preferences updated successfully",
		"emailNotifications": preferences.EmailNotifications,
		"timezone":           preferences.Timezone,
		"language":           preferences.Language,
	}); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	GivenName     string `json:"givenName"`
	FamilyName    string `json:"familyName"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GoogleTokenResponse struct {
	AccessToken string `json:"accessToken"`
	IDToken     string `json:"idToken"`
	ExpiresIn   int    `json:"expiresIn"`
	TokenType   string `json:"tokenType"`
	Scope       string `json:"scope"`
}

type GoogleUserInfoResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	Name          string `json:"name"`
	GivenName     string `json:"givenName"`
	FamilyName    string `json:"familyName"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verifiedEmail"`
}
