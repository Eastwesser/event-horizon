ЧАСТЬ 2. Глубокий разбор ключевых сервисов
Ты спросил: «какие ещё прям ключевые?» Вот что мы не разобрали глубоко:

Сервис	Что не разобрали	Почему ключевой
Gateway	CB + Rate Limiter + WS Ticket sequence	Весь трафик через него
Shop + Inventory	Взаимодействие (inventory.item.created → shop.items)	Каталог, деньги
Balancer + Gateway	Least Connections + sticky WS	Front door
Auth + Roles	RBAC matrix (user/author/admin)	Все ручки
Billing + Shop	SpendCycle + Conditional UPDATE	Деньги
Разберу 4 самых ключевых: Gateway (CB + Rate Limiter + WS Ticket), Shop + Inventory (каталог), Auth + Roles (RBAC), Balancer + Gateway (sticky WS).

2.1. Gateway: CB + Rate Limiter + WS Ticket (sequence)
CB + Rate Limiter (sync path)
sequenceDiagram
  autonumber
  participant C as Client
  participant B as Balancer
  participant GW as Gateway
  participant RL as Rate Limiter (Redis)
  participant A as Auth
  participant CB as Circuit Breaker
  participant S as Shop (gRPC)

  C->>B: POST /api/v1/shop/purchase<br/>Authorization: Bearer
  B->>GW: HTTP (least connections)
  GW->>RL: INCR rate:{user}:{path}
  RL-->>GW: count=1
  alt count > limit
    GW-->>C: 429 Too Many Requests
  end

  GW->>A: gRPC ValidateToken
  A->>A: HMAC + jti в Redis
  alt invalid
    A-->>GW: Unauthenticated
    GW-->>C: 401
  end
  A-->>GW: user_id + role

  GW->>CB: call Shop.PurchaseItem
  alt CB Open
    CB-->>GW: 503 fail-fast
    GW-->>C: 503 service temporarily unavailable
  end

  CB->>S: gRPC PurchaseItem
  alt timeout / error
    S--x CB: error
    CB->>CB: fail count++
    alt fail count >= 5
      CB->>CB: Open
    end
    GW-->>C: 503
  else ok
    S-->>CB: ok
    CB->>CB: reset fail count
    CB-->>GW: ok
    GW-->>C: 200
  end
WS Ticket (async path)
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant A as Auth
  participant R as Redis
  participant H as WS Hub
  participant N as NATS

  Note over C,R: Шаг 1: получить билет
  C->>GW: POST /api/v1/ws-ticket<br/>Authorization: Bearer
  GW->>A: ValidateToken
  A-->>GW: user_id
  GW->>R: SETEX ws:ticket:{uuid} 30 1
  R-->>GW: ok
  GW-->>C: {ticket, expires_in: 30}

  Note over C,H: Шаг 2: открыть сокет
  C->>GW: GET /ws/leaderboard?ticket=...
  GW->>GW: Origin check (same host / localhost / allowlist)
  alt Origin bad
    GW-->>C: 403
  end
  GW->>R: GETDEL ws:ticket:{uuid}
  alt ticket not found
    R-->>GW: nil
    GW-->>C: 401
  end
  R-->>GW: ok (билет сгорел)
  GW->>H: upgrade + register client
  H-->>C: 101 Switching Protocols

  Note over N,H: Real-time updates
  N->>H: score.updated
  H->>H: broadcast to all WS clients
  H-->>C: JSON frame

  Note over C,H: Disconnect
  C--xH: close
  H->>H: unregister
Ключевые моменты:

Rate Limiter до Auth — дешёвый отказ.

CB после Auth — не тратим CB на неавторизованных.

WS Ticket: SETEX 30s → GETDEL (одноразовый). ?token= с JWT → 401.

Origin check до upgrade.

