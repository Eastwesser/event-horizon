# Event Horizon — техдолг (финальный, без дубликатов)

> После большого ревью. Убраны дубликаты. P0 — блокеры, P1 — надёжность, P2 — гигиена.
> Источник: EVENT_HORIZON_PATTERNS_15.10.2026.md, FINISHERS.md, FUTURE_EVENT_HORIZON.md.

## P0 — блокеры (нельзя в прод)

| # | Сервис | Проблема | Решение | Почему P0 |
|---|--------|----------|---------|-----------|
| 1 | **Fulfillment** | Нет Outbox — событие теряется между `purchase.paid` и `purchase.fulfilled` | Transactional Outbox | Заказ «завис», `fulfilled_at` null |
| 2 | **Fulfillment** | `time.After(10s)` — заглушка вместо WMS | Заменить на warehouse/delivery API | Не реальная сборка |
| 3 | **Gateway** | `runGateway` — God Function > 1000 строк | Разбить на `bootstrap/clients.go`, `routes.go`, `middleware.go` | Долгая инициализация, сложно поддерживать |
| 4 | **Shop** | Saga `started` > 60s не берётся `resumeSagas` | Воркер берёт `started`, проверяет Billing `GetTransactions(spend_ref)` | **Деньги списаны, компенсации нет** |
| 5 | **Billing** | `RewardConsumer` `time.Now().UnixNano()` для `reference_id` | `event.ReferenceID` из события | **Двойное начисление** |
| 6 | **Authors** | `RevertApplication` hard delete — FK падает | Soft delete + check FKs | Approve не откатывается |
| 7 | **Balancer** | Single point of failure (нет резервного) | Несколько Balancer + VIP/DNS | **Всё упало — всё недоступно** |

## P1 — надёжность (перед production)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 8 | **Gateway** | Два Hub (`app/gateway.go` + `websocket/hub.go`) | Унифицировать |
| 9 | **Gateway** | Async gRPC в admin/users без CB | Обернуть в `throughBreaker` |
| 10 | **Shop** | `paymentClient` nil — мерч-гейт деградирует | Метрика `merch_gate_disabled` |
| 11 | **Shop** | `kafkaProd` создаётся, не используется | Убрать или подключить |
| 12 | **Shop** | `cancelPurchase` — `reference_id = cancel-{id}` может задвоить | Idempotency Key |
| 13 | **Inventory** | Два драйвера (PG + Mongo) — двойная поддержка | Определиться с одним |
| 14 | **Inventory** | `toPostgresFilter` через `fmt.Sprintf` | Переписать без `fmt.Sprintf` для имён колонок |
| 15 | **Inventory** | Outbox через type assertion — хрупко | Явный интерфейс `TransactionalRepository` |
| 16 | **MCP** | SQL Sandbox string matching — обходится | Read-only роль + `pgx` prepared |
| 17 | **MCP** | `err.Error()` — утечка внутренних деталей | Санитизировать сообщения |
| 18 | **Balancer** | `/ready` не проверяет бэкенды | Реальная проверка хотя бы одного backend |
| 19 | **Balancer** | Static registry — нет service discovery | etcd/consul/k8s |
| 20 | **Balancer** | Нет health checks бэкендов | Периодический ping `/health` |
| 21 | **Notification** | Poison message из Kafka — бесконечный retry | Poison message → Ack + DLQ |
| 22 | **Analytics** | `fmt.Sprintf` для db name | Параметризовать |
| 23 | **Leaderboard** | `RestoreFromPostgres` при падении PG — пустой leaderboard | Retry / exponential backoff |
| 24 | **Payment** | Нет сверки суммы webhook с `payments.amount` | Сверять `amount` |
| 25 | **History** | RetentionWorker в процессе — при рестарте сброс | Kubernetes CronJob |
| 26 | **Auth** | Нет ротации `JWT_SECRET` | Ротация |
| 27 | **Все** | Нет WAF | WAF перед Gateway |

