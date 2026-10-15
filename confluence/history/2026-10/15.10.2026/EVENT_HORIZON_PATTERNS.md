# Event Horizon — Programming Patterns

> Каталог паттернов по всем 17 сервисам. Найден в бэкенде 2026-10-15.
> Per-service детали: `SERVICES_CAT_COMMANDS/<N>.<service>/<SERVICE>_PATTERNS.md`.

## Содержание
- [Сводка](#сводка)
- [Таблица паттернов](#таблица-паттернов)
- [Уникальные фичи по сервисам](#уникальные-фичи-по-сервисам)
- [Потенциальные улучшения / предпродовый релиз](#потенциальные-улучшения--предпродовый-релиз)

## Сводка

| # | Сервис | Паттернов | Роль |
|---|--------|-----------|------|
| 1 | auth | 13 | Идентификация (register, login, JWT) |
| 2 | profile | 14 | Агрегация профиля + достижения |
| 3 | history | 16 | Журнал событий + retention |
| 4 | notification | 16 | Inbox + Telegram |
| 5 | analytics | 14 | DAU / MAU / Retention (ClickHouse) |
| 6 | leaderboard | 20 | Топ игроков (Redis + Postgres) |
| 7 | inventory | 28 | Каталог товаров (Postgres + MongoDB) |
| 8 | shop | 30 | Магазин + Saga покупок |
| 9 | payment | 21 | Подписки Boosty + HMAC webhook |
| 10 | billing | 24 | Валюты (лампы, билетики) + Outbox |
| 11 | game | 20 | Игровое ядро + Boost + Validators |
| 12 | authors | 20 | Заявки на автора + State Machine |
| 13 | fulfillment | 15 | Эмуляция сборки заказа |
| 14 | gateway | 41 | API Gateway (BFF) |
| 15 | balancer | 15 | Least Connections LB |
| 16 | nats-hub | 11 | Bootstrap JetStream EVENTS |
| 17 | mcp | 15 | MCP-сервер для AI (RAG) |

**Итого: 333 паттерна/приёма.** Рекорд — gateway (41), shop (30), inventory (28).

## Таблица паттернов

### Архитектурные паттерны

| Паттерн | Где встречается |
|---------|-----------------|
| **DI Container** | auth, profile, history, notification, analytics, leaderboard, inventory, shop, payment, billing, game, authors |
| **Pipeline / Init Chain** | auth, profile, history, notification, analytics, leaderboard, inventory, shop, payment, billing, game, authors |
| **Manual DI (без контейнера)** | fulfillment, balancer, nats-hub |
| **Facade** | auth, profile, history, notification, analytics, leaderboard, inventory, shop, payment, billing, game, authors |
| **Port (Hexagonal)** | history, notification, analytics, leaderboard, inventory, shop, payment, billing, game, authors |
| **Repository** | auth, profile, inventory, shop, payment, billing, game, authors |
| **Converter / Mapper** | auth, profile, inventory, shop, billing, game, authors |
| **DTO Mapper** | gateway |
| **Adapter** | auth (jwt), analytics (clickhouse), inventory (mongo), mcp (natsx, pgxtool, redisx) |
| **Strategy** | auth (JWT generator), game (по играм), balancer (least conn), inventory (driver) |
| **Decorator** | inventory (CachedRepository) |
| **Middleware Chain** | auth, profile, history, notification, analytics, leaderboard, inventory, shop, billing, game, authors, gateway |
| **Circuit Breaker + Bulkhead** | gateway |
| **Retry Interceptor** | gateway |
| **Rate Limiter (Sliding Window)** | gateway |
| **Repository как Port** | history (HistoryStore), analytics (AnalyticsStore), notification (InboxStore) |

### Data / Persistence

| Паттерн | Где |
|---------|-----|
| **Cache-Aside** | auth, profile, leaderboard, inventory, shop, payment, billing |
| **Write-Through Cache** | inventory, leaderboard |
| **Cache Invalidation** | auth, profile, inventory, shop, billing |
| **Idempotent Write (ON CONFLICT)** | profile (achievements), notification, shop, payment, game |
| **Optimistic Locking (version)** | inventory, shop, billing |
| **Pessimistic Lock (FOR UPDATE)** | authors, shop, payment |
| **Conditional UPDATE** | billing |
| **Atomic Upsert-Increment** | billing |
| **Pagination / Clamp** | history, leaderboard, billing, authors |
| **Soft Delete / Restore** | inventory |
| **SQL Builder** | auth (pkg/sqb) |
| **Filter Builder / Parameterized Query** | inventory |
| **Template SQL** | analytics (fmt.Sprintf + ClickHouse params) |
| **Batch INSERT (COPY)** | billing |
| **Multi-Level Key** | leaderboard |

### Event-Driven / Messaging

| Паттерн | Где |
|---------|-----|
| **Event-Driven Consumer** | profile, history, notification, analytics, leaderboard, billing, game, fulfillment |
| **Durable Consumer** | profile, history, notification, analytics, leaderboard, billing, game, fulfillment, nats-hub |
| **Manual Ack + Nak** | profile, history, notification, analytics, leaderboard, game |
| **Transactional Outbox** | inventory, shop, payment, billing, game, authors |
| **Outbox Worker (Polling Publisher)** | inventory, shop, payment, billing, game, authors |
| **Event Chaining / Choreography** | leaderboard (publish `record_beaten`), gateway (publish `user.registered`) |
| **NATS → WebSocket Broadcast** | gateway |
| **Dual-Transport Consumer** | fulfillment (NATS primary, Kafka fallback) |
| **Dual-Bus Consumer** | notification (NATS + Kafka) |
| **Idempotency Guard (transport)** | fulfillment |
| **Saga Resume Worker** | shop |
| **Compensation Handler (Outbox → Callback)** | shop |
| **Reward Consumer** | billing |
| **Wildcard Subscription** | nats-hub |
| **Bootstrap / Stream Init** | nats-hub |

### Saga / Distributed Transactions

| Паттерн | Где |
|---------|-----|
| **Saga (Orchestration)** | shop |
| **Compensation Action** | shop, authors |
| **Idempotency Key (reference_id)** | shop, payment, billing |
| **Spend Cycle** | shop |
| **Idempotent Refund** | shop |
| **Idempotent Payment Completion** | payment |
| **Idempotent State Transition** | authors |
| **Subscription Expiry** | payment |
| **Revenue Split** | payment |

### Web / HTTP

| Паттерн | Где |
|---------|-----|
| **BFF (Backend For Frontend)** | gateway |
| **API Gateway** | gateway |
| **Reverse Proxy** | balancer |
| **WebSocket Hub (Broadcast)** | gateway |
| **WebSocket Ticket Auth** | gateway |
| **Origin Check** | gateway |
| **Auth Middleware (gRPC)** | gateway |
| **Role-Based Access Control** | inventory, gateway, authors |
| **Security Headers** | gateway |
| **Path Traversal Guard** | gateway |
| **Body Limit Middleware** | gateway |
| **Docs Gate (Basic Auth)** | gateway |
| **Upload with Validation** | gateway |
| **Legacy API Rewrite** | gateway |
| **gRPC → HTTP Error Mapping** | gateway |
| **Partial Response Fallback** | gateway |
| **MCP Server** | mcp |
| **Tool Registration** | mcp |
| **stdio Transport** | mcp |
| **Health / Ready** | все сервисы |

### Observability

| Паттерн | Где |
|---------|-----|
| **Prometheus Metrics (Counter)** | game, gateway, balancer, billing (косвенно) |
| **Prometheus Metrics (Histogram)** | game, gateway |
| **Prometheus Metrics (Gauge)** | balancer |
| **OTel / Jaeger Tracing** | auth, leaderboard, inventory, shop, billing, game, gateway |
| **OTel gRPC StatsHandler** | auth, leaderboard, inventory, shop, billing, game |
| **otelgin Middleware** | gateway |
| **Context Propagation to gRPC** | gateway, inventory |
| **X-Balancer-Backend Header** | balancer |

### Reliability

| Паттерн | Где |
|---------|-----|
| **Graceful Shutdown / Closer** | все сервисы |
| **Context Cancellation** | все сервисы с воркерами |
| **Graceful Degradation** | profile (Redis), notification (Redis), analytics, shop (Payment), game (Billing), fulfillment (NATS), mcp (per-tool), gateway (Redis) |
| **Retry with Fixed Delay** | notification |
| **Retry (без backoff)** | inventory (mongo), shop (grpc dial) |
| **Fail-Fast (breaker)** | gateway |
| **Fail-Open (rate limit)** | gateway |
| **FOR UPDATE SKIP LOCKED** | shop, game |

### Security

| Паттерн | Где |
|---------|-----|
| **HMAC-SHA256 Verification** | payment, gateway |
| **ConstantTimeCompare** | payment, gateway |
| **Dual Auth Path (HMAC + Secret)** | payment |
| **Role Check (Privilege Escalation)** | authors |
| **SQL Sandbox** | mcp |
| **Crypto-Secure UUID** | fulfillment |
| **Session Store (Redis)** | auth |

### Testing

| Паттерн | Где |
|---------|-----|
| **Mock Object** | auth, profile, history, notification, analytics, leaderboard, inventory, shop, payment, billing, game, authors |
| **Testcontainers** | history, inventory, shop, payment, billing, game, authors |
| **Build Tags (`//go:build integration`)** | history, inventory, shop, payment, billing, game, authors |
| **Race Condition Test** | billing |
| **Failover Test** | game |
| **Thorough Testing (breaker, retry)** | gateway |

## Уникальные фичи по сервисам

### auth (13)
- **Session Store** — JWT хранятся в Redis для возможности отзыва.
- **Config View** — `ConfigProvider` с узкими view-интерфейсами (Interface Segregation).
- **SQL Builder** — свой `pkg/sqb`.

### profile (14)
- **Rule Engine** — `EvaluateAchievements` — чистая функция вычисления кодов.
- **Achievement Unlock Pipeline** — `EvaluateAndUnlock` → `UnlockCodes` → `Unlock`.

### history (16)
- **Port как `HistoryStore`** (Hexagonal).
- **Retention Policy** — самоочистка данных по TTL.
- **Payload Sanitization** — если невалидный JSON, оборачивается в `{"raw": "..."}`.
- **Scheduled Worker** (`RetentionWorker`) — cron внутри сервиса.
- **Nak Pattern** для обработки ошибок.

### notification (16)
- **Два независимых сервиса в одном процессе** — `InboxService` + `Notifier`.
- **Два шины одновременно** — NATS + Kafka с fallback.
- **Fan-Out** — одно событие → N уведомлений.
- **Idempotency Key** через `source_event`.

### analytics (14)
- **ClickHouse** — единственный сервис с аналитической БД.
- **Cohort Retention Analysis** — классический приём.
- **Parameterized Query через ClickHouse-параметры**.

### leaderboard (20)
- **Ack Batcher** — подтверждения NATS собираются в батчи.
- **Event Chaining** — publish `record_beaten` внутри consumer'а.
- **Write-Through Cache** Redis + Postgres backup.
- **Restore-on-Startup** — warm-up кеша.
- **Max-Only Write** через `ZAddArgs{GT: true}`.

### inventory (28)
- **Два драйвера БД** — Postgres + MongoDB (Strategy).
- **Decorator для кеша** — самый чистый Cache-Aside.
- **Change Streams (CDC)** для MongoDB.
- **MongoDB Transaction + Aggregation + $lookup**.
- **RBAC через интерсепторы**.
- **Outbox через интерфейс + Type Assertion**.

### shop (30)
- **Saga Pattern (Orchestration)** — единственный сервис с явной сагой.
- **Spend Cycle** — оригинальный приём для идемпотентности.
- **Saga Resume Worker** — фоновое доведение саг.
- **Pessimistic + Optimistic Lock** в одной транзакции.
- **Outbox → Compensate Callback**.

### payment (21)
- **HMAC-SHA256 Verification** + **subtle.ConstantTimeCompare**.
- **Dual Auth Path** (HMAC + Shared Secret).
- **Subscription Expiry** при новом платеже.
- **Revenue Split** (85/15).
- **Cache TTL based on Expiry**.

### billing (24)
- **Atomic Upsert-Increment** — `ON CONFLICT DO UPDATE SET balance = balance + amount`.
- **Conditional UPDATE** — списание через `WHERE balance >= $1`.
- **Double-Entry Ledger** — транзакции с `balance_after`.
- **Batch INSERT через COPY**.
- **Transaction Batch (in-memory)**.
- **Check-Only Mode**.
- **Race Condition Test**.

### game (20)
- **Strategy по играм** — `switch req.GameID`.
- **Domain Validators** — для hexagon и memory.
- **Boost System** — 3-фазный цикл с идемпотентным переиспользованием.
- **Transactional Outbox с Fallback**.
- **Prometheus Metrics в handler**.
- **Failover Test**.

### authors (20)
- **State Machine** для заявок.
- **Pessimistic Lock** через `FOR UPDATE` в бизнес-логике.
- **Compensating Action (Revert)**.
- **Idempotent Application Submission** с обработкой race.
- **Role Check** для privilege escalation.

### fulfillment (15)
- **Dual-Transport Consumer** — NATS primary, Kafka fallback.
- **Idempotency Guard** на уровне транспорта.
- **Delay-Based Simulation**.
- **Transport Selection в Service**.
- **Нет gRPC** — только NATS/Kafka + HTTP.
- **Нет DI-контейнера** — manual DI.

### gateway (41)
- **Circuit Breaker + Bulkhead** — единственный сервис с этим.
- **Retry Interceptor с full jitter**.
- **Sliding Window Rate Limiter**.
- **Partial Response Fallback**.
- **DTO Mapper (proto3 omitempty fix)**.
- **WebSocket Ticket Auth**.
- **Parallel Fan-Out + Semaphore**.
- **Docs Gate**.
- **Анти-паттерны:** Code Duplication (Hub ×2), God Function (runGateway).

### balancer (15)
- **Least Connections** — единственный сервис с балансировкой.
- **Atomic Counter** для ActiveConns.
- **Reverse Proxy** — `httputil.ReverseProxy`.
- **Prometheus Gauge** — единственный сервис с Gauge.
- **Анти-паттерны:** пустые пакеты `backend/`, `proxy/`.

### nats-hub (11)
- **Bootstrap / Initializer** — единственный сервис, создающий инфраструктуру.
- **Idempotent Stream Creation** (`AddStream` → `UpdateStream`).
- **Stream Config as Code**.
- **Реальный health check** через `nc.IsConnected()`.

### mcp (15)
- **MCP Server** для AI-интеграции.
- **Tool Registration** (6 tools).
- **Graceful Degradation per-tool**.
- **SQL Sandbox** (`ValidateSelect`).
- **Offline RAG (TF-IDF)** без embedding API.
- **stdio Transport**.

## Потенциальные улучшения / предпродовый релиз

### Критичные (Pre-Prod)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 1 | **fulfillment** | Заглушка `time.After(10s)` вместо реального WMS. | Заменить на вызов warehouse/delivery API. Добавить Outbox для гарантии. |
| 2 | **fulfillment** | Нет Outbox — если сервис упадёт между `purchase.paid` и `purchase.fulfilled`, событие потеряется. | Внедрить Transactional Outbox. |
| 3 | **shop** | `kafkaProd` создаётся, но не используется для публикации (только NATS). | Убрать или подключить. |
| 4 | **shop** | `paymentClient` может быть `nil` — мерч-гейт деградирует. | Добавить метрику `merch_gate_disabled`. |
| 5 | **billing** | `RewardConsumer` использует `time.Now().UnixNano()` для `reference_id` — при повторной доставке события может начислить дважды. | Использовать `event.ReferenceID` из события. |
| 6 | **gateway** | Два Hub (`app/gateway.go` и `websocket/hub.go`) — дублирование. | Унифицировать. |
| 7 | **gateway** | `runGateway` — God Function (больше 1000 строк). | Разбить на модули: `bootstrap/clients.go`, `bootstrap/routes.go`, `bootstrap/middleware.go`. |
| 8 | **gateway** | `Async` gRPC-вызовы в admin/users без circuit breaker. | Обернуть в `throughBreaker`. |
| 9 | **balancer** | Пустые пакеты `backend/` и `proxy/` — мёртвый код. | Удалить или наполнить. |
| 10 | **balancer** | `/ready` не проверяет бэкенды — всегда `ready`. | Реальная проверка доступности хотя бы одного backend. |
| 11 | **balancer** | Static registry — нет service discovery. | Интегрировать с etcd/consul/k8s. |
| 12 | **balancer** | Нет health checks бэкендов. | Периодический ping `/health` каждого backend. |
| 13 | **notification** | `Notifier` (Telegram) в `Handle` для Kafka падает без Ack при unmarshal error. | Poison message → Ack. |
| 14 | **analytics** | `sql := fmt.Sprintf(...)` с `%s` для db name — если станет user input → SQL injection. | Параметризовать db name. |
| 15 | **mcp** | `postgres_query` — SQL sandbox через string matching. Легко обходится. | Использовать read-only роль в Postgres + `pgx` prepared statements. |
| 16 | **mcp** | RAG — TF-IDF без embedding. | Опционально: добавить embedding fallback (OpenAI). |
| 17 | **mcp** | `mcp.NewToolResultError(err.Error())` — утечка внутренних деталей. | Санитизировать сообщения. |

### Средней важности

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 18 | **auth** | `boosty.to` URL захардкожен. | Config. |
| 19 | **payment** | `subscriptionDays` в `Config` — defaults 30. | Проверить, что 30 корректно. |
| 20 | **shop** | `cancelPurchase` в `creditCancel` — `reference_id = cancel-{purchaseID}`. Если покупка дважды отменена, вернёт `AlreadyRefunded`, но `AddCurrency` может начислить дважды. | Использовать Idempotency Key. |
| 21 | **inventory** | Два драйвера (Postgres + MongoDB) — двойная поддержка. | Определиться с одним. |
| 22 | **inventory** | `toPostgresFilter` собирает SQL через `fmt.Sprintf` — если появятся новые фильтры, может утечь. | Уже защищено через `$1, $2`. |
| 23 | **game** | `SubmitScore` — god-function `switch` на 8 игр. | Вынести в per-game strategy. |
| 24 | **authors** | `RevertApplication` удаляет `authors` — если там есть заказы, FK упадёт. | Soft delete + check FKs. |
| 25 | **leaderboard** | `RestoreFromPostgres` при старте — если Postgres упал, leaderboard пустой. | Retry / exponential backoff. |
| 26 | **billing** | `GetBalance` возвращает `-1` при cache miss — можно спутать с реальным `-1`. | Использовать `(int, bool)`. |

### Мелкие

| # | Сервис | Проблема |
|---|--------|----------|
| 27 | **gateway** | `log.Printf` везде — нет структурированного логирования. |
| 28 | **gateway** | `//go:build integration` — нет тестов на circuit breaker в prod-режиме. |
| 29 | **balancer** | `log.Printf` в hot path. |
| 30 | **fulfillment** | `newUUID` (crypto/rand) vs `uuid.NewString()` (google/uuid) — несогласованность. |

### Что **не** улучшать

- **Saga в `shop`** — работает корректно.
- **Circuit Breaker в `gateway`** — минимальный, но работает.
- **Least Connections в `balancer`** — правильный алгоритм.
- **Double-Entry Ledger в `billing`** — фундамент.

## Ссылки

- Per-service: `SERVICES_CAT_COMMANDS/<N>.<service>/<SERVICE>_PATTERNS.md`
- Raw terminal snapshots: `SERVICES_CAT_COMMANDS/<N>.<service>/RAW_TERMINAL_SNAPSHOT.md`
- CAT commands: `SERVICES_CAT_COMMANDS/<N>.<service>/CAT_COMMANDS.md`