package services

import (
	"U-235/repositories"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
)

type ExpirationService interface {
	InitializeKeyspaceNotifications(ctx context.Context) error
	StartExpirationListener(ctx context.Context)
	Stop()
}

type RedisExpirationService struct {
	redisClient *redis.Client
	psqlRepo    repositories.UrlsPsql
	subscriber  *redis.PubSub
	stopChan    chan struct{}
}

func NewRedisExpirationService(redisClient *redis.Client, psqlRepo repositories.UrlsPsql) ExpirationService {
	return &RedisExpirationService{
		redisClient: redisClient,
		psqlRepo:    psqlRepo,
		stopChan:    make(chan struct{}),
	}
}

func (s *RedisExpirationService) InitializeKeyspaceNotifications(ctx context.Context) error {
	// Enable keyspace notifications for expired events
	_, err := s.redisClient.ConfigSet(ctx, "notify-keyspace-events", "Ex").Result()
	if err != nil {
		return fmt.Errorf("failed to enable keyspace notifications: %w", err)
	}

	s.subscriber = s.redisClient.PSubscribe(ctx, "__keyevent@0__:expired")

	return nil
}

func (s *RedisExpirationService) StartExpirationListener(ctx context.Context) {
	if s.subscriber == nil {
		log.Printf("Warning: subscriber not initialized, call InitializeKeyspaceNotifications first")
		return
	}

	go func() {
		defer s.subscriber.Close()

		ch := s.subscriber.Channel()
		log.Println("Redis expiration listener started")

		for {
			select {
			case msg := <-ch:
				if msg != nil {
					shortUrl := msg.Payload
					log.Printf("Redis TTL expired for key: %s", shortUrl)

					if err := s.handleExpiredUrl(ctx, shortUrl); err != nil {
						log.Printf("Failed to handle expired URL %s: %v", shortUrl, err)
					}
				}
			case <-s.stopChan:
				log.Println("Expiration listener stopped")
				return
			case <-ctx.Done():
				log.Println("Expiration listener stopped due to context cancellation")
				return
			}
		}
	}()
}

func (s *RedisExpirationService) Stop() {
	close(s.stopChan)
	if s.subscriber != nil {
		s.subscriber.Close()
	}
}

func (s *RedisExpirationService) handleExpiredUrl(ctx context.Context, shortUrl string) error {
	err := s.psqlRepo.MarkUrlAsExpired(ctx, shortUrl)
	if err != nil {
		return fmt.Errorf("failed to mark URL as expired in database: %w", err)
	}

	log.Printf("Successfully marked URL %s as expired in database", shortUrl)
	return nil
}
