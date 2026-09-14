package main

import (
	"net/http"
	"sync"
)

func main() {
	urls := []string{"a.com", "b.com", "c.com"}
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func() { // сюда надо передать значение go func(url string) {
			defer wg.Done()
			http.Get(url) // нужно обработать ошибку, если Get повиснет
			//_, err := http.Get(url)
			//if err != nil {
			//	fmt.Println("Error", err)
			//}
		}() // сюда надо передать значение }(url)
	}
	wg.Wait()
}
