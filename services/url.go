package services

import (
	"U-235/models"
	"U-235/repositories"
	"context"
	"github.com/google/uuid"
)

type UrlServices interface {
	CreateUrlService(originalUrl string, customShortUrl string, expireTime int64, ctx context.Context) (*models.CreateShortUrlRes, error)
}

type ShortUrlService struct {
	RedisRepo repositories.RedisRepo
	// PsqlRepo here
}

func NewShortUrlService(repo repositories.RedisRepo) *ShortUrlService {
	return &ShortUrlService{
		RedisRepo: repo,
	}
}

func (r *ShortUrlService) CreateUrlService(userId uuid.UUID, originalUrl string, customShortUrl string, expireTime int64, ctx context.Context) (*models.CreateShortUrlRes, error) {

}
