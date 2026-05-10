package repository

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisBillingRepo struct {
    client *redis.Client
}

func NewRedisBillingRepo(addr string, db int) *RedisBillingRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisBillingRepo{client: client}
}

func (r *RedisBillingRepo) GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error) {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return -1, nil // not found in cache
    }
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(val)
}

func (r *RedisBillingRepo) SetBalance(ctx context.Context, userID string, currency CurrencyType, balance int, ttl time.Duration) error {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    return r.client.Set(ctx, key, balance, ttl).Err()
}

func (r *RedisBillingRepo) DeleteBalance(ctx context.Context, userID string, currency CurrencyType) error {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    return r.client.Del(ctx, key).Err()
}
