package utils

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func ExtractTokenFromHeader(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Token not provided")
	}

	return token, nil
}
