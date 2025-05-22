package middleware

import "github.com/labstack/echo/v4"

func UrlCache(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Path() == "/:shortId" {
			c.Response().Header().Set("Cache-Control", "public, max-age=3600")
		}
		return next(c)
	}
}
