package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb_technoschool/level_3/task_1/models"
)

type Cache interface {
	Set(ctx context.Context, notification *models.Notification) error
	Get(ctx context.Context, id string) (*models.Notification, error)
	Delete(ctx context.Context, id string) error
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
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

	return &RedisCache{
		client: client,
		ttl:    24 * time.Hour,
	}, nil
}

func (c *RedisCache) Set(ctx context.Context, notification *models.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	key := fmt.Sprintf("notification:%s", notification.ID)
	return c.client.SetWithExpiration(ctx, key, data, c.ttl)
}

func (c *RedisCache) Get(ctx context.Context, id string) (*models.Notification, error) {
	key := fmt.Sprintf("notification:%s", id)
	data, err := c.client.Get(ctx, key)
	if err == redis.NoMatches {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var notification models.Notification
	if err := json.Unmarshal([]byte(data), &notification); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	return &notification, nil
}

func (c *RedisCache) Delete(ctx context.Context, id string) error {
	key := fmt.Sprintf("notification:%s", id)
	return c.client.Del(ctx, key)
}

func (c *RedisCache) Close() error {
	return nil
}

type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Set(ctx context.Context, notification *models.Notification) error {
	return nil
}

func (c *NoOpCache) Get(ctx context.Context, id string) (*models.Notification, error) {
	return nil, nil
}

func (c *NoOpCache) Delete(ctx context.Context, id string) error {
	return nil
}
