Event Horizon — сводный пакет для собеса
Три артефакта: сводная таблица, 10-минутный рассказ, одна Mermaid-схема на все 17 сервисов.

ЧАСТЬ 1. Сводная таблица «сервис → порт → ручки → паттерны → SLO»
1.1. Полная таблица
#	Сервис	gRPC	Metrics	Хранилища	Публичные ручки	Паттерны	SLO (цель)
1	Auth	:50051	:9091	PG :5460, Redis :6379	POST /register, /login, /refresh, /logout, /update-nickname, /update-role; GET /whoami, /user, /admin/users	13	p95 login < 500ms, error < 1%
2	Game	:50052	:9092	PG :5461 (Redis :6380 не подключён)	POST /game/submit, /boost/start, /companion/daily-gift	20	p95 submit < 500ms, outbox_pending не растёт
3	Billing	:50053	:9093	PG :5462, Redis :6381	GET /billing/balance/all, /balance, /transactions	24	p95 spend < 200ms, deadlocks = 0
4	Leaderboard	:50054	:9094	PG :5463, Redis :6382	GET /leaderboard (публичный), WS /ws/leaderboard	20	p95 top < 50ms, consumer_lag не растёт
5	Shop	:50055	:9095	PG :5465, Redis :6383	GET /shop/items, /shop/inventory; POST /shop/purchase, /purchase/:id/cancel	30	p95 purchase < 500ms, refund_failures не растёт
6	Inventory	:50059	:9096	PG :5466, Redis :6384, Mongo :27017 (experimental)	GET/POST/PUT/DELETE /inventory/items, bulk, reserve, soft, restore, stats	28	p95 create < 300ms, p95 get < 50ms
7	Profile	:50060	:9099	PG :5464, Redis :6385	GET /profile (агрегат)	14	p95 get < 100ms, _partial не растёт
8	Payment	:50058	:9103	PG :5467, Redis :6386	POST /payment/checkout, /webhook, /yookassa/webhook; GET /subscription, /can-purchase-merch	21	p95 webhook < 500ms, webhook_invalid низкий
9	Authors	:50061	:9104	PG :5468, Redis :6387	GET /authors, /:user_id (публичные), /me, /me/sales; PUT /me; POST /apply, /applications/:id/approve, /reject	20	p95 apply < 200ms, p95 approve < 500ms
10	History	:50062	:9105	PG :5469	GET /history (своя лента)	16	p95 get < 100ms, rows_total не растёт бесконечно
11	Analytics	:50057	:9106	ClickHouse :8123/:9000	GET /analytics/dau, /mau, /retention (admin)	14	p95 dau < 500ms, p95 mau < 1s
12	Notification	:50063	:9102	PG :5470	GET /notifications; POST /notifications/:id/read	16	p95 list < 100ms, telegram_failed низкий
13	Fulfillment	нет	:9101	нет	нет (NATS/Kafka only)	15	p95 assembly < 12s, consumer_lag не растёт
14	Gateway	HTTP :8081–8083	:9095–9097 (внутри)	Redis :6379	все /api/v1/* (~50+) + /ws/leaderboard + /docs	41	p95 < 200ms, p99 < 500ms, CB state = 0
15	Balancer	HTTP :8079	:9098	нет	Проксирует всё в Gateway; /health, /ready, /metrics	15	p95 < 50ms, active_connections > 0
16	NATS Hub	нет	:9097	NATS :4222/4223/4224	/health, /ready, /metrics	11	connected = 1, stream_exists = 1
17	MCP	stdio	нет	PG (read-only), Redis (read-only), NATS (read-only), Prydwen (TF-IDF)	6 tools (не HTTP)	15	tool call p95 < 1s
1.2. Сводка по портам
gRPC-сервисы
text
Auth        50051 | Game      50052 | Billing   50053 | Leaderboard 50054
Shop        50055 | Analytics 50057 | Payment   50058 | Inventory   50059
Profile     50060 | Authors   50061 | History   50062 | Notification 50063
Metrics (на хосте)
text
Auth        9091 | Game      9092 | Billing   9093 | Leaderboard 9094
Shop        9095 | Inventory 9096 | NATS Hub  9097 | Balancer    9098
Profile     9099 | Fulfillment 9101 | Notification 9102
Payment     9103 | Authors   9104 | History   9105 | Analytics   9106
Важно: Gateway :9095/:9096/:9097 — внутри контейнера, на хост не проброшены. На хосте это Shop/Inventory/NATS Hub.

HTTP edge
text
Balancer    :8079  (front door)
Gateway     :8081, :8082, :8083  (×3 реплики)
React       :5173  (dev)
PostgreSQL
text
Auth         5460 | Game        5461 | Billing    5462 | Leaderboard 5463
Profile      5464 | Shop        5465 | Inventory  5466 | Payment     5467
Authors      5468 | History     5469 | Notification 5470
Redis
text
Auth         6379 (общий с Gateway) | Game 6380 (не подключён)
Billing      6381 | Leaderboard 6382 | Shop 6383 | Inventory 6384
Profile      6385 | Payment     6386 | Authors 6387
Инфра
text
NATS        4222 / 4223 / 4224 (3 ноды)
ClickHouse  8123 (HTTP) / 9000 (native)
Prometheus  9090 | Grafana 3000 | Jaeger 16686 | Alertmanager 9193
Kafka       19092 (heavy, opt-in) | Mongo 27017 (experimental)
1.3. Сводка по паттернам
Паттерн	Кто	Что значит
Saga	Shop	purchase_sagas, единственный с явной сагой
Outbox	Game, Billing, Shop, Inventory, Payment, Authors	В одной TX с бизнес-строкой
Circuit Breaker + Bulkhead	Gateway	Единственный, fail-fast 503
WebSocket Ticket	Gateway	Одноразовый билет вместо JWT в query
Strategy	Game (по играм), Inventory (PG/Mongo)	Выбор поведения
Decorator	Inventory (CachedRepository)	Cache-Aside поверх Repository
Rule Engine	Profile (EvaluateAchievements)	Чистая функция
State Machine	Authors (pending → approved / rejected)	Заявки
HMAC + ConstantTimeCompare	Payment	Webhook Boosty
Revenue Split	Payment	85/15 — журнал, не перевод
Retention Policy	History	30 дней, RetentionWorker
Cohort Retention	Analytics	Классический SQL в ClickHouse
Dual-Bus	Notification	NATS primary, Kafka fallback
Dual-Transport	Fulfillment	NATS primary, Kafka fallback
Least Connections	Balancer	Atomic counter
Bootstrap	NATS Hub	Единственный, создаёт stream
MCP Server	MCP	6 tools, stdio
1.4. Сводка по брокерам
Брокер	Кто публикует	Кто потребляет
NATS score.updated	Game	Leaderboard, Profile, Billing, History, Analytics, Gateway WS
NATS user.registered	Auth/Gateway	History, Profile, Analytics
NATS shop.purchased	Shop	History, Analytics
NATS purchase.paid	Shop	Fulfillment, Notification
NATS purchase.fulfilled	Fulfillment	Shop, Notification
NATS payment.completed	Payment	History, Analytics, Notification
NATS author.upserted	Authors	History, Analytics
NATS balance.updated	Billing	(интересующиеся)
NATS record_beaten	Leaderboard	Notification
NATS inventory.item.created	Inventory	Shop
Kafka purchase.paid	Shop (heavy)	Fulfillment, Notification
1.5. Сводка по хранилищам
Хранилище	Кто	Что хранит
Postgres	13 сервисов	Source of truth
Redis	Auth, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors	Сессии, кэш, Sorted Set, rate limit
ClickHouse	Analytics	OLAP, DAU/MAU/retention
Mongo	Inventory (experimental)	Гибкие атрибуты
NATS	Все (через NATS Hub)	Stream EVENTS
Kafka	Shop, Fulfillment, Notification (heavy)	purchase.paid, purchase.fulfilled
ЧАСТЬ 2. Как рассказать про Event Horizon за 10 минут
2.1. Структура (10 минут)
Минута	Что	Кому
0:00–1:00	Elevator pitch + масштаб	Всем
1:00–3:00	Edge + sync path (Schema A)	Всем
3:00–5:00	Event spine + Outbox (Schema B)	Всем
5:00–7:00	Polyglot persistence (Schema C)	Всем
7:00–8:30	Money path + Saga (Schema D)	Senior+
8:30–10:00	Ops + trade-offs	Всем
2.2. Текст по минутам
0:00–1:00. Elevator pitch + масштаб
«Event Horizon — игровая микросервисная платформа на Go + React. 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд. 17 сервисов, Clean Architecture, gRPC mesh, NATS JetStream, ClickHouse. Целевая модель — ~10k DAU, десятки RPS в среднем, пик ×2. Не 100k RPS — это осознанно: я проектирую под correctness и ops, а не под фантазию. Локально измерено: 20 VU browse PASS, ~50 RPS, p95 429ms на 8 GiB VM.»

Ключевая фраза: «10k DAU — это daily active users, не 10k concurrent VUs. На 8 GiB VM 10k VUs убьют стек. Модель требует 8c/32G.»

1:00–3:00. Edge + sync path (Schema A)
Рисуй:

text
React :5173 → Balancer :8079 → Gateway ×3 :8081–8083 → gRPC-сервисы
«Клиент бьёт в Balancer :8079 — самописный L7, Least Connections, atomic counter. Дальше Gateway ×3 :8081–8083 — единственная публичная HTTP-точка. Gateway проверяет JWT через Auth ValidateToken на каждом запросе, кладёт user_id + role в контекст. Дальше — gRPC в доменные сервисы. На каждом gRPC-клиенте — Circuit Breaker + Bulkhead: после серии ошибок — быстрый 503, half-open через 10s. Это единственный сервис с CB. Плюс Retry с full jitter и Sliding Window Rate Limiter: login 5/s IP (fail-closed), общий ~100/s (fail-open).»

Ключевая фраза: «Gateway не разбирает подпись сам — зовёт Auth. Это лишний хоп, но Auth — единственная точка правды о сессии. Отзыв мгновенный.»

3:00–5:00. Event spine + Outbox (Schema B)
Рисуй:

text
Service TX: UPDATE + INSERT outbox → COMMIT → worker → NATS → consumers
«Асинхронщина — через NATS JetStream, stream EVENTS. Сервисы не публикуют напрямую после commit — это dual-write проблема. Вместо этого Transactional Outbox: бизнес-строка + outbox-строка в одной транзакции. Воркер публикует после commit, FOR UPDATE SKIP LOCKED, mark after publish. Outbox есть у Game, Billing, Shop, Inventory, Payment, Authors. Шесть сервисов. Ключевые субъекты: score.updated (Game → 6 потребителей), purchase.paid (Shop → Fulfillment + Notification), payment.completed (Payment → History + Analytics).»

Ключевая фраза: «Не publish в NATS внутри транзакции — это внешний IO, держит соединение. Outbox пишет в БД атомарно с бизнес-строкой, воркер публикует позже. Дубликаты возможны — consumer идемпотентный.»

5:00–7:00. Polyglot persistence (Schema C)
Рисуй:

text
Postgres (OLTP) | Redis (cache) | ClickHouse (OLAP) | Mongo (experimental)
«Четыре хранилища. Postgres — source of truth, 13 сервисов, у каждого своя база. Redis — сессии, кэш балансов, Sorted Set для лидерборда, rate limit, WS ticket. ClickHouse — единственный сервис Analytics, DAU/MAU/retention, колоночный OLAP. Mongo — experimental в Inventory, гибкие атрибуты. Anti-pattern, который избегаю: не ставить MAU SQL на History Postgres. OLTP и OLAP — разные движки.»

Ключевая фраза: «Redis никогда не единственная копия. Заказ, баланс, склад — только Postgres. Redis — кэш.»

7:00–8:30. Money path + Saga (Schema D)
Рисуй:

text
Shop → Billing (Spend) → Shop TX (stock + purchase + outbox) → NATS
                    ↓ fail
              Compensate (AddCurrency)
«Деньги — самая интересная часть. Shop покупка — сага на две базы: списание в Billing (Postgres billing), склад и purchases в Shop (Postgres shop). Общей транзакции нет. Строка purchase_sagas — оркестратор: started → spent → stocked, либо spent → compensating → compensated. Ключ списания один на цикл: shop:{user}:{item}:{cycle} — два клика делят один ключ. Если процесс умер после списания — воркер resumeSagas через 30s дочитывает. purchase.paid в outbox в той же транзакции, воркер публикует FOR UPDATE SKIP LOCKED.»

Ключевая фраза: «Не 2PC — он ломает доступность. Сага даёт eventual business consistency. Компенсация идемпотентна по ключу.»

8:30–10:00. Ops + trade-offs
Рисуй:

text
Prometheus :9090 | Grafana :3000 | Jaeger :16686 | Alertmanager :9193
«Observability: Prometheus скрейпит /metrics, Grafana — дашборды, Jaeger — трейсы через OTLP :4317, Alertmanager — алерты. Каждый сервис: /health + /ready. SOS-дашборд: P99 latency, 5xx rate, NATS consumer lag, free RAM. Три trade-off, которые назову честно: Auth write path — узкое место by design (bcrypt 12 + rate limit). runGateway — God Function > 1000 строк, техдолг. MCP SQL Sandbox — string matching, легко обходится, техдолг. Что улучшил бы: разбить runGateway, Outbox в Fulfillment, health checks в Balancer.»

Ключевая фраза: «Я не обещаю 10k одновременных логинов. Обещаю модель 10k DAU при нормальном RPS и жёстком rate limit на auth. Read path проверен CORE k6.»

2.3. Что говорить, если спросят «это ты писал?»
«В Event Horizon — 17 сервисов, но не все мои. Game, Billing, Shop, Inventory, Payment, Authors, History, Analytics, Notification, Fulfillment, Leaderboard, Profile, Gateway, Balancer, NATS Hub, MCP — это мой пет-проект, я писал всё. Auth — тоже мой. Если спросят про AdTime — там 5 из 18 мои: auth, generation, order, deschat, monitoring. Roolz — OSRM, PostGIS, CDC, Kafka, Redis, Helm. Не путаю.»

2.4. Что НЕ говорить
Не говори «10k RPS» — это 10k DAU, десятки RPS.

Не говори «Kafka в AdTime» — там NATS JetStream. Kafka — Roolz.

Не говори «Mongo в Shop» — Mongo experimental в Inventory, Shop — Postgres.

Не говори «Profile без Redis» — Profile подключён к Redis :6385.

Не говори «Game кэширует в Redis» — redis-game поднят, но не подключён.

Не говори «Notification :50056» — актуальный :50063.

Не говори «Gateway :9095 на хосте» — это Shop. Gateway внутри контейнера.

Не говори «MCP в player path» — MCP stdio, не в player path.

2.5. Вопросы, к которым готов
Вопрос	Ответ
«Почему NATS, а не Kafka?»	Простота, latency, Go-native. Kafka — Roolz.
«Почему не 2PC?»	Ломает доступность. Saga + компенсация.
«Почему Outbox?»	Dual-write проблема. Атомарность в одной TX.
«Почему ClickHouse?»	OLAP vs OLTP. DAU/MAU на миллиардах — только колонки.
«Что если Redis упал?»	Auth — 503 (сессия не живёт без Redis). Billing — читаем из PG.
«Что если NATS упал?»	Outbox pending растёт. Fulfillment — Kafka fallback.
«Как тестируешь?»	Unit (mock), integration (testcontainers), k6 CORE (20 VU), race test (Billing).
«Что улучшил бы?»	runGateway, Outbox в Fulfillment, health checks в Balancer, SQL Sandbox в MCP.
ЧАСТЬ 3. Mermaid-схема всех 17 сервисов на одном холсте
3.1. Полная схема (для mermaid.live → SVG → Miro)

flowchart TB
  subgraph clients["🖥️ Clients"]
    REACT["React SPA<br/>:5173"]
    CURSOR["Cursor / Claude Desktop<br/>(MCP stdio)"]
  end

  subgraph edge["🚪 Edge"]
    BAL["Balancer<br/>:8079<br/>Least Connections"]
    GW1["Gateway-1<br/>:8081"]
    GW2["Gateway-2<br/>:8082"]
    GW3["Gateway-3<br/>:8083"]
  end

  subgraph auth["🔐 Identity"]
    AUTH["Auth<br/>:50051<br/>JWT + Redis"]
  end

  subgraph game["🎮 Game"]
    GAME["Game<br/>:50052<br/>Strategy by game"]
    LEAD["Leaderboard<br/>:50054<br/>Redis Sorted Set"]
    PROF["Profile<br/>:50060<br/>Rule Engine"]
  end

  subgraph money["💰 Money"]
    BILL["Billing<br/>:50053<br/>Double-Entry"]
    SHOP["Shop<br/>:50055<br/>Saga"]
    INV["Inventory<br/>:50059<br/>PG/Mongo"]
    PAY["Payment<br/>:50058<br/>HMAC"]
    AUTHORS["Authors<br/>:50061<br/>State Machine"]
  end

  subgraph data["📊 Data"]
    HIST["History<br/>:50062<br/>Retention"]
    ANALYTICS["Analytics<br/>:50057<br/>ClickHouse"]
  end

  subgraph async["📨 Async"]
    NOTIF["Notification<br/>:50063<br/>Dual-Bus"]
    FULFILL["Fulfillment<br/>:9101<br/>Delay 10s"]
    NatsHub["NATS Hub<br/>:9097<br/>Bootstrap"]
  end

  subgraph dev["🛠️ Dev"]
    MCP["MCP<br/>stdio<br/>6 tools"]
  end

  subgraph infra["⚙️ Infra"]
    NATS["NATS JetStream<br/>:4222/:4223/:4224<br/>Stream EVENTS"]
    PG["PostgreSQL ×13<br/>:5460–5470"]
    REDIS["Redis ×8<br/>:6379–6387"]
    CH["ClickHouse<br/>:8123/:9000"]
    MONGO["MongoDB<br/>:27017<br/>(experimental)"]
    KAFKA["Kafka<br/>:19092<br/>(heavy)"]
    PROM["Prometheus<br/>:9090"]
    GRAF["Grafana<br/>:3000"]
    JAEG["Jaeger<br/>:16686"]
  end

  %% Clients → Edge
  REACT -->|HTTP/WS| BAL
  BAL --> GW1 & GW2 & GW3

  %% Gateway → Services (sync gRPC)
  GW1 & GW2 & GW3 -->|gRPC| AUTH
  GW1 & GW2 & GW3 -->|gRPC| GAME
  GW1 & GW2 & GW3 -->|gRPC| BILL
  GW1 & GW2 & GW3 -->|gRPC| LEAD
  GW1 & GW2 & GW3 -->|gRPC| SHOP
  GW1 & GW2 & GW3 -->|gRPC| INV
  GW1 & GW2 & GW3 -->|gRPC| PROF
  GW1 & GW2 & GW3 -->|gRPC| PAY
  GW1 & GW2 & GW3 -->|gRPC| AUTHORS
  GW1 & GW2 & GW3 -->|gRPC| HIST
  GW1 & GW2 & GW3 -->|gRPC| ANALYTICS
  GW1 & GW2 & GW3 -->|gRPC| NOTIF

  %% Sync dependencies (money)
  SHOP -->|SpendCurrency| BILL
  SHOP -->|CanPurchaseMerch| PAY
  GAME -->|SpendCurrency| BILL
  AUTHORS -->|UpdateRole| AUTH

  %% Auth validation
  GW1 & GW2 & GW3 -.->|ValidateToken| AUTH

  %% NATS publishing (async via outbox)
  GAME -.->|score.updated| NATS
  BILL -.->|balance.updated| NATS
  SHOP -.->|shop.purchased<br/>purchase.paid<br/>billing.refund| NATS
  INV -.->|inventory.item.created| NATS
  PAY -.->|payment.completed| NATS
  AUTHORS -.->|author.upserted| NATS
  LEAD -.->|record_beaten| NATS

  %% NATS consuming
  NATS -.->|score.updated| LEAD
  NATS -.->|score.updated| PROF
  NATS -.->|score.updated| BILL
  NATS -.->|score.updated| HIST
  NATS -.->|score.updated| ANALYTICS
  NATS -.->|score.updated| GW1
  NATS -.->|shop.purchased| HIST
  NATS -.->|shop.purchased| ANALYTICS
  NATS -.->|purchase.paid| FULFILL
  NATS -.->|purchase.paid| NOTIF
  NATS -.->|purchase.fulfilled| SHOP
  NATS -.->|purchase.fulfilled| NOTIF
  NATS -.->|payment.completed| HIST
  NATS -.->|payment.completed| ANALYTICS
  NATS -.->|author.upserted| HIST
  NATS -.->|author.upserted| ANALYTICS
  NATS -.->|record_beaten| NOTIF
  NATS -.->|inventory.item.created| SHOP

  %% NATS Hub
  NatsHub -->|creates stream EVENTS| NATS

  %% Kafka (heavy)
  SHOP -.->|purchase.paid| KAFKA
  KAFKA -.->|purchase.paid| FULFILL
  KAFKA -.->|purchase.paid| NOTIF
  FULFILL -.->|purchase.fulfilled| KAFKA

  %% Storage
  AUTH --> PG & REDIS
  GAME --> PG
  BILL --> PG & REDIS
  LEAD --> PG & REDIS
  SHOP --> PG & REDIS
  INV --> PG & REDIS & MONGO
  PROF --> PG & REDIS
  PAY --> PG & REDIS
  AUTHORS --> PG & REDIS
  HIST --> PG
  ANALYTICS --> CH
  NOTIF --> PG

  %% MCP (dev only)
  CURSOR -.->|stdio| MCP
  MCP -.->|read-only| PG
  MCP -.->|read-only| REDIS
  MCP -.->|read-only| NATS

  %% Observability
  PROM -.->|scrape /metrics| GW1 & GW2 & GW3
  GRAF --> PROM
  JAEG -.->|OTLP| GW1

  %% Styling
  classDef edge fill:#dbeafe,stroke:#2563eb,stroke-width:2px
  classDef money fill:#fef3c7,stroke:#d97706,stroke-width:2px
  classDef game fill:#dcfce7,stroke:#16a34a,stroke-width:2px
  classDef data fill:#f3e8ff,stroke:#9333ea,stroke-width:2px
  classDef async fill:#fee2e2,stroke:#dc2626,stroke-width:2px
  classDef infra fill:#f1f5f9,stroke:#64748b
  classDef dev fill:#e0e7ff,stroke:#4f46e5,stroke-dasharray: 5 5

  class BAL,GW1,GW2,GW3 edge
  class BILL,SHOP,INV,PAY,AUTHORS money
  class GAME,LEAD,PROF game
  class HIST,ANALYTICS data
  class NOTIF,FULFILL,NatsHub async
  class NATS,PG,REDIS,CH,MONGO,KAFKA,PROM,GRAF,JAEG infra
  class MCP dev

3.2. Упрощённая схема (для быстрого рисования на доске)

flowchart LR
  subgraph edge["Edge"]
    R["React"] --> B["Balancer<br/>:8079"] --> G["Gateway ×3<br/>:8081–8083"]
  end

  subgraph sync["Sync (gRPC)"]
    G --> AUTH["Auth"]
    G --> GAME["Game"]
    G --> BILL["Billing"]
    G --> SHOP["Shop"]
    G --> INV["Inventory"]
    G --> MORE["... +8 сервисов"]
  end

  subgraph async["Async (NATS)"]
    GAME -.->|score.updated| N["NATS<br/>Stream EVENTS"]
    SHOP -.->|purchase.paid| N
    N -.->|score.updated| LEAD["Leaderboard"]
    N -.->|purchase.paid| FULF["Fulfillment"]
    N -.->|purchase.paid| NOTIF["Notification"]
  end

  subgraph data["Data"]
    AUTH --> PG["Postgres ×13"]
    BILL --> REDIS["Redis ×8"]
    ANALYTICS --> CH["ClickHouse"]
    INV --> MONGO["Mongo<br/>(experimental)"]
  end

  style edge fill:#dbeafe
  style sync fill:#dcfce7
  style async fill:#fee2e2
  style data fill:#f3e8ff

3.3. Как использовать
Для Miro
Открой mermaid.live

Вставь блок ```mermaid ... ```

Экспорт → SVG

Вставь в Miro

Для доски на собесе
Рисуй упрощённую (3.2). Полная (3.1) — для Miro / PDF.

Легенда цветов
🔵 Edge — Balancer, Gateway

🟡 Money — Billing, Shop, Inventory, Payment, Authors

🟢 Game — Game, Leaderboard, Profile

🟣 Data — History, Analytics

🔴 Async — Notification, Fulfillment, NATS Hub

⚪ Infra — NATS, PG, Redis, CH, Mongo, Kafka, Prometheus, Grafana, Jaeger

🟦 Dev (пунктир) — MCP

Ключевые потоки
Sync: React → Balancer → Gateway → gRPC → сервис → ответ

Async: сервис → outbox → NATS → consumer’ы

Money: Shop → Billing (sync) + Shop TX (stock + outbox)

Auth: Gateway → Auth ValidateToken на каждом запросе

WS: Gateway → NATS score.updated → WebSocket → клиент

Dev: Cursor → MCP (stdio) → read-only PG/Redis/NATS


Ответ на вопрос «как у вас работает сервис аутентификации?» + Redis + NATS + техдолг + резюме
Разберу всё по порядку. Начну с главного — как отвечать на вопросы про Auth, Redis, NATS с учётом паттернов. Потом — два файла (сервисы с плюсами/минусами, техдолг). Потом — резюме.

ЧАСТЬ 1. Auth — как отвечать (с необычными кейсами)
1.1. Одна фраза
Auth — единственный, кто знает пароль. Регистрация/логин, bcrypt cost 12, пара JWT (access 15 мин + refresh 7 дней), HMAC-SHA256. jti каждого токена в Redis — пока ключ есть, сессия жива. Gateway на каждом закрытом запросе зовёт Auth ValidateToken (не разбирает подпись сам). Redis упал — 503, не «токен жив».

1.2. Порты
Что	Значение
gRPC	:50051
Metrics	:9091
PG	:5460 (eventhorizon)
Redis	:6379 (общий с Gateway)
1.3. Ручки
Метод	Путь	Роль
POST	/api/v1/auth/register	—
POST	/api/v1/auth/login	—
POST	/api/v1/auth/refresh	—
POST	/api/v1/auth/logout	user+
POST	/api/v1/auth/update-nickname	user+
POST	/api/v1/auth/update-role	admin
GET	/api/v1/auth/whoami	user+
GET	/api/v1/auth/user	user+
GET	/api/v1/admin/users	admin
1.4. Как работает (по шагам)
Логин
POST /login → Gateway → gRPC Login.

Auth достаёт password_hash из PG, сверяет bcrypt cost 12 (~100–300ms).

Выпускает пару: access (15 мин) + refresh (7 дней). HMAC-SHA256.

jti access → Redis 15 мин. jti refresh → Redis 7 дней.

Возвращает {access_token, refresh_token, expires_in}.

Проверка на каждом закрытом запросе
Gateway RequireAuth вырезает Authorization: Bearer.

Зовёт Auth gRPC ValidateToken.

Auth проверяет: подпись HMAC, тип access, jti в Redis.

Нет ключа → Unavailable → Gateway 401.

Роль в контекст. RequireRole пускает только нужную.

Refresh (одноразовый)
Проверка подписи refresh, тип, jti в Redis.

Удаляет старый jti refresh.

Новая пара, роль перечитывает из PG.

Новые jti → Redis.

Logout
Удаляет jti из Redis.

Подпись верна до exp, но проверка сессии не пускает.

Redis упал
ValidateToken: Unavailable → Gateway 503 session store unavailable.

Refresh: то же.

Подпись без живого Redis не считается сессией.

1.5. Необычные кейсы (главное на собесе)
Session Store в Redis — не «просто JWT». Обычный JWT stateless: подпись верна — токен жив. Здесь наоборот: подпись верна, но jti нет — токен мёртв. Мгновенный отзыв (logout, смена роли), цена — Redis в hot path.

Gateway не разбирает подпись сам. Каждый запрос — gRPC ValidateToken. Auth — единственная точка правды о сессии. Минус — лишний хоп. Плюс — нельзя обойти даже при компрометации Gateway.

Refresh одноразовый. Старый jti удаляется, пара заново, роль из PG. Защита от replay refresh-токена.

Login fail-closed, общий лимит fail-open. Redis на логине упал — не пускаем (брутфорс важнее). Общий лимит — пускаем (не гасим весь вход).

bcrypt cost 12 — by design медленно. 500 VU login blast в k6 «краснеет» намеренно.

1.6. Паттерны (13)
Паттерн	Где
DI Container	internal/app
Pipeline / Init Chain	bootstrap
Facade	service layer
Repository	internal/repository
Converter / Mapper	internal/converter
Middleware Chain	gRPC interceptors
Strategy	JWT generator (HMAC)
Cache-Aside	Redis sessions
Cache Invalidation	logout / refresh
Session Store	Redis — необычный
Config View	ConfigProvider
SQL Builder	pkg/sqb
Graceful Shutdown	main
1.7. Метрики
Метрика	Цель
go_goroutines{job="auth"}	< 1000
auth_logins_total	—
auth_registrations_total	—
auth_token_validations_total	—
auth_refresh_total	—
redis_connected_clients{job="redis"}	—
pg_stat_database_numbackends{datname="eventhorizon"}	< 100
p95 login	< 500ms
p95 ValidateToken	< 50ms
error rate	< 1%
1.8. Команды
bash
# Health
curl -fsS http://localhost:9091/health
curl -fsS http://localhost:9091/ready

# Логин
curl -sS -X POST http://localhost:8079/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventhorizon.local","password":"YOUR_PASSWORD"}' | jq .

# БД
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon \
  -c "SELECT id, email, nickname FROM users LIMIT 10;"

# Пересобрать
cd ~/event_horizon/services/auth && go build -o auth-service ./cmd/main.go
bash scripts/rebuild-services.sh auth
ЧАСТЬ 2. Redis — как отвечать («где у вас Redis, как работает?»)
2.1. Одна фраза
Redis — не база, а кэш + сессии + Sorted Set + rate limit + pub/sub. Никогда не единственная копия заказа или баланса. Восемь инстансов :6379–6387, у каждого своя роль. Упал — деградация, но не потеря данных.

2.2. Где Redis (8 инстансов)
Инстанс	Порт	Кто	Что хранит
redis	:6379	Auth + Gateway	Сессии (jti), rate limit, WS ticket
redis-game	:6380	—	Поднят, но Game не подключён
redis-billing	:6381	Billing	Кэш балансов (Cache-Aside)
redis-leaderboard	:6382	Leaderboard	Sorted Set топа
redis-shop	:6383	Shop	Кэш витрины
redis-inventory	:6384	Inventory	Кэш каталога (Decorator)
redis-profile	:6385	Profile	Кэш профиля
redis-payment	:6386	Payment	Кэш подписки (TTL from Expiry)
redis-authors	:6387	Authors	Кэш карточек
2.3. Как работает (по ролям)
Сессии (Auth :6379)
jti каждого токена. SETEX jti:{id} 900 1.

Logout → DEL jti:{id}.

Gateway ValidateToken → Auth проверяет EXISTS jti:{id}.

Redis упал → 503. Не «токен жив».

Rate limit (Gateway :6379)
Sliding Window: INCR rate:{user_or_ip}:{path}, EXPIRE.

Login 5/s IP — fail-closed.

Общий ~100/s — fail-open.

WS 100/min IP.

WS ticket (Gateway :6379)
SETEX ws:ticket:{id} 30 {ticket}.

GETDEL ws:ticket:{id} — одноразовый.

Билет сгорает на первом GET.

Кэш балансов (Billing :6381)
Cache-Aside: GET balance:{user}:{currency} → miss → PG → SETEX.

После AddCurrency / SpendCurrency → DEL balance:{user}:{currency}.

Redis упал → читаем из PG (медленнее, но работает).

Sorted Set топа (Leaderboard :6382)
ZADD leaderboard:{game}:{level} GT {score} {user_id}.

GT: true — Max-Only Write. Хуже не перезаписывает.

ZREVRANGE leaderboard:{game}:{level} 0 9 WITHSCORES.

Redis упал → топ пустой (Restore-on-Startup из PG при рестарте).

Кэш витрины (Shop :6383)
Cache-Aside: GET shop:items:{hash} → miss → PG → SETEX.

Инвалидация после мутации.

Кэш каталога (Inventory :6384)
Decorator CachedRepository: GET item:{id} → miss → PG/Mongo → SETEX.

Самый чистый Cache-Aside.

Кэш профиля (Profile :6385)
Cache-Aside: GET profile:{user_id} → miss → PG → SETEX.

Инвалидация после score.updated.

Кэш подписки (Payment :6386)
Cache TTL based on Expiry: SETEX subscription:{user_id} {ttl_from_expires_at} 1.

TTL вычисляется из expires_at — не константа.

Когда подписка истекает, кэш тоже.

Кэш карточек (Authors :6387)
Cache-Aside: GET author:{user_id} → miss → PG → SETEX.

2.4. Необычные кейсы
Redis как Session Store, не просто кэш. Auth: jti в Redis — сессия жива. Это не кэш, а source of truth для сессии.

Cache TTL based on Expiry (Payment). TTL не константа, а вычисляется из expires_at подписки.

ZAddArgs{GT: true} (Leaderboard). Max-Only Write — хуже не перезаписывает.

redis-game поднят, но не подключён. HOW_DOES_EH_WORK.md: «Game config has no Redis usage now». Техдолг: либо подключить, либо убрать из compose.

Rate limit fail-closed vs fail-open. Login — fail-closed (Redis упал → не пускаем). Общий — fail-open (Redis упал → пускаем).

Pub/sub не хранит сообщение. В AdTime deschat Redis Pub/Sub — кто не был подписан, тот не увидит. В Event Horizon pub/sub не используется (NATS вместо).

2.5. Что НЕ хранить в Redis
Заказ как единственную копию (только PG).

Баланс как единственную копию (только PG, Redis — кэш).

Досье на 560 млн (только PG/CH).

Медиа (только S3).

2.6. Метрики
Метрика	Что значит
redis_connected_clients{job="redis-*"}	Клиенты
redis_memory_used_bytes	Память
redis_keyspace_hits_total / misses_total	Hit rate
redis_commands_total{cmd="get"}	Команды
2.7. Команды
bash
# Сессии (Auth)
docker exec -it event-horizon-redis redis-cli KEYS "jti:*"

# Топ (Leaderboard)
docker exec -it event-horizon-redis-leaderboard redis-cli \
  ZREVRANGE leaderboard:hexagon:1 0 9 WITHSCORES

# Баланс (Billing)
docker exec -it event-horizon-redis-billing redis-cli \
  GET balance:{user_id}:tickets

# Rate limit (Gateway)
docker exec -it event-horizon-redis redis-cli KEYS "rate:*"

# WS ticket (Gateway)
docker exec -it event-horizon-redis redis-cli KEYS "ws:ticket:*"
ЧАСТЬ 3. NATS — как отвечать («а где NATS?»)
3.1. Одна фраза
NATS JetStream — event spine. Stream EVENTS со всеми субъектами, FileStorage, MaxAge 7 дней, MaxBytes 1 ГБ. Кластер из 3 нод. Публикация — через Outbox (6 сервисов). Потребление — durable consumers. NATS Hub создаёт stream. Kafka — fallback для Fulfillment/Notification в heavy.

3.2. Где NATS
Компонент	Порт
nats-1	:4222 (клиент), :8222 (monitoring), :6222 (cluster)
nats-2	:4223, :8223, :6223
nats-3	:4224, :8224, :6224
nats-hub	:9097 (metrics)
nats-exporter	:7777 (/metrics)
3.3. Stream EVENTS
text
Name:     "EVENTS"
Subjects: ["score.updated", "user.registered", "event.>",
           "shop.purchased", "purchase.paid", "purchase.fulfilled",
           "payment.completed", "inventory.item.*",
           "author.upserted", "balance.updated"]
Storage:  FileStorage
MaxAge:   7 * 24 * time.Hour
MaxBytes: 1 * 1024 * 1024 * 1024  (1 ГБ)
3.4. Кто публикует / потребляет
Субъект	Публикует	Потребляет
score.updated	Game (outbox)	Leaderboard, Profile, Billing, History, Analytics, Gateway WS
user.registered	Auth/Gateway	History, Profile, Analytics
shop.purchased	Shop (outbox)	History, Analytics
purchase.paid	Shop (outbox)	Fulfillment, Notification
purchase.fulfilled	Fulfillment	Shop, Notification
payment.completed	Payment (outbox)	History, Analytics, Notification
author.upserted	Authors (outbox)	History, Analytics
balance.updated	Billing (outbox)	(интересующиеся)
record_beaten	Leaderboard	Notification
inventory.item.created	Inventory (outbox)	Shop
3.5. Как работает Outbox
В одной транзакции: бизнес-строка + INSERT outbox(event_type, payload).

Commit.

Воркер: SELECT … FROM outbox WHERE processed = false ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED.

Publish в NATS.

UPDATE outbox SET processed = true — только после успешного Publish.

Если NATS упал — строка остаётся processed = false, уйдёт со следующей попытки.

Outbox есть у: Game, Billing, Shop, Inventory, Payment, Authors. Нет у: Fulfillment (техдолг).

3.6. Необычные кейсы
Outbox через FOR UPDATE SKIP LOCKED. Два воркера берут разные строки. Строка processed только после Publish. Защита от потери и дубля.

Dual-Bus Consumer (Notification). NATS primary, Kafka fallback. Kafka только когда JetStream не поднят.

Dual-Transport Consumer (Fulfillment). То же, но без БД.

Idempotent consumer. Каждый consumer терпит повтор: ON CONFLICT, source_event, event_id.

NATS Hub — единственный, кто создаёт stream. Idempotent Stream Creation: AddStream → если AlreadyExists → UpdateStream.

MaxAge 7 дней, не 30. Для stream — 7 дней. Для History — 30 (другое хранилище).

3.7. Kafka (heavy, opt-in)
Топик	Публикует	Потребляет
purchase.paid	Shop	Fulfillment, Notification
purchase.fulfilled	Fulfillment	Notification
Kafka не в make deploy. Только make deploy-heavy. В обычном деплое Kafka молчит, NATS работает.

3.8. Метрики
Метрика	Что значит
gnatsd_varz_connections	Подключения
gnatsd_varz_in_msgs / out_msgs	Входящие/исходящие
gnatsd_streaming_store_msgs	Сообщений в stream
gnatsd_streaming_store_bytes	Байт в stream (MaxBytes 1 ГБ)
nats_consumer_lag{consumer="..."}	Лаг consumer’а
3.9. Команды
bash
# Streams
nats stream ls --server localhost:4222
nats stream info EVENTS --server localhost:4222

# Consumer'ы
nats consumer ls EVENTS --server localhost:4222

# Подписка
nats sub "score.updated" --server localhost:4222
nats sub "event.>" --server localhost:4222

# Публикация
nats pub "score.updated" '{"user_id":"test","score":150}' --server localhost:4222

# Outbox pending (Game)
docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game \
  -c "SELECT count(*) FROM outbox WHERE processed = false;"

# Consumer lag
curl -s 'http://localhost:9090/api/v1/query?query=nats_consumer_lag' | jq '.data.result'
ЧАСТЬ 4. Файл 1 — «Сервисы: преимущества и недостатки»
Создам файл EVENT_HORIZON_SERVICES_PRO_CON.md.

markdown
# Event Horizon — сервисы: преимущества и недостатки

> 17 сервисов, 333 паттерна. Для собеса: что хвалить, что честно признать.
> Источник: `EVENT_HORIZON_PATTERNS_15.10.2026.md`, `FINISHERS.md`, `README.md`.

## Сводная таблица

| # | Сервис | Паттерны | ✅ Преимущества | ❌ Недостатки / техдолг |
|---|--------|----------|-----------------|------------------------|
| 1 | **Auth** | 13 | Session Store в Redis (мгновенный отзыв), refresh одноразовый, bcrypt 12, login fail-closed | Redis в hot path каждого запроса, нет ротации JWT_SECRET |
| 2 | **Game** | 20 | Strategy по играм, Outbox `FOR UPDATE SKIP LOCKED`, Domain Validators (Hexagon/Memory), буст не pay-to-win | `switch` на 8 игр — god-function, anti-cheat не для всех игр, `redis-game` не подключён |
| 3 | **Billing** | 24 | Conditional UPDATE, Double-Entry Ledger, Race Condition Test, Idempotent Write | `RewardConsumer` использует `time.Now()` — двойное начисление, `GetBalance` возвращает `-1` при miss |
| 4 | **Leaderboard** | 20 | `ZAddArgs{GT: true}`, Ack Batcher, Event Chaining, Restore-on-Startup | Restore при падении PG — пустой leaderboard, нет защиты от спама `record_beaten` |
| 5 | **Shop** | 30 | Saga `purchase_sagas`, Spend Cycle, Resume Worker, Pessimistic + Optimistic Lock, Outbox → Compensate | `kafkaProd` не используется, `paymentClient` nil, `cancelPurchase` может задвоить |
| 6 | **Inventory** | 28 | Strategy (PG/Mongo), Decorator (Cache-Aside), Change Streams, RBAC interceptor, Outbox type assertion | Два драйвера (двойная поддержка), `toPostgresFilter` через `fmt.Sprintf`, outbox type assertion хрупок |
| 7 | **Profile** | 14 | Rule Engine (чистая функция), Unlock Pipeline, Idempotent Write, `_partial` Graceful Degradation | `UpdateProfile` наружу не открыт, нет отдельного gRPC для редактирования |
| 8 | **Payment** | 21 | HMAC + ConstantTimeCompare, Dual Auth Path, Idempotent Completion, Revenue Split, Cache TTL based on Expiry | Revenue Split — журнал, не перевод, `subscriptionDays` default 30, нет сверки суммы |
| 9 | **Authors** | 20 | State Machine, Pessimistic Lock, Compensating Action, Idempotent Submission, Role Check | `RevertApplication` hard delete (FK падает), UX gap: JWT старый после approve |
| 10 | **History** | 16 | Retention Policy, Payload Sanitization, Nak Pattern, HistoryStore Port | RetentionWorker в процессе (при рестарте сброс), PII в payload, 5 субъектов в одном consumer’е |
| 11 | **Analytics** | 14 | ClickHouse OLAP, Cohort Retention, Parameterized Query, Adapter, AnalyticsStore Port | `fmt.Sprintf` для db name, PII в событиях, 5 субъектов в одном consumer’е |
| 12 | **Notification** | 16 | Dual-Bus, Fan-Out, Idempotency Key `source_event`, Graceful Degradation | Poison message из Kafka — бесконечный retry, Fixed Delay для Telegram |
| 13 | **Fulfillment** | 15 | Dual-Transport, Idempotency Guard, Delay-Based Simulation, Manual DI | Нет Outbox (событие теряется при падении), `time.After(10s)` заглушка, UUID несогласованность |
| 14 | **Gateway** | 41 | Circuit Breaker + Bulkhead, Retry full jitter, Sliding Window Rate Limiter, WS Ticket, BFF, DTO Mapper, Docs Gate | `runGateway` God Function > 1000 строк, два Hub (дублирование), Async gRPC в admin/users без CB |
| 15 | **Balancer** | 15 | Least Connections, Atomic Counter, Reverse Proxy, Prometheus Gauge, X-Balancer-Backend | Нет health checks, static registry, `/ready` не проверяет, пустые пакеты, `log.Printf` в hot path |
| 16 | **NATS Hub** | 11 | Bootstrap, Idempotent Stream Creation, Stream Config as Code, Wildcard Subscription, реальный health check | Нет Graceful Degradation (падает при падении NATS) |
| 17 | **MCP** | 15 | MCP Server, Tool Registration, Graceful Degradation per-tool, SQL Sandbox, Offline RAG (TF-IDF), stdio | SQL Sandbox string matching (обходится), утечка ошибок, TF-IDF хуже embedding, DSN читает целиком |

## Ключевые преимущества (что хвалить)

### Архитектурные
- **Saga в Shop** — единственный с явной сагой, `purchase_sagas` оркестратор.
- **Outbox в 6 сервисах** — Game, Billing, Shop, Inventory, Payment, Authors.
- **Circuit Breaker + Bulkhead в Gateway** — единственный, fail-fast 503.
- **Session Store в Auth** — мгновенный отзыв сессии.
- **Least Connections в Balancer** — atomic counter, не mutex.

### Данные
- **Double-Entry Ledger в Billing** — `balance_after` в каждой транзакции.
- **Conditional UPDATE в Billing** — `WHERE balance >= $1`, атомарно.
- **ClickHouse в Analytics** — OLAP, не OLTP.
- **Redis Sorted Set в Leaderboard** — `ZAddArgs{GT: true}`.
- **Cache TTL based on Expiry в Payment** — TTL из данных, не константа.

### Безопасность
- **HMAC + ConstantTimeCompare в Payment** — защита от timing attack.
- **Refresh одноразовый в Auth** — защита от replay.
- **WS Ticket в Gateway** — не JWT в query.
- **SQL Sandbox в MCP** — string matching (техдолг, но есть).

### Ops
- **Stream Config as Code в NATS Hub** — версионируется.
- **Prometheus Gauge в Balancer** — единственный с Gauge.
- **Graceful Degradation per-tool в MCP** — один tool упал, остальные работают.

## Ключевые недостатки (что честно признать)

### Критичные (Pre-Prod)
1. **Fulfillment без Outbox** — событие теряется при падении.
2. **`runGateway` God Function** — > 1000 строк.
3. **`RewardConsumer` `time.Now()`** — двойное начисление.
4. **`paymentClient` nil** — мерч-гейт деградирует.
5. **`RevertApplication` hard delete** — FK падает.

### Средние
6. **Два драйвера БД в Inventory** — двойная поддержка.
7. **Два Hub в Gateway** — дублирование.
8. **`toPostgresFilter` `fmt.Sprintf`** — потенциальная injection.
9. **MCP SQL Sandbox** — string matching, обходится.
10. **Balancer static registry** — нет service discovery.

### Мелкие
11. **`redis-game` не подключён** — мёртвый код.
12. **Poison message в Notification** — бесконечный retry.
13. **`time.After(10s)` в Fulfillment** — заглушка.
14. **RetentionWorker в History** — при рестарте сброс.
15. **`fmt.Sprintf` для db name в Analytics** — injection.

## Итого

| Метрика | Значение |
|---------|----------|
| Всего сервисов | 17 |
| Всего паттернов | 333 |
| Outbox | 6 сервисов (Game, Billing, Shop, Inventory, Payment, Authors) |
| Circuit Breaker | 1 (Gateway) |
| Saga | 1 (Shop) |
| ClickHouse | 1 (Analytics) |
| Критичных техдолгов | 5 |
| Средних | 5 |
| Мелких | 5 |
ЧАСТЬ 5. Файл 2 — «Техдолг на 15 октября 2026»
Создам файл EVENT_HORIZON_TECH_DEBT_15.10.2026.md.

markdown
# Event Horizon — техдолг на 15.10.2026 (v1.1.3)

> Что обязательно доделать. Разделено по приоритетам.
> Источник: `EVENT_HORIZON_PATTERNS_15.10.2026.md`, `FINISHERS.md`, `FUTURE_EVENT_HORIZON.md`.

## 🔴 Критичные (Pre-Prod) — блокеры

| # | Сервис | Проблема | Решение | Приоритет |
|---|--------|----------|---------|-----------|
| 1 | **Fulfillment** | Нет Outbox — событие теряется между `purchase.paid` и `purchase.fulfilled` | Внедрить Transactional Outbox | P0 |
| 2 | **Fulfillment** | `time.After(10s)` вместо реального WMS | Заменить на warehouse/delivery API | P0 |
| 3 | **Gateway** | `runGateway` — God Function > 1000 строк | Разбить на `bootstrap/clients.go`, `routes.go`, `middleware.go` | P0 |
| 4 | **Billing** | `RewardConsumer` `time.Now().UnixNano()` для `reference_id` | Использовать `event.ReferenceID` | P0 |
| 5 | **Shop** | `paymentClient` может быть `nil` — мерч-гейт деградирует | Метрика `merch_gate_disabled` | P0 |
| 6 | **Authors** | `RevertApplication` hard delete — FK падает | Soft delete + check FKs | P0 |

## 🟡 Средние — качество и надёжность

| # | Сервис | Проблема | Решение | Приоритет |
|---|--------|----------|---------|-----------|
| 7 | **Gateway** | Два Hub (`app/gateway.go` + `websocket/hub.go`) — дублирование | Унифицировать | P1 |
| 8 | **Gateway** | Async gRPC в admin/users без CB | Обернуть в `throughBreaker` | P1 |
| 9 | **Inventory** | Два драйвера БД (PG + Mongo) — двойная поддержка | Определиться с одним | P1 |
| 10 | **Inventory** | `toPostgresFilter` через `fmt.Sprintf` | Переписать без `fmt.Sprintf` для имён колонок | P1 |
| 11 | **MCP** | SQL Sandbox string matching — обходится | Read-only роль + `pgx` prepared | P1 |
| 12 | **MCP** | `err.Error()` — утечка внутренних деталей | Санитизировать сообщения | P1 |
| 13 | **Balancer** | `/ready` не проверяет бэкенды | Реальная проверка хотя бы одного backend | P1 |
| 14 | **Balancer** | Static registry — нет service discovery | Интегрировать с etcd/consul/k8s | P1 |
| 15 | **Balancer** | Нет health checks бэкендов | Периодический ping `/health` | P1 |
| 16 | **Notification** | Poison message из Kafka — бесконечный retry | Poison message → Ack | P1 |
| 17 | **Analytics** | `fmt.Sprintf` для db name | Параметризовать db name | P1 |
| 18 | **Shop** | `kafkaProd` создаётся, не используется | Убрать или подключить | P1 |

## 🟢 Мелкие — гигиена

| # | Сервис | Проблема | Решение | Приоритет |
|---|--------|----------|---------|-----------|
| 19 | **Gateway** | `log.Printf` везде — нет structured logging | Перейти на `slog` | P2 |
| 20 | **Balancer** | Пустые пакеты `backend/`, `proxy/` | Удалить или наполнить | P2 |
| 21 | **Balancer** | `log.Printf` в hot path | Structured logging | P2 |
| 22 | **Fulfillment** | `newUUID` (crypto/rand) vs `uuid.NewString()` | Унифицировать | P2 |
| 23 | **Game** | `redis-game` `:6380` поднят, не подключён | Подключить или убрать | P2 |
| 24 | **Billing** | `GetBalance` возвращает `-1` при cache miss | Использовать `(int, bool)` | P2 |
| 25 | **Leaderboard** | `RestoreFromPostgres` при падении PG — пустой leaderboard | Retry / exponential backoff | P2 |
| 26 | **Shop** | `cancelPurchase` `reference_id = cancel-{id}` — может задвоить | Idempotency Key | P2 |
| 27 | **History** | RetentionWorker в процессе — при рестарте сброс | Kubernetes CronJob | P2 |
| 28 | **Payment** | `subscriptionDays` default 30 — проверить | Явно в config | P2 |
| 29 | **Auth** | `boosty.to` URL захардкожен | Config | P2 |
| 30 | **Authors** | UX gap: JWT старый после approve | Автоперевыпуск JWT или инструкция | P2 |

## 🔵 Roadmap (не техдолг, но запланировано)

| # | Что | Зачем | Приоритет |
|---|-----|-------|-----------|
| 31 | **i18n** (RU/EN/DE/ES) | Смена языка сессии | P3 |
| 32 | **Пагинация каталога** на сервере | До 1000+ карт | P3 |
| 33 | **Уровни лидерборда 2–3** | Сейчас все `level=1` | P3 |
| 34 | **Порог покрытия в CI** | ≥70% | P3 |
| 35 | **Retry с джиттером, bulkhead** | У Gateway уже CB | P3 |
| 36 | **108 авторов** | Нет текстов/ассетов | P4 |
| 37 | **3D** | Нет ассетов | P4 |
| 38 | **Helm как повседневный деплой** | Сейчас compose | P4 |
| 39 | **Loki / Tempo** | Не в `make deploy` | P4 |
| 40 | **WAF / пентест** | Не флаг в репо | P4 |

## ✅ Уже закрыто (v1.1.0–v1.1.3)

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

## Что делать прямо сейчас (P0)

1. **Fulfillment Outbox** — событие теряется. Критично.
2. **`runGateway`** — разбить на модули.
3. **`RewardConsumer`** — `event.ReferenceID` вместо `time.Now()`.
4. **`RevertApplication`** — soft delete.
5. **`paymentClient` nil** — метрика.

## Что делать после P0 (P1)

1. Унифицировать два Hub в Gateway.
2. Определиться с одним драйвером в Inventory.
3. SQL Sandbox в MCP — read-only роль.
4. Health checks в Balancer.
5. Poison message в Notification — Ack.

## Итого

| Приоритет | Количество |
|-----------|-----------|
| P0 (критичные) | 6 |
| P1 (средние) | 12 |
| P2 (мелкие) | 12 |
| P3 (roadmap) | 5 |
| P4 (долгосрок) | 5 |
| Закрыто | 15 |

