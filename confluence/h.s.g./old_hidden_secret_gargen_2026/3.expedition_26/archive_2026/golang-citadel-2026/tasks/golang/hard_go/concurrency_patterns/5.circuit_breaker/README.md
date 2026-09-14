# Паттерн: Circuit Breaker — Защита от каскадных сбоев

## 🎯 Задача

Прекратить попытки обращения к **упавшему сервису**, чтобы избежать накопления ошибок и таймаутов. Периодически проверять восстановление.

## 🗡️ Аналогия

Представь **автоматический выключатель** в электросети:
- **Closed** (Закрыт) — электричество течет (запросы проходят)
- **Open** (Открыт) — выключатель сработал (запросы блокируются)
- **Half-Open** (Полуоткрыт) — пробуем включить (тестовые запросы)

## 🔍 Состояния Circuit Breaker

```
           SUCCESS
    ┌──────────────────┐
    │                  │
    │  ┌───────────┐   │
    └─▶│  CLOSED   │◀──┼─────────┐
       └─────┬─────┘   │         │
             │         │         │
        FAILURE        │      SUCCESS
        THRESHOLD      │         │
             │         │         │
             ▼         │         │
       ┌───────────┐   │    ┌────────────┐
       │   OPEN    │───┼───▶│ HALF-OPEN  │
       └───────────┘   │    └────────────┘
             │         │         │
        TIMEOUT    FAILURE   FAILURE
             │                  │
             └──────────────────┘
```

### 1. **CLOSED** (Нормальная работа)
- Все запросы проходят
- Счетчик ошибок: `failures < threshold`
- При достижении порога → **OPEN**

### 2. **OPEN** (Сервис недоступен)
- Запросы **не выполняются**, сразу возвращается ошибка
- Экономим ресурсы (нет таймаутов)
- После `timeout` → **HALF-OPEN**

### 3. **HALF-OPEN** (Проверка восстановления)
- Пропускаем **тестовые запросы** (1-3 шт)
- Если успех → **CLOSED**
- Если ошибка → **OPEN**

## 💡 Правило Додзё

> **"Circuit Breaker — это мудрый самурай, который не атакует упавшего врага. Он ждет, пока враг встанет, и только тогда возобновляет бой."**

## 🧪 Базовая реализация

```go
package main

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	timeout      time.Duration
	lastFailTime time.Time
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Проверка таймаута для перехода в Half-Open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	// Вызов функции
	err := fn()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 2 {  // 2 успешных запроса → CLOSED
			cb.state = StateClosed
			cb.failures = 0
		}
	} else {
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
```

## 🛠️ Использование

```go
func main() {
	cb := NewCircuitBreaker(3, 5*time.Second)

	for i := 0; i < 10; i++ {
		err := cb.Call(func() error {
			return callExternalService()
		})

		if err != nil {
			fmt.Printf("Request %d failed: %v (state: %v)\n", i, err, cb.State())
		} else {
			fmt.Printf("Request %d success\n", i)
		}

		time.Sleep(1 * time.Second)
	}
}

func callExternalService() error {
	// Имитация сервиса
	if time.Now().Unix()%2 == 0 {
		return errors.New("service unavailable")
	}
	return nil
}
```

**Сценарий**:
```
Request 0: fail (state: CLOSED, failures: 1/3)
Request 1: fail (state: CLOSED, failures: 2/3)
Request 2: fail (state: OPEN,   failures: 3/3) ← Circuit открыт
Request 3: fail (state: OPEN) ← Запрос не выполняется
Request 4: fail (state: OPEN)
Request 5: fail (state: OPEN)
... через 5 секунд ...
Request 6: success (state: HALF-OPEN, successes: 1/2)
Request 7: success (state: CLOSED, successes: 2/2) ← Восстановлено
```

## 🧪 Продвинутая реализация (с метриками)

