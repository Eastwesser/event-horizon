package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Cache — фоновое обновление прогноза под RWMutex.
type Cache struct {
	mu   sync.RWMutex
	data int
}

func aiWeatherForecast() int {
	time.Sleep(5 * time.Millisecond) // «нейронка ~1с» — ускорено для демо
	return 21
}

func NewCache() *Cache {
	c := &Cache{data: aiWeatherForecast()}
	go c.Updater(50 * time.Millisecond)
	return c
}

func (c *Cache) Updater(every time.Duration) {
	t := time.NewTicker(every)
	for range t.C {
		res := aiWeatherForecast()
		c.mu.Lock()
		c.data = res
		c.mu.Unlock()
	}
}

func (c *Cache) Get() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

func main() {
	c := NewCache()
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"temperature":%d}`+"\n", c.Get())
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/weather")
	defer resp.Body.Close()
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	fmt.Print(string(buf[:n]))
}
