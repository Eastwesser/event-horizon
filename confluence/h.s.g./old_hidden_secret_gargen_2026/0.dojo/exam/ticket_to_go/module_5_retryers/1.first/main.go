package main

import (
	"errors"
	"fmt"
	"time"
)

/*
Retry + exponential backoff (без jitter).

backoff: base * 2^attempt  →  50ms, 100ms, 200ms...
На собесе: MaxRetries обязателен; без потолка = бесконечный шторм.
*/
func retryWithBackoff(fn func() error, maxRetries int, base time.Duration) error {
	var last error
	for i := 0; i < maxRetries; i++ {
		last = fn()
		if last == nil {
			return nil
		}
		if i == maxRetries-1 {
			break
		}
		backoff := base * time.Duration(1<<i)
		fmt.Printf("attempt %d failed (%v), sleep %v\n", i+1, last, backoff)
		time.Sleep(backoff)
	}
	return last
}

func main() {
	n := 0
	err := retryWithBackoff(func() error {
		n++
		if n < 3 {
			return errors.New("temporary")
		}
		return nil
	}, 5, 40*time.Millisecond)

	fmt.Println("result:", err, "calls:", n)
}
