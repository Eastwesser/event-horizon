# Event Horizon — полная архитектура проекта (актуально 13.08.2026)

## Что это

Игровая/мерч платформа на Go-микросервисах: **React (:5173)** → **Balancer (:8079)** → **Gateway×3 (:8081–8083)** → доменные **gRPC**-сервисы. События: **NATS JetStream** (+ Kafka для purchase path). OLAP: **ClickHouse** (Analytics). Документация для агентов: Prydwen + MCP.

Это **реальный** учебный/пет-прод контур репозитория `event_horizon` — рассказывайте как hands-on проект.

## Топология (словами)

Клиент ходит в HTTP Gateway (Gin). Gateway валидирует JWT, RBAC, дергает gRPC downstream, на критичных вызовах — circuit breaker. Сервисы пишут в свои Postgres; кеш/сессии — Redis; асинхронщина — outbox → NATS subjects; History/Analytics/Fulfillment/Notification — consumers.

## Порты (host)

| Сервис | gRPC | Metrics | Postgres host | Redis / прочее |
|--------|------|---------|---------------|----------------|
| Auth | 50051 | 9091 | 5460 | Redis 6379 (sessions) |
| Game | 50052 | 9092 | 5461 | — |
| Billing | 50053 | 9093 | 5462 | 6381 |
| Leaderboard | 50054 | 9094 | 5463 | 6382 |
| Shop | 50055 | 9095 | 5465 | 6383 |
| Notification | 50056 | 9102 | — | — |
| Analytics | 50057 | 9106 | ClickHouse 8123/9000 | — |
| Payment | 50058 | 9103 | 5467 | 6386 |
| Inventory | 50059 | 9096 | 5466 | Redis 6384 + Mongo 27017 |
| Profile | 50060 | 9099 | 5464 | Redis |
| Authors | 50061 | 9104 | 5468 | 6387 |
| History | 50062 | 9105 | 5469 | — |
| Gateway | — | (см. compose) | — | HTTP 8081–8083 |
| Balancer | — | 9098 | — | HTTP 8079 |
| NATS | — | — | — | 4222–4224 |
| Prometheus / Grafana / Jaeger | 9090 / 3000 / 16686 | | | |

Health: metrics HTTP `/health` + `/ready` (ping зависимостей). Образы: `eastwesser/<svc>:latest`. Compose: `deployments/docker-compose.cluster.yml`.

## Clean Architecture

Слои: `internal/handler` (gRPC) → `internal/service` → `internal/repository` → `internal/model`. DI в `internal/app`. Зависимости не вверх. Interceptors: `Recovery`, `Logger`, `Validate`, role где нужно.

## Outbox (transactional messaging)

В одной PG-транзакции: бизнес-изменение + insert в `outbox`. После commit worker публикует в JetStream и помечает отправленным.

Где есть: **Shop, Billing, Inventory, Payment, Authors**.

Важные subjects: `shop.purchased`, `payment.completed`, `author.upserted`, `inventory.item.created`, `score.updated`, `user.registered`, `balance.updated` (+ ingest History/Analytics). Consumers должны быть идемпотентны (at-least-once).

## Redis

- Auth: sessions / refresh.
- Cache-Aside: Inventory `CachedRepository`; горячие чтения Profile/Shop/Billing/Leaderboard/Payment/Authors.
- Инвалидация после мутаций (shop keys, inventory update/delete).
- Game и History — без лишнего Redis (осознанно убран dead config).

## Payment / Authors / History / Analytics

- **Payment**: merch gate — Shop `PurchaseItem` при `category=merch` → Payment `CanPurchaseMerch`; outbox `payment.completed`.
- **Authors**: домен авторов, Redis cache, outbox `author.upserted`.
- **History**: Postgres trail событий (аудит/история пользователя).
- **Analytics**: ClickHouse HTTP, named params `{name:Type}` / `param_*` — без конкатенации user input; DAU/MAU/retention-стиль агрегаций.

## RBAC

JWT roles: `user` | `author` | `admin`. Gateway: `RequireAuth` + `RequireRole`. Inventory gRPC interceptor читает `x-user-role`. bcrypt cost **12** в Auth.

## Circuit breaker

Gateway `internal/circuit` на billing/shop/payment: после серии ошибок — open, быстрый **503**, half-open (Timeout≈10s, MaxRequests≈3). Fail-fast вместо каскадного таймаута.

## Optimistic locking

Колонка `version` на `inventory_items`, `user_currencies`, shop `items` — борьба с lost update.

## Saga-лайт (Shop purchase)

Spend в Billing → покупка со стоком; при фейле стока — компенсация `AddCurrency` + событие failure. Не полный orchestrator.

## MCP / Prydwen

`services/mcp` (stdio): `nats_list_streams`, `nats_list_consumers`, `postgres_query` (SELECT-only), `redis_get`/`redis_keys`, `search_prydwen` (TF-IDF RAG). База знаний — этот каталог.

## Пулы БД (стандарт)

MaxOpen/MaxConns=25, MinIdle/MinConns=10, MaxConnLifetime=5m. Секреты и порты — только env через `internal/config`.

## Как поднимать и что не путать

`docker compose -f deployments/docker-compose.cluster.yml up -d`. Источник правды по портам/паттернам — **этот файл** + compose + `services/*/internal/config`. Устаревшие заметки «Payment ещё план» / «Profile без Redis» игнорировать.

## Типичные вопросы на собесе

1. Зачем Outbox, если можно Publish после commit в сервисе?
2. Как у вас устроен RBAC от Gateway до gRPC Inventory?
3. Что происходит при открытом circuit breaker?
4. Почему Analytics в ClickHouse, а History в Postgres?
5. Как Shop связан с Payment для merch?
6. Cache-Aside в Inventory: инвалидация и stampedе?
7. Какие subjects NATS критичны для purchase flow?
8. Как MCP `postgres_query` защищён от мутаций?
9. Optimistic locking: какой ответ клиенту при конфликте version?
10. Чем ваш e2e/smoke отличается от unit на converters?
