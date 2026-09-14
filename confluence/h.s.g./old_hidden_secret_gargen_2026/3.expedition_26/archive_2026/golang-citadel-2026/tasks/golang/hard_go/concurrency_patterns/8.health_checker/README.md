# Паттерн: Health Checker — Мониторинг состояния сервисов

## 🎯 Задача

Периодически проверять **доступность N сервисов** и обновлять их статус, чтобы знать, какие сервисы доступны для использования.

## 🗡️ Классическая задача (из WB)

> Реализовать **Health Checker**, который:
> 1. Проверяет N хостов параллельно
> 2. Обновляет статус каждые T секунд
> 3. Не блокирует основной поток
> 4. Показывает прогресс проверки

## 🔍 Архитектура

```
                   ┌──────────────────┐
                   │  Health Checker  │
                   └────────┬─────────┘
                            │
                   ┌────────▼─────────┐
                   │   CheckAllHosts  │
                   │  (parallel check)│
                   └────────┬─────────┘
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
    ┌─────▼─────┐     ┌─────▼─────┐     ┌─────▼─────┐
    │ Host A    │     │ Host B    │     │ Host C    │
    │ GET /health│     │ GET /health│     │ GET /health│
    └───────────┘     └───────────┘     └───────────┘
          │                 │                 │
        Alive             Dead              Alive
```

## 🧪 Базовая реализация

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	hosts   = make(map[string]bool)
	hostsMu sync.RWMutex
)

// HealthCheck выполняет GET-запрос к URL
func HealthCheck(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// CheckAllHosts проверяет все хосты параллельно
func CheckAllHosts(ctx context.Context) map[string]error {
	hostsMu.RLock()
	urls := make([]string, 0, len(hosts))
	for url := range hosts {
		urls = append(urls, url)
	}
	hostsMu.RUnlock()

	if len(urls) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		errors  = make(map[string]error)
		checked int32
		total   = int32(len(urls))
	)

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			// Индивидуальный timeout для каждой проверки
			checkCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()

			status, err := HealthCheck(checkCtx, u)

			mu.Lock()
			if err != nil {
				errors[u] = err
				hosts[u] = false  // Помечаем как dead
			} else {
				hosts[u] = status
			}
			mu.Unlock()

			// Прогресс
			atomic.AddInt32(&checked, 1)
			progress := float64(atomic.LoadInt32(&checked)) / float64(total) * 100
			fmt.Printf("[Progress: %.0f%%] Checked %s: %v\n",
				progress, u, err == nil && status)
		}(url)
	}

	wg.Wait()
	return errors
}

// StartHealthChecker запускает периодические проверки
func StartHealthChecker(ctx context.Context, interval time.Duration) {
	fmt.Printf("Starting health checker (interval: %v)\n", interval)

	// Начальная проверка
	fmt.Println("\n=== Initial health check ===")
	if errs := CheckAllHosts(ctx); len(errs) > 0 {
		fmt.Println("Some hosts had errors:")
		for url, err := range errs {
			fmt.Printf("  %s: %v\n", url, err)
		}
	}
	printStatuses()

	// Периодические проверки
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for checkCount := 1; ; checkCount++ {
		select {
		case <-ctx.Done():
			fmt.Println("\nHealth checker stopped")
			return
		case <-ticker.C:
			fmt.Printf("\n=== Health check #%d ===\n", checkCount)
			start := time.Now()

			errs := CheckAllHosts(ctx)

			duration := time.Since(start)
			fmt.Printf("Check completed in %v\n", duration)

			if len(errs) > 0 {
				fmt.Printf("%d hosts had errors:\n", len(errs))
				for url, err := range errs {
					fmt.Printf("  %s: %v\n", url, err)
				}
			}

			printStatuses()
		}
	}
}

func printStatuses() {
	hostsMu.RLock()
	defer hostsMu.RUnlock()

	alive := 0
	dead := 0
	for _, status := range hosts {
		if status {
			alive++
		} else {
			dead++
		}
	}
	fmt.Printf("Status: %d alive, %d dead\n", alive, dead)
}

func AddHost(url string) {
	hostsMu.Lock()
	defer hostsMu.Unlock()
	hosts[url] = true  // По умолчанию считаем alive
}

func GetStatuses() map[string]bool {
	hostsMu.RLock()
	defer hostsMu.RUnlock()

	result := make(map[string]bool, len(hosts))
	for url, status := range hosts {
		result[url] = status
	}
	return result
}

func main() {
	// Добавляем хосты
	AddHost("https://httpbin.org/status/200")
	AddHost("https://httpbin.org/status/404")
	AddHost("https://httpbin.org/delay/3")  // Таймаут
	AddHost("https://google.com")
	AddHost("https://example.com")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем health checker
	go StartHealthChecker(ctx, 15*time.Second)

	// Работаем 1 минуту
	time.Sleep(1 * time.Minute)
}
```

## 💡 Правило Додзё

> **"Health Checker — это разведчик, который регулярно проверяет мосты через реку. Он не ждет, пока путник упадет с моста, а предупреждает заранее."**

## 🛠️ Продвинутая реализация с метриками

```go
type HealthChecker struct {
	mu       sync.RWMutex
	hosts    map[string]*HostStatus
	interval time.Duration
	client   *http.Client
}

