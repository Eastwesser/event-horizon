# 🔍 Deep Dive: Альтернативные решения и Trade-offs

---

## 1. Выбор БД: Альтернативы

### PostgreSQL vs MySQL

| Критерий | PostgreSQL | MySQL |
|----------|------------|-------|
| JSON поддержка | ✅ JSONB (индексируемый) | ⚠️ JSON (без индексов до 8.0) |
| Партиционирование | ✅ Нативное (RANGE, LIST, HASH) | ⚠️ Ограниченное |
| Full-text search | ✅ Встроенный (tsvector) | ⚠️ Слабый |
| Replication | ✅ Streaming (async/sync) | ✅ GTID |
| Window Functions | ✅ Полная поддержка | ⚠️ С 8.0 |
| Concurrent writes | ✅ MVCC (без блокировок) | ⚠️ Table locks |

**Вывод:** PostgreSQL лучше для сложных запросов и аналитики.

### Cassandra vs MongoDB

| Критерий | Cassandra | MongoDB |
|----------|-----------|---------|
| Модель данных | Wide-column | Document |
| Масштабирование | ✅ Линейное | ⚠️ Sharding сложнее |
| Consistency | Eventual (tunable) | Strong (по умолчанию) |
| Query flexibility | ⚠️ Только по Partition Key | ✅ Гибкие запросы |
| TTL | ✅ Встроенный | ✅ Есть |
| Операционная сложность | ⚠️ Высокая | ✅ Проще |

**Вывод:** Cassandra для write-heavy архива, MongoDB если нужны гибкие запросы.

---

## 2. WebSocket: Альтернативы

### WebSocket vs Server-Sent Events (SSE)

| | WebSocket | SSE |
|-|-----------|-----|
| Двунаправленность | ✅ Да | ❌ Только Server → Client |
| Протокол | ws:// (отдельный) | ✅ HTTP/HTTPS |
| Firewall/Proxy | ⚠️ Может блокироваться | ✅ Работает везде |
| Reconnect | ⚠️ Вручную | ✅ Автоматический |
| Binary data | ✅ Да | ❌ Только текст |
| Use case | Chat, gaming | Notifications, live updates |

**Выбор:** WebSocket, т.к. нужна двунаправленность (клиент отправляет сообщения).

### Long Polling vs WebSocket

**Long Polling:**
```
Client → Server (HTTP GET /messages?last_id=123)
         ↓ (server holds connection until new message)
Client ← Server (response with new messages)
Client → Server (next poll)
```

**Проблема:** Неэффективно (много HTTP overhead).

**WebSocket:**
```
Client → Server (HTTP Upgrade to WebSocket)
Client ↔ Server (persistent bidirectional connection)
```

**Вывод:** WebSocket эффективнее для real-time.

---

## 3. Message Queue: Kafka vs RabbitMQ

| Критерий | Kafka | RabbitMQ |
|----------|-------|----------|
| Throughput | ✅ Millions msg/s | ⚠️ Tens of thousands msg/s |
| Persistence | ✅ Да (disk log) | ⚠️ Опционально |
| Message ordering | ✅ По partition | ⚠️ По queue (сложнее) |
| Consumer groups | ✅ Нативно | ⚠️ Через exchanges |
| Latency | ⚠️ Выше (batch) | ✅ Ниже |
| Use case | Event streaming, logs | Task queues, RPC |

**Выбор:** Kafka для event-driven (audit log, реплики, множество consumers).

**Альтернатива:** RabbitMQ + PostgreSQL для transactional outbox.

### Transactional Outbox Pattern

**Проблема:** Как гарантировать, что сообщение попало и в БД, и в Kafka?

**Решение:**
```sql
-- В одной транзакции
BEGIN;
INSERT INTO messages (...);
INSERT INTO outbox_events (event_type, payload, created_at);
COMMIT;

-- Отдельный worker читает outbox и публикует в Kafka
SELECT * FROM outbox_events WHERE published_at IS NULL ORDER BY created_at LIMIT 100;
-- Публикуем в Kafka
-- UPDATE outbox_events SET published_at = NOW() WHERE id IN (...);
```

