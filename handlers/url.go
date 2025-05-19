package handlers

import (
	"U-235/middleware"
	"U-235/models"
	"U-235/services"
	"U-235/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type UrlHandlers interface {
	CreateUrlHandler(c echo.Context) error
	GetUrlHandler(c echo.Context) error
	DeleteUrlHandler(c echo.Context) error
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
	claims, err := middleware.ValidateToken(token)
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

func (u *UrlHandler) GetUrlHandler(c echo.Context) error {
	// Get user ID from context (assuming it's set by authentication middleware)
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized access",
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Parse active status filter if provided
	var isActive *bool
	if activeStr := c.QueryParam("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			isActive = &active
		}
	}

	// Call the service to get URLs
	response, err := u.UrlService.GetUserUrls(c.Request().Context(), userID, page, limit, isActive)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve URLs: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (u *UrlHandler) DeleteUrlHandler(c echo.Context) error {
	var DelReq models.DeleteShortUrlReq
	if err := c.Bind(&DelReq); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&DelReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}

	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized access",
		})
	}
	DelReq.UserId = userID

	ctx := c.Request().Context()

	err := u.UrlService.DeleteUrlService(&DelReq, ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete URL: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "successfully deleted URL",
	})
}
