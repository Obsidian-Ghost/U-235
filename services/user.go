package services

import (
	"U-235/models"
	"U-235/repositories"
	"U-235/utils"
	"context"
	"fmt"
)

type UserServices interface {
	UserRegistrationService(user models.UserRegister, ctx context.Context) (*models.User, error)
	UserLoginService(login models.UserLogin, ctx context.Context) (string, error)
}

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserServices {
	return &UserService{
		repo: repo,
	}
}

func (u *UserService) UserRegistrationService(user models.UserRegister, ctx context.Context) (*models.User, error) {
	// Hash the user's password
	hashPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Call the repository function with the hashed password
	registeredUser, err := u.repo.UserRegistrationService(user.Name, user.Email, hashPassword, ctx)
	if err != nil {
		return nil, err
	}

	return registeredUser, nil
}

func (u *UserService) UserLoginService(user models.UserLogin, ctx context.Context) (string, error) {
	email := user.Email
	password := user.Password
	token, err := u.repo.UserLoginService(email, password, ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}
