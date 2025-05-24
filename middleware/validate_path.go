package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
)

// ValidateShortId Middleware to validate short ID and reject known system paths
func ValidateShortId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		shortId := c.Param("shortId")

		// Reject known system paths that might conflict
		reservedPaths := []string{
			"api", "health", "favicon.ico", "robots.txt", "sitemap.xml",
			"admin", "dashboard", "login", "register", "static", "assets",
			"js", "css", "img", "images", "fonts", "docs", "help", "about",
			"contact", "privacy", "terms", "www", "ftp", "mail", "blog",
		}

		lowerShortId := strings.ToLower(shortId)
		for _, reserved := range reservedPaths {
			if lowerShortId == reserved {
				return echo.NewHTTPError(http.StatusNotFound, "Page not found")
			}
		}

		// Validate short ID format (adjust based on your short ID generation)
		// Example: 6-10 alphanumeric characters
		if len(shortId) < 4 || len(shortId) > 12 {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid short URL")
		}

		// Only allow alphanumeric characters (adjust based on your encoding)
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, shortId); !matched {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid short URL format")
		}

		return next(c)
	}
}
