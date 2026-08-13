package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func Connect(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("REDIS_ADDR not set")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DialTimeout: 3 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() {
	if c != nil && c.rdb != nil {
		_ = c.rdb.Close()
	}
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", fmt.Errorf("redis not connected")
	}
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Keys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	if c == nil || c.rdb == nil {
		return nil, fmt.Errorf("redis not connected")
	}
	if pattern == "" {
		pattern = "*"
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var (
		cursor uint64
		out    []string
	)
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, limit).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		cursor = next
		if cursor == 0 || int64(len(out)) >= limit {
			break
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}
