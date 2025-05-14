package services

import (
	"U-235/models"
	"U-235/repositories"
	"context"
	"github.com/labstack/echo/v4"
)

func NewUserService(repo repositories.UserRepository) UserServices {

}

type UserServices interface {
	UserRegistrationService(c echo.Context, user models.UserRegister, ctx context.Context) error
	UserLoginService(c echo.Context, login models.UserLogin, ctx context.Context) error
}
