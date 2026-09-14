// Async HTTP: параллельные запросы + timeout через select на ctx.Done().
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Err        error
}

func fetch(ctx context.Context, client *http.Client, raw string) Result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return Result{URL: raw, Err: err}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: raw, Err: err}
	}
	defer resp.Body.Close()
	return Result{URL: raw, StatusCode: resp.StatusCode}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/yandex", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/google", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(40 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	urls := []string{srv.URL + "/yandex", srv.URL + "/google"}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	client := &http.Client{} // deadline уже в ctx запроса
	out := make(chan Result, len(urls))
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			select {
			case out <- fetch(ctx, client, u):
			case <-ctx.Done():
				out <- Result{URL: u, Err: ctx.Err()}
			}
		}(u)
	}
	go func() { wg.Wait(); close(out) }()

	for r := range out {
		if r.Err != nil {
			fmt.Printf("%s err=%v\n", r.URL, r.Err)
			continue
		}
		fmt.Printf("%s status=%d\n", r.URL, r.StatusCode)
	}
}
