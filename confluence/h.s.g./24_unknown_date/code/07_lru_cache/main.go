package main

import "fmt"

type node struct {
	key, val   int
	prev, next *node
}

// LRU — map + двусвязный список, O(1) get/put.
type LRU struct {
	cap        int
	m          map[int]*node
	head, tail *node
}

func NewLRU(capacity int) *LRU {
	h, t := &node{}, &node{}
	h.next, t.prev = t, h
	return &LRU{cap: capacity, m: make(map[int]*node), head: h, tail: t}
}

func (l *LRU) remove(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (l *LRU) pushFront(n *node) {
	n.next = l.head.next
	n.prev = l.head
	l.head.next.prev = n
	l.head.next = n
}

func (l *LRU) Get(key int) (int, bool) {
	n, ok := l.m[key]
	if !ok {
		return 0, false
	}
	l.remove(n)
	l.pushFront(n)
	return n.val, true
}

func (l *LRU) Put(key, val int) {
	if n, ok := l.m[key]; ok {
		n.val = val
		l.remove(n)
		l.pushFront(n)
		return
	}
	n := &node{key: key, val: val}
	l.m[key] = n
	l.pushFront(n)
	if len(l.m) > l.cap {
		old := l.tail.prev
		l.remove(old)
		delete(l.m, old.key)
	}
}

func main() {
	c := NewLRU(2)
	c.Put(1, 1)
	c.Put(2, 2)
	fmt.Println(c.Get(1))
	c.Put(3, 3)
	fmt.Println(c.Get(2)) // false — вытеснен
}
