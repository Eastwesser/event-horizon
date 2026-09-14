package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

/*
Конкурентность: retry внешнего API с коротким backoff (демо в ms).
Ретраить 5xx/сеть; 4xx обычно нет (кроме 429).
*/
func retry(ctx context.Context, max int, fn func(context.Context) error) error {
	var last error
	for attempt := 0; attempt < max; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = fn(ctx)
		if last == nil {
			return nil
		}
		if attempt == max-1 {
			break
		}
		wait := time.Duration(1<<attempt) * 30 * time.Millisecond
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
		fmt.Printf("retry #%d after %v (err=%v)\n", attempt+1, wait, last)
	}
	return last
}

func main() {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{}
	err := retry(context.Background(), 5, func(ctx context.Context) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		fmt.Println("FAIL:", err)
		return
	}
	fmt.Println("OK after retries, hits=", hits.Load())
}
