package services

import (
	"U-235/middleware"
	"U-235/models"
	"U-235/repositories"
	"U-235/utils"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type UserServices interface {
	UserRegistrationService(user models.UserRegister, ctx context.Context) (*models.User, error)
	UserLoginService(login models.UserLogin, ctx context.Context) (*models.AuthResponse, error)
	UserProfileService(userId uuid.UUID, ctx context.Context) (*models.UserProfile, error)
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
	registeredUser, err := u.repo.UserRegistration(user.Name, user.Email, hashPassword, ctx)
	if err != nil {
		return nil, err
	}

	return registeredUser, nil
}

func (u *UserService) UserLoginService(user models.UserLogin, ctx context.Context) (*models.AuthResponse, error) {
	email := user.Email
	password := user.Password

	// Get user ID from repo
	userID, err := u.repo.UserLogin(email, password, ctx)
	if err != nil {
		return nil, err
	}

	// Get full user details
	userDetails, err := u.repo.GetUserByID(userID, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	// Generate token
	token, err := middleware.CreateToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &models.AuthResponse{
		User:  *userDetails,
		Token: token,
	}, nil
}

func (u *UserService) UserProfileService(userId uuid.UUID, ctx context.Context) (*models.UserProfile, error) {
	res, err := u.repo.UserProfileService(userId, ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}