2.2. Shop + Inventory: каталог (sequence)
sequenceDiagram
  autonumber
  participant AU as Author
  participant GW as Gateway
  participant INV as Inventory
  participant PG as PG Inventory
  participant N as NATS
  participant SH as Shop
  participant PGS as PG Shop

  Note over AU,PGS: Автор создаёт карточку
  AU->>GW: POST /api/v1/inventory/items<br/>Authorization: Bearer
  GW->>GW: RequireRole(author|admin)
  GW->>INV: gRPC CreateItem<br/>metadata: x-user-role=author
  INV->>INV: RBAC interceptor: role ok
  INV->>PG: BEGIN
  INV->>PG: INSERT items (version=1)
  INV->>PG: INSERT outbox(inventory.item.created)
  INV->>PG: COMMIT
  INV->>INV: DEL cache list:*
  INV-->>GW: ok (item_id)
  GW-->>AU: 201 Created

  Note over INV,SH: Async: Inventory → Shop
  INV->>N: (outbox worker) publish inventory.item.created
  N->>SH: consume inventory.item.created<br/>(durable shop-inventory-sync)
  SH->>PGS: INSERT shop.items (id=item_id, price, stock)
  SH->>SH: DEL cache shop:items:*
  Note over SH: Карточка появилась в витрине

  Note over AU,SH: Игрок покупает
  AU->>GW: GET /api/v1/shop/items
  GW->>SH: gRPC GetItems
  SH->>SH: Cache-Aside: GET shop:items:hash
  alt cache hit
    SH-->>GW: items from Redis
  else cache miss
    SH->>PGS: SELECT * FROM shop.items
    SH->>SH: SETEX shop:items:hash 300
  end
  GW-->>AU: 200 {items}
Ключевые моменты:

Inventory — каталог автора. Shop — витрина.

inventory.item.created → Shop создаёт shop.items.

Прямого gRPC Inventory → Shop нет. Только через NATS.

Cache-Aside в обоих.

2.3. Auth + Roles: RBAC matrix
Матрица ролей
Ручка	user	author	admin
POST /auth/register	—	—	—
POST /auth/login	—	—	—
GET /auth/whoami	✓	✓	✓
POST /auth/update-nickname	✓	✓	✓
POST /auth/update-role	✗	✗	✓
GET /admin/users	✗	✗	✓
POST /game/submit	✓	✓	✓
POST /game/boost/start	✓	✓	✓
GET /billing/balance	✓	✓	✓
GET /shop/items	✓	✓	✓
POST /shop/purchase	✓	✓	✓
GET /shop/inventory	✓	✓	✓
GET /inventory/items	✓	✓	✓
POST /inventory/items	✗	✓	✓
PUT/DELETE /inventory/items/:id	✗	✓	✓
GET /inventory/stats	✗	✗	✓
GET /profile	✓	✓	✓
POST /payment/checkout	✓	✓	✓
GET /payment/subscription	✓	✓	✓
GET /payment/can-purchase-merch	✓	✓	✓
POST /payment/webhook	—	—	— (HMAC)
GET /authors	—	—	— (публичный)
GET /authors/:user_id	—	—	— (публичный)
GET /authors/me	✗	✓	✓
PUT /authors/me	✗	✓	✓
GET /authors/me/sales	✗	✓	✓
POST /authors/apply	✓	✓	✓
GET /authors/applications	✗	✗	✓
POST /authors/applications/:id/approve	✗	✗	✓
POST /authors/applications/:id/reject	✗	✗	✓
GET /history	✓	✓	✓
GET /analytics/dau|mau|retention	✗	✗	✓
GET /notifications	✓	✓	✓
POST /notifications/:id/read	✓	✓	✓
WS /ws/leaderboard	✓ (ticket)	✓	✓
Sequence: RBAC (update-role)
sequenceDiagram
  autonumber
  participant U as User
  participant GW as Gateway
  participant A as Auth
  participant PG as PG Auth
  participant I as Inventory

  Note over U,I: Admin меняет роль
  U->>GW: POST /api/v1/auth/update-role<br/>Authorization: Bearer (admin)
  GW->>A: ValidateToken
  A-->>GW: user_id, role=admin
  GW->>GW: RequireRole(admin) → ok
  GW->>A: gRPC UpdateRole(user_id=target, role=author)
  A->>PG: UPDATE users SET role='author'<br/>WHERE id=target
  PG-->>A: ok
  A-->>GW: ok
  GW-->>U: 200

  Note over U,I: После смены роли
  U->>GW: POST /api/v1/inventory/items<br/>Authorization: Bearer (target user)
  GW->>A: ValidateToken
  A-->>GW: user_id=target, role=user<br/>(JWT старый!)
  GW->>GW: RequireRole(author|admin) → 403
  GW-->>U: 403 insufficient permissions

  Note over U,I: UX gap: JWT старый<br/>Нужен refresh / re-login
  U->>GW: POST /api/v1/auth/refresh
  GW->>A: gRPC Refresh
  A->>PG: SELECT role FROM users WHERE id=target
  PG-->>A: role=author
  A->>A: выпускает новый access с role=author
  A-->>GW: new pair
  GW-->>U: 200 {access_token}

  U->>GW: POST /api/v1/inventory/items<br/>Authorization: Bearer (new)
  GW->>A: ValidateToken
  A-->>GW: role=author
  GW->>I: gRPC CreateItem<br/>metadata: x-user-role=author
  I->>I: RBAC interceptor: ok
  I-->>GW: ok
  GW-->>U: 201
