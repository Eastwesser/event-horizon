package redisclient

import (
	"context"
	"time"
)

// Cache is a minimal key-value cache (Week 6 shared Redis facade).
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
