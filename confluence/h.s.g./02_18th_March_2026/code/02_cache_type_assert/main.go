// Баг обёртки над LRU: Put(value) vs Get assert *value → вечный miss.
package main

import "fmt"

type Warehouse struct {
	ID   int
	Name string
}

// «сломанный» кэш как в конспекте
type BrokenStorage struct {
	cache map[int]any
}

func (s *BrokenStorage) Set(wh *Warehouse) {
	s.cache[wh.ID] = *wh // value
}

func (s *BrokenStorage) Get(id int) *Warehouse {
	item, ok := s.cache[id]
	if !ok {
		return nil
	}
	if wh, ok := item.(*Warehouse); ok { // assert pointer → fail
		return wh
	}
	fmt.Println("broken: type assert failed, cache useless")
	return nil
}

// исправление: один тип туда и обратно
type FixedStorage struct {
	cache map[int]*Warehouse
}

func (s *FixedStorage) Set(wh *Warehouse) {
	s.cache[wh.ID] = wh
}

func (s *FixedStorage) Get(id int) *Warehouse {
	return s.cache[id]
}

func main() {
	b := &BrokenStorage{cache: map[int]any{}}
	w := &Warehouse{ID: 1, Name: "MSK"}
	b.Set(w)
	fmt.Println("broken get:", b.Get(1))

	f := &FixedStorage{cache: map[int]*Warehouse{}}
	f.Set(w)
	fmt.Println("fixed get:", f.Get(1))
}
