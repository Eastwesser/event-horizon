# Паттерн: Retryer — Повторные попытки с backoff

## 🎯 Задача

Повторить упавший запрос **N раз** с **экспоненциальной задержкой** (backoff), чтобы дать сервису время восстановиться.

## 🗡️ Типы ошибок

Не все ошибки стоит retry:

| Ошибка | Retry? | Причина |
|--------|--------|---------|
| `connection timeout` | ✅ | Временная сетевая проблема |
| `503 Service Unavailable` | ✅ | Сервис перегружен |
| `429 Too Many Requests` | ✅ | Rate limit (с увеличенным backoff) |
| `500 Internal Server Error` | ✅ | Может быть временным |
| `404 Not Found` | ❌ | Ресурс не существует |
| `400 Bad Request` | ❌ | Неправильный запрос |
| `401 Unauthorized` | ❌ | Неправильные credentials |

## 🔍 Стратегии Backoff

### 1. Fixed Delay (фиксированная задержка)

```go
func retryFixed(fn func() error, maxAttempts int, delay time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		
		if !isRetryable(err) {
			return err
		}
		
		if i < maxAttempts-1 {
			time.Sleep(delay)  // Всегда одинаковая задержка
		}
	}
	return errors.New("max retries exceeded")
}

// Использование
retryFixed(func() error {
	return callAPI()
}, 3, 1*time.Second)
```

**Проблема**: Не дает сервису время восстановиться при высокой нагрузке.

### 2. Exponential Backoff (экспоненциальная задержка)

```go
func retryExponential(fn func() error, maxAttempts int, baseDelay time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		
		if !isRetryable(err) {
			return err
		}
		
		if i < maxAttempts-1 {
			delay := baseDelay * time.Duration(1<<i)  // 1s, 2s, 4s, 8s, ...
			time.Sleep(delay)
		}
	}
	return errors.New("max retries exceeded")
}

// Использование
retryExponential(func() error {
	return callAPI()
}, 5, 1*time.Second)
// Попытки: 0s → 1s → 2s → 4s → 8s
```

**Преимущества**:
- Быстрая первая попытка
- Дает время на восстановление при повторных ошибках

### 3. Exponential Backoff + Jitter (случайность)

```go
func retryWithJitter(fn func() error, maxAttempts int, baseDelay, maxDelay time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		
		if !isRetryable(err) {
			return err
		}
		
		if i < maxAttempts-1 {
			// Exponential backoff
			delay := baseDelay * time.Duration(1<<i)
			
			// Cap max delay
			if delay > maxDelay {
				delay = maxDelay
			}
			
			// Add jitter (±25%)
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			delay = delay - delay/4 + jitter
			
			time.Sleep(delay)
		}
	}
	return errors.New("max retries exceeded")
}

// Использование
retryWithJitter(func() error {
	return callAPI()
}, 5, 1*time.Second, 30*time.Second)
```

**Зачем jitter?**
- Предотвращает **thundering herd** (одновременные retry от всех клиентов)
- Распределяет нагрузку во времени

## 💡 Правило Додзё

> **"Retry — это мудрость самурая: после падения он не встает мгновенно, а ждет, давая противнику успокоиться. С каждым падением пауза увеличивается."**

## 🧪 Полная реализация

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   1 * time.Second,
	MaxDelay:    30 * time.Second,
	Multiplier:  2.0,
	Jitter:      true,
}

func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Проверка контекста
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Выполнение функции
		lastErr = fn()
		
		// Успех
		if lastErr == nil {
			if attempt > 0 {
				fmt.Printf("Retry succeeded on attempt %d\n", attempt+1)
			}
			return nil
		}
		
		// Проверка retryable
		if !isRetryable(lastErr) {
			return lastErr
		}
		
		// Последняя попытка
		if attempt == config.MaxAttempts-1 {
			break
		}
		
		// Backoff
		delay := calculateBackoff(attempt, config)
		fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt+1, lastErr, delay)
		
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.BaseDelay) * (config.Multiplier * float64(attempt))
	
	// Cap max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	// Add jitter
	if config.Jitter {
		jitterRange := delay * 0.25  // ±25%
		jitter := (rand.Float64()*2 - 1) * jitterRange
		delay += jitter
	}
	
	return time.Duration(delay)
}

func isRetryable(err error) bool {
	// Temporary errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	// HTTP errors
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}
	
	return false
}

type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}
```

## 🛠️ Использование

```go
func main() {
	ctx := context.Background()
	
	err := Retry(ctx, DefaultRetryConfig, func() error {
		return callAPI()
	})
	
	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	}
}

func callAPI() error {
	resp, err := http.Get("https://httpbin.org/status/503")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 500 {
		return &HTTPError{StatusCode: resp.StatusCode}
	}
	
	return nil
}
```

## 🎓 Реальные примеры

### 1. HTTP Client с Retry

```go
type ResilientHTTPClient struct {
	client *http.Client
}

func (c *ResilientHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	var resp *http.Response
	
	err := Retry(ctx, DefaultRetryConfig, func() error {
		var err error
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err = c.client.Do(req)
		
		if err != nil {
			return err
		}
		
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			return &HTTPError{StatusCode: resp.StatusCode}
		}
		
		return nil
	})
	
	return resp, err
}
```

### 2. Database Retry (deadlock, connection loss)

```go
func queryWithRetry(db *sql.DB, query string) (*sql.Rows, error) {
	var rows *sql.Rows
	
	err := Retry(context.Background(), RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}, func() error {
		var err error
		rows, err = db.Query(query)
		
		// Retry on connection errors or deadlocks
		if err != nil && (isDeadlock(err) || isConnError(err)) {
			return err
		}
		
		return nil
	})
	
	return rows, err
}
```

### 3. gRPC Retry (с unary interceptor)

```go
func RetryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return Retry(ctx, DefaultRetryConfig, func() error {
			err := invoker(ctx, method, req, reply, cc, opts...)
			
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded {
					return err  // Retryable
				}
			}
			
			return err
		})
	}
}
```

## 🎓 Вопросы для медитации

1. Почему нельзя retry 401 Unauthorized?
2. Что такое thundering herd и как jitter помогает?
3. Как выбрать оптимальный `maxAttempts` для API с SLA 99.9%?
4. Чем отличается retry от circuit breaker?

## 🔗 Связанные паттерны

- **Circuit Breaker**: Retry может усугубить проблему (накопление таймаутов)
- **Timeout**: Каждая попытка должна иметь свой timeout
- **Idempotency**: Безопасные retry только для идемпотентных операций

## 📚 Best Practices

1. **Не retry всё подряд**: проверяй `isRetryable(err)`
2. **Используй context**: `context.WithTimeout` для общего лимита времени
3. **Exponential backoff + jitter**: избегай thundering herd
4. **Логируй попытки**: для debugging и мониторинга
5. **Идемпотентность**: GET/PUT безопасны, POST — нет (без идемпотентного ключа)

---

**Задание**: Реализуй Retryer, который:
1. Поддерживает **per-error backoff** (429 → больший delay, 503 → обычный)
2. Логирует каждую попытку с метриками (attempt, delay, error)
3. Работает с `context.Context` и прерывается при `ctx.Done()`
4. Возвращает **wrapped error** с количеством попыток

Объясни, почему retry POST-запросов без идемпотентного ключа опасен.