## P2 — гигиена (когда будет время)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 28 | **Gateway** | `log.Printf` везде — нет structured logging | `slog` |
| 29 | **Balancer** | Пустые пакеты `backend/`, `proxy/` | Удалить или наполнить |
| 30 | **Balancer** | `log.Printf` в hot path | Structured logging |
| 31 | **Fulfillment** | `newUUID` (crypto/rand) vs `uuid.NewString()` | Унифицировать |
| 32 | **Game** | `redis-game` `:6380` поднят, не подключён | Подключить или убрать |
| 33 | **Billing** | `GetBalance` возвращает `-1` при cache miss | `(int, bool)` |
| 34 | **Auth** | `boosty.to` URL захардкожен | Config |
| 35 | **Payment** | `subscriptionDays` default 30 — проверить | Явно в config |
| 36 | **Authors** | UX gap: JWT старый после approve | Автоперевыпуск JWT или инструкция |
| 37 | **Analytics** | 5 субъектов в одном consumer’е | Разделить на 5 consumer’ов |
| 38 | **History** | 5 субъектов в одном consumer’е | Разделить на 5 consumer’ов |
| 39 | **Notification** | Retry Fixed Delay для Telegram | Exponential backoff |
| 40 | **Profile** | `UpdateProfile` наружу не открыт | Отдельный gRPC без `total_score` |

## Roadmap (не техдолг, но запланировано)

| # | Что | Зачем |
|---|-----|-------|
| 41 | i18n (RU/EN/DE/ES) | Смена языка сессии |
| 42 | Пагинация каталога | До 1000+ карт |
| 43 | Уровни лидерборда 2–3 | Сейчас все `level=1` |
| 44 | Порог покрытия в CI | ≥70% |
| 45 | Retry с джиттером, bulkhead | У Gateway уже CB |
| 46 | 108 авторов | Нет текстов/ассетов |
| 47 | 3D | Нет ассетов |
| 48 | Helm как повседневный деплой | Сейчас compose |
| 49 | Loki / Tempo | Не в `make deploy` |
| 50 | Пентест | Отчёт человека |

## Уже закрыто (v1.1.0–v1.1.3)

| # | Что | Версия |
|---|-----|--------|
| 1 | `/api/v1` публичный REST | v1.1.0 |
| 2 | Буст за лампы, достижения | v1.1.0 |
| 3 | Grafana provision fix | v1.1.0 |
| 4 | CORE k6 browse/purchase | v1.1.0 |
| 5 | `purchase.paid` с настоящим `purchases.id` | v1.1.1 |
| 6 | Spend Cycle — один ключ на цикл | v1.1.1 |
| 7 | Outbox `FOR UPDATE SKIP LOCKED` в Shop/Game | v1.1.1 |
| 8 | WS Ticket (не JWT в query) | v1.1.1 |
| 9 | Docs Gate (`DOCS_PASSWORD`) | v1.1.1 |
| 10 | Revenue Split 85/15 | v1.1.1 |
| 11 | Redis упал → 503 | v1.1.1 |
| 12 | Kafka молчит при живом JetStream | v1.1.1 |
| 13 | Grafana: paid → fulfilled, refund failures | v1.1.1 |
| 14 | Patterns catalog (333) | v1.1.3 |
| 15 | Frontend v1.1.2 polish | v1.1.3 |

## Итого

| Приоритет | Количество |
|-----------|-----------|
| P0 (блокеры) | 7 |
| P1 (надёжность) | 20 |
| P2 (гигиена) | 13 |
| Roadmap | 10 |
| Закрыто | 15 |

## Что делать прямо сейчас (P0)

1. **Fulfillment Outbox** — событие теряется.
2. **Fulfillment WMS** — заглушка.
3. **`runGateway`** — разбить.
4. **Shop Saga `started` > 60s** — дыра.
5. **`RewardConsumer` `time.Now()`** — двойное начисление.
6. **`RevertApplication` hard delete** — FK.
7. **Balancer single point** — резервный.