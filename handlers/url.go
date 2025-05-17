package handlers

import (
	"U-235/middlewares"
	"U-235/models"
	"U-235/services"
	"U-235/utils"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UrlHandlers interface {
	CreateUrlHandler(c echo.Context) error
}

type UrlHandler struct {
	UrlService services.UrlServices
}

func NewUrlHandler(UrlService services.UrlServices) UrlHandlers {
	return &UrlHandler{
		UrlService: UrlService,
	}
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

	//Extract Token
	token, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//Extract Claims from token for UserId
	claims, err := middlewares.GetClaims(token)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userId := claims.UserID

	// Validate user's UUID
	if !utils.IsValidUUID(userId.String()) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user id")
	}

	ctx := c.Request().Context()

	// Get the final response here (models/ShortenedUrlInfoRes)
	res, err := u.UrlService.CreateUrlService(userId, &url, ctx)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}
