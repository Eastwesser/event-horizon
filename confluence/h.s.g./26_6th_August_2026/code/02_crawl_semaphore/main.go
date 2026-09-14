package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// crawl — не больше k одновременных запросов (семафор на chan).
func crawl(urls []string, k int) []int {
	sem := make(chan struct{}, k)
	res := make([]int, len(urls))
	var wg sync.WaitGroup
	client := &http.Client{}

	for i, url := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, url string) {
			defer wg.Done()
			defer func() { <-sem }()
			resp, err := client.Get(url)
			if err != nil {
				res[i] = 0
				return
			}
			res[i] = resp.StatusCode
			_ = resp.Body.Close()
		}(i, url)
	}
	wg.Wait()
	return res
}

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	urls := make([]string, 10)
	for i := range urls {
		urls[i] = srv.URL
	}
	fmt.Println(crawl(urls, 3))
}
