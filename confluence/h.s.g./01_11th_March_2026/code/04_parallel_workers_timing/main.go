// Тайминг worker(): важно КАК вызываешь.
// 1) ch1, ch2 := worker(), worker() → оба стартуют → ~3с
// 2) _, _ = <-worker(), <-worker() → RHS слева направо → ~6с
package main

import (
	"fmt"
	"time"
)

func worker() <-chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch <- 1
	}()
	return ch
}

func main() {
	// параллельно: оба worker() до receive
	t0 := time.Now()
	ch1, ch2 := worker(), worker()
	_, _ = <-ch1, <-ch2
	fmt.Println("parallel (assign chans first):", time.Since(t0).Round(10*time.Millisecond))

	// последовательно: каждый <-worker() ждёт своего завершения
	t1 := time.Now()
	_, _ = <-worker(), <-worker()
	fmt.Println("sequential (recv in RHS):     ", time.Since(t1).Round(10*time.Millisecond))
}
