// Concurrent cache: shared RWMutex, double-checked GetOrCreate.
// Оригинал: локальный mutex + write под RLock — data race / бесполезная защита.
package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

// GetOrCreate: read path быстрый; write один раз под exclusive lock.
func (c *Cache) GetOrCreate(key, value string) string {
	c.mu.RLock()
	if v, ok := c.data[key]; ok {
		c.mu.RUnlock()
		return v
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.data[key]; ok { // другой мог успеть
		return v
	}
	c.data[key] = value
	return value
}

func main() {
	c := NewCache()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(c.GetOrCreate("hello", "world"))
		}()
	}
	wg.Wait()
	v, ok := c.Get("hello")
	fmt.Println(v, ok)
}
