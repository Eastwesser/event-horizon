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

func userSessionSetKey(userID string) string {
	return fmt.Sprintf("auth:user_session:%s", userID)
}

// CreateSession регистрирует выданный JWT (по jti) как активный, TTL = сроку жизни токена.
// Also tracks jti in auth:user_session:{userID} so UpdateRole can revoke all access tokens.
func (r *RedisAuthRepo) CreateSession(ctx context.Context, jti, userID string, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, sessionKey(jti), userID, ttl)
	if userID != "" {
		pipe.SAdd(ctx, userSessionSetKey(userID), jti)
		pipe.Expire(ctx, userSessionSetKey(userID), ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// SessionExists проверяет, не отозвана ли сессия (logout/бан удаляют ключ раньше TTL).
func (r *RedisAuthRepo) SessionExists(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, sessionKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// DeleteSession отзывает одну access-сессию (logout).
func (r *RedisAuthRepo) DeleteSession(ctx context.Context, jti string) error {
	return r.client.Del(ctx, sessionKey(jti)).Err()
}

// DeleteSessionForUser removes one access jti and drops it from the user's session set.
func (r *RedisAuthRepo) DeleteSessionForUser(ctx context.Context, jti, userID string) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, sessionKey(jti))
	if userID != "" {
		pipe.SRem(ctx, userSessionSetKey(userID), jti)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteAllSessionsForUser revokes every access session for the user (role change).
// Refresh tokens are intentionally left alive so FE can 401 → refresh → new role.
func (r *RedisAuthRepo) DeleteAllSessionsForUser(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	jtis, err := r.client.SMembers(ctx, userSessionSetKey(userID)).Result()
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, sessionKey(jti))
	}
	pipe.Del(ctx, userSessionSetKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}