**Trade-off:** Сложность vs гарантия доставки.

---

## 4. Sharding: Стратегии

### 4.1 Sharding по `user_id`

**Плюсы:**
- ✅ Равномерное распределение
- ✅ Все чаты пользователя на одном шарде

**Минусы:**
- ❌ Запрос "история чата" → может быть на 2 шардах (если собеседник на другом)

### 4.2 Sharding по `chat_id`

**Плюсы:**
- ✅ Все сообщения чата на одном шарде
- ✅ Эффективные запросы истории

**Минусы:**
- ❌ Запрос "все мои чаты" → нужен broadcast на все шарды

### 4.3 Hybrid: Sharding + Routing Table

**Решение:**
```sql
-- Routing table (маленькая, можно в Redis)
CREATE TABLE chat_shards (
    chat_id UUID PRIMARY KEY,
    shard_id INT NOT NULL
);

-- При создании чата
shard_id = hash(seller_id + buyer_id) % total_shards
INSERT INTO chat_shards (chat_id, shard_id) VALUES (?, ?);

-- При запросе
SELECT shard_id FROM chat_shards WHERE chat_id = ?;
-- Идём на нужный шард
```

**Trade-off:** Дополнительный lookup vs эффективные запросы.

---

## 5. Caching: Многоуровневая стратегия

### Level 1: Client-side (Mobile App)

**SQLite для локального кэша:**
```sql
CREATE TABLE cached_messages (
    message_id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    content TEXT,
    sent_at INTEGER,
    synced INTEGER DEFAULT 0 -- 0 = не синхронизировано с сервером
);
```

**Offline-first:**
- Пользователь отправляет сообщение → сохраняем в SQLite
- Background sync → отправляем на сервер
- При успехе → `synced = 1`

### Level 2: Server-side (Redis)

**Hot cache:**
```
# Последние N сообщений чата
chat:messages:123 -> [msg1, msg2, msg3] (LIST)
EXPIRE 300 (5 минут)

# Unread counts
unread:user:456 -> {chat_123: 5, chat_789: 12} (HASH)

# Online users
online:users -> {user_456, user_789, ...} (SET)
TTL 60 секунд
```

### Level 3: PostgreSQL (L3 Cache)

**Partition pruning:**
```sql
-- Запрос только нужной партиции
SELECT * FROM messages WHERE chat_id = ? AND sent_at > '2025-01-01';
-- PostgreSQL автоматически выберет partition messages_2025_01
```

### Level 4: Cassandra (Cold Storage)

**Компрессия + TTL:**
```cql
CREATE TABLE messages_archive (...)
WITH compression = {'class': 'LZ4Compressor'}
AND default_time_to_live = 63072000;
```

---

## 6. Push Notifications: Проблемы и решения

### 6.1 FCM (Firebase Cloud Messaging) Limits

**Quota:**
- 500 msg/s per project (бесплатно)
- 1,500 msg/s (платный план)

**Проблема:** У нас ~1,000 msg/s → нужно несколько FCM projects.

**Решение:**
```
Notification Service → Round-robin между 3 FCM projects
Project 1: 500 msg/s
Project 2: 500 msg/s
Project 3: 500 msg/s
Total: 1,500 msg/s
```

### 6.2 Silent Push vs Alert Push

**Silent Push:**
- Фоновое обновление (без уведомления)
- Для синхронизации данных

**Alert Push:**
- С уведомлением (звук, баннер)
- Для новых сообщений

**Стратегия:**
```
IF user online (WebSocket connected):
    → Отправляем через WebSocket (не нужен push)
ELSE:
    → Отправляем Alert Push
```

### 6.3 Badge Count

**Проблема:** iOS показывает badge (красный кружок) с количеством непрочитанных.

**Решение:**
```json
{
  "to": "device_token",
  "notification": {
    "title": "New message from Иван",
    "body": "Привет! Когда встретимся?",
    "badge": 5  // Получаем из Redis unread:user:{user_id}
  }
}
```