Ключевые моменты:

update-role — только admin.

JWT старый после смены роли (UX gap). Refresh перечитывает роль из PG.

Inventory RBAC interceptor читает x-user-role из metadata.

2.4. Balancer + Gateway: sticky WS
sequenceDiagram
  autonumber
  participant C1 as Client 1
  participant C2 as Client 2
  participant B as Balancer :8079
  participant G1 as Gateway-1 :8081
  participant G2 as Gateway-2 :8082
  participant N as NATS
  participant R as Redis

  Note over C1,G2: HTTP: least connections
  C1->>B: GET /api/v1/leaderboard
  B->>B: ActiveConns: G1=3, G2=1 → select G2
  B->>G2: proxy
  G2-->>B: response
  B-->>C1: response

  Note over C1,G2: WS: sticky (same node)
  C1->>B: POST /ws-ticket
  B->>G2: proxy
  G2->>R: SETEX ws:ticket:1 30
  G2-->>C1: {ticket}

  C1->>B: GET /ws/leaderboard?ticket=...<br/>Upgrade: websocket
  B->>B: ActiveConns: G1=3, G2=0 → select G2
  B->>G2: proxy upgrade
  G2->>R: GETDEL ws:ticket:1
  G2->>G2: register client in Hub
  G2-->>C1: 101 Switching Protocols

  Note over C1,G2: Sticky: C1 держится на G2
  C2->>B: GET /ws/leaderboard?ticket=...
  B->>B: ActiveConns: G1=3, G2=1 → select G2
  B->>G2: proxy upgrade
  G2->>G2: register client
  G2-->>C2: 101

  Note over N,G2: Real-time: score.updated
  N->>G1: score.updated
  N->>G2: score.updated (durable gateway-websocket)
  G1->>G1: broadcast to G1 clients (none)
  G2->>G2: broadcast to C1, C2
  G2-->>C1: JSON frame
  G2-->>C2: JSON frame

  Note over B,G2: Gateway-2 упал
  B->>B: ActiveConns: G1=5, G2=0
  C1--xC2: connection lost
  Note over C1,C2: clients reconnect<br/>new ticket, new node
Ключевые моменты:

HTTP: least connections (G2 с меньшим ActiveConns).

WS: sticky — C1 и C2 на G2 (least connections на момент upgrade).

score.updated — durable consumer gateway-websocket на каждой Gateway. Все реплики получают событие, broadcast своим клиентам.

Gateway-2 упал → C1/C2 disconnect → reconnect (новый ticket, новый node).

ЧАСТЬ 3. Все 17 sequence-диаграмм (happy path)
3.1. Auth
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant A as Auth
  participant PG as PG
  participant R as Redis

  C->>GW: POST /auth/login {email, password}
  GW->>A: gRPC Login
  A->>PG: SELECT password_hash FROM users
  PG-->>A: hash
  A->>A: bcrypt.Compare (cost 12, ~200ms)
  A->>A: выпускает access (15m) + refresh (7d)
  A->>R: SETEX jti:access 900
  A->>R: SETEX jti:refresh 604800
  A-->>GW: {access, refresh}
  GW-->>C: 200

  C->>GW: GET /profile (Bearer)
  GW->>A: ValidateToken
  A->>R: EXISTS jti:access
  R-->>A: 1
  A-->>GW: user_id, role
  GW->>GW: RequireRole(user+) → ok
  GW-->>C: 200
3.2. Game
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant G as Game
  participant PG as PG Game
  participant B as Billing
  participant N as NATS

  C->>GW: POST /game/submit {game_id, score, moves}
  GW->>G: gRPC SubmitScore (user_id from token)
  G->>G: Strategy by game_id
  G->>G: Domain Validator (hexagon/memory)
  alt boost active
    G->>G: not ranked
  end
  G->>B: (if boost) SpendCurrency(lamp)
  B-->>G: ok
  G->>PG: BEGIN
  G->>PG: INSERT/UPDATE highscores
  G->>PG: INSERT outbox(score.updated)
  G->>PG: COMMIT
  G-->>GW: ok
  GW-->>C: 200

  G->>N: (outbox worker) publish score.updated
  N->>N: Leaderboard, Profile, Billing, History, Analytics, WS
