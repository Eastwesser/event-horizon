package main

import "fmt"

/*
LRU Cache — главный паттерн модуля (LeetCode 146).

map + doubly linked list + sentinels head/tail.
Get/Put O(1). Evict tail.prev при переполнении.
*/
type Node struct {
	key, value int
	prev, next *Node
}

type LRUCache struct {
	capacity int
	cache    map[int]*Node
	head     *Node
	tail     *Node
}

func Constructor(capacity int) LRUCache {
	h, t := &Node{}, &Node{}
	h.next, t.prev = t, h
	return LRUCache{capacity: capacity, cache: make(map[int]*Node), head: h, tail: t}
}

func (c *LRUCache) moveToFront(n *Node) {
	c.remove(n)
	c.addToFront(n)
}

func (c *LRUCache) remove(n *Node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (c *LRUCache) addToFront(n *Node) {
	n.next = c.head.next
	n.prev = c.head
	c.head.next.prev = n
	c.head.next = n
}

func (c *LRUCache) removeLRU() {
	lru := c.tail.prev
	c.remove(lru)
	delete(c.cache, lru.key)
}

func (c *LRUCache) Get(key int) int {
	if n, ok := c.cache[key]; ok {
		c.moveToFront(n)
		return n.value
	}
	return -1
}

func (c *LRUCache) Put(key, value int) {
	if n, ok := c.cache[key]; ok {
		n.value = value
		c.moveToFront(n)
		return
	}
	n := &Node{key: key, value: value}
	c.cache[key] = n
	c.addToFront(n)
	if len(c.cache) > c.capacity {
		c.removeLRU()
	}
}

func main() {
	c := Constructor(2)
	c.Put(1, 1)
	c.Put(2, 2)
	fmt.Println(c.Get(1)) // 1
	c.Put(3, 3)           // evicts 2
	fmt.Println(c.Get(2)) // -1
	c.Put(4, 4)           // evicts 1
	fmt.Println(c.Get(1)) // -1
	fmt.Println(c.Get(3)) // 3
	fmt.Println(c.Get(4)) // 4
}
