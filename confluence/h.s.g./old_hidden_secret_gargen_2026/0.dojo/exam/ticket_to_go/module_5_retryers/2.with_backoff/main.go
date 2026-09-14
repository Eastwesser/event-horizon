package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

/*
Exponential backoff + jitter.

Jitter размазывает клиенты → нет thundering herd после общего outage.
Full jitter (AWS): sleep = rand(0, backoff) часто лучше, чем fixed ±25%.

Retry ≠ Circuit Breaker:
  Retry — ещё раз этот запрос.
  CB — временно не слать запросы зависимости вообще.
*/
var ErrPermanent = errors.New("permanent: do not retry")

func retryable(err error) bool {
	return err != nil && !errors.Is(err, ErrPermanent)
}

func retryBackoffJitter(fn func() error, maxRetries int, base, max time.Duration) error {
	var last error
	for i := 0; i < maxRetries; i++ {
		last = fn()
		if last == nil {
			return nil
		}
		if !retryable(last) {
			return last
		}
		if i == maxRetries-1 {
			break
		}
		exp := base * time.Duration(1<<i)
		if exp > max {
			exp = max
		}
		// full jitter
		wait := time.Duration(rand.Int63n(int64(exp) + 1))
		fmt.Printf("attempt %d err=%v wait=%v\n", i+1, last, wait)
		time.Sleep(wait)
	}
	return last
}

func main() {
	n := 0
	err := retryBackoffJitter(func() error {
		n++
		if n < 4 {
			return errors.New("503 upstream")
		}
		return nil
	}, 6, 30*time.Millisecond, 200*time.Millisecond)
	fmt.Println("flaky OK:", err, "calls:", n)

	err = retryBackoffJitter(func() error {
		return fmt.Errorf("bad request: %w", ErrPermanent)
	}, 5, 30*time.Millisecond, 200*time.Millisecond)
	fmt.Println("permanent:", err)
}