3.3. Billing
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant B as Billing
  participant PG as PG Billing
  participant R as Redis
  participant N as NATS

  C->>GW: GET /billing/balance/all
  GW->>B: gRPC GetAllBalances
  B->>R: GET balance:{user}
  alt miss
    B->>PG: SELECT balance FROM user_currencies
    B->>R: SETEX balance:{user} 300
  end
  B-->>GW: {lamps, tickets}
  GW-->>C: 200

  Note over C,N: Spend (from Shop/Game)
  B->>PG: BEGIN
  B->>PG: UPDATE user_currencies<br/>SET balance = balance - $1<br/>WHERE user_id=$2 AND balance >= $1
  alt 0 rows
    B->>PG: ROLLBACK
    B-->>C: insufficient funds
  end
  B->>PG: INSERT transactions (reference_id unique, balance_after)
  B->>PG: INSERT outbox(balance.updated)
  B->>PG: COMMIT
  B->>R: DEL balance:{user}
  B->>N: (outbox worker) publish balance.updated
3.4. Leaderboard
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant L as Leaderboard
  participant R as Redis
  participant N as NATS

  N->>L: consume score.updated
  L->>R: ZADD leaderboard:{game}:{level} GT {score} {user_id}
  alt score > current
    R-->>L: 1 (added)
    L->>N: publish record_beaten (if other user)
  else score <= current
    R-->>L: 0 (skipped)
  end
  L->>L: Ack Batcher: accumulate acks
  L->>N: batch ack

  C->>GW: GET /leaderboard?game_id=hexagon&limit=10
  GW->>L: gRPC GetLeaderboard
  L->>R: ZREVRANGE leaderboard:hexagon:1 0 9 WITHSCORES
  R-->>L: top
  L-->>GW: top
  GW-->>C: 200
3.5. Shop
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant S as Shop
  participant PG as PG Shop
  participant B as Billing
  participant P as Payment
  participant N as NATS

  C->>GW: POST /shop/purchase {item_id}
  GW->>S: gRPC PurchaseItem
  alt category=merch
    S->>P: CanPurchaseMerch (sync)
    alt no subscription
      P-->>S: false
      S-->>GW: 403 subscription_required
    end
  end
  S->>PG: SELECT cycle FROM purchase_claims
  S->>PG: INSERT purchase_sagas (started)
  S->>B: SpendCurrency(spend_ref)
  B-->>S: ok
  S->>PG: UPDATE saga → spent
  S->>PG: BEGIN
  S->>PG: SELECT stock FOR UPDATE
  S->>PG: UPDATE stock -1 (version)
  S->>PG: INSERT inventory
  S->>PG: INSERT purchases
  S->>PG: INSERT outbox(shop.purchased, purchase.paid)
  S->>PG: COMMIT
  S->>PG: UPDATE saga → stocked
  S-->>GW: ok
  GW-->>C: 200

  S->>N: (outbox) publish shop.purchased + purchase.paid
3.6. Inventory
sequenceDiagram
  autonumber
  participant A as Author
  participant GW as Gateway
  participant I as Inventory
  participant PG as PG Inventory
  participant R as Redis
  participant N as NATS

  A->>GW: POST /inventory/items
  GW->>GW: RequireRole(author|admin)
  GW->>I: gRPC CreateItem (x-user-role)
  I->>I: RBAC interceptor
  I->>I: Strategy: PG or Mongo
  I->>PG: BEGIN
  I->>PG: INSERT items (version=1)
  I->>PG: INSERT outbox(inventory.item.created)
  I->>PG: COMMIT
  I->>R: DEL list:*
  I-->>GW: ok
  GW-->>A: 201

  I->>N: (outbox) publish inventory.item.created
  N->>N: Shop создаёт shop.items
3.7. Profile
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant P as Profile
  participant PG as PG Profile
  participant B as Billing
  participant G as Game
  participant R as Redis

  C->>GW: GET /profile
  GW->>P: gRPC GetProfile
  P->>R: GET profile:{user_id}
  alt miss
    P->>PG: SELECT profile
    P->>B: GetBalance (sync)
    alt Billing down
      P->>P: _partial = true
    end
    P->>G: GetBestScores (sync)
    P->>R: SETEX profile:{user_id} 300
  end
  P-->>GW: profile (+ _partial)
  GW-->>C: 200

  Note over P,PG: On score.updated
  N->>P: consume score.updated
  P->>P: Rule Engine EvaluateAchievements
  P->>PG: Pipeline: Evaluate → Filter → Unlock (ON CONFLICT)
  P->>R: DEL profile:{user_id}
