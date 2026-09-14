package main

import (
	"fmt"
	"sync"
)

/*
Thread-safe LRU: Mutex на Get/Put (RWMutex почти не помогает —
Get тоже мутирует порядок → почти всегда Lock).
*/
type node struct {
	key, val   int
	prev, next *node
}

type LRUCache struct {
	cap        int
	items      map[int]*node
	head, tail *node
}

func NewLRU(capacity int) *LRUCache {
	h, t := &node{}, &node{}
	h.next, t.prev = t, h
	return &LRUCache{cap: capacity, items: make(map[int]*node), head: h, tail: t}
}

func (c *LRUCache) remove(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (c *LRUCache) pushFront(n *node) {
	n.next = c.head.next
	n.prev = c.head
	c.head.next.prev = n
	c.head.next = n
}

func (c *LRUCache) Get(key int) int {
	n, ok := c.items[key]
	if !ok {
		return -1
	}
	c.remove(n)
	c.pushFront(n)
	return n.val
}

func (c *LRUCache) Put(key, value int) {
	if n, ok := c.items[key]; ok {
		n.val = value
		c.remove(n)
		c.pushFront(n)
		return
	}
	n := &node{key: key, val: value}
	c.items[key] = n
	c.pushFront(n)
	if len(c.items) > c.cap {
		lru := c.tail.prev
		c.remove(lru)
		delete(c.items, lru.key)
	}
}

type SafeLRU struct {
	mu    sync.Mutex
	cache *LRUCache
}

func NewSafeLRU(cap int) *SafeLRU {
	return &SafeLRU{cache: NewLRU(cap)}
}

func (s *SafeLRU) Get(key int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cache.Get(key)
}

func (s *SafeLRU) Put(key, value int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache.Put(key, value)
}

func main() {
	s := NewSafeLRU(2)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Put(i%3, i)
			_ = s.Get(i % 3)
		}(i)
	}
	wg.Wait()
	fmt.Println("get0:", s.Get(0), "get1:", s.Get(1), "get2:", s.Get(2))
}
