package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/parkertr/tipping/internal/auth"
	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
)

// contextKey is the key used to store the user in the request context
type contextKey string

const (
	// userContextKey is the key used to store the user in the request context
	userContextKey contextKey = "user"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(tokenManager *auth.TokenManager, userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Check if the header has the Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
				return
			}

			// Validate the token
			claims, err := tokenManager.ValidateToken(parts[1])
			if err != nil {
				if err == auth.ErrExpiredToken {
					http.Error(w, "Token has expired", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get the user from the repository
			user, err := userRepo.GetByID(r.Context(), claims.UserID)
			if err != nil {
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Add the user to the request context
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(userContextKey).(*domain.User); ok {
		return user
	}
	return nil
}

// RequireAuth is a middleware that ensures the request has a valid authenticated user
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
