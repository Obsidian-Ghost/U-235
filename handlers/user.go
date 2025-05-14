package handlers

import (
	"U-235/services"
	"github.com/labstack/echo/v4"
)

type UserHandlers interface {
	UserRegistrationHandler(c echo.Context) error
	UserLoginHandler(c echo.Context) error
}

type userHandler struct {
	userService services.UserServices
}

func NewUserHandler(services services.UserServices) UserHandlers {
	return &userHandler{
		userService: services,
	}
}

func (u *userHandler) UserRegistrationHandler(c echo.Context) error {
	// TODO: implement registration logic
	panic("implement me")
}

func (u *userHandler) UserLoginHandler(c echo.Context) error {
	// TODO: implement login logic
	panic("implement me")
}
