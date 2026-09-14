# 💬 System Design: Мессенджер Avito

**Задача:** Спроектировать real-time мессенджер для маркетплейса с 3M DAU

---

## 📊 1. Математика и расчёты

### Исходные данные
- **DAU (Daily Active Users):** 3,000,000
- **Средний размер сообщения:** 10 KB
- **Сообщений на пользователя в день:** 30
- **Соотношение write/read:** 1:1
- **Хранение данных:** 2 года + бекапы

### 1.1 RPS (Requests Per Second)

**Формула:** `RPS = DAU × Messages per user / 86400`

```
RPS (write) = 3,000,000 × 30 / 86,400 = ~1,042 write/s
RPS (read)  = 3,000,000 × 30 / 86,400 = ~1,042 read/s

Total RPS = 2,084 r/s
```

**Peak RPS (учитываем неравномерность, коэффициент 3x):**
```
Peak write = 1,042 × 3 = 3,126 write/s
Peak read  = 1,042 × 3 = 3,126 read/s

Total Peak RPS = 6,252 r/s
```

### 1.2 Traffic per second

```
Traffic (write) = 1,042 × 10 KB = 10.42 MB/s = ~10 MB/s
Traffic (read)  = 1,042 × 10 KB = 10.42 MB/s = ~10 MB/s

Total traffic = ~20 MB/s
Peak traffic  = ~60 MB/s (3x)
```

### 1.3 Storage (2 года + бекапы)

**Сообщений в день:**
```
Daily messages = 3,000,000 × 30 = 90,000,000 messages/day
```

**Сообщений за 2 года:**
```
Messages in 2 years = 90,000,000 × 730 = 65,700,000,000 messages
```

**Объём данных (без индексов):**
```
Storage = 65.7B × 10 KB = 657 TB
```

**С учётом индексов (коэффициент 1.5x):**
```
Storage with indexes = 657 TB × 1.5 = ~986 TB
```

**С бекапами (full + incremental):**
```
Backups = 986 TB × 0.5 = 493 TB
Total storage = 986 + 493 = 1,479 TB ≈ 1.5 PB
```

**✅ Итого: 1500 TB за 2 года с бекапами** (соответствует требованию!)

---

## 🗄️ 2. Выбор базы данных

### 2.1 Основная БД для сообщений

**Выбор: PostgreSQL (для транзакционных данных) + Cassandra/MongoDB (для архива сообщений)**

#### PostgreSQL (Hot Data — последние 30 дней)

**Почему:**
- ✅ ACID транзакции (важно для статусов доставки)
- ✅ Сложные запросы (поиск по чатам, фильтрация)
- ✅ Поддержка JSON (метаданные сообщений)
- ✅ Foreign Keys (связи users ↔ chats ↔ messages)
- ✅ Репликация (Primary-Replica для read)

**Структура:**
```sql
-- Чаты (диалоги между продавцом и покупателем)
CREATE TABLE chats (
    chat_id UUID PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES users(user_id),
    buyer_id BIGINT NOT NULL REFERENCES users(user_id),
    item_id BIGINT, -- товар, по которому общаемся
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(seller_id, buyer_id, item_id)
);

-- Сообщения (hot data — последние 30 дней)
CREATE TABLE messages (
    message_id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(chat_id),
    sender_id BIGINT NOT NULL REFERENCES users(user_id),
    content TEXT NOT NULL, -- текст сообщения
    media_urls JSONB, -- прикреплённые файлы
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP, -- время доставки
    read_at TIMESTAMP, -- время прочтения
    
    INDEX idx_chat_sent (chat_id, sent_at DESC), -- для получения истории чата
    INDEX idx_sender (sender_id, sent_at DESC) -- для отправленных пользователем
);

-- Партиционирование по времени (по месяцам)
CREATE TABLE messages_2025_01 PARTITION OF messages
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

**Масштабирование:**
- Партиционирование по времени (по месяцам)
- Read Replicas (3-5 реплик для чтения)
- Connection Pooling (PgBouncer)
- Sharding по `chat_id` (если нужно > 10TB hot data)

#### Cassandra/MongoDB (Cold Data — архив > 30 дней)

**Почему:**
- ✅ Горизонтальное масштабирование (до петабайтов)
- ✅ Partition Key = `chat_id` (локальность данных)
- ✅ TTL (автоматическое удаление после 2 лет)
- ✅ Дешёвое хранение холодных данных

**Cassandra Schema:**
```cql
CREATE TABLE messages_archive (
    chat_id UUID,
    message_id UUID,
    sender_id BIGINT,
    content TEXT,
    media_urls TEXT, -- JSON as string
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    PRIMARY KEY (chat_id, sent_at, message_id)
) WITH CLUSTERING ORDER BY (sent_at DESC)
  AND default_time_to_live = 63072000; -- 2 года в секундах
