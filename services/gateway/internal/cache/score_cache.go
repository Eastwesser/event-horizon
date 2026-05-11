package cache

import (
    "sync"
    "time"
)

type ScoreCacheEntry struct {
    Response []byte
    ExpiresAt time.Time
}

type ScoreCache struct {
    cache map[string]ScoreCacheEntry
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewScoreCache(ttl time.Duration) *ScoreCache {
    c := &ScoreCache{
        cache: make(map[string]ScoreCacheEntry),
        ttl:   ttl,
    }
    go c.cleanup()
    return c
}

func (c *ScoreCache) Get(key string) ([]byte, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    entry, exists := c.cache[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    return entry.Response, true
}

func (c *ScoreCache) Set(key string, response []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache[key] = ScoreCacheEntry{
        Response:  response,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

func (c *ScoreCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        c.mu.Lock()
        for key, entry := range c.cache {
            if time.Now().After(entry.ExpiresAt) {
                delete(c.cache, key)
            }
        }
        c.mu.Unlock()
    }
}