3.8. Payment
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant P as Payment
  participant PG as PG Payment
  participant BB as Boosty
  participant N as NATS

  C->>GW: POST /payment/checkout {plan}
  GW->>P: gRPC CreateCheckout
  P->>PG: INSERT payments (pending)
  P-->>GW: {checkout_url}
  GW-->>C: 200

  C->>BB: pay
  BB->>P: POST /webhook (HMAC)
  P->>P: HMAC verify + ConstantTimeCompare
  P->>PG: BEGIN
  P->>PG: SELECT FOR UPDATE payment
  alt status=completed
    P->>PG: COMMIT
    P-->>BB: 200 (idempotent)
  end
  P->>PG: UPDATE payment completed
  P->>PG: UPDATE old subscription expires
  P->>PG: INSERT new subscription
  P->>PG: INSERT outbox(payment.completed)
  P->>PG: INSERT revenue_splits (85/15)
  P->>PG: COMMIT
  P-->>BB: 200
  P->>N: (outbox) publish payment.completed
3.9. Authors
sequenceDiagram
  autonumber
  participant U as User
  participant A as Admin
  participant GW as Gateway
  participant AU as Authors
  participant PG as PG Authors
  participant Auth as Auth
  participant N as NATS

  U->>GW: POST /authors/apply
  GW->>AU: gRPC Apply
  AU->>PG: check existing (idempotent)
  AU->>PG: INSERT author_applications (pending)
  AU-->>GW: ok
  GW-->>U: 200

  A->>GW: POST /authors/applications/:id/approve
  GW->>GW: RequireRole(admin)
  GW->>AU: gRPC Approve
  AU->>PG: SELECT FOR UPDATE application
  alt status != pending
    AU-->>GW: AlreadyExists
  end
  AU->>PG: BEGIN
  AU->>PG: UPDATE application approved
  AU->>PG: INSERT authors
  AU->>PG: INSERT outbox(author.upserted)
  AU->>PG: COMMIT
  AU->>Auth: UpdateRole(user_id, author)
  alt Auth fails
    AU->>AU: RevertApplication (compensating)
    AU-->>GW: error
  end
  AU-->>GW: ok
  GW-->>A: 200
  AU->>N: (outbox) publish author.upserted
3.10. History
sequenceDiagram
  autonumber
  participant N as NATS
  participant H as History
  participant PG as PG History

  Note over N,PG: 5 subjects
  N->>H: score.updated
  N->>H: user.registered
  N->>H: shop.purchased
  N->>H: payment.completed
  N->>H: author.upserted

  loop each event
    H->>H: Payload Sanitization (invalid → raw)
    H->>PG: INSERT history ON CONFLICT DO NOTHING
    alt db error
      H->>N: Nak (retry)
    else ok
      H->>N: Ack
    end
  end

  Note over H,PG: RetentionWorker (cron)
  H->>PG: DELETE FROM history WHERE created_at < now()-30d LIMIT 10000
3.11. Analytics
sequenceDiagram
  autonumber
  participant N as NATS
  participant A as Analytics
  participant CH as ClickHouse
  participant Admin as Admin
  participant GW as Gateway

  N->>A: 5 subjects
  loop each event
    A->>A: buffer (batch)
  end
  A->>CH: INSERT INTO events VALUES (batch)
  A->>N: Ack

  Admin->>GW: GET /analytics/dau
  GW->>GW: RequireRole(admin)
  GW->>A: gRPC GetDAU
  A->>CH: SELECT count(DISTINCT user_id) FROM events
  CH-->>A: dau
  A-->>GW: {dau}
  GW-->>Admin: 200
3.12. Notification
sequenceDiagram
  autonumber
  participant N as NATS
  participant K as Kafka (fallback)
  participant No as Notification
  participant PG as PG Notification
  participant T as Telegram
  participant C as Client
  participant GW as Gateway

  Note over N,K: Dual-Bus (NATS primary, Kafka fallback)
  N->>No: score.updated / purchase.paid / etc.
  No->>No: Fan-Out (1 → N)
  No->>PG: INSERT notifications<br/>ON CONFLICT (user_id, source_event) DO NOTHING
  alt Telegram subscribed
    No->>T: send message
    alt Telegram fails
      No->>No: Retry Fixed Delay
    end
  end
  No->>N: Ack

  C->>GW: GET /notifications
  GW->>No: gRPC ListNotifications
  No->>PG: SELECT
  No-->>GW: list
  GW-->>C: 200
