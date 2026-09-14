package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

type Cache interface {
	Get(k string) int
	Set(k string, v int)
}

type shard struct {
	mu   sync.RWMutex
	data map[string]int
}

type PartitionedCache struct {
	shards []*shard
}

func NewPartitionedCache(n int) *PartitionedCache {
	c := &PartitionedCache{shards: make([]*shard, n)}
	for i := range c.shards {
		c.shards[i] = &shard{data: make(map[string]int)}
	}
	return c
}

func (c *PartitionedCache) shard(k string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(k))
	return c.shards[int(h.Sum32())%len(c.shards)]
}

func (c *PartitionedCache) Get(k string) int {
	s := c.shard(k)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[k]
}

func (c *PartitionedCache) Set(k string, v int) {
	s := c.shard(k)
	s.mu.Lock()
	s.data[k] = v
	s.mu.Unlock()
}

func main() {
	var c Cache = NewPartitionedCache(8)
	c.Set("a", 10)
	fmt.Println(c.Get("a"))
}
