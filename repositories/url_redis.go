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
	ExistsInRedis(ctx context.Context, shortUrl string) (bool, error)
	DeleteKeys(ctx context.Context, shortUrl string) error
	ExtendExpiry(ctx context.Context, originalUrl string, shortUrl string, duration time.Duration) error
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

	_, err := pipe.Exec(ctx)
	return err
}

func (u *UrlRedis) ExistsInRedis(ctx context.Context, shortUrl string) (bool, error) {
	ShortExists, err := u.RedisClient.Exists(ctx, shortUrl).Result()

	if err != nil {
		return false, fmt.Errorf("error checking key: %v", err)
	}
	if ShortExists == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("key not exists")
	}

}

func (u *UrlRedis) DeleteKeys(ctx context.Context, shortUrl string) error {
	pipe := u.RedisClient.TxPipeline()

	del := pipe.Del(ctx, shortUrl)

	_, err := pipe.Exec(ctx)

	if err != nil || del.Err() != nil {
		return fmt.Errorf("failed to delete keys: exec=%v, short=%v", err, del.Err())
	}

	return nil
}

func (u *UrlRedis) ExtendExpiry(ctx context.Context, originalUrl string, shortUrl string, duration time.Duration) error {
	// Handle shortUrl expiry
	ttl, err := u.RedisClient.TTL(ctx, shortUrl).Result()
	if err != nil {
		return err
	}

	if ttl < 0 {
		err = u.RedisClient.Expire(ctx, shortUrl, duration).Err()
	} else {
		newExpiry := ttl + duration
		err = u.RedisClient.Expire(ctx, shortUrl, newExpiry).Err()
	}

	if err != nil {
		return err
	}

	// Handle originalUrl expiry
	ttl2, err := u.RedisClient.TTL(ctx, originalUrl).Result()
	if err != nil {
		return err
	}

	if ttl2 < 0 {
		return u.RedisClient.Expire(ctx, originalUrl, duration).Err()
	} else {
		newExpiry2 := ttl2 + duration
		return u.RedisClient.Expire(ctx, originalUrl, newExpiry2).Err()
	}
}