```

**Архивация:**
- Раз в день ETL job переносит сообщения старше 30 дней из PostgreSQL в Cassandra
- PostgreSQL освобождается от старых партиций (DROP PARTITION)

### 2.2 Кэш (Redis)

**Для чего:**
- ✅ Online users (список активных пользователей)
- ✅ Typing indicators ("пользователь печатает...")
- ✅ Unread counts (счётчики непрочитанных)
- ✅ Last seen (время последней активности)
- ✅ Rate limiting (защита от спама)

**Структуры:**
```
SET online:users:chat:{chat_id} -> {user_id1, user_id2}
STRING typing:chat:{chat_id}:user:{user_id} -> timestamp (TTL 5s)
HASH unread:user:{user_id} -> {chat_id: count}
STRING last_seen:user:{user_id} -> timestamp
```

### 2.3 Message Queue (Kafka/RabbitMQ)

**Kafka** — для event-driven архитектуры:
- Topic `chat.messages.sent` — новые сообщения
- Topic `chat.messages.delivered` — подтверждения доставки
- Topic `chat.messages.read` — подтверждения прочтения
- Topic `notifications.push` — пуш-уведомления

**Зачем:**
- ✅ Decoupling (Chat Service → Notification Service → Push Service)
- ✅ At-least-once доставка (retry для пушей)
- ✅ Audit log (все события сохраняются)

---

## 🏗️ 3. Архитектура системы

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Mobile App / Web                            │
└─────────────────────────────────────────────────────────────────────────┘
                    │                                    │
                    │ HTTP/REST                          │ WebSocket
                    ▼                                    ▼
┌──────────────────────────────────┐   ┌────────────────────────────────┐
│      API Gateway (nginx)         │   │   WebSocket Gateway (Go)       │
│  - Auth (JWT validation)         │   │   - Sticky sessions            │
│  - Rate limiting                 │   │   - Connection pooling         │
│  - Load balancing                │   │   - Heartbeat (ping/pong)      │
└──────────────────────────────────┘   └────────────────────────────────┘
                    │                                    │
                    ▼                                    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                          Chat Service (Go)                                │
│  - Send message (POST /chats/{id}/messages)                              │
│  - Get messages (GET /chats/{id}/messages?cursor=...)                    │
│  - Mark as read (POST /chats/{id}/messages/{msg_id}/read)                │
│  - Get chat list (GET /chats?unread_only=true)                           │
└──────────────────────────────────────────────────────────────────────────┘
            │                    │                    │
            ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐
│   PostgreSQL    │  │      Redis      │  │       Kafka         │
│  (Hot: 30 days) │  │   (Cache)       │  │  (Event Stream)     │
└─────────────────┘  └─────────────────┘  └─────────────────────┘
            │                                         │
            │                                         ▼
            │                              ┌─────────────────────┐
            │                              │ Notification Service│
            │                              │  - Push (FCM/APNs)  │
            │                              │  - Email (optional) │
            │                              └─────────────────────┘
            ▼
┌─────────────────────┐
│   Cassandra/MongoDB │
│  (Cold: > 30 days)  │
└─────────────────────┘
```

### 3.1 Компоненты

#### API Gateway (nginx/Envoy)
- JWT validation
- Rate limiting (по user_id)
- SSL termination
- Load balancing (Round Robin / Least Connections)

#### WebSocket Gateway (Go)
- Sticky sessions (один user → один сервер)
- Connection pooling (до 10K connections per server)
- Heartbeat (ping/pong каждые 30 секунд)
- Graceful shutdown (закрытие соединений при деплое)

**Особенности:**
- Horizontal scaling (10 серверов × 10K connections = 100K одновременных)
- Redis Pub/Sub для broadcast (сообщение из Chat Service → всем WS серверам)

