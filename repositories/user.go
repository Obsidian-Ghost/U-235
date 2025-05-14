package repositories

import (
	"U-235/models"
	"context"
	"database/sql"
)

type UserRepository interface {
	UserRegistrationService(Name, email, passwordHashed string, ctx context.Context) error
	UserLoginService(user models.UserLogin, ctx context.Context) error
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{
		db: db,
	}
}

func (u userRepo) UserRegistrationService(Name, email, passwordHashed string, ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (u userRepo) UserLoginService(user models.UserLogin, ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
