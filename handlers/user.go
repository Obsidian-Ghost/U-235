package handlers

import (
	"U-235/models"
	"U-235/services"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type UserHandlers interface {
	UserRegistrationHandler(c echo.Context) error
	UserLoginHandler(c echo.Context) error
	UserProfileHandler(c echo.Context) error
}

type userHandler struct {
	UserService services.UserServices
}

func NewUserHandler(services services.UserServices) UserHandlers {
	return &userHandler{
		UserService: services,
	}
}

func (u *userHandler) UserRegistrationHandler(c echo.Context) error {
	// Bind the request body to the UserRegister model
	var userRegister models.UserRegister
	if err := c.Bind(&userRegister); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate the user input
	if err := c.Validate(&userRegister); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}

	// Get the request context
	ctx := c.Request().Context()

	// Call the service layer to register the user
	registeredUser, err := u.UserService.UserRegistrationService(userRegister, ctx)
	if err != nil {
		// Check for specific errors to provide better responses
		if strings.Contains(err.Error(), "already exists") {
			return echo.NewHTTPError(http.StatusConflict, "User with this email already exists")
		}
		// Log the actual error for debugging
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	// Create response object - we don't want to return the UserRegister object
	// as it contains the password
	response := map[string]interface{}{
		"id":      registeredUser.Id,
		"name":    registeredUser.Name,
		"email":   registeredUser.Email,
		"message": "User registered successfully",
	}

	// Return successful response with status code 201 Created
	return c.JSON(http.StatusCreated, response)
}

func (u *userHandler) UserLoginHandler(c echo.Context) error {
	var user models.UserLogin
	err := c.Bind(&user)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	err = c.Validate(user)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
	}
	ctx := c.Request().Context()
	token, err := u.UserService.UserLoginService(user, ctx)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Return token in the response
	return c.JSON(http.StatusOK, map[string]string{
		"token":   token,
		"message": "User logged in successfully",
	})
}

func (u *userHandler) UserProfileHandler(c echo.Context) error {
	userId := c.Get("userID").(uuid.UUID)
	ctx := c.Request().Context()

	profile, err := u.UserService.UserProfileService(userId, ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, profile)
}
