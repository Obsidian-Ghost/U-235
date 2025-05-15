package handlers

import (
	"U-235/models"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UrlHandlers interface {
	CreateUrlHandler(c echo.Context) error
}

type UrlHandler struct {
}

func (u *UrlHandler) CreateUrlHandler(c echo.Context) error {
	var url models.CreateShortUrlReq
	if err := c.Bind(&url); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&url); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}

	ctx := c.Request().Context()

	shortUrl, err := // call service

	return c.JSON(http.StatusCreated, url)
}