#### Chat Service (Go)
- CRUD операции с чатами
- Send message → DB + Kafka + WebSocket broadcast
- Cursor-based pagination (для истории чата)
- Idempotency key (для дедупликации повторных отправок)

#### Notification Service (Go)
- Consumer из Kafka
- Отправка push через FCM (Android) / APNs (iOS)
- Retry с exponential backoff
- DLQ для failed notifications

---

## 🔥 4. Подводные камни и нюансы

### 4.1 WebSocket: Sticky Sessions

**Проблема:** Если user переключается между WS серверами → теряет соединение.

**Решение:**
- Sticky sessions (по `user_id` через consistent hashing)
- Или Redis Pub/Sub (сообщение broadcast на все WS сервера)

**Пример (Redis Pub/Sub):**
```go
// Chat Service публикует в Redis
redis.Publish("chat:messages", json.Marshal(message))

// WS Gateway подписан на канал
redis.Subscribe("chat:messages", func(msg string) {
    // Отправляем всем подключённым клиентам из этого чата
    ws.BroadcastToChat(chatID, msg)
})
```

### 4.2 "Typing indicator"

**Проблема:** Миллионы событий "user is typing..." → нагрузка на сервер.

**Решение:**
- Throttling на клиенте (отправляем не чаще 1 раза в 2 секунды)
- TTL в Redis (5 секунд)
- Не пишем в БД (только в Redis)

```go
// При получении "typing" события
redis.Set("typing:chat:123:user:456", time.Now(), 5*time.Second)

// При запросе "кто печатает"
keys := redis.Keys("typing:chat:123:user:*")
// Возвращаем только те, которые не истекли
```

### 4.3 Unread Counts

**Проблема:** Пересчитывать `SELECT COUNT(*) WHERE read_at IS NULL` при каждом запросе → медленно.

**Решение:**
- Денормализация: храним счётчик в Redis
- При новом сообщении → `HINCRBY unread:user:{user_id} {chat_id} 1`
- При прочтении → `HSET unread:user:{user_id} {chat_id} 0`

```go
// Получить все непрочитанные
unread := redis.HGetAll("unread:user:123")
// {"chat_456": 5, "chat_789": 12}
```

### 4.4 Message Delivery Statuses

**Проблема:** Как отслеживать "доставлено" и "прочитано"?

**Решение:**
- Клиент отправляет ACK через WebSocket при получении сообщения
- Chat Service обновляет `delivered_at` в БД
- При открытии чата → отправляем "mark as read" → обновляем `read_at`

**Flow:**
```
1. Sender → Chat Service (send message)
2. Chat Service → DB (insert)
3. Chat Service → Kafka (event)
4. Chat Service → WS Gateway (broadcast)
5. WS Gateway → Receiver (WebSocket push)
6. Receiver → WS Gateway (ACK "delivered")
7. WS Gateway → Chat Service (update delivered_at)
8. Receiver opens chat → Chat Service (mark as read)
```

### 4.5 Rate Limiting (защита от спама)

**Проблема:** User может заспамить чат тысячами сообщений.

**Решение:**
- Token Bucket (30 сообщений в минуту на user)
- Redis: `INCR rate_limit:user:{user_id}` + `EXPIRE 60`

```go
func RateLimit(userID int64) bool {
    key := fmt.Sprintf("rate_limit:user:%d", userID)
    count := redis.Incr(key)
    if count == 1 {
        redis.Expire(key, 60) // 1 минута
    }
    return count <= 30 // максимум 30 сообщений в минуту
}
```

### 4.6 Media Files (фото, видео)

**Проблема:** Сообщения могут содержать файлы → нельзя хранить в БД.

**Решение:**
- S3/MinIO для хранения файлов
- В БД хранится только URL: `{"media_urls": ["https://cdn.avito.ru/chat/abc123.jpg"]}`
- CDN (CloudFront / CloudFlare) для быстрой доставки

**Upload flow:**
```
1. Client → Chat Service (request presigned URL)
2. Chat Service → S3 (generate presigned URL)
3. Client → S3 (upload file directly)
4. Client → Chat Service (send message with S3 URL)
```

### 4.7 Search по сообщениям

**Проблема:** `LIKE '%keyword%'` в PostgreSQL → медленно на миллионах строк.

