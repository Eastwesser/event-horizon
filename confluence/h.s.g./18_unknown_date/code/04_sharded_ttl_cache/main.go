package main

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type entry struct {
	val       string
	expiresAt time.Time
}

type shard struct {
	mu   sync.RWMutex
	data map[string]entry
}

// ShardedCache — []*shard + RWMutex + фоновый janitor. Без sync.Map.
type ShardedCache struct {
	shards []*shard
	ttl    time.Duration
	stop   chan struct{}
}

func NewShardedCache(n int, defaultTTL time.Duration) *ShardedCache {
	c := &ShardedCache{shards: make([]*shard, n), ttl: defaultTTL, stop: make(chan struct{})}
	for i := range c.shards {
		c.shards[i] = &shard{data: make(map[string]entry)}
	}
	go c.janitor(30 * time.Millisecond)
	return c
}

func (c *ShardedCache) shardFor(key string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return c.shards[int(h.Sum32())%len(c.shards)]
}

func (c *ShardedCache) Set(key, val string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.ttl
	}
	s := c.shardFor(key)
	s.mu.Lock()
	s.data[key] = entry{val: val, expiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
}

func (c *ShardedCache) Get(key string) (string, bool) {
	s := c.shardFor(key)
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.val, true
}

func (c *ShardedCache) janitor(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			now := time.Now()
			for _, s := range c.shards {
				s.mu.Lock()
				for k, e := range s.data {
					if now.After(e.expiresAt) {
						delete(s.data, k)
					}
				}
				s.mu.Unlock()
			}
		}
	}
}

func (c *ShardedCache) Close() { close(c.stop) }

func main() {
	c := NewShardedCache(4, time.Second)
	defer c.Close()
	c.Set("msk", "55.75,37.61", 40*time.Millisecond)
	fmt.Println(c.Get("msk"))
	time.Sleep(60 * time.Millisecond)
	fmt.Println(c.Get("msk")) // false после TTL
}
