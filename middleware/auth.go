package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWT Configuration
var (
	jwtSecret       []byte
	TokenExpiration = 72 * time.Hour
)

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable not set")
	}
	jwtSecret = []byte(secret)
}

// Custom errors
var (
	ErrInvalidToken         = errors.New("invalid authentication credentials")
	ErrExpiredToken         = errors.New("authentication credentials expired")
	ErrMissingToken         = errors.New("missing authentication credentials")
	ErrInvalidSigningMethod = errors.New("unexpected token signing method")
	ErrInvalidClaims        = errors.New("invalid token claims")
)

// TokenClaims represents the JWT claims structure
type TokenClaims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

// ValidateToken validates and returns claims from a token string
func ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || claims.UserID == uuid.Nil {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// CreateToken generates a new JWT token
func CreateToken(userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("cannot create token without valid user ID")
	}

	claims := &TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app-name",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthMiddleware validates JWT tokens in incoming requests
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		// Validate header format
		if authHeader == "" {
			c.Response().Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error":   ErrMissingToken.Error(),
				"code":    "missing_auth_header",
				"details": "Authorization header required with 'Bearer <token>' format",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error":   "Invalid authorization format",
				"code":    "invalid_auth_header",
				"details": "Expected format: 'Bearer <token>'",
			})
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenStr == "" {
			c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error":   "Empty bearer token",
				"code":    "empty_bearer_token",
				"details": "Token cannot be empty",
			})
		}

		// Validate token
		claims, err := ValidateToken(tokenStr)
		if err != nil {
			status := http.StatusUnauthorized
			errorCode := "invalid_token"
			wwwAuthHeader := `Bearer error="invalid_token"`

			if errors.Is(err, ErrExpiredToken) {
				errorCode = "expired_token"
				wwwAuthHeader = `Bearer error="invalid_token", error_description="Token expired"`
			}

			c.Response().Header().Set("WWW-Authenticate", wwwAuthHeader)
			return c.JSON(status, echo.Map{
				"error":   err.Error(),
				"code":    errorCode,
				"details": "Requires valid authentication credentials",
			})
		}

		// Set user context
		c.Set("userID", claims.UserID)

		return next(c)
	}
}
