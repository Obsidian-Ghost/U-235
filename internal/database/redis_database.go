package database

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

// Env variable for Redis URL
const envRedisURL = "REDIS_URL"

// NewRedisDatabase creates a Redis client using a Redis URL (e.g., redis://user:password@host:port/db)
func NewRedisDatabase() (*redis.Client, error) {
	redisURL := os.Getenv(envRedisURL)
	if redisURL == "" {
		return nil, fmt.Errorf("environment variable %s is required", envRedisURL)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	return client, nil
}
