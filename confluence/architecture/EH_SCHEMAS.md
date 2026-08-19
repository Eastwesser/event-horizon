# Event Horizon — схемы архитектуры (v1.0.7)

Источник: Miro `SYSTEM_DESIGN/event-horizon-v1.0.6.png` + **[`SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md`](./SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md)** (полные Mermaid-диаграммы) + актуальный `docker-compose.cluster.yml` / gateway.

Стрелки: **сплошные** = sync (HTTP/gRPC), **пунктир** = async (NATS/Kafka/Outbox/WS).

---

## 1. Entry path (клиент → сервисы)

```text
[React Client :5173]
        │ HTTP (JSON)
        ▼
[Balancer :8079]  Least Connections
        │ HTTP
        ▼
[Gateway ×3 :8081–8083]  JWT · HTTP→gRPC · circuit breaker (all gRPC clients)
        │ gRPC
        ▼
   микросервисы (см. блок 2)
```

Gateway также: `GET /openapi.yaml`, `GET /docs`, `WS /ws/leaderboard`.

---

## 2. Микросервисы + хранилища (актуальные порты)

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│ CORE                                                                        │
│  Auth :50051 ── PG :5460 · Redis :6379 (sessions)                           │
│  Game :50052 ── PG :5461                                                    │
│  Billing :50053 ── PG :5462 · Redis :6381 · version (optimistic lock)       │
│  Leaderboard :50054 ── PG :5463 · Redis :6382 (Sorted Set)                  │
│  Shop :50055 ── PG :5465 · Redis :6383 · Outbox → NATS/Kafka                │
│  Inventory :50059 ── PG (+ Mongo legacy) · Redis · Outbox                   │
│  Profile :50060 ── PG :5464 · NATS consumer (score/user events)             │
├─────────────────────────────────────────────────────────────────────────────┤
│ STAGE-1 / COMMERCE & CONTENT                                                │
│  Payment :50058 ── metrics :9103 · subscription / merch gate                │
│  Authors :50061 ── PG :5468 · Redis :6387 · metrics :9104                   │
│  History :50062 ── PG :5469 · retention 30d · metrics :9105                 │
│  Analytics :50057 ── ClickHouse :8123/:9000 · metrics :9106                 │
├─────────────────────────────────────────────────────────────────────────────┤
│ PLATFORM                                                                    │
│  Notification · Fulfillment · NATS Hub (Stream EVENTS)                      │
│  MCP server (local Cursor / RAG over Prydwen)                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Сервис | gRPC | Metrics | Store |
|--------|------|---------|-------|
| Auth | 50051 | 9091 | PG 5460, Redis 6379 |
| Game | 50052 | 9092 | PG 5461 |
| Billing | 50053 | 9093 | PG 5462, Redis 6381 |
| Leaderboard | 50054 | 9094 | PG 5463, Redis 6382 |
| Shop | 50055 | 9095 | PG 5465, Redis 6383 |
| Analytics | 50057 | 9106 | ClickHouse 8123/9000 |
| Payment | 50058 | 9103 | (subscription store per deploy) |
| Inventory | 50059 | 9096 | PG / Mongo, Redis |
| Profile | 50060 | 9099 | PG 5464 |
| Authors | 50061 | 9104 | PG 5468, Redis 6387 |
| History | 50062 | 9105 | PG 5469 |
| Gateway | HTTP 8081–8083 | 9095–9097 | — |
| Balancer | HTTP 8079 | 9098 | — |

---

## 3. Событийная шина

```text
[Shop / Inventory / Game / …]
        ╎ Outbox / publish (async)
        ▼
┌──────────────────────────────────────────────────────────┐
│  NATS JetStream cluster :4222 / :4223 / :4224            │
│  Stream EVENTS (NATS Hub)                                │
│  subjects: score.updated, user.registered, shop.*, …     │
└──────────────────────────────────────────────────────────┘
        ╎ subscribe
        ▼
[Profile] [Leaderboard] [Notification] [Fulfillment] …

Опционально / purchase path: Kafka (см. contracts/events/PURCHASE_KAFKA.md)
```

Leaderboard real-time:

```text
Game ──(async)──► NATS ──(async)──► Leaderboard
                                      │ Redis Sorted Set
                                      │ WS push
                                      ▼
                               [WebSocket :8080 / via Gateway]
                                      │
                                      ▼
                               [React Client]
```

---

## 4. Ключевые sync-потоки (бизнес)

```text
Shop purchase (tickets)
  Gateway → Shop gRPC → Billing (debit) → Inventory / Outbox

Shop merch
  Gateway → Shop → Payment.CanPurchaseMerch → (if ok) fulfill

Analytics (admin)
  Gateway → Analytics gRPC → ClickHouse (named params)

Authors / History
  Gateway → Authors | History gRPC → PG
```

---

## 5. Observability

| Компонент | Порт | Назначение |
|-----------|------|------------|
| Prometheus | 9090 | scrape `/metrics` |
| Grafana | 3000 | dashboards |
| Jaeger | 16686 | traces |
| Alertmanager | (compose) | alerts |
| NATS Exporter | 7777 | NATS metrics |

Каждый сервис: `/health` + `/ready` на metrics HTTP.

---

## 6. Frontend vs backend (разрыв)

Уже в UI: Auth, Game(s), Billing, Shop, Inventory, Leaderboard, Profile, Towers(=Башенки).

Клиенты есть, UI нет: Payment, Authors, Analytics → см. `FRONTEND_REMAINING.md`.

Игра **Ханойская башня** — отдельная от Башенок; спека `FRONTEND_KHANOY_TOWERS.md`.

---

## 7. Легенда для Miro

| Элемент | Форма |
|---------|--------|
| React Client | шестиугольник |
| Balancer / Gateway | прямоугольник |
| Микросервис | скруглённый прямоугольник |
| PostgreSQL / ClickHouse | цилиндр |
| Redis | ромб |
| NATS / Kafka bus | шестиугольник / шина |
| WebSocket | пунктир WS |

Цвета: синий клиент · зелёный LB/GW · жёлтый сервисы · оранжевый шина.

---

## 8. Собес-резюме

Gateway — единая HTTP-точка: JWT, map на gRPC, OpenAPI/docs.  
Сервисы не ходят друг к другу sync «как попало» — деньги/мерч через явные gRPC вызовы из Shop/Gateway; события — через NATS/Outbox.  
Топ — Redis Sorted Set + WS.  
Analytics — ClickHouse; Payment — подписка и gate на merch; Authors/History — контент и аудит.
