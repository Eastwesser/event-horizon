// Есть 100 url'ов. Нужно скачать их все, но одновременно может выполняться не более 5 запросов.
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	totalURLs        = 100
	concurrencyLimit = 5
)

// download имитирует долгий сетевой запрос
func download(url int) {
	// Имитация работы (от 100 до 500 мс)
	time.Sleep(time.Duration(100+url%5*100) * time.Millisecond)
	fmt.Printf("Downloaded: %d\n", url)
}

func main() {

	// 1. Подготовка данных (просто список ID)
	urls := make([]int, totalURLs)

	for i := 0; i < totalURLs; i++ {
		urls[i] = i
	}

	fmt.Println("Start downloading...")

	start := time.Now()

	// TODO: Реализуйте запуск download(url) параллельно,
	// но с ограничением в concurrencyLimit (5) одновременных запросов.
	// Ваш код здесь:
	var wg sync.WaitGroup

	ch := make(chan struct{}, concurrencyLimit)

	for _, url := range urls {
		wg.Add(1)

		go func(id int) {
			ch <- struct{}{}
			defer func() {
				<-ch
				wg.Done()
			}()
			download(id)
		}(url)

	}

	wg.Wait()

	fmt.Printf("Took: %v\n", time.Since(start))
}
