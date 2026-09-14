package main

import (
	"fmt"
	"sync"
	"time"
)

/*
Разминка: замыкания в горутинах.

На собесе всегда объясняй классику (до Go 1.22):
  for i := 0; i < 3; i++ { go func() { fmt.Println(i) }() }
→ 3 3 3: замыкание держит *переменную* i.

Фикс (универсальный, говори его всегда):
  go func(id int){ fmt.Println(id) }(i)

С Go 1.22 у `i` per-iteration семантика — «broken» ниже может печатать 0 1 2.
Интервьюер всё равно ждёт понимание замыкания + приём с аргументом.
*/
func broken() {
	fmt.Println("--- broken (ожидай 3 3 3 на старом семантическом контракте) ---")
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Println("broken:", i)
		}()
	}
	time.Sleep(200 * time.Millisecond)
}

func fixed() {
	fmt.Println("--- fixed ---")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Println("fixed:", id)
		}(i)
	}
	wg.Wait()
}

func main() {
	broken()
	fixed()
}
