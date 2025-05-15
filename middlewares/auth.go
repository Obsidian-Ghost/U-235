package middlewares

import (
	"errors"
	"github.com/google/uuid"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret string = os.Getenv("JWT_SECRET")

var jwtSecret = []byte(secret) // Replace with your actual secret

// errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenClaims represents the claims in the JWT
type TokenClaims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

// IsValidToken validates the provided token string
func IsValidToken(tokenString string) (bool, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtSecret, nil
	})

	// Handle parsing errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return false, ErrExpiredToken
		}
		return false, ErrInvalidToken
	}

	// Verify token is valid
	if !token.Valid {
		return false, ErrInvalidToken
	}

	return true, nil
}

// GetClaims extracts claims from a token string
func GetClaims(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// CreateToken generates a new JWT token
func CreateToken(userID uuid.UUID) (string, error) {
	// Create token with claims
	claims := &TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
