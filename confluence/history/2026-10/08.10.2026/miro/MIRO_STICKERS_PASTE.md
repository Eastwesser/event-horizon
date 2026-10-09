# Miro stickers — paste sheet (09.10.2026)

## Где живёт `/api/v1`?

**Только на HTTP Gateway** (`services/gateway`).  
Микросервисы говорят **gRPC** (`:5005x`) — у них **нет** своих HTTP `/api/v1/...`.

Сделано: все Gin-ручки → `/api/v1/...` + legacy rewrite `/api/*` → `/api/v1/*`.  
FE: `baseURL: '/api/v1'`. OpenAPI / k6 — тоже.

На стикерах сервисов пиши **публичные HTTP-ручки через GW**, не «ручки внутри gRPC».

---

## MCP — сделан ли?

**Да.** `services/mcp` + бинарь `mcp-server` + `.cursor/mcp.json`.

| Что | Факт |
|-----|------|
| Транспорт | **stdio** (Cursor / Claude Desktop), не HTTP за LB |
| Tools | `search_prydwen` (RAG), `nats_list_*`, `postgres_query` (SELECT-only), `redis_get` / `redis_keys` |
| Runtime игроков | **нет** — клиент в MCP не ходит |

**На Miro:** оставь блок MCP, но стрелка **не** Client→GW→MCP.  
Правильно: `Cursor / agent` → `MCP (stdio)` → NATS / PG(read) / Redis / Prydwen RAG.  
Подпись: `tools: search_prydwen, nats_*, postgres_query, redis_*` · `not in player path`.

---

## Порты — сверка с compose (исправь стикеры)

| Сервис | gRPC | PG host | Redis host | Mongo |
|--------|------|---------|------------|-------|
| Auth | **50051** | **5460** | **6379** | — |
| Game | **50052** | **5461** | **6380** | — |
| Billing | **50053** | **5462** | **6381** | — |
| Leaderboard | **50054** | **5463** | **6382** | — |
| Shop | **50055** | **5465** | **6383** | — |
| Analytics | **50057** | ClickHouse **8123/9000** | — | — |
| Payment | **50058** | **5467** | **6386** | — |
| Inventory | **50059** | **5466** (не 5446) | **6384** (не 6364) | **27017** (compose inventory; polyglot) |
| Profile | **50060** | **5464** | **6385** | — |
| Authors | **50061** | **5468** | **6387** | **нет Mongo** |
| History | **50062** | **5469** | — | — |
| Notification | **50063** (не 50056) | **5470** | — | — |
| Fulfillment | worker | (своя/нет публичного gRPC в таблице) | opt Kafka | — |
| LB (nginx) | HTTP **8079** | — | — | — |
| Gateway ×3 | HTTP **8081–8083** | — | Redis limiter shared | — |

---

## Outbox — где рисовать цилиндр

### Есть Outbox (+ worker)
- Game  
- Billing  
- Shop  
- Inventory  
- Payment  
- Authors  

### Нет Outbox — **убери** цилиндр с Miro
- Auth  
- Profile  
- Leaderboard  
- History (consumer NATS → PG)  
- Analytics (NATS → ClickHouse)  
- Notification (consumer → PG inbox)  
- Fulfillment (NATS/Kafka consumer; **нет** таблицы outbox в коде)  
- Gateway / Balancer / MCP  

---

## Порядок блоков на доске (нумерацию свою оставь)

Слева → направо / сверху → вниз по смыслу edge→core→async:

1. Client → LB → GW  
2. Auth  
3. Game → Billing → Leaderboard → Profile  
4. Shop → Inventory → Authors  
5. Payment → History → Analytics → Notification  
6. Fulfillment (+ opt Kafka)  
7. NATS hub по центру  
8. MCP сбоку у Cursor (не в player path)  
9. Observability  

---

## API routes — копипаст на стикеры

### LOAD BALANCER
```
http: 8079
metrics: 9098
proxy: /api/v1/* → GW
proxy: /ws/leaderboard → GW
```

### API GW
```
http: 8081-8083
prefix: /api/v1/*
also: /health /ready /docs /openapi.yaml
ws: /ws/leaderboard
legacy: /api/* → /api/v1/*
```

### AUTH
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/whoami
GET  /api/v1/auth/user
POST /api/v1/auth/update-nickname
POST /api/v1/auth/update-role   (admin)
GET  /api/v1/admin/users        (admin; via GW→Auth)
```

### GAME
```
POST /api/v1/game/submit
POST /api/v1/game/boost/start
```

### BILLING
```
GET /api/v1/billing/balance/all
```

### LEADERBOARD
```
GET /api/v1/leaderboard
WS  /ws/leaderboard
```

### PROFILE
```
GET /api/v1/profile
```

### SHOP
```
GET  /api/v1/shop/items
POST /api/v1/shop/purchase
POST /api/v1/shop/purchase/:id/cancel
GET  /api/v1/shop/inventory
```

### INVENTORY
```
GET    /api/v1/inventory/items
POST   /api/v1/inventory/items
POST   /api/v1/inventory/items/bulk
GET    /api/v1/inventory/items/:id
PUT    /api/v1/inventory/items/:id
DELETE /api/v1/inventory/items/:id
POST   /api/v1/inventory/items/:id/reserve
DELETE /api/v1/inventory/items/:id/soft
POST   /api/v1/inventory/items/:id/restore
GET    /api/v1/inventory/stats
POST   /api/v1/uploads
```

### AUTHORS
```
GET  /api/v1/authors
GET  /api/v1/authors/:user_id
PUT  /api/v1/authors/me
GET  /api/v1/authors/me
GET  /api/v1/authors/me/sales
POST /api/v1/authors/apply
GET  /api/v1/authors/me/application
GET  /api/v1/authors/applications
POST /api/v1/authors/applications/:id/approve
POST /api/v1/authors/applications/:id/reject
```
(PG+Redis only · **no Mongo**)

### PAYMENT
```
POST /api/v1/payment/checkout
GET  /api/v1/payment/subscription
GET  /api/v1/payment/can-purchase-merch
POST /api/v1/payment/webhook
POST /api/v1/payment/yookassa/webhook
```

### HISTORY
```
GET /api/v1/history
```

### ANALYTICS
```
GET /api/v1/analytics/dau
GET /api/v1/analytics/mau
GET /api/v1/analytics/retention
```

### NOTIFICATION
```
GET  /api/v1/notifications
POST /api/v1/notifications/:id/read
gRPC: 50063
```

### FULFILLMENT
```
(no public /api/v1 — NATS/Kafka consumer only)
```

### MCP (Cursor tooling)
```
stdio (not /api/v1)
tools: search_prydwen, nats_*, postgres_query, redis_*
```