---

## 7. Monitoring: Детальные метрики

### 7.1 Business Metrics

```promql
# Conversion rate (сколько чатов приводят к сделкам)
chat_deals_completed / chat_conversations_started

# Ответное время (среднее время ответа продавца)
avg(chat_response_time_seconds) by (seller_id)

# Engagement (сколько сообщений в среднем в чате)
avg(chat_messages_count) by (chat_id)
```

### 7.2 SLI/SLO (Service Level Indicators/Objectives)

**Availability:**
```
SLO: 99.9% uptime
SLI: (successful_requests / total_requests) × 100
```

**Latency:**
```
SLO: P95 latency < 500ms
SLI: histogram_quantile(0.95, chat_messages_sent_duration_seconds)
```

**Error Rate:**
```
SLO: Error rate < 0.1%
SLI: (error_responses / total_responses) × 100
```

### 7.3 Custom Dashboards (Grafana)

**Dashboard 1: Real-time Overview**
- Active WebSocket connections (gauge)
- Messages sent per second (graph)
- Unread messages (heatmap by user)

**Dashboard 2: Database Health**
- PostgreSQL queries latency (histogram)
- Replication lag (graph)
- Connection pool usage (gauge)

**Dashboard 3: Kafka Lag**
- Consumer lag (по topics)
- Message throughput (graph)

---

## 8. Security: Дополнительные меры

### 8.1 End-to-End Encryption (E2EE)

**Проблема:** Сервер видит содержимое сообщений → риск утечки.

**Решение (Signal Protocol):**
```
1. Client A генерирует ключевую пару (public/private)
2. Client B генерирует ключевую пару
3. Обмен публичными ключами через сервер
4. Client A шифрует сообщение ключом Client B (RSA/ECC)
5. Сервер передаёт зашифрованное сообщение
6. Client B расшифровывает своим приватным ключом
```

**Trade-off:** Сервер не может модерировать контент (нужен client-side фильтр).

### 8.2 XSS Protection

**Проблема:** Пользователь вставляет `<script>alert('XSS')</script>` в сообщение.

**Решение:**
```go
import "html"

func SanitizeMessage(content string) string {
    // Экранируем HTML теги
    return html.EscapeString(content)
}
```

### 8.3 Rate Limiting: Advanced

**По IP адресу:**
```redis
INCR rate_limit:ip:192.168.1.1
EXPIRE rate_limit:ip:192.168.1.1 60
```

**По user_id + endpoint:**
```redis
# Разные лимиты для разных действий
INCR rate_limit:user:123:send_message  # 30/min
INCR rate_limit:user:123:get_messages  # 100/min
```

**Sliding Window:**
```go
func SlidingWindowRateLimit(userID int64, limit int) bool {
    now := time.Now().Unix()
    key := fmt.Sprintf("rate_limit:user:%d", userID)
    
    // Удаляем старые записи (старше 1 минуты)
    redis.ZRemRangeByScore(key, 0, now-60)
    
    // Добавляем текущую запись
    redis.ZAdd(key, now, fmt.Sprintf("%d", now))
    
    // Проверяем количество
    count := redis.ZCard(key)
    return count <= limit
}
```

---

## 9. Disaster Recovery

### 9.1 Backup Strategy

**PostgreSQL:**
```bash
# Full backup (раз в день)
pg_basebackup -D /backups/full/2025-02-10

# WAL archiving (continuous)
archive_command = 'cp %p /backups/wal/%f'

# Point-in-Time Recovery (PITR)
# Восстановление до любого момента времени за последние 7 дней
```

**Cassandra:**
```bash
# Snapshot (раз в день)
nodetool snapshot messages_archive

# S3 Glacier для долгосрочного хранения
aws s3 cp /var/lib/cassandra/data s3://backups/cassandra/2025-02-10 --recursive
```

### 9.2 Disaster Recovery Plan

**RTO (Recovery Time Objective):** 1 час  
**RPO (Recovery Point Objective):** 5 минут

