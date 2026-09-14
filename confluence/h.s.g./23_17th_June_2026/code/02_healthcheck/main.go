package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

var (
	mu         sync.RWMutex
	CacheHosts = map[string]bool{}
)

func healthcheck(client *http.Client) {
	mu.RLock()
	urls := make([]string, 0, len(CacheHosts))
	for u := range CacheHosts {
		urls = append(urls, u)
	}
	mu.RUnlock()

	for _, url := range urls {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var resp *http.Response
		var err error
		for try := 0; try < 3; try++ {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			resp, err = client.Do(req)
			if err == nil {
				break
			}
		}
		cancel()
		ok := err == nil && resp != nil && resp.StatusCode == http.StatusOK
		if resp != nil {
			_ = resp.Body.Close()
		}
		mu.Lock()
		CacheHosts[url] = ok
		mu.Unlock()
	}
}

func main() {
	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer okSrv.Close()
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer badSrv.Close()

	CacheHosts[okSrv.URL] = false
	CacheHosts[badSrv.URL] = true

	client := &http.Client{Timeout: time.Second}
	healthcheck(client)

	mu.RLock()
	fmt.Println(CacheHosts)
	mu.RUnlock()
}
