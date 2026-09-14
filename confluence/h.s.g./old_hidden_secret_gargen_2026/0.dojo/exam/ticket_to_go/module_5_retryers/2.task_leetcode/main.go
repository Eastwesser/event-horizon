package main

import "fmt"

/*
LeetCode 146 — LRU Cache (минимум для собеса).
map[key]*node + doubly linked list. Get/Put O(1).
Полный разбор также в Module 6.
*/
type node struct {
	key, val   int
	prev, next *node
}

type LRUCache struct {
	cap        int
	items      map[int]*node
	head, tail *node // sentinel
}

func Constructor(capacity int) LRUCache {
	h, t := &node{}, &node{}
	h.next, t.prev = t, h
	return LRUCache{cap: capacity, items: make(map[int]*node), head: h, tail: t}
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

func main() {
	c := Constructor(2)
	c.Put(1, 1)
	c.Put(2, 2)
	fmt.Println(c.Get(1)) // 1
	c.Put(3, 3)           // evict 2
	fmt.Println(c.Get(2)) // -1
	fmt.Println(c.Get(3)) // 3
}