type HostStatus struct {
	URL            string
	Alive          bool
	LastCheck      time.Time
	LastError      error
	CheckCount     int
	SuccessCount   int
	FailureCount   int
	AvgResponseTime time.Duration
}

func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		hosts:    make(map[string]*HostStatus),
		interval: interval,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (hc *HealthChecker) AddHost(url string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	hc.hosts[url] = &HostStatus{
		URL:   url,
		Alive: true,
	}
}

func (hc *HealthChecker) Check(ctx context.Context, url string) (bool, time.Duration, error) {
	start := time.Now()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, 0, err
	}
	
	resp, err := hc.client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		return false, duration, err
	}
	defer resp.Body.Close()
	
	alive := resp.StatusCode >= 200 && resp.StatusCode < 300
	return alive, duration, nil
}

func (hc *HealthChecker) CheckAll(ctx context.Context) {
	hc.mu.RLock()
	urls := make([]string, 0, len(hc.hosts))
	for url := range hc.hosts {
		urls = append(urls, url)
	}
	hc.mu.RUnlock()
	
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			alive, duration, err := hc.Check(checkCtx, u)
			
			hc.mu.Lock()
			status := hc.hosts[u]
			status.LastCheck = time.Now()
			status.LastError = err
			status.CheckCount++
			status.Alive = alive
			
			if alive {
				status.SuccessCount++
			} else {
				status.FailureCount++
			}
			
			// Скользящее среднее времени ответа
			status.AvgResponseTime = (status.AvgResponseTime + duration) / 2
			
			hc.mu.Unlock()
		}(url)
	}
	
	wg.Wait()
}

func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()
	
	// Начальная проверка
	hc.CheckAll(ctx)
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.CheckAll(ctx)
		}
	}
}

func (hc *HealthChecker) GetStatus(url string) (*HostStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	status, ok := hc.hosts[url]
	return status, ok
}

func (hc *HealthChecker) GetAliveHosts() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	var alive []string
	for url, status := range hc.hosts {
		if status.Alive {
			alive = append(alive, url)
		}
	}
	return alive
}
```

## 🎓 Реальные примеры

### 1. Load Balancer с Health Checks

```go
type LoadBalancer struct {
	healthChecker *HealthChecker
	current       int
	mu            sync.Mutex
}

func (lb *LoadBalancer) GetNextHost() string {
	hosts := lb.healthChecker.GetAliveHosts()
	if len(hosts) == 0 {
		return ""
	}
	
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	host := hosts[lb.current%len(hosts)]
	lb.current++
	return host
}
```

### 2. HTTP Server с /health endpoint

```go
func (hc *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	type Response struct {
		Status string                  `json:"status"`
		Hosts  map[string]*HostStatus `json:"hosts"`
	}
	
	allAlive := true
	for _, status := range hc.hosts {
		if !status.Alive {
			allAlive = false
			break
		}
	}
	
	statusCode := http.StatusOK
	statusStr := "healthy"
	if !allAlive {
		statusCode = http.StatusServiceUnavailable
		statusStr = "unhealthy"
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Status: statusStr,
		Hosts:  hc.hosts,
	})
}
```

### 3. Circuit Breaker интеграция

```go
type SmartCircuitBreaker struct {
	healthChecker *HealthChecker
	breaker       *CircuitBreaker
}

func (scb *SmartCircuitBreaker) Call(url string, fn func() error) error {
	// Если Health Checker говорит, что хост dead, сразу fail
	if status, ok := scb.healthChecker.GetStatus(url); ok && !status.Alive {
		return errors.New("host is marked as dead by health checker")
	}
	
	return scb.breaker.Call(fn)
}
```

## 🎓 Вопросы для медитации

1. Почему нужен индивидуальный `context.WithTimeout` для каждой проверки?
2. Как выбрать оптимальный `interval` для health checks?
3. Что делать, если все хосты недоступны?
4. Как реализовать exponential backoff для повторных проверок упавших хостов?

## 🔗 Связанные паттерны

- **Load Balancer**: использует Health Checker для выбора alive хостов
- **Circuit Breaker**: комбинируется с Health Checker для защиты
- **Service Discovery**: Health Checker обновляет список доступных сервисов

## 📚 Best Practices

1. **Timeout per check**: не блокировать весь Health Checker из-за одного медленного хоста
2. **Метрики**: success rate, avg response time, last error
3. **Graceful degradation**: возвращать кешированные данные, если все хосты dead
4. **Логирование переходов**: alive → dead, dead → alive
5. **Prometheus /metrics**: `host_alive{host="example.com"} 1`

---

**Задание**: Реализуй Health Checker, который:
1. Проверяет хосты с **разными интервалами** (critical hosts — чаще, non-critical — реже)
2. Использует **exponential backoff** для dead хостов (проверять реже)
3. Отправляет **уведомления** (webhook/email) при изменении статуса
4. Поддерживает **custom health checks** (не только HTTP GET)

Объясни, почему нельзя использовать один глобальный `context.WithTimeout` для всех проверок.
