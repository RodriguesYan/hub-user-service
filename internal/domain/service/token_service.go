package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// ITokenService defines the interface for JWT token operations
type ITokenService interface {
	CreateToken(userID, email string) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
}

// TokenClaims represents the claims stored in JWT tokens
type TokenClaims struct {
	UserID string
	Email  string
	jwt.StandardClaims
}

// TokenService implements JWT token creation and validation
type TokenService struct {
	secretKey       string
	tokenExpiration time.Duration
}

// NewTokenService creates a new TokenService instance
func NewTokenService(secretKey string, tokenExpiration time.Duration) ITokenService {
	return &TokenService{
		secretKey:       secretKey,
		tokenExpiration: tokenExpiration,
	}
}

// CreateToken creates a new JWT token for a user
func (s *TokenService) CreateToken(userID, email string) (string, error) {
	if userID == "" {
		return "", errors.New("userID cannot be empty")
	}

	if email == "" {
		return "", errors.New("email cannot be empty")
	}

	claims := &TokenClaims{
		UserID: userID,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenExpiration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "hub-user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token string cannot be empty")
	}

	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
