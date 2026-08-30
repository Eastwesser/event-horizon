package cache

import (
	"sync"
	"time"
)

// ResponseCache stores short-lived JSON HTTP responses (whoami, profile, leaderboard submit).
type ResponseCache struct {
	cache map[string]entry
	mu    sync.RWMutex
	ttl   time.Duration
}

type entry struct {
	response  []byte
	expiresAt time.Time
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	c := &ResponseCache{
		cache: make(map[string]entry),
		ttl:   ttl,
	}
	go c.cleanup()
	return c
}

// NewScoreCache is an alias for game submit score caching (2s default at call site).
func NewScoreCache(ttl time.Duration) *ResponseCache {
	return NewResponseCache(ttl)
}

func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.cache[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.response, true
}

func (c *ResponseCache) Set(key string, response []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = entry{
		response:  response,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ResponseCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, e := range c.cache {
			if now.After(e.expiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}
