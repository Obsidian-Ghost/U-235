package repositories

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRepo interface {
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool)
	GetShortUrl(ctx context.Context, originalUrl string) (string, bool)
	SaveUrl(ctx context.Context, originalUrl string, shortUrl string, time time.Duration) error
}

type UrlRedis struct {
	RedisClient *redis.Client
}

func NewUrlRedis(client *redis.Client) (RedisRepo, error) {

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.New("redis ping failed")
	}

	return &UrlRedis{
		RedisClient: client,
	}, nil
}

func (u *UrlRedis) GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool) {
	OriginalUrl, err := u.RedisClient.Get(ctx, shortUrl).Result()
	if err != nil {
		return "", false
	}
	if OriginalUrl == "" {
		return "", false
	}
	return OriginalUrl, true

}

func (u *UrlRedis) GetShortUrl(ctx context.Context, originalUrl string) (string, bool) {
	shortUrl, err := u.RedisClient.Get(ctx, originalUrl).Result()

	if err == redis.Nil {
		return "", false
	}

	if err != nil {
		return "", false
	}

	return shortUrl, true
}

func (u *UrlRedis) SaveUrl(ctx context.Context, originalUrl string, shortUrl string, time time.Duration) error {
	// Use a pipeline for atomic operations
	pipe := u.RedisClient.Pipeline()

	// Store shortID -> originalURL mapping (for redirects)
	pipe.Set(ctx, shortUrl, originalUrl, time)

	// Store originalURL -> shortID mapping (to prevent duplicates)
	pipe.Set(ctx, originalUrl, shortUrl, time)

	_, err := pipe.Exec(ctx)
	return err
}
