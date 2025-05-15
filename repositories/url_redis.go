package repositories

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type UrlRedis struct {
	RedisClient *redis.Client
}

func NewUrlRedis(client *redis.Client) *UrlRedis {
	return &UrlRedis{
		RedisClient: client,
	}
}

func (u *UrlRedis) TestRedis() {
	ctx := context.Background()
	err := u.RedisClient.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := u.RedisClient.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("foo", val)
}
