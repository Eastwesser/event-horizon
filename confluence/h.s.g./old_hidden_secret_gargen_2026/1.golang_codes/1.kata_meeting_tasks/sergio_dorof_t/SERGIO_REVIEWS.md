```go
// 1) "Задачка на append()" ============================================================================================

package main

import "fmt"

func a() {
	x := []int{}      // len=0, cap=0
	x = append(x, 0)  // len=1, cap=1, [0]
	x = append(x, 1)  // len=2, cap=2, [0,1]
	x = append(x, 2)  // len=3, cap=4, [0,1,2] (выделился новый массив с cap=4)
	y := append(x, 3) // x все еще len=3, cap=4, добавляем 3 → [0,1,2,3]
	z := append(x, 4) // x все еще len=3, cap=4, добавляем 4 → [0,1,2,4]
	fmt.Println(y, z) // [0 1 2 4] [0 1 2 4] - объясни, почему? Из-за ссылки на базовый массив
}

/*
    Ключевой момент: x остается неизменным (len=3, cap=4). Оба append работают с одним и тем же базовым массивом x, но:
    y := append(x, 3) — записывает 3 в 4-ю позицию массива
    z := append(x, 4) — записывает 4 в 4-ю позицию массива (перезаписывая 3)
    Поэтому y и z показывают [0 1 2 4] — они оба ссылаются на один массив, который был изменен последним append.
*/

// "Напиши foo()"
func foo(sl [6]int) {
	fmt.Println(sl) // [1 2 3 4 5 6]
}

func main() {
	a() // [0 1 2 4] [0 1 2 4]
	sl := [6]int{1, 2, 3, 4, 5, 6}
	foo(sl)
}

```

```go
// 2) "Скачай видосы по URLs" ==========================================================================================
package main

import (
	"fmt"
	"net/http"
	"sync"
)


type Result struct {
	url string
	err error
	StatusCode string
}

func main() {
	urls := []string{
		"https://www.youtube.com/watch?v=2cxmJUJ2Ge0", // Козырев про контекст в Golang   
		"https://www.youtube.com/watch?v=4aTt9E-EG-o&t=202s", // skill issue про конкурентность
		"https://www.youtube.com/watch?v=k-1OEYl7N8Q", // skill issue про chan
		"https://www.youtube.com/watch?v=hbseyn-CfXY", // Влад Мишустин про Кафка
		"https://www.youtube.com/watch?v=zvdZXO8GWd4", // Балун Как работают профессионалы с базами данных и кешем в Golang?        
		"https://www.youtube.com/watch?v=8NcufYa2HrQ&t=211s", // Николай Павлин (Про EXPLAIN ANALYZE) 
	}

	ch := make(chan Result, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) { // передаем url как параметр
			defer wg.Done()
			resp, err := http.Get(u)
			if err != nil {
				ch <- Result{
					url: u, 
					err: err,
				}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ch <- Result{
					url: u,
					err: fmt.Errorf("HTTP error: %s", resp.Status),
				}
				return
			}
			ch <- Result{
				url: u,
				StatusCode: resp.Status,
			}
		}(url) // передаем текущий url
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for c := range ch {
		fmt.Println(c)
	}
}

```

```go
// 3) "CACHE" ==========================================================================================================
package main

import (
	"sync"
)

type Cache interface {
	Set(k, v string)
	Get(k string) (v string, ok bool)
}

type ICache struct {
	storage map[string]string
	mu      sync.RWMutex
}

func NewICache() Cache {  // ← возвращаем интерфейс!
	return &ICache{
		storage: make(map[string]string),
	}
}

func (c *ICache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storage[k] = v
}

func (c *ICache) Get(k string) (v string, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok = c.storage[k]
	return v, ok
}

```

```go
// 4) "LRU CACHE"  =====================================================================================================

// LRU (Least Recently Used) — вытесняет наименее недавно использованный элемент

LRU.cache(capacity 2) // cache{} cap = 2

cache.PUT(A, 1) // {A:1}                  // A - последний использованный
cache.PUT(B, 2) // {A:1 B:2}              // B - последний
cache.GET(A)    // {1}                     // A - последний, B становится "старым"
// Получаем A, он становится последним использованным
cache.PUT(C, 3) // {A:1 C:3}               // B вытеснен (наименее недавний)
// B удален, C добавлен
cache.GET(B)    // {ERROR}                  // B уже нет в кэше
cache.GET(C)    // {3}                      // C есть, он последний использованный
cache.PUT(D, 4) // {C:3 D:4}               // A вытеснен
// A удален (наименее недавний после GET(C)), D добавлен
cache.GET(A)    // {ERROR}                  // A уже нет в кэше

// Что будет выведено на каждой строке?
```