# 🎫 Модуль 5: Retry with Backoff + Кэширование

**Цель:** Повторять неудачные запросы с умной стратегией задержек.

---

## 📋 Билет

1. **Разминка:** Работа с defer, порядок выполнения
2. **LeetCode:** LRU Cache (LeetCode 146)
3. **Конкурентность:** Retry механизм с backoff для внешнего API
4. **Паттерн:** Экспоненциальный backoff + jitter
5. **SQL:** Кэширование результатов тяжёлых запросов
6. **Проект:** Retry в Adtime для генерации подарков

---

## 🎯 Задачи

### 1. Defer порядок выполнения

```go
// Что выведет?
func test() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    fmt.Println("Start")
}

// Output:
// Start
// 3
// 2
// 1
```

**Правило:** defer = **LIFO** (Last In, First Out)

---

### 2. LeetCode: LRU Cache

**См.** Module 6 (следующий модуль)

---

### 3. Retry с Exponential Backoff

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if i == maxRetries-1 {
            return err
        }
        
        // Exponential backoff: 1s, 2s, 4s, 8s...
        backoff := time.Duration(1<<i) * time.Second
        
        // Jitter (случайная задержка ±25%)
        jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
        time.Sleep(backoff + jitter)
    }
    
    return nil
}
```

**Зачем jitter?**
- Избегаем "thundering herd" (все клиенты повторяют одновременно)
- Распределяем нагрузку

---

### 4. SQL: Кэширование

```sql
-- Материализованное представление (Materialized View)
CREATE MATERIALIZED VIEW expensive_report AS
SELECT 
    user_id,
    COUNT(*) as total_orders,
    SUM(amount) as total_spent
FROM orders
GROUP BY user_id;

-- Обновление кэша
REFRESH MATERIALIZED VIEW CONCURRENTLY expensive_report;
```

---

**Стратегии:**
- **Exponential Backoff** — удвоение задержки
- **Jitter** — случайная добавка
- **Max Retries** — лимит попыток

Удачи! 🔄
