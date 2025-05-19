package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRepo interface {
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool)
	GetShortUrl(ctx context.Context, originalUrl string) (string, bool)
	SaveUrl(ctx context.Context, originalUrl string, shortUrl string, time time.Duration) error
	ExistsInRedis(ctx context.Context, originalUrl string, shortUrl string) (bool, error)
	DeleteKeys(ctx context.Context, originalUrl string, shortUrl string) error
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

func (u *UrlRedis) SaveUrl(ctx context.Context, originalUrl string, shortUrl string, ExpiryTime time.Duration) error {
	// Use a pipeline for atomic operations
	pipe := u.RedisClient.Pipeline()

	// Store shortID -> originalURL mapping (for redirects)
	pipe.Set(ctx, shortUrl, originalUrl, ExpiryTime)

	// Store originalURL -> shortID mapping (to prevent duplicates)
	pipe.Set(ctx, originalUrl, shortUrl, ExpiryTime)

	_, err := pipe.Exec(ctx)
	return err
}

func (u *UrlRedis) ExistsInRedis(ctx context.Context, originalUrl string, shortUrl string) (bool, error) {
	ShortExists, err := u.RedisClient.Exists(ctx, shortUrl).Result()
	OriginalExists, err := u.RedisClient.Exists(ctx, originalUrl).Result()

	if err != nil {
		return false, fmt.Errorf("error checking key: %v", err)
	}
	if ShortExists == 1 && OriginalExists == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("key not exists")
	}

}

func (u *UrlRedis) DeleteKeys(ctx context.Context, originalUrl string, shortUrl string) error {
	pipe := u.RedisClient.TxPipeline()

	del1 := pipe.Del(ctx, originalUrl)
	del2 := pipe.Del(ctx, shortUrl)

	_, err := pipe.Exec(ctx)
	if err != nil || del1.Err() != nil || del2.Err() != nil {
		return fmt.Errorf("failed to delete keys: exec=%v, original=%v, short=%v", err, del1.Err(), del2.Err())
	}

	return nil
}
