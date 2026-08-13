package repository

import (
	"context"
	"fmt"
	"time"
)

func refreshKey(jti string) string {
	return fmt.Sprintf("auth:refresh:%s", jti)
}

func userRefreshSetKey(userID string) string {
	return fmt.Sprintf("auth:user_refresh:%s", userID)
}

// SaveRefresh stores a refresh token jti in Redis (Week 6).
func (r *RedisAuthRepo) SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, refreshKey(jti), userID, ttl)
	pipe.SAdd(ctx, userRefreshSetKey(userID), jti)
	pipe.Expire(ctx, userRefreshSetKey(userID), ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisAuthRepo) RefreshExists(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, refreshKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisAuthRepo) DeleteRefresh(ctx context.Context, jti, userID string) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, refreshKey(jti))
	if userID != "" {
		pipe.SRem(ctx, userRefreshSetKey(userID), jti)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteAllRefreshForUser revokes every refresh token for the user (password change).
func (r *RedisAuthRepo) DeleteAllRefreshForUser(ctx context.Context, userID string) error {
	jtis, err := r.client.SMembers(ctx, userRefreshSetKey(userID)).Result()
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, refreshKey(jti))
	}
	pipe.Del(ctx, userRefreshSetKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}
