package main

import (
	"fmt"
	"sync"
)

/*
Разминка: map.

1) Порядок range по map — не гарантирован (рандомизация с Go 1.0+).
2) Concurrent write без синхронизации → fatal: concurrent map writes.
   Фикс: sync.Mutex / sync.RWMutex, либо sync.Map (редко нужен на собесе — чаще Mutex).
*/
func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	fmt.Println("--- range order (может меняться между запусками) ---")
	for k, v := range m {
		fmt.Println(k, v)
	}

	fmt.Println("--- safe concurrent writes ---")
	var (
		safe sync.Mutex
		sm   = make(map[string]int)
		wg   sync.WaitGroup
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			safe.Lock()
			sm[fmt.Sprintf("key%d", id)] = id
			safe.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Println("len:", len(sm))
}