```go
type CircuitBreakerMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	RejectedRequests int64
}

type AdvancedCircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	timeout      time.Duration
	lastFailTime time.Time
	metrics      CircuitBreakerMetrics
}

func (cb *AdvancedCircuitBreaker) Call(fn func() error) error {
	atomic.AddInt64(&cb.metrics.TotalRequests, 1)

	cb.mu.RLock()
	state := cb.state
	lastFail := cb.lastFailTime
	cb.mu.RUnlock()

	// Fast path: reject if OPEN
	if state == StateOpen && time.Since(lastFail) <= cb.timeout {
		atomic.AddInt64(&cb.metrics.RejectedRequests, 1)
		return ErrCircuitOpen
	}

	// Execute
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		atomic.AddInt64(&cb.metrics.FailedRequests, 1)
		cb.onFailure()
		return err
	}

	atomic.AddInt64(&cb.metrics.SuccessRequests, 1)
	cb.onSuccess()
	return nil
}

func (cb *AdvancedCircuitBreaker) Metrics() CircuitBreakerMetrics {
	return CircuitBreakerMetrics{
		TotalRequests:    atomic.LoadInt64(&cb.metrics.TotalRequests),
		SuccessRequests:  atomic.LoadInt64(&cb.metrics.SuccessRequests),
		FailedRequests:   atomic.LoadInt64(&cb.metrics.FailedRequests),
		RejectedRequests: atomic.LoadInt64(&cb.metrics.RejectedRequests),
	}
}
```

## 🎓 Реальные примеры

### 1. HTTP Client с Circuit Breaker

```go
type ResilientHTTPClient struct {
	client *http.Client
	cb     *CircuitBreaker
}

func (c *ResilientHTTPClient) Get(url string) (*http.Response, error) {
	var resp *http.Response
	
	err := c.cb.Call(func() error {
		var err error
		resp, err = c.client.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}
		return nil
	})

	return resp, err
}
```

### 2. Database с Circuit Breaker

```go
type ResilientDB struct {
	db *sql.DB
	cb *CircuitBreaker
}

func (rdb *ResilientDB) Query(query string) (*sql.Rows, error) {
	var rows *sql.Rows
	
	err := rdb.cb.Call(func() error {
		var err error
		rows, err = rdb.db.Query(query)
		return err
	})

	return rows, err
}
```

### 3. gRPC с Circuit Breaker

```go
type ResilientGRPCClient struct {
	client pb.ServiceClient
	cb     *CircuitBreaker
}

func (c *ResilientGRPCClient) Process(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	var resp *pb.Response
	
	err := c.cb.Call(func() error {
		var err error
		resp, err = c.client.Process(ctx, req)
		return err
	})

	return resp, err
}
```

## 🎓 Вопросы для медитации

1. Почему Circuit Breaker должен быть **per-service**, а не глобальным?
2. Как выбрать оптимальный `maxFailures` и `timeout`?
3. Что произойдет, если Half-Open пропустит только 1 запрос вместо 2?
4. Как комбинировать Circuit Breaker с Retry Pattern?

## 🔗 Связанные паттерны

- **Retry**: Circuit Breaker предотвращает бесполезные retry
- **Timeout**: Circuit Breaker дополняет таймауты
- **Bulkhead**: изоляция ресурсов для каждого сервиса

## 📚 Best Practices

1. **Один Circuit Breaker на сервис** (не на метод)
2. **Логирование переходов состояний** для мониторинга
3. **Метрики в Prometheus** (`circuit_breaker_state`, `circuit_breaker_requests_total`)
4. **Настраиваемые параметры** через конфиг (не хардкод)
5. **Graceful degradation** (возвращать кешированные данные)

---

**Задание**: Реализуй Circuit Breaker, который:
1. Поддерживает **sliding window** (считает ошибки за последние N секунд, а не все подряд)
2. Отправляет метрики в Prometheus
3. Логирует каждый переход состояния
4. Работает с `context.Context` для отмены

Объясни, почему нельзя просто использовать Retry вместо Circuit Breaker.