3.13. Fulfillment
sequenceDiagram
  autonumber
  participant N as NATS
  participant K as Kafka (fallback)
  participant F as Fulfillment
  participant S as Shop
  participant No as Notification

  Note over N,K: Dual-Transport (NATS primary, Kafka fallback)
  N->>F: purchase.paid
  F->>F: Idempotency Guard (map)
  F->>F: time.After(10s) — эмуляция
  F->>N: publish purchase.fulfilled
  N->>S: consume purchase.fulfilled
  S->>S: UPDATE purchases SET fulfilled_at
  N->>No: consume purchase.fulfilled
  No->>No: in-app bell
3.14. Gateway
sequenceDiagram
  autonumber
  participant C as Client
  participant B as Balancer
  participant GW as Gateway
  participant RL as Rate Limiter
  participant A as Auth
  participant CB as Circuit Breaker
  participant S as Service

  C->>B: HTTP
  B->>GW: proxy (least conn)
  GW->>RL: INCR rate
  alt over limit
    GW-->>C: 429
  end
  GW->>A: ValidateToken
  alt invalid
    GW-->>C: 401
  end
  GW->>GW: RequireRole
  alt wrong role
    GW-->>C: 403
  end
  GW->>CB: call
  alt CB open
    GW-->>C: 503
  end
  CB->>S: gRPC
  alt error
    CB->>CB: fail++
  else ok
    CB->>CB: reset
  end
  S-->>GW: response
  GW->>GW: gRPC → HTTP mapping
  GW-->>C: response
3.15. Balancer
sequenceDiagram
  autonumber
  participant C as Client
  participant B as Balancer
  participant G1 as Gateway-1
  participant G2 as Gateway-2
  participant G3 as Gateway-3

  C->>B: HTTP
  B->>B: read ActiveConns (atomic)
  alt G2 has min
    B->>G2: proxy
    B->>B: atomic Add(G2.ActiveConns, 1)
    G2-->>B: response
    B->>B: atomic Add(G2.ActiveConns, -1)
    B-->>C: response (X-Balancer-Backend: gateway-2)
  end
3.16. NATS Hub
sequenceDiagram
  autonumber
  participant NH as NATS Hub
  participant N as NATS cluster

  NH->>N: connect
  N-->>NH: ok
  NH->>N: js.AddStream(EVENTS)
  alt already exists
    N-->>NH: AlreadyExists
    NH->>N: js.UpdateStream(EVENTS)
    N-->>NH: ok
  else created
    N-->>NH: ok
  end
  NH->>N: js.Subscribe("event.>")
  N-->>NH: ok
  NH->>NH: /health → nc.IsConnected()
3.17. MCP
sequenceDiagram
  autonumber
  participant Cu as Cursor
  participant M as MCP
  participant PG as Postgres
  participant R as Redis
  participant N as NATS
  participant P as Prydwen

  Note over Cu,M: stdio (JSON-RPC)
  Cu->>M: tools/list
  M-->>Cu: 6 tools

  Cu->>M: tools/call search_prydwen {query}
  M->>P: TF-IDF search
  P-->>M: top-N
  M-->>Cu: results

  Cu->>M: tools/call postgres_query {sql}
  M->>M: ValidateSelect (no UNION)
  alt invalid
    M-->>Cu: error
  end
  M->>PG: SELECT (read-only)
  PG-->>M: rows
  M-->>Cu: rows

  Cu->>M: tools/call nats_list_streams
  M->>N: js.Streams()
  N-->>M: streams
  M-->>Cu: streams

  Cu->>M: tools/call redis_get {key}
  M->>R: GET
  R-->>M: value
  M-->>Cu: value
Итого: три артефакта
Артефакт	Что внутри
1. Техдолг без дубликатов	7 P0, 20 P1, 13 P2, 10 roadmap, 15 закрыто. Убраны дубликаты. Добавлена Shop Saga started > 60s (P0), Balancer single point (P0), сверка webhook (P1), ротация JWT (P1), WAF (P1)
2. Ключевые сервисы глубоко	Gateway (CB + Rate Limiter + WS Ticket), Shop + Inventory (каталог), Auth + Roles (RBAC matrix), Balancer + Gateway (sticky WS)
3. 17 sequence-диаграмм	Все happy path: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Notification, Fulfillment, Gateway, Balancer, NATS Hub, MCP