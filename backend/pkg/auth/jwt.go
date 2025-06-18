package auth

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/parkertr/tipping/internal/constants"
	"github.com/parkertr/tipping/pkg/utils"
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
		secretKey: []byte(utils.GetEnvOrDefault("JWT_SECRET", "your-secret-key")),
	}
}

// GenerateToken generates a new JWT token for a user.
func (m *TokenManager) GenerateToken(userID string) (string, error) {
	expiration := time.Now().Add(
		utils.GetEnvOrDefaultDuration("JWT_EXPIRATION", constants.DefaultTokenExpiration),
	)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiration.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (m *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
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
func (m *TokenManager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with extended expiration
	return m.GenerateToken(claims.UserID)
}
