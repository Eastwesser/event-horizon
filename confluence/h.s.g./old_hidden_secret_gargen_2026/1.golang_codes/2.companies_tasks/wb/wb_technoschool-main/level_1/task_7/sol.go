package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeMap - безопасная для конкуренции структура на основе map
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]int
}

// NewSafeMap создает новый экземпляр SafeMap
func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]int),
	}
}

// Set добавляет или обновляет значение в map
func (sm *SafeMap) Set(key string, value int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
	fmt.Printf("Записано: %s = %d\n", key, value)
}

// Get получает значение из map
func (sm *SafeMap) Get(key string) (int, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, exists := sm.data[key]
	return value, exists
}

// Delete удаляет элемент из map
func (sm *SafeMap) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
	fmt.Printf("Удален ключ: %s\n", key)
}

// Len возвращает количество элементов в map
func (sm *SafeMap) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.data)
}

// GetAll возвращает копию всех данных
func (sm *SafeMap) GetAll() map[string]int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]int)
	for k, v := range sm.data {
		result[k] = v
	}
	return result
}

// Демонстрация работы с безопасной map
func demonstrateSafeMap() {
	fmt.Println("=== Демонстрация безопасной map ===")
	safeMap := NewSafeMap()
	var wg sync.WaitGroup

	// Запускаем несколько горутин для записи
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := id*10 + j
				safeMap.Set(key, value)
				time.Sleep(time.Millisecond * 10)
			}
		}(i)
	}

	// Запускаем горутины для чтения
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 15; j++ {
				key := fmt.Sprintf("key_%d_%d", id%5, j%10)
				if value, exists := safeMap.Get(key); exists {
					fmt.Printf("Прочитано: %s = %d\n", key, value)
				}
				time.Sleep(time.Millisecond * 15)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("Итого элементов в безопасной map: %d\n", safeMap.Len())
}

// Пример использования sync.Map (встроенная concurrent-safe map)
func demonstrateSyncMap() {
	fmt.Println("\n=== Демонстрация sync.Map ===")
	var syncMap sync.Map
	var wg sync.WaitGroup

	// Запись в sync.Map
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				key := fmt.Sprintf("sync_key_%d_%d", id, j)
				value := id*8 + j
				syncMap.Store(key, value)
				fmt.Printf("sync.Map записано: %s = %d\n", key, value)
				time.Sleep(time.Millisecond * 8)
			}
		}(i)
	}

	// Чтение из sync.Map
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("sync_key_%d_%d", id%5, j%8)
				if value, ok := syncMap.Load(key); ok {
					fmt.Printf("sync.Map прочитано: %s = %v\n", key, value)
				}
				time.Sleep(time.Millisecond * 12)
			}
		}(i)
	}

	wg.Wait()

	// Подсчет элементов в sync.Map
	count := 0
	syncMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	fmt.Printf("Итого элементов в sync.Map: %d\n", count)
}

func main() {
	fmt.Println("Демонстрация безопасной для конкуренции работы с map")
	fmt.Println("Для проверки на гонки данных запустите: go run -race sol.go")
	fmt.Println("")

	// Демонстрация безопасной map
	demonstrateSafeMap()

	// Демонстрация встроенной sync.Map
	demonstrateSyncMap()

	fmt.Println("\n=== Завершено ===")
}
