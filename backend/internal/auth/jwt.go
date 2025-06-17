// Package auth provides JWT-based authentication functionality for the tipping application.
// It includes token generation, validation, and refresh operations using the JWT standard.
package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/parkertr/tipping/internal/constants"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT token operations.
type TokenManager struct {
	secretKey []byte
}

// NewTokenManager creates a new token manager.
func NewTokenManager() *TokenManager {
	return &TokenManager{
		secretKey: []byte(os.Getenv("JWT_SECRET")),
	}
}

// GenerateToken generates a new JWT token for a user.
func (manager *TokenManager) GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(constants.DefaultTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tipping-app",
			Subject:   userID,
			ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
			Audience:  []string{"tipping-app"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(manager.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (manager *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return manager.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken generates a new token with extended expiration.
func (manager *TokenManager) RefreshToken(tokenString string) (string, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with extended expiration
	return manager.GenerateToken(claims.UserID)
}
