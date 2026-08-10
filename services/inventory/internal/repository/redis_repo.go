package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/redis/go-redis/v9"
)

// RedisCacheRepo — кеш для товаров и результатов поиска.
// Использует TTL для автоматической инвалидации.
type RedisCacheRepo struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacheRepo создает новый Redis кеш.
func NewRedisCacheRepo(addr string, ttl time.Duration) *RedisCacheRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		PoolSize: 10, // размер пула соединений
	})
	return &RedisCacheRepo{client: client, ttl: ttl}
}

// GetItem получает товар из кеша по ID.
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

// SetItem сохраняет товар в кеше.
func (r *RedisCacheRepo) SetItem(ctx context.Context, item *model.Item) error {
	key := fmt.Sprintf("inventory:item:%s", item.ID)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// DeleteItem удаляет товар из кеша (при обновлении или удалении).
func (r *RedisCacheRepo) DeleteItem(ctx context.Context, id string) error {
	key := fmt.Sprintf("inventory:item:%s", id)
	return r.client.Del(ctx, key).Err()
}

// GetSearchResult получает закешированный результат поиска.
func (r *RedisCacheRepo) GetSearchResult(ctx context.Context, query string, limit, offset int) ([]*model.Item, error) {
	key := fmt.Sprintf("inventory:search:%s:%d:%d", query, limit, offset)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var items []*model.Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// SetSearchResult сохраняет результат поиска в кеше.
func (r *RedisCacheRepo) SetSearchResult(ctx context.Context, query string, limit, offset int, items []*model.Item) error {
	key := fmt.Sprintf("inventory:search:%s:%d:%d", query, limit, offset)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// InvalidateSearchCache удаляет все закешированные результаты поиска (при изменении данных).
func (r *RedisCacheRepo) InvalidateSearchCache(ctx context.Context) error {
	// Используем паттерн для удаления всех ключей с префиксом "inventory:search:*"
	iter := r.client.Scan(ctx, 0, "inventory:search:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Close закрывает соединение с Redis.
func (r *RedisCacheRepo) Close() error {
	return r.client.Close()
}
