package main
import "fmt"
type node struct{ k, v int; prev, next *node }
type LRU struct {
	cap int
	m map[int]*node
	head, tail *node
}
func NewLRU(c int) *LRU {
	h, t := &node{}, &node{}
	h.next, t.prev = t, h
	return &LRU{cap: c, m: map[int]*node{}, head: h, tail: t}
}
func (l *LRU) rem(n *node) { n.prev.next = n.next; n.next.prev = n.prev }
func (l *LRU) front(n *node) {
	n.next = l.head.next; n.prev = l.head
	l.head.next.prev = n; l.head.next = n
}
func (l *LRU) Get(k int) int {
	n, ok := l.m[k]; if !ok { return -1 }
	l.rem(n); l.front(n); return n.v
}
func (l *LRU) Put(k, v int) {
	if n, ok := l.m[k]; ok { n.v = v; l.rem(n); l.front(n); return }
	n := &node{k: k, v: v}; l.m[k] = n; l.front(n)
	if len(l.m) > l.cap {
		x := l.tail.prev; l.rem(x); delete(l.m, x.k)
	}
}
func main() {
	c := NewLRU(2); c.Put(1,1); c.Put(2,2); fmt.Println(c.Get(1)); c.Put(3,3); fmt.Println(c.Get(2))
}