**Сценарий 1: Primary PostgreSQL упал**
```
1. Promote Replica to Primary (automatic failover)
2. Update connection string in services
3. Start new Replica from backup
Time: ~5 минут
```

**Сценарий 2: Регион AWS упал**
```
1. Failover на secondary region (multi-region setup)
2. Update DNS (Route53)
3. Replication catch-up
Time: ~30 минут
```

**Сценарий 3: Data corruption**
```
1. Stop writes
2. Restore from latest backup (pg_basebackup)
3. Apply WAL logs (PITR)
4. Validate data
5. Resume writes
Time: ~1 час
```

---

## 10. A/B Testing Infrastructure

### 10.1 Feature Flags

**Зачем:** Тестировать новые фичи на части пользователей.

**Пример:**
```go
func ShouldUseNewChatUI(userID int64) bool {
    // 10% пользователей видят новый UI
    return userID % 10 == 0
}

// Или через сервис (LaunchDarkly, Unleash)
if featureFlags.IsEnabled("new-chat-ui", userID) {
    // Новый UI
} else {
    // Старый UI
}
```

### 10.2 Metrics по когортам

```promql
# Сравнение метрик между контрольной и тестовой группой
chat_messages_sent_total{cohort="control"}
chat_messages_sent_total{cohort="test"}

# Statistical significance (p-value)
# Проверяем в Jupyter Notebook / Python
```

---

## 11. Cost Optimization

### 11.1 Auto-scaling

**Kubernetes HPA:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: chat-service
spec:
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

**Экономия:** В ночное время (low traffic) → 3 pods вместо 10 → -70% cost.

### 11.2 Spot Instances

**AWS Spot Instances:** Скидка до 90% от On-Demand цены.

**Риск:** Могут отозвать в любой момент (с предупреждением за 2 минуты).

**Решение:**
```
Chat Service:
  - 50% On-Demand (стабильно)
  - 50% Spot (экономия)

WS Gateway:
  - 100% On-Demand (нельзя терять соединения)
```

### 11.3 Reserved Instances

**PostgreSQL/Redis:** Резервируем на 1 год → скидка 40%.

**Экономия:** `$207,000 × 0.4 = $82,800/year`

---

## 12. Альтернативная архитектура: Serverless

### 12.1 AWS Lambda + DynamoDB

**Плюсы:**
- ✅ Автоматическое масштабирование
- ✅ Платишь только за использование
- ✅ Нет управления серверами

**Минусы:**
- ❌ Cold start latency (до 1 секунды)
- ❌ WebSocket сложнее (API Gateway WebSocket)
- ❌ Лимиты (1000 concurrent executions по умолчанию)

**Когда использовать:**
- Малый трафик (< 100 RPS)
- Нерегулярная нагрузка
- Быстрый MVP

**Наш случай:** ❌ Не подходит (RPS > 1000, нужна низкая latency).

---

## Итог: Рекомендации

### ✅ Начните с простого

**MVP (Minimum Viable Product):**
1. PostgreSQL (без Cassandra)
2. Redis
3. Kafka
4. Go (Chat Service + WS Gateway)
5. FCM (push notifications)

**Масштабирование позже:**
- Cassandra (когда PostgreSQL > 10 TB)
- Sharding (когда write IOPS > 20K)
- Multi-region (когда нужна geo-распределённость)

### 📊 Мониторинг с первого дня

- Prometheus + Grafana
- ELK Stack (логи)
- Jaeger (tracing)
- PagerDuty (alerts)

### 🛡️ Безопасность

- JWT для auth
- Rate limiting (Token Bucket)
- Input sanitization (XSS protection)
- E2EE (опционально)

### 💰 Оптимизация затрат

- Auto-scaling (Kubernetes HPA)
- Spot Instances (50%)
- Reserved Instances (БД)
- CDN (CloudFront)

---

**Финальный совет:** Начните с простой архитектуры, измеряйте всё, масштабируйте по мере роста. Premature optimization — корень всех зол! 🚀
