package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type Cache interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.New(addr, password, db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	strategy := retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2}
	_, err := client.GetWithRetry(ctx, strategy, "ping")
	if err != nil && err != redis.NoMatches {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.SetWithExpiration(ctx, key, value, ttl)
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key)
	if err == redis.NoMatches {
		return "", nil
	}
	return val, err
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key)
}

func (c *RedisCache) Close() error {
	return nil
}

type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}
