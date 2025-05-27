package services

import (
	"U-235/core"
	"U-235/models"
	"U-235/repositories"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

type UrlServices interface {
	CreateUrlService(userID uuid.UUID, req *models.CreateShortUrlReq, ctx context.Context) (*models.ShortenedUrlInfoRes, error)
	GetUserUrls(ctx context.Context, userID uuid.UUID, page, limit int, isActive *bool) (*models.PaginatedUrlsResponse, error)
	SoftDeleteUrlService(DelReq *models.DeleteShortUrlReq, ctx context.Context) error
	ExtendExpiryService(userId uuid.UUID, Req *models.ExtendExpiry, ctx context.Context) error
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, error)
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
	//var Domain = os.Getenv("DOMAIN")
	CustomUrlTag := req.CustomShortUrl
	urlInfo := models.ShortenedUrlInfoReq{
		UserId:      userID,
		OriginalUrl: req.OriginalUrl,
	}

	if CustomUrlTag != "" && len(CustomUrlTag) < 5 {
		return nil, errors.New("invalid length of the custom url suffix")
	} else if CustomUrlTag == "" {
		shortID := core.GenerateShortID(req.OriginalUrl)
		urlInfo.ShortUrl = shortID
	} else {
		_, exists := r.RedisRepo.GetOriginalUrl(ctx, CustomUrlTag)
		if exists {
			return nil, errors.New("custom short url already exists")
		}
		urlInfo.ShortUrl = CustomUrlTag
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

func (r *ShortUrlService) SoftDeleteUrlService(DelReq *models.DeleteShortUrlReq, ctx context.Context) error {
	// First, check if the URL record exists at all (regardless of owner)
	exists, err := r.PsqlRepo.UrlRecordExists(ctx, DelReq.UrlRecordId)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if !exists {
		return models.NewAppError("URL_NOT_FOUND", "The requested URL does not exist", http.StatusNotFound)
	}

	// Now check if this URL belongs to the requesting user
	urlInfo, err := r.PsqlRepo.GetUrlInfoByUserIdAndUrlRecordId(ctx, DelReq.UserId, DelReq.UrlRecordId)
	if err != nil {
		// If no rows found, it means URL exists but belongs to another user
		if errors.Is(err, sql.ErrNoRows) {
			return models.NewAppError("ACCESS_DENIED", "You do not have permission to delete this URL", http.StatusForbidden)
		}
		return fmt.Errorf("database error: %w", err)
	}

	ok, err := r.RedisRepo.ExistsInRedis(ctx, urlInfo.ShortUrl)
	if err != nil {
		return models.NewAppError("ACTIVE_URL_NOT_FOUND", "The requested URL does not exist", http.StatusNotFound)
	}

	// if isActive, then it must be inside redis.
	if urlInfo.IsActive && ok {
		// Soft delete: deactivate URL and set expiration to current time
		err = r.PsqlRepo.SoftDeleteUrl(ctx, urlInfo.UserId, urlInfo.Id)
		if err != nil {
			return models.NewAppError("DEACTIVATION_FAILED", "Failed to deactivate the URL", http.StatusInternalServerError)
		}

		// delete url from redis
		err = r.RedisRepo.DeleteKeys(ctx, urlInfo.ShortUrl)
		if err != nil {
			// Attempt rollback - reactivate the URL
			rollbackErr := r.PsqlRepo.SetUrlState(ctx, urlInfo.UserId, urlInfo.Id, true)
			if rollbackErr != nil {
				log.Printf("Failed to rollback URL state: %v", rollbackErr)
			}
			return models.NewAppError("CACHE_DELETE_FAILED", "Failed to remove URL from cache", http.StatusInternalServerError)
		}
	} else if urlInfo.IsActive {
		// URL is active but not in Redis, just soft delete in database
		err = r.PsqlRepo.SoftDeleteUrl(ctx, urlInfo.UserId, urlInfo.Id)
		if err != nil {
			return models.NewAppError("SOFT_DELETE_FAILED", "Failed to soft delete URL", http.StatusInternalServerError)
		}
	}
	// If URL is already inactive, no action needed

	return nil
}

func (r *ShortUrlService) ExtendExpiryService(userId uuid.UUID, Req *models.ExtendExpiry, ctx context.Context) error {
	// First update the PostgreSQL database (this will also validate ownership)
	err := r.PsqlRepo.ExtendExpiry(ctx, userId, Req.UrlId, Req.Hours)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "URL not found or you don't have permission")
		}
		return err
	}

	// Now get the URL info to update Redis
	urlInfo, err := r.PsqlRepo.GetUrlInfoByUserIdAndUrlRecordId(ctx, userId, Req.UrlId)
	if err != nil {
		return err
	}

	if !urlInfo.IsActive {
		return errors.New("URL is Inactive or Deleted")
	}

	// Update Redis expiry
	duration := time.Duration(Req.Hours) * time.Hour
	err = r.RedisRepo.ExtendExpiry(ctx, urlInfo.OriginalUrl, urlInfo.ShortUrl, duration)
	if err != nil {
		return err
	}

	return nil
}

func (r *ShortUrlService) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	original, exists := r.RedisRepo.GetOriginalUrl(ctx, shortUrl)
	if !exists {
		return "", errors.New("URL not found in cache")
	}
	return original, nil
}
