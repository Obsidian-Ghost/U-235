package database

import (
	"errors"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
)

const (
	envRedisAddress  = "REDIS_DB_ADDR"
	envRedisPassword = "REDIS_DB_PASS"
	envRedisDB       = "REDIS_DB_NUMBER"
	envRedisProtocol = "REDIS_DB_PROTOCOL"
)

// NewRedisDatabase creates a new Redis configuration from environment variables.
// Returns initialized RedisConfig or error if any environment variable is invalid.
func NewRedisDatabase() (*redis.Client, error) {
	// Validate and parse database number
	dbNumberStr := os.Getenv(envRedisDB)
	dbNumber, err := strconv.Atoi(dbNumberStr)
	if err != nil {
		return nil, fmt.Errorf("invalid redis database number: %w", err)
	}

	// Validate and parse protocol version
	protocolStr := os.Getenv(envRedisProtocol)
	protocolVersion, err := strconv.Atoi(protocolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid redis protocol version: %w", err)
	}

	// Get connection parameters
	address := os.Getenv(envRedisAddress)
	if address == "" {
		return nil, errors.New("redis address must be specified")
	}

	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: envRedisPassword, // No password set
		DB:       dbNumber,         // Use default DB
		Protocol: protocolVersion,  // Connection protocol
	}), nil
}
