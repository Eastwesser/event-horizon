// services/gateway/internal/ratelimit/limiter.go
package ratelimit

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
    rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
    return &RateLimiter{rdb: rdb}
}

// Allow проверяет, не превышен ли лимит (sliding window алгоритм)
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now().UnixMilli()
    windowStart := now - window.Milliseconds()

    pipe := rl.rdb.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    pipe.ZCard(ctx, key)
    pipe.Expire(ctx, key, window)

    cmds, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := cmds[2].(*redis.IntCmd).Val()
    return count <= int64(limit), nil
}

// AllowSubmit для POST /api/game/submit (10 запросов в секунду)
func (rl *RateLimiter) AllowSubmit(userID string) bool {
    key := fmt.Sprintf("rl:submit:%s", userID)
    allowed, _ := rl.Allow(context.Background(), key, 10, time.Second)
    return allowed
}

// AllowLogin для POST /api/auth/login (5 запросов в секунду с одного IP)
func (rl *RateLimiter) AllowLogin(ip string) bool {
    key := fmt.Sprintf("rl:login:%s", ip)
    allowed, _ := rl.Allow(context.Background(), key, 5, time.Second)
    return allowed
}

// AllowWebSocket для WebSocket соединений (100 соединений в минуту с одного IP)
func (rl *RateLimiter) AllowWebSocket(ip string) bool {
    key := fmt.Sprintf("rl:ws:%s", ip)
    allowed, _ := rl.Allow(context.Background(), key, 100, time.Minute)
    return allowed
}
