package services

import (
	"U-235/core"
	"U-235/models"
	"U-235/repositories"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"
)

var Domain = os.Getenv("DOMAIN")

type UrlServices interface {
	CreateUrlService(userID uuid.UUID, req *models.CreateShortUrlReq, ctx context.Context) (*models.ShortenedUrlInfoRes, error)
}

type ShortUrlService struct {
	RedisRepo repositories.RedisRepo
	PsqlRepo  repositories.UrlsPsql
}

func NewShortUrlService(repo repositories.RedisRepo, psql repositories.UrlsPsql) *ShortUrlService {
	return &ShortUrlService{
		RedisRepo: repo,
		PsqlRepo:  psql,
	}
}

func (r *ShortUrlService) CreateUrlService(userID uuid.UUID, req *models.CreateShortUrlReq, ctx context.Context) (*models.ShortenedUrlInfoRes, error) {
	CustomUrlTag := req.CustomShortUrl

	existingShortUrl, found := r.RedisRepo.GetShortUrl(ctx, req.OriginalUrl)
	if found {
		//get data from psql using "userId AND originalUrl" , then return the existing Mapping
	}

	urlInfo := models.ShortenedUrlInfoReq{
		UserId:      userID,
		OriginalUrl: req.OriginalUrl,
	}

	if CustomUrlTag != "" && len(CustomUrlTag) < 5 {
		return nil, errors.New("invalid length of the custom url suffix")
	} else if CustomUrlTag == "" {
		shortID := core.GenerateShortID(req.OriginalUrl)
		urlInfo.ShortUrl = Domain + shortID
	} else {
		_, exists := r.RedisRepo.GetOriginalUrl(ctx, Domain+CustomUrlTag)
		if exists {
			return nil, errors.New("custom short url already exists")
		}
		urlInfo.ShortUrl = Domain + CustomUrlTag
	}

	urlInfo.ExpiresAt = time.Now().Add(time.Duration(req.ExpireTime) * time.Hour)
	urlInfo.IsActive = true

	// 1. Save to PostgreSQL first
	finalUrlRes, rollback, err := r.PsqlRepo.SaveUrl(ctx, &urlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to save url to DB: %w", err)
	}

	// 2. Save to Redis
	redisErr := r.RedisRepo.SaveUrl(ctx, req.OriginalUrl, urlInfo.ShortUrl, time.Duration(req.ExpireTime)*time.Hour)
	if redisErr != nil {
		// Rollback: delete from PostgreSQL
		_ = r.PsqlRepo.DeleteUrlRecord(ctx, rollback.UserId, rollback.UrlRecordId)
		return nil, fmt.Errorf("failed to save url to Redis, rolled back DB: %w", redisErr)
	}
	return finalUrlRes, nil
}
