# 🎫 Модуль 3: Circuit Breaker + Обработка ошибок

**Цель:** Защитить систему от каскадных отказов при проблемах с внешними сервисами.

---

## 📋 Билет

1. **Разминка:** Паника и recover
2. **LeetCode:** Merge Two Sorted Lists (LeetCode 21)
3. **Конкурентность:** Graceful shutdown с контекстом
4. **Паттерн:** Circuit Breaker (Closed/Open/Half-Open)
5. **SQL:** Анализ ошибок запросов по времени
6. **Проект:** Circuit breaker в Roolz для внешних API поставщиков

---

## 🎯 Задачи

### 1. Разминка: Panic & Recover

**Вопросы:**
```go
// Что выведет?
func test() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered:", r)
        }
    }()
    
    panic("boom!")
    fmt.Println("After panic")  // Не выполнится!
}
```

**Правила:**
- `panic` → разворачивает стек
- `recover` работает только в `defer`
- `recover` в другой горутине **не поймает** панику!

---

### 2. LeetCode: Merge Two Sorted Lists

**Задача:** [LeetCode 21](https://leetcode.com/problems/merge-two-sorted-lists/)

```go
func mergeTwoLists(l1, l2 *ListNode) *ListNode {
    dummy := &ListNode{}
    current := dummy
    
    for l1 != nil && l2 != nil {
        if l1.Val < l2.Val {
            current.Next = l1
            l1 = l1.Next
        } else {
            current.Next = l2
            l2 = l2.Next
        }
        current = current.Next
    }
    
    if l1 != nil {
        current.Next = l1
    }
    if l2 != nil {
        current.Next = l2
    }
    
    return dummy.Next
}
```

---

### 3. Graceful Shutdown с Context

**См.:** `6.six-stickers/2.sema_context/`

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

<-ctx.Done()  // Ждём Ctrl+C
log.Println("Shutting down...")
```

---

### 4. Circuit Breaker Pattern

**3 состояния:**
- **Closed** (закрыт) → запросы проходят
- **Open** (открыт) → запросы блокируются
- **Half-Open** (полуоткрыт) → тестовые запросы

**Реализация:**
```go
type CircuitBreaker struct {
    state         State  // Closed/Open/Half-Open
    failureCount  int
    successCount  int
    threshold     int
    timeout       time.Duration
    lastFailTime  time.Time
    mu            sync.Mutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case Open:
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.state = HalfOpen
            cb.successCount = 0
        } else {
            return ErrCircuitOpen
        }
    }
    
    err := fn()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailTime = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = Open
        }
        return err
    }
    
    // Success
    cb.failureCount = 0
    
    if cb.state == HalfOpen {
        cb.successCount++
        if cb.successCount >= 3 {
            cb.state = Closed
        }
    }
    
    return nil
}
```

---

### 5. SQL: Анализ ошибок

```sql
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    COUNT(*) FILTER (WHERE status >= 500) as errors,
    COUNT(*) as total,
    (COUNT(*) FILTER (WHERE status >= 500)::float / COUNT(*)) * 100 as error_rate
FROM requests
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY error_rate DESC;
```

---

### 6. Проект: Roolz API

**Где применить:**
- API поставщиков (защита от перегрузки)
- Платёжные системы (защита от каскадных отказов)
- Внутренние микросервисы

**Библиотеки:**
- `github.com/sony/gobreaker`
- Custom implementation

Удачи! ⚡
