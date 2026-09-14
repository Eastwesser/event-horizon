# 🎫 Модуль 4: Health Checker + Мониторинг

**Цель:** Мониторить здоровье сервисов и быстро реагировать на проблемы.

---

## 📋 Билет

1. **Разминка:** Селекты на каналах с таймаутами
2. **LeetCode:** Linked List Cycle (LeetCode 141)
3. **Конкурентность:** Периодические health checks нескольких сервисов
4. **Паттерн:** Health Checker с экспоненциальным backoff
5. **SQL:** SLA расчёты (uptime, error rate)
6. **Проект:** Мониторинг в Adtime (Prometheus, Grafana)

---

## 🎯 Задачи

### 1. Select с таймаутом

```go
select {
case result := <-ch:
    fmt.Println("Got result:", result)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout!")
}
```

---

### 2. LeetCode: Linked List Cycle

```go
func hasCycle(head *ListNode) bool {
    slow, fast := head, head
    
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
        
        if slow == fast {
            return true
        }
    }
    
    return false
}
```

---

### 3. Health Checker Pattern

```go
type HealthChecker struct {
    url      string
    interval time.Duration
    timeout  time.Duration
    status   bool
    mu       sync.RWMutex
}

func (hc *HealthChecker) Start(ctx context.Context) {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            hc.check()
        case <-ctx.Done():
            return
        }
    }
}

func (hc *HealthChecker) check() {
    ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "GET", hc.url, nil)
    resp, err := http.DefaultClient.Do(req)
    
    hc.mu.Lock()
    hc.status = (err == nil && resp.StatusCode == 200)
    hc.mu.Unlock()
}
```

---

### 4. SQL: SLA расчёты

```sql
-- Uptime за последние 24 часа
SELECT 
    service_name,
    COUNT(*) FILTER (WHERE status = 'up') * 100.0 / COUNT(*) as uptime_percent,
    AVG(response_time_ms) as avg_response_time
FROM health_checks
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY service_name;
```

---

**См. также:** Prometheus, Grafana для мониторинга

Удачи! 📊
