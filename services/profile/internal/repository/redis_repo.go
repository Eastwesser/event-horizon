package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProfileRepo implements cache-aside for aggregated user profiles.
// Profile previously had no Redis at all, so every GetProfile call hit
// Postgres directly — this was explicitly flagged as the project's biggest
// caching gap.
type RedisProfileRepo struct {
	client *redis.Client
}

func NewRedisProfileRepo(addr string) *RedisProfileRepo {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisProfileRepo{client: client}
}

func (r *RedisProfileRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisProfileRepo) Close() error {
	return r.client.Close()
}

func cacheKey(userID string) string {
	return fmt.Sprintf("profile:%s", userID)
}

func (r *RedisProfileRepo) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	val, err := r.client.Get(ctx, cacheKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var profile UserProfile
	if err := json.Unmarshal(val, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *RedisProfileRepo) SetProfile(ctx context.Context, profile *UserProfile, ttl time.Duration) error {
	payload, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, cacheKey(profile.UserID), payload, ttl).Err()
}

func (r *RedisProfileRepo) InvalidateProfile(ctx context.Context, userID string) error {
	return r.client.Del(ctx, cacheKey(userID)).Err()
}
