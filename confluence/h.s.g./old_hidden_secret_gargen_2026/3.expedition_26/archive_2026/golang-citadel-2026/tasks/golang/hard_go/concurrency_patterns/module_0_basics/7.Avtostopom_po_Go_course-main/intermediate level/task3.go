package main

import (
	"fmt"
	"sync"
	"time"
)

// Расскажи подробно что происходит
// Сколько времени будет выполняться?
const numRequests = 10000

var count int

var m sync.Mutex

func networkRequest() {
	time.Sleep(time.Millisecond) // Эмуляция сетевого запроса.
	m.Lock()
	count++
	m.Unlock()
}

func main() {
	var wg sync.WaitGroup

	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			networkRequest()
		}()
	}

	wg.Wait()
	fmt.Println(count)
}
