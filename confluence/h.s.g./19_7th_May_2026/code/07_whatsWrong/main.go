package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Классика «что не так»:
// 1) loop var (до 1.22) — передаём аргументом
// 2) http.Get без timeout
// 3) игнор ошибки / Body.Close
func main() {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ok.Close()

	urls := []string{ok.URL, ok.URL + "/a", ok.URL + "/b"}
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 2 * time.Second}

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			resp, err := client.Get(u)
			if err != nil {
				fmt.Println("Error:", u, err)
				return
			}
			defer resp.Body.Close()
			fmt.Println("OK:", u, resp.StatusCode)
		}(url)
	}
	wg.Wait()
}
