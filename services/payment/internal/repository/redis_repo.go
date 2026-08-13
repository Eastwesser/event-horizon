package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepo(addr string, ttl time.Duration) *RedisRepo {
	return &RedisRepo{
		client: redis.NewClient(&redis.Options{Addr: addr, PoolSize: 10}),
		ttl:    ttl,
	}
}

func (r *RedisRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisRepo) Close() error {
	return r.client.Close()
}

func subKey(userID string) string {
	return fmt.Sprintf("payment:sub:%s", userID)
}

func (r *RedisRepo) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	data, err := r.client.Get(ctx, subKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var s model.Subscription
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *RedisRepo) SetSubscription(ctx context.Context, s *model.Subscription) error {
	if s == nil {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	ttl := r.ttl
	if rem := time.Until(s.ExpiresAt); rem > 0 && rem < ttl {
		ttl = rem
	}
	return r.client.Set(ctx, subKey(s.UserID), data, ttl).Err()
}

func (r *RedisRepo) DeleteSubscription(ctx context.Context, userID string) error {
	return r.client.Del(ctx, subKey(userID)).Err()
}
