package main

import (
	"fmt"
	"sync"
)

/*
Write-Through vs Write-Back (Cache-Aside рядом — как в EH Inventory).

Write-Through: Put пишет в cache + store синхронно → консистентно, медленнее.
Write-Back:    Put только в cache, dirty flush позже → быстро, риск потери.
Cache-Aside:   app читает cache→miss→DB→fill; пишет DB→invalidate/update cache.
*/
type Store struct {
	mu   sync.Mutex
	data map[int]int
}

func NewStore() *Store { return &Store{data: make(map[int]int)} }

func (s *Store) Set(k, v int) {
	s.mu.Lock()
	s.data[k] = v
	s.mu.Unlock()
}

func (s *Store) Get(k int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[k]
	return v, ok
}

type memCache struct {
	mu    sync.Mutex
	data  map[int]int
	dirty map[int]bool
}

func newMem() *memCache {
	return &memCache{data: make(map[int]int), dirty: make(map[int]bool)}
}

func (c *memCache) get(k int) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *memCache) set(k, v int, markDirty bool) {
	c.mu.Lock()
	c.data[k] = v
	if markDirty {
		c.dirty[k] = true
	}
	c.mu.Unlock()
}

// WriteThroughCache
type WriteThrough struct {
	cache *memCache
	store *Store
}

func (w *WriteThrough) Put(k, v int) {
	w.store.Set(k, v)
	w.cache.set(k, v, false)
}

func (w *WriteThrough) Get(k int) (int, bool) {
	if v, ok := w.cache.get(k); ok {
		return v, true
	}
	v, ok := w.store.Get(k)
	if ok {
		w.cache.set(k, v, false)
	}
	return v, ok
}

// WriteBackCache
type WriteBack struct {
	cache *memCache
	store *Store
}

func (w *WriteBack) Put(k, v int) {
	w.cache.set(k, v, true) // только кэш
}

func (w *WriteBack) Flush() {
	w.cache.mu.Lock()
	defer w.cache.mu.Unlock()
	for k := range w.cache.dirty {
		w.store.Set(k, w.cache.data[k])
		delete(w.cache.dirty, k)
	}
}

func (w *WriteBack) Get(k int) (int, bool) {
	if v, ok := w.cache.get(k); ok {
		return v, true
	}
	v, ok := w.store.Get(k)
	if ok {
		w.cache.set(k, v, false)
	}
	return v, ok
}

func main() {
	fmt.Println("--- write-through ---")
	st := NewStore()
	wt := &WriteThrough{cache: newMem(), store: st}
	wt.Put(1, 100)
	fmt.Println("store has", must(st.Get(1)), "cache", must(wt.Get(1)))

	fmt.Println("--- write-back ---")
	st2 := NewStore()
	wb := &WriteBack{cache: newMem(), store: st2}
	wb.Put(2, 200)
	_, inStore := st2.Get(2)
	fmt.Println("before flush in store?", inStore, "from cache", must(wb.Get(2)))
	wb.Flush()
	fmt.Println("after flush store", must(st2.Get(2)))
}

func must(v int, ok bool) int {
	if !ok {
		return -1
	}
	return v
}
