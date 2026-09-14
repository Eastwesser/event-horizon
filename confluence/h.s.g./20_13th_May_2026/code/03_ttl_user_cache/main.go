package main

import (
	"fmt"
	"sync"
	"time"
)

type User struct {
	ID   int64
	Name string
}

type item struct {
	user      User
	expiresAt time.Time
}

// Cache с кастомным TTL на элемент; Get не продлевает TTL.
type Cache struct {
	mu   sync.RWMutex
	data map[int64]item
}

func NewCache() *Cache { return &Cache{data: make(map[int64]item)} }

func (c *Cache) Set(u User, ttl time.Duration) {
	c.mu.Lock()
	c.data[u.ID] = item{user: u, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *Cache) Get(id int64) (User, bool) {
	c.mu.RLock()
	it, ok := c.data[id]
	c.mu.RUnlock()
	if !ok || time.Now().After(it.expiresAt) {
		return User{}, false
	}
	return it.user, true
}

func (c *Cache) Delete(id int64) {
	c.mu.Lock()
	delete(c.data, id)
	c.mu.Unlock()
}

func main() {
	c := NewCache()
	c.Set(User{ID: 1, Name: "Ann"}, 40*time.Millisecond)
	u, ok := c.Get(1)
	fmt.Println(u, ok)
	time.Sleep(50 * time.Millisecond)
	_, ok = c.Get(1)
	fmt.Println("after ttl:", ok)
	c.Set(User{ID: 2, Name: "Bob"}, time.Second)
	c.Delete(2)
	_, ok = c.Get(2)
	fmt.Println("after delete:", ok)
}