**Решение:**
- ElasticSearch для full-text search
- Индексируем только последние 30 дней (hot data)
- Async indexing через Kafka

```
Chat Service → Kafka (message.sent) → ES Indexer → ElasticSearch
```

### 4.8 Archiving (перенос в Cassandra)

**Проблема:** PostgreSQL разрастается → медленные запросы.

**Решение:**
- Cron job раз в день: `SELECT * FROM messages WHERE sent_at < NOW() - INTERVAL '30 days'`
- Batch insert в Cassandra (по 1000 сообщений)
- `DROP PARTITION` в PostgreSQL

```sql
-- PostgreSQL: удаляем старую партицию
DROP TABLE messages_2024_11;

-- Cassandra: данные уже там через ETL
SELECT * FROM messages_archive WHERE chat_id = ? AND sent_at < ?;
```

---

## 🚀 5. Масштабирование

### 5.1 Horizontal Scaling

**Chat Service:**
- Stateless (можно запустить 10+ инстансов)
- Load Balancer (nginx/Envoy) перед ними
- Auto-scaling на основе CPU/Memory (Kubernetes HPA)

**WebSocket Gateway:**
- 10 серверов × 10K connections = 100K одновременных
- Redis Pub/Sub для broadcast между серверами
- Sticky sessions (consistent hashing по `user_id`)

**PostgreSQL:**
- Primary (write) + 5 Replicas (read)
- Read queries → Replicas (через pgpool или service mesh)
- Write queries → Primary

### 5.2 Database Sharding (если нужно)

**По `user_id`:**
- Shard 1: `user_id % 10 = 0` (users 0, 10, 20, ...)
- Shard 2: `user_id % 10 = 1` (users 1, 11, 21, ...)
- ...
- Shard 10: `user_id % 10 = 9`

**Проблема:** Запрос "все чаты пользователя" → может быть на разных шардах (если чат участников из разных шардов).

**Решение:** Шардируем по `chat_id`, а не по `user_id`.

### 5.3 Caching Strategy

**Redis (L1 Cache):**
- Hot data: последние N чатов пользователя
- TTL: 5 минут
- Cache invalidation: при новом сообщении обновляем кэш

**PostgreSQL (L2 Cache):**
- Hot data: последние 30 дней
- Уже быстро (партиции + индексы)

**Cassandra (L3 Storage):**
- Cold data: > 30 дней
- Медленнее, но дёшево

---

## 📊 6. Monitoring & Observability

### 6.1 Метрики (Prometheus)

**Ключевые метрики:**
```
# Chat Service
chat_messages_sent_total (Counter)
chat_messages_sent_duration_seconds (Histogram)
chat_messages_delivered_total (Counter)
chat_messages_read_total (Counter)

# WebSocket Gateway
ws_connections_active (Gauge) — сколько сейчас подключено
ws_messages_sent_total (Counter)
ws_messages_received_total (Counter)

# Database
postgres_queries_duration_seconds (Histogram)
postgres_connections_active (Gauge)

# Redis
redis_commands_duration_seconds (Histogram)
redis_memory_used_bytes (Gauge)
```

### 6.2 Логирование (ELK Stack)

**Обязательные поля:**
```json
{
  "timestamp": "2025-02-10T12:34:56Z",
  "level": "INFO",
  "service": "chat-service",
  "request_id": "req-abc123",
  "user_id": 456,
  "chat_id": "chat-789",
  "action": "send_message",
  "duration_ms": 45
}
```

### 6.3 Tracing (Jaeger)

**Распределённая трассировка:**
```
Client → API Gateway → Chat Service → PostgreSQL
                    ↓
                  Kafka → Notification Service → FCM
```

**Span IDs:**
- `send_message` (root span)
  - `db.insert` (child span)
  - `kafka.publish` (child span)
  - `ws.broadcast` (child span)

### 6.4 Alerts

**SLA: 99.9% uptime (8.76 часов downtime в год)**

**Критические алёрты:**
```
# RPS упал > 50% от baseline
- alert: HighMessageDropRate
  expr: rate(chat_messages_sent_total[5m]) < 500
  for: 5m

# Latency > 1 секунда
- alert: HighLatency
  expr: histogram_quantile(0.95, chat_messages_sent_duration_seconds) > 1
  for: 5m

# WebSocket connections падают
- alert: WebSocketConnectionDrop
  expr: rate(ws_connections_active[5m]) < -100
  for: 2m

# PostgreSQL реплики отстали
- alert: PostgreSQLReplicationLag
  expr: pg_replication_lag_seconds > 60
  for: 5m
```

