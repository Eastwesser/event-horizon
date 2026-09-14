package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Что было не так (классика собеса):
// 1) loop variable: до Go 1.22 все горутины видели один и тот же url → передаём аргументом.
// 2) http.Get без timeout / context — может повиснуть навсегда.
// 3) ошибка Get игнорировалась.
// 4) нет ограничения concurrency (здесь ок для 3 URL; в проде — семафор из задачи 1).
func main() {
	urls := []string{
		"https://example.com",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
	}
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 3 * time.Second}

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
