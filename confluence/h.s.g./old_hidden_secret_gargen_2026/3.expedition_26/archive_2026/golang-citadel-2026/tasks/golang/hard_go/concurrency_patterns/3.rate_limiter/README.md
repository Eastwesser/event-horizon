# Паттерн: Rate Limiter — Ограничение частоты запросов

## 🎯 Задача

Ограничить **частоту выполнения операций** до N операций в единицу времени (обычно в секунду).

**Отличие от Worker Pool**: Worker Pool ограничивает **параллелизм** (сколько горутин), Rate Limiter ограничивает **throughput** (сколько операций в секунду).

## 🗡️ Классическая задача

> Telegram Bot API позволяет отправлять не более **30 сообщений в секунду**. Как обработать 1000 сообщений без бана?

## 🔍 Способ 1: Token Bucket (time.Tick)

```go
func main() {
	limiter := time.Tick(100 * time.Millisecond)  // 10 req/sec

	requests := generateRequests(100)
	
	for _, req := range requests {
		<-limiter  // Ждем токен
		go processRequest(req)
	}
}
```

**Как работает**:
- Каждые 100ms канал `limiter` получает новое значение
- `<-limiter` блокируется, пока не пройдет 100ms
- Результат: **ровно 10 запросов в секунду**

## 🔍 Способ 2: golang.org/x/time/rate (Token Bucket Algorithm)

```go
import "golang.org/x/time/rate"

func main() {
	limiter := rate.NewLimiter(10, 1)  // 10 req/s, burst=1

	for _, req := range requests {
		if err := limiter.Wait(context.Background()); err != nil {
			log.Fatal(err)
		}
		processRequest(req)
	}
}
```

**Параметры**:
- `rate.Limit(10)` — 10 запросов в секунду
- `burst=1` — размер "взрыва" (сколько запросов можно сделать сразу)

### Методы:
- `Wait(ctx)` — блокируется до получения токена
- `Allow()` — возвращает `true`, если токен доступен (non-blocking)
- `Reserve()` — резервирует токен на будущее

## 🔍 Способ 3: Sliding Window (кастомная реализация)

```go
type RateLimiter struct {
	mu       sync.Mutex
	requests []time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make([]time.Time, 0, limit),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Удаляем старые запросы
	var i int
	for i = 0; i < len(rl.requests); i++ {
		if rl.requests[i].After(cutoff) {
			break
		}
	}
	rl.requests = rl.requests[i:]

	// Проверяем лимит
	if len(rl.requests) >= rl.limit {
		return false
	}

	rl.requests = append(rl.requests, now)
	return true
}

func main() {
	limiter := NewRateLimiter(10, time.Second)

	for _, req := range requests {
		for !limiter.Allow() {
			time.Sleep(10 * time.Millisecond)
		}
		processRequest(req)
	}
}
```

**Преимущества**:
- Точный подсчет запросов за последнюю секунду
- Нет "всплесков" на границе секунд

## 💡 Правило Додзё

> **"Rate Limiter — это песочные часы. Token Bucket — это ведро с дыркой: вода (токены) капает с постоянной скоростью, но ведро может накопить несколько капель (burst)."**

## 🛠️ Сравнение алгоритмов

| Алгоритм | Простота | Точность | Burst | Память |
|----------|----------|----------|-------|--------|
| time.Tick | Высокая | Средняя | Нет | O(1) |
| Token Bucket | Средняя | Высокая | Да | O(1) |
| Sliding Window | Низкая | Максимальная | Нет | O(N) |
| Fixed Window | Высокая | Низкая | Да | O(1) |

## 🧪 Эксперименты

### Тест 1: time.Tick — равномерное распределение

```go
func main() {
	limiter := time.Tick(200 * time.Millisecond)
	start := time.Now()

	for i := 0; i < 10; i++ {
		<-limiter
		fmt.Printf("[%v] Request %d\n", time.Since(start), i)
	}
}
```

**Вывод**:
```
[200ms] Request 0
[400ms] Request 1
[600ms] Request 2
...
```

### Тест 2: rate.Limiter с burst

```go
func main() {
	limiter := rate.NewLimiter(5, 10)  // 5 req/s, burst=10
	start := time.Now()

	for i := 0; i < 20; i++ {
		limiter.Wait(context.Background())
		fmt.Printf("[%v] Request %d\n", time.Since(start), i)
	}
}
```

**Вывод**:
```
[0s] Request 0-9     ← burst (10 сразу)
[200ms] Request 10   ← затем по 5 req/s
[400ms] Request 11
...
```

### Тест 3: Allow() vs Wait()

```go
func main() {
	limiter := rate.NewLimiter(10, 1)

	// Allow() — non-blocking
	for i := 0; i < 100; i++ {
		if limiter.Allow() {
			processRequest(i)
		} else {
			fmt.Printf("Request %d dropped\n", i)
		}
	}

	// Wait() — blocking
	for i := 0; i < 100; i++ {
		limiter.Wait(context.Background())
		processRequest(i)
	}
}
```

**Разница**:
- `Allow()` — отбрасывает лишние запросы (DDoS protection)
- `Wait()` — ждет, пока не появится токен (гарантированная доставка)

## 🎓 Реальные примеры

### 1. Telegram Bot (30 msg/s)

```go
limiter := rate.NewLimiter(30, 5)  // 30 msg/s, burst=5

func sendMessage(chatID int, text string) error {
	if err := limiter.Wait(context.Background()); err != nil {
		return err
	}
	return bot.Send(chatID, text)
}
```

### 2. GitHub API (5000 req/hour)

```go
limiter := rate.NewLimiter(rate.Every(720*time.Millisecond), 1)  // 5000/3600

func fetchRepo(owner, repo string) (*Repo, error) {
	limiter.Wait(context.Background())
	return github.GetRepo(owner, repo)
}
```

### 3. DDoS Protection (100 req/s per IP)

```go
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(100, 10)
		rl.limiters[ip] = limiter
	}
	return limiter
}

func middleware(next http.Handler) http.Handler {
	limiter := NewIPRateLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if !limiter.GetLimiter(ip).Allow() {
			http.Error(w, "Too Many Requests", 429)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

## 🎓 Вопросы для медитации

1. Чем отличается `rate.NewLimiter(10, 1)` от `rate.NewLimiter(10, 100)`?
2. Что произойдет, если вызвать `limiter.Allow()` 1000 раз подряд?
3. Как реализовать Rate Limiter для разных типов запросов (GET=10/s, POST=5/s)?
4. Можно ли использовать `time.Tick` для ограничения запросов к БД?

## 🔗 Связанные паттерны

- **Worker Pool**: ограничивает параллелизм, Rate Limiter — throughput
- **Circuit Breaker**: защита от перегрузки сервиса
- **Token Bucket**: основа для многих rate limiters

---

**Задание**: Реализуй Rate Limiter, который:
1. Ограничивает запросы до 100 req/s
2. Поддерживает burst=20 (разрешает 20 запросов сразу)
3. Логирует отклоненные запросы
4. Работает с контекстом (context.Context) для отмены

Объясни, почему нельзя использовать `time.Sleep(10 * time.Millisecond)` между запросами для достижения 100 req/s.