---

## 🛡️ 7. Security

### 7.1 Authentication

**JWT (JSON Web Token):**
- Client → Auth Service (login) → JWT token
- JWT содержит: `user_id`, `role` (seller/buyer), `exp` (expiration)
- API Gateway валидирует JWT перед каждым запросом

### 7.2 Authorization

**Проверка доступа к чату:**
```go
func CanAccessChat(userID int64, chatID string) bool {
    chat := db.GetChat(chatID)
    return chat.SellerID == userID || chat.BuyerID == userID
}
```

### 7.3 Rate Limiting

- 30 сообщений в минуту (Token Bucket)
- 100 API запросов в минуту
- DDoS protection (Cloudflare)

### 7.4 Content Moderation

**Проблема:** Спам, мошенничество, неприемлемый контент.

**Решение:**
- ML модель для детекции спама (async обработка через Kafka)
- Blacklist слов (проверка на клиенте и сервере)
- User reports → модерация

---

## 💰 8. Cost Estimation

### 8.1 Servers (AWS/DigitalOcean)

```
Chat Service:     10 × c5.2xlarge ($0.34/h) = $29,376/year
WS Gateway:       10 × c5.2xlarge ($0.34/h) = $29,376/year
API Gateway:       3 × c5.xlarge  ($0.17/h) = $4,468/year
Notification:      5 × c5.xlarge  ($0.17/h) = $7,446/year

Total compute: ~$70,000/year
```

### 8.2 Database

```
PostgreSQL (Primary + 5 Replicas):
  - 6 × r5.4xlarge (128 GB RAM) = $70,000/year
  - Storage: 10 TB × $0.10/GB/month = $12,000/year

Cassandra (3-node cluster):
  - 3 × i3.4xlarge (1.9 TB NVMe) = $45,000/year
  - Storage: 1.5 PB на 2 года = $50,000/year (S3 Glacier for backups)

Redis (ElastiCache):
  - 3 × r5.2xlarge (64 GB) = $30,000/year

Total storage: ~$207,000/year
```

### 8.3 Network & CDN

```
Traffic: 20 MB/s × 86,400 × 30 = 51.84 TB/month
CloudFront: 51.84 TB × $0.085/GB = $4,400/month = $52,800/year
```

### 8.4 Total

```
Compute:   $70,000
Storage:   $207,000
Network:   $52,800
Kafka:     $20,000
ELK Stack: $15,000

Total: ~$365,000/year ≈ $30,000/month
```

**Per user per year:** `$365,000 / 3,000,000 = $0.12/user/year`

---

## 📋 9. Summary

### Архитектура
```
Mobile/Web → API Gateway → Chat Service → PostgreSQL (hot 30d)
                                       ↓ Kafka → Notification Service
          → WS Gateway ────────────────┘
                       → Redis (cache)
                       → Cassandra (cold > 30d)
```

### Ключевые цифры
- **RPS:** ~2K r/s (average), ~6K r/s (peak)
- **Traffic:** ~20 MB/s (average), ~60 MB/s (peak)
- **Storage:** 1.5 PB за 2 года с бекапами
- **Cost:** ~$365K/year ($0.12/user/year)

### Выбор технологий
- **PostgreSQL:** Hot data (30 дней), ACID транзакции
- **Cassandra:** Cold data (> 30 дней), горизонтальное масштабирование
- **Redis:** Cache, online users, typing indicators
- **Kafka:** Event streaming, decoupling, audit log
- **Go:** Chat Service, WebSocket Gateway (высокая производительность)

### Подводные камни
1. Sticky sessions для WebSocket
2. Redis Pub/Sub для broadcast между WS серверами
3. Rate limiting (Token Bucket)
4. Unread counts в Redis (денормализация)
5. Media files в S3/CDN
6. Archiving в Cassandra (ETL job)
7. Search через ElasticSearch
8. Content moderation (ML + blacklist)

---

**Итог:** Система готова к масштабированию до 3M DAU с возможностью роста до 10M+ при добавлении серверов и шардировании БД. 🚀
