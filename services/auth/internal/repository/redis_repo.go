package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss сигнализирует о промахе кеша (ключ не найден) — это не ошибка Redis,
// вызывающая сторона должна сходить в Postgres.
var ErrCacheMiss = errors.New("cache miss")

// RedisAuthRepo — Redis-хранилище для Auth Service.
// Два независимых назначения одного клиента:
//  1. Cache-Aside для user:{id} (TTL 5 минут) — снижает нагрузку на Postgres.
//  2. Хранилище активных сессий (auth:session:{jti}) — авторизационные ключи
//     ОБЯЗАНЫ жить в Redis (см. confluence/history/2026-08/13.08.2026/THE_VOICE_MESSAGE.md),
//     что даёт возможность отзыва токена (logout) без ожидания истечения exp.
type RedisAuthRepo struct {
	client  *redis.Client
	userTTL time.Duration
}

func NewRedisAuthRepo(addr string, userTTL time.Duration) *RedisAuthRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		PoolSize: 10,
	})
	return &RedisAuthRepo{client: client, userTTL: userTTL}
}

func (r *RedisAuthRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisAuthRepo) Close() error {
	return r.client.Close()
}

// --- Cache-Aside для User ---

func userCacheKey(id string) string {
	return fmt.Sprintf("auth:user:%s", id)
}

func (r *RedisAuthRepo) GetUserCache(ctx context.Context, id string) (*User, error) {
	data, err := r.client.Get(ctx, userCacheKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *RedisAuthRepo) SetUserCache(ctx context.Context, u *User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, userCacheKey(u.ID), data, r.userTTL).Err()
}

func (r *RedisAuthRepo) InvalidateUserCache(ctx context.Context, id string) error {
	return r.client.Del(ctx, userCacheKey(id)).Err()
}

// --- Хранилище сессий (авторизационные ключи) ---

func sessionKey(jti string) string {
	return fmt.Sprintf("auth:session:%s", jti)
}

// CreateSession регистрирует выданный JWT (по jti) как активный, TTL = сроку жизни токена.
func (r *RedisAuthRepo) CreateSession(ctx context.Context, jti, userID string, ttl time.Duration) error {
	return r.client.Set(ctx, sessionKey(jti), userID, ttl).Err()
}

// SessionExists проверяет, не отозвана ли сессия (logout/бан удаляют ключ раньше TTL).
func (r *RedisAuthRepo) SessionExists(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, sessionKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// DeleteSession отзывает сессию (logout).
func (r *RedisAuthRepo) DeleteSession(ctx context.Context, jti string) error {
	return r.client.Del(ctx, sessionKey(jti)).Err()
}
