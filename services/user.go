package services

import (
	"U-235/models"
	"context"
	"github.com/labstack/echo/v4"
)

func NewUserService() UserServices {

}

type UserServices interface {
	UserRegistrationService(c echo.Context, user models.UserRegister, ctx context.Context) error
	UserLoginService(c echo.Context, login models.UserLogin, ctx context.Context) error
}
