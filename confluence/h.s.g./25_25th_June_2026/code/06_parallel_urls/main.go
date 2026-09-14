package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type Response struct {
	url    string
	status int
	err    error
}

func process(client *http.Client, url string) Response {
	r := Response{url: url}
	resp, err := client.Get(url)
	if err != nil {
		r.err = err
		return r
	}
	defer resp.Body.Close()
	r.status = resp.StatusCode
	return r
}

func main() {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer slow.Close()
	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer fast.Close()

	urls := []string{slow.URL, fast.URL, slow.URL + "/x"}
	client := &http.Client{Timeout: 2 * time.Second}
	resp := make([]Response, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			resp[i] = process(client, url)
		}(i, url)
	}
	wg.Wait()
	for _, r := range resp {
		if r.err != nil {
			fmt.Println(r.url, "ERR", r.err)
		} else {
			fmt.Println(r.url, r.status)
		}
	}
}
