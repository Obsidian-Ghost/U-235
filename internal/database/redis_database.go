package database

import (
	"errors"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strconv"
)

const (
	envRedisAddress  = "REDIS_DB_ADDR"
	envRedisPassword = "REDIS_DB_PASS"
	envRedisDB       = "REDIS_DB_NUMBER"
	envRedisProtocol = "REDIS_DB_PROTOCOL"
)

// RedisConfig holds connection parameters for Redis database
type RedisConfig struct {
	Address  string
	Password string
	DB       int
	Protocol int
}

// NewRedisDatabase creates a new Redis configuration from environment variables.
// Returns initialized RedisConfig or error if any environment variable is invalid.
func NewRedisDatabase() (*RedisConfig, error) {
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

	return &RedisConfig{
		Address:  address,
		Password: os.Getenv(envRedisPassword),
		DB:       dbNumber,
		Protocol: protocolVersion,
	}, nil
}
