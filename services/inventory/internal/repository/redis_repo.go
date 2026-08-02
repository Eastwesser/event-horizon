package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
    "github.com/redis/go-redis/v9"
)

type RedisCacheRepo struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisCacheRepo(addr string, ttl time.Duration) *RedisCacheRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   0,
    })
    return &RedisCacheRepo{client: client, ttl: ttl}
}

func (r *RedisCacheRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
    key := fmt.Sprintf("inventory:item:%s", id)
    data, err := r.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }
    var item model.Item
    if err := json.Unmarshal(data, &item); err != nil {
        return nil, err
    }
    return &item, nil
}

func (r *RedisCacheRepo) SetItem(ctx context.Context, item *model.Item) error {
    key := fmt.Sprintf("inventory:item:%s", item.ID)
    data, err := json.Marshal(item)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisCacheRepo) DeleteItem(ctx context.Context, id string) error {
    key := fmt.Sprintf("inventory:item:%s", id)
    return r.client.Del(ctx, key).Err()
}

func (r *RedisCacheRepo) Close() error {
    return r.client.Close()
}
