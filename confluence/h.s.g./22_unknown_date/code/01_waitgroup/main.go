package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done() // без Done — Wait повиснет навсегда
			time.Sleep(20 * time.Millisecond)
			fmt.Println("worker", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("all done")
}
