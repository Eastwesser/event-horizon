// First success: N реплик → первый успешный ответ; cancel остальным.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Result struct {
	Docs []string
	Err  error
}

func search(server, query string) Result {
	delay := map[string]time.Duration{
		"s1": 80 * time.Millisecond,
		"s2": 30 * time.Millisecond, // быстрее и успех
		"s3": 100 * time.Millisecond,
	}
	time.Sleep(delay[server])
	if server == "s1" {
		return Result{Err: errors.New("s1 down")}
	}
	return Result{Docs: []string{server + ":" + query}}
}

func firstSuccess(ctx context.Context, servers []string, query string) (Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan Result, len(servers)) // буфер ≥ N — без утечки на send
	var wg sync.WaitGroup
	for _, ser := range servers {
		wg.Add(1)
		go func(ser string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			r := search(ser, query)
			select {
			case ch <- r:
			case <-ctx.Done():
			}
		}(ser)
	}
	go func() { wg.Wait(); close(ch) }()

	var last error
	for r := range ch {
		if r.Err == nil {
			cancel()
			return r, nil
		}
		last = r.Err
	}
	if last == nil {
		last = errors.New("all failed")
	}
	return Result{}, last
}

func main() {
	r, err := firstSuccess(context.Background(), []string{"s1", "s2", "s3"}, "go")
	fmt.Println(r.Docs, err)
}
