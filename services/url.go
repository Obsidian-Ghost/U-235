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
	GetUserUrls(ctx context.Context, userID uuid.UUID, page, limit int, isActive *bool) (*models.PaginatedUrlsResponse, error)
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

func (r *ShortUrlService) GetUserUrls(ctx context.Context, userID uuid.UUID, page, limit int, isActive *bool) (*models.PaginatedUrlsResponse, error) {
	// Set default pagination values if not provided
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		// Cap the maximum limit to prevent excessive resource usage
		limit = 100
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get total count of URLs for this user with the applied filter
	totalCount, err := r.PsqlRepo.CountUserUrls(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count user URLs: %w", err)
	}

	// Get paginated URLs for this user
	urls, err := r.PsqlRepo.GetUserUrls(ctx, userID, offset, limit, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user URLs: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrevious := page > 1

	// Create the response
	response := &models.PaginatedUrlsResponse{
		Urls: urls,
		Meta: models.PaginationMeta{
			CurrentPage: page,
			TotalPages:  totalPages,
			PageSize:    limit,
			TotalCount:  totalCount,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}

	return response, nil
}
