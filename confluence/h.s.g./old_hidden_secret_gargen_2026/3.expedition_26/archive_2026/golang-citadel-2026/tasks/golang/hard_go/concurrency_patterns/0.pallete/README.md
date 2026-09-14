# 🎨 Палитра Паттернов Конкурентности — Дорожная Карта

## 🗡️ Философия

Конкурентность в Go — это не про скорость. Это про **структуру программы** и **координацию независимых задач**.

> _"Concurrency is about dealing with lots of things at once. Parallelism is about doing lots of things at once."_ — Rob Pike

## 📚 Каталог Паттернов

### 1️⃣ **WaitGroup** — Синхронизация горутин
**Задача**: Дождаться завершения N горутин

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        work(id)
    }(i)
}
wg.Wait()
```

**Когда использовать**: 
- Параллельная обработка массива
- Fan-out задач без ограничений
- Простая координация без результатов

---

### 2️⃣ **Worker Pool** — Ограниченный параллелизм
**Задача**: Обработать N задач с M воркерами (M < N)

```go
jobs := make(chan Job, 100)
var wg sync.WaitGroup

// Запускаем 5 воркеров
for i := 0; i < 5; i++ {
    wg.Add(1)
    go worker(jobs, &wg)
}

// Отправляем задачи
for _, job := range tasks {
    jobs <- job
}
close(jobs)

wg.Wait()
```

**Когда использовать**:
- Ограничение нагрузки на ресурсы (БД, API)
- Контроль использования CPU/памяти
- Очереди задач

---

### 3️⃣ **Rate Limiter** — Ограничение частоты запросов
**Задача**: Не более X запросов в секунду

```go
limiter := time.Tick(100 * time.Millisecond)  // 10 req/sec

for req := range requests {
    <-limiter  // Ждем разрешения
    processRequest(req)
}
```

**Когда использовать**:
- API с лимитами (GitHub, Telegram)
- Защита от DDoS
- Контроль нагрузки на внешние сервисы

---

### 4️⃣ **Load Balancer** — Распределение нагрузки
**Задача**: Распределить N задач между M воркерами равномерно

```go
workers := []chan Job{
    make(chan Job),
    make(chan Job),
    make(chan Job),
}

for i, job := range jobs {
    workers[i%len(workers)] <- job  // Round-robin
}
```

**Когда использовать**:
- Распределение между серверами
- Балансировка между горутинами
- Оптимизация нагрузки

---

### 5️⃣ **Circuit Breaker** — Защита от каскадных сбоев
**Задача**: Прекратить запросы к упавшему сервису

**Состояния**:
- **Closed**: Нормальная работа
- **Open**: Сервис недоступен, возвращаем ошибку сразу
- **Half-Open**: Пробуем восстановить связь

```go
if cb.State == Open {
    return ErrServiceUnavailable
}

err := doRequest()
if err != nil {
    cb.RecordFailure()
} else {
    cb.RecordSuccess()
}
```

**Когда использовать**:
- Микросервисы с сетевыми вызовами
- Защита от зависших сервисов
- Retry с умом

---

### 6️⃣ **Cacher** — Конкурентный кеш
**Задача**: Потокобезопасный доступ к кешу

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}
```

**Когда использовать**:
- In-memory кеш для API
- Дедупликация запросов
- Ленивая инициализация (`sync.Once`)

---

### 7️⃣ **Retryer** — Повторные попытки
**Задача**: Повторить запрос N раз с backoff

```go
for i := 0; i < maxRetries; i++ {
    err := doRequest()
    if err == nil {
        return nil
    }
    
    if !isRetryable(err) {
        return err
    }
    
    time.Sleep(backoff(i))
}
```

**Когда использовать**:
- Временные сетевые ошибки
- Rate-limit errors (429)
- Eventual consistency

---

### 8️⃣ **Health Checker** — Проверка состояния сервисов
**Задача**: Периодически проверять доступность N сервисов

```go
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        checkAllServices()
    case <-ctx.Done():
        return
    }
}
```

**Когда использовать**:
- Мониторинг микросервисов
- Prometheus /healthz endpoint
- Service discovery

---

## 🧭 Как выбрать паттерн?

| Задача | Паттерн |
|--------|---------|
| Обработать список параллельно | **WaitGroup** |
| Обработать список с ограничением | **Worker Pool** |
| Защитить внешний API от перегрузки | **Rate Limiter** |
| Распределить задачи равномерно | **Load Balancer** |
| Защититься от упавшего сервиса | **Circuit Breaker** |
| Кешировать результаты | **Cacher** |
| Повторить упавший запрос | **Retryer** |
| Следить за состоянием сервисов | **Health Checker** |

## 🎓 Комбинации паттернов

### Worker Pool + Rate Limiter
```go
// 5 воркеров, но каждый ограничен rate limiter
for i := 0; i < 5; i++ {
    go worker(jobs, rateLimiter)
}
```

### Circuit Breaker + Retryer
```go
// Retry только если Circuit Breaker не открыт
if cb.State == Open {
    return ErrServiceUnavailable
}
return retryWithBackoff(doRequest, 3)
```

### Cacher + WaitGroup
```go
// Параллельная прогрузка кеша
var wg sync.WaitGroup
for key := range keys {
    wg.Add(1)
    go func(k string) {
        defer wg.Done()
        cache.Set(k, fetchData(k))
    }(key)
}
wg.Wait()
```

## 💡 Правило Додзё

> **"Паттерн — это не код, который ты копируешь. Это решение проблемы, которое ты понимаешь."**

---

**Задание**: Прочитай каждую папку паттернов. Реализуй один из них **с нуля**, не глядя на код. Объясни, какую проблему он решает и где бы ты его применил.
