package repository

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisShopRepo struct {
    client *redis.Client
}

func NewRedisShopRepo(addr string, db int) *RedisShopRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisShopRepo{client: client}
}

func (r *RedisShopRepo) GetItems(ctx context.Context, key string) ([]Item, error) {
    data, err := r.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }
    var items []Item
    if err := json.Unmarshal(data, &items); err != nil {
        return nil, err
    }
    return items, nil
}

func (r *RedisShopRepo) SetItems(ctx context.Context, key string, items []Item, ttl time.Duration) error {
    data, err := json.Marshal(items)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisShopRepo) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}
