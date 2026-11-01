Три задачи: 7 сервисов глубоко, финальный FULL.md, follow-up
ЧАСТЬ 1. 7 оставшихся сервисов — глубокий разбор
1.1. Game — глубоко
Sequence: submit с boost
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant G as Game
  participant B as Billing
  participant PG as PG Game
  participant W as Outbox worker
  participant N as NATS

  C->>GW: POST /game/boost/start
  GW->>G: gRPC StartBoost
  G->>B: SpendCurrency(lamp, boost_ref)
  alt no lamps
    B-->>G: insufficient
    G-->>GW: 402
  else ok
    B-->>G: ok
    G->>PG: INSERT boost (user_id, expires_at)
    G-->>GW: ok
  end

  C->>GW: POST /game/submit {game_id, score, moves}
  GW->>G: gRPC SubmitScore
  G->>G: Strategy by game_id
  G->>G: Domain Validator (hexagon/memory)
  G->>PG: SELECT boost WHERE user_id AND expires_at > now()
  alt boost active
    G->>G: not ranked (в LB не идёт)
    G->>PG: INSERT highscores (не в LB)
  else no boost
    G->>PG: INSERT/UPDATE highscores
  end
  G->>PG: INSERT outbox(score.updated)
  G->>PG: COMMIT
  G-->>GW: ok
  GW-->>C: 200

  W->>PG: SELECT * FROM outbox FOR UPDATE SKIP LOCKED
  W->>N: publish score.updated
  N->>N: Leaderboard, Profile, Billing, History, Analytics, WS
Кейсы
Кейс	Что происходит	Решение
Два воркера	SKIP LOCKED — разные строки	Один рекорд не публикуется дважды
NATS упал	outbox_pending растёт	Воркер ретраит, после рестарта — догоняет
Billing упал при boost	Буст не стартует	Обычный submit работает (Game не зависит)
Domain Validator fail	InvalidArgument → 400	Клиент видит ошибку
switch на 8 игр	God-function	Техдолг: per-game strategy
Anti-cheat не для всех	Только Hexagon/Memory	Техдолг: для остальных
redis-game не подключён	Redis поднят, не используется	Техдолг: подключить или убрать
Что улучшил бы
switch → per-game strategy interface.

Anti-cheat для всех 8 игр.

redis-game — подключить для кэша или убрать из compose.

1.2. Leaderboard — глубоко
Sequence: record_beaten + Ack Batcher
sequenceDiagram
  autonumber
  participant N as NATS
  participant L as Leaderboard
  participant R as Redis
  participant PG as PG Leaderboard
  participant No as Notification

  N->>L: score.updated {user_id, game_id, score}
  L->>R: ZADD leaderboard:{game}:1 GT {score} {user_id}
  alt added
    R-->>L: 1
    L->>R: ZREVRANK leaderboard:{game}:1 {user_id}
    R-->>L: rank
    alt rank == 1 (побил чужой)
      L->>R: ZREVRANGE 0 1 WITHSCORES
      R-->>L: [new, old]
      L->>N: publish record_beaten {victim: old, game_id, score}
    end
    L->>PG: UPSERT leaderboard_backup
  else skipped (score <= current)
    R-->>L: 0
  end
  L->>L: Ack Batcher: accumulate (100 events or 1s)
  L->>N: batch Ack
Кейсы
Кейс	Что происходит	Решение
Redis упал	Топ пустой	RestoreFromPostgres при рестарте (техдолг: retry)
PG упал	Redis жив, топ работает	Restore при рестарте PG
record_beaten лавина	1000 рекордов → 1000 record_beaten	Нет rate limit (техдолг)
Ack Batcher NATS упал	Подтверждения копятся	При рестарте — догоняет
WS reconnect storm	100/min IP → 429	By design
Max-Only Write	GT: true — хуже не перезаписывает	By design
Что улучшил бы
Rate limit на record_beaten.

RestoreFromPostgres — retry/backoff.

Уровни 2–3 (сейчас все level=1).

1.3. Profile — глубоко
Sequence: Rule Engine + Pipeline
sequenceDiagram
  autonumber
  participant N as NATS
  participant P as Profile
  participant PG as PG Profile
  participant R as Redis

  N->>P: score.updated {user_id, game_id, score}
  P->>PG: SELECT user_stats (total_score, games_played, ...)
  PG-->>P: stats
  P->>P: EvaluateAchievements(stats, event) → [codes]
  Note over P: чистая функция, без IO
  P->>PG: SELECT unlocked_codes WHERE user_id
  PG-->>P: [already_unlocked]
  P->>P: filter new codes
  loop for each new code
    P->>PG: INSERT achievements<br/>ON CONFLICT (user_id, code) DO NOTHING
  end
  P->>PG: UPDATE user_profiles SET total_score = MAX(total_score, $new)
  P->>R: DEL profile:{user_id}
  P->>N: Ack
Sequence: _partial при Billing down
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant P as Profile
  participant PG as PG Profile
  participant B as Billing
  participant G as Game

  C->>GW: GET /profile
  GW->>P: gRPC GetProfile
  P->>PG: SELECT email, nickname, total_score
  PG-->>P: ok
  P->>B: GetBalance (timeout 1s)
  B--xP: timeout
  P->>P: partial = true
  P->>G: GetBestScores (timeout 1s)
  G-->>P: ok
  P-->>GW: {email, nickname, total_score,<br/>lamps: null, tickets: null,<br/>_partial: true}
  GW-->>C: 200
Кейсы
Кейс	Что происходит	Решение
Billing/Game down	_partial: true	Graceful Degradation
Rule Engine	Чистая функция	Легко тестировать
Дублирование достижений	ON CONFLICT DO NOTHING	Idempotent
Redis упал	Cache-Aside miss → PG	Медленнее, но работает
UpdateProfile наружу	Не открыт (умеет переписать total_score)	Защита
Ник	Через Auth update-nickname	Не через Profile
Что улучшил бы
Отдельный gRPC для редактирования профиля без total_score.

Алерт на _partial резко растёт.

1.4. Authors — глубоко
Sequence: Apply + Approve + Compensating
sequenceDiagram
  autonumber
  participant U as User
  participant GW as Gateway
  participant A as Authors
  participant PG as PG Authors
  participant Auth as Auth

  U->>GW: POST /authors/apply
  GW->>A: gRPC Apply
  A->>PG: SELECT * FROM author_applications<br/>WHERE user_id AND status='pending' FOR UPDATE
  alt pending exists
    A-->>GW: existing (idempotent)
  else new
    A->>PG: INSERT author_applications (pending)
    A-->>GW: ok
  end

  Note over U,Auth: Admin approve
  U->>GW: POST /applications/:id/approve (admin)
  GW->>A: gRPC Approve
  A->>PG: SELECT FOR UPDATE application
  alt status != pending
    A-->>GW: AlreadyExists
  end
  A->>PG: BEGIN
  A->>PG: UPDATE application approved
  A->>PG: INSERT authors
  A->>PG: INSERT outbox(author.upserted)
  A->>PG: COMMIT
  A->>Auth: UpdateRole(user_id, author)
  alt Auth fails
    A->>PG: BEGIN
    A->>PG: UPDATE application pending
    A->>PG: DELETE authors (hard — техдолг)
    A->>PG: DELETE outbox
    A->>PG: COMMIT
    A-->>GW: error
  else ok
    A-->>GW: ok
  end
Кейсы
Кейс	Что происходит	Решение
Два Apply	FOR UPDATE на user_id	Второй видит pending
Два админа approve	FOR UPDATE на заявке	Первый approved, второй AlreadyExists
Auth упал	RevertApplication	Компенсация
RevertApplication hard delete	FK падает	Техдолг: soft delete
UX gap	JWT старый после approve	Refresh / re-login
Privilege escalation	Только admin	RequireRole(admin)
Что улучшил бы
RevertApplication — soft delete + check FKs.

UX gap: автоперевыпуск JWT или инструкция.

1.5. History — глубоко
Sequence: 5 субъектов + Retention
sequenceDiagram
  autonumber
  participant N as NATS
  participant H as History
  participant PG as PG History
  participant RW as RetentionWorker

  Note over N,H: 5 subjects (durable history-ingest)
  N->>H: score.updated
  N->>H: user.registered + event.user.registered
  N->>H: shop.purchased
  N->>H: payment.completed
  N->>H: author.upserted

  loop each event
    alt invalid JSON
      H->>H: payload = {"raw": "..."}
    end
    H->>PG: INSERT history<br/>ON CONFLICT (user_id, event_id) DO NOTHING
    alt db error
      H->>N: Nak
    else ok
      H->>N: Ack
    end
  end

  Note over RW,PG: Раз в сутки
  RW->>PG: DELETE FROM history<br/>WHERE created_at < now()-30d<br/>LIMIT 10000
  PG-->>RW: ok
  Note over RW: VACUUM (autovacuum или вручную)
Кейсы
Кейс	Что происходит	Решение
Битый JSON	Payload Sanitization → {"raw": "..."}	Не poison
БД упала	Nak → повтор	At-least-once
5 субъектов в одном consumer’е	Lag по всем	Техдолг: разделить
RetentionWorker рестарт	Cron сброс	Техдолг: K8s CronJob
user.registered + event.user.registered	Оба, дедуп	Один факт — одна строка
PII в payload	Не фильтруется	Техдолг
Что улучшил бы
RetentionWorker → K8s CronJob.

Разделить 5 consumer’ов.

Фильтрация PII.

1.6. Analytics — глубоко
Sequence: батчинг + Parameterized Query
sequenceDiagram
  autonumber
  participant N as NATS
  participant A as Analytics
  participant BUF as Buffer (memory)
  participant CH as ClickHouse
  participant Admin as Admin
  participant GW as Gateway

  loop 5 subjects
    N->>A: event
    A->>BUF: append
  end
  Note over BUF: раз в 1s или 100 events
  A->>CH: INSERT INTO events VALUES<br/>({user_id:String}, {event_type:String}, ...)<br/>PARAM param_user_id=...
  Note over CH: named params — не конкатенация
  CH-->>A: ok
  A->>N: Ack (batch)

  Admin->>GW: GET /analytics/dau
  GW->>GW: RequireRole(admin)
  GW->>A: gRPC GetDAU
  A->>CH: SELECT toDate(created_at), count(DISTINCT user_id)<br/>FROM events WHERE event_type='score.updated'<br/>GROUP BY 1
  CH-->>A: [{day, dau}]
  A-->>GW: dau
  GW-->>Admin: 200
Кейсы
Кейс	Что происходит	Решение
ClickHouse down	Nak → повтор	At-least-once
Too many parts	Батчинг сломан	clickhouse_parts_current алерт
fmt.Sprintf db name	Потенциальная injection	Техдолг: параметризовать
5 субъектов в одном consumer’е	Lag по всем	Техдолг: разделить
PII в событиях	Не фильтруется	Техдолг
Все ручки — admin	Защита	RequireRole(admin)
Что улучшил бы
Параметризовать db name.

Разделить 5 consumer’ов.

Фильтрация PII.

1.7. MCP — глубоко
Sequence: 6 tools + Graceful Degradation per-tool
sequenceDiagram
  autonumber
  participant Cu as Cursor
  participant M as MCP
  participant P as Prydwen (TF-IDF)
  participant PG as Postgres (read-only)
  participant R as Redis (read-only)
  participant N as NATS (read-only)

  Cu->>M: tools/list (stdio JSON-RPC)
  M-->>Cu: [search_prydwen, nats_list_streams,<br/>nats_list_consumers, postgres_query,<br/>redis_get, redis_keys]

  Cu->>M: search_prydwen {query}
  M->>P: TF-IDF search
  alt Prydwen loaded
    P-->>M: top-N
    M-->>Cu: results
  else not loaded
    M-->>Cu: error (per-tool degradation)
  end

  Cu->>M: postgres_query {sql: "SELECT * FROM users LIMIT 10"}
  M->>M: ValidateSelect (no UNION/EXCEPT/INTERSECT)
  M->>PG: SELECT (read-only role)
  PG-->>M: rows
  M-->>Cu: rows

  Cu->>M: nats_list_streams
  M->>N: js.Streams()
  alt NATS down
    N--xM: error
    M-->>Cu: error (per-tool)
  else ok
    N-->>M: streams
    M-->>Cu: streams
  end

  Cu->>M: redis_get {key}
  M->>R: GET
  alt Redis down
    R--xM: error
    M-->>Cu: error (per-tool)
  else ok
    R-->>M: value
    M-->>Cu: value
  end
Кейсы
Кейс	Что происходит	Решение
Один tool упал	Graceful Degradation per-tool	Остальные работают
SQL Sandbox	String matching обходится	Техдолг: read-only роль + pgx
err.Error()	Утечка DSN/SQL	Техдолг: санитизация
Prydwen не загружен	search_prydwen ошибка	Per-tool degradation
TF-IDF	Хуже embedding	Техдолг: embedding fallback
В сеть нельзя	DSN читает целиком	Только stdio, локально
Что улучшил бы
SQL Sandbox → read-only роль + pgx prepared.

Санитизация ошибок.

Embedding fallback для RAG.

ЧАСТЬ 2. Финальный EVENT_HORIZON_FULL.md — целиком
markdown
# Event Horizon — полный пакет для собеса

> Всё в одном файле: техдолг, шпаргалка, sequence-диаграммы, глубокие разборы, follow-up.
> Для печати / Miro / Cursor.

---

## СОДЕРЖАНИЕ

1. Elevator pitch
2. 17 сервисов — сводная таблица
3. Техдолг (финальный)
4. 17 вопросов — 17 ответов (простыми словами)
5. 17 sequence-диаграмм (happy path)
6. 17 sequence-диаграмм (edge cases)
7. Глубокие разборы (15 сервисов)
8. Thundering herd / Retry storm / Cache stampede
9. Follow-up: «как тестируешь?», «что если 10k DAU?»
10. Что НЕ говорить

---

## 1. Elevator pitch

> Event Horizon — игровая микросервисная платформа на Go + React. 17 сервисов, 333 паттерна. Целевая модель ~10k DAU, десятки RPS в среднем, пик ×2. Не 100k RPS — осознанно: correctness + ops. Локально: 20 VU browse PASS, ~50 RPS, p95 429ms на 8 GiB VM.

**Ключевая фраза:** «10k DAU — это daily active users, не 10k concurrent VUs. Модель требует 8c/32G.»

---

## 2. 17 сервисов — сводная таблица

| # | Сервис | gRPC | Metrics | Хранилища | Паттерны | SLO |
|---|--------|------|---------|-----------|----------|-----|
| 1 | Auth | 50051 | 9091 | PG 5460, Redis 6379 | 13 | p95 login < 500ms |
| 2 | Game | 50052 | 9092 | PG 5461 | 20 | p95 submit < 500ms |
| 3 | Billing | 50053 | 9093 | PG 5462, Redis 6381 | 24 | p95 spend < 200ms |
| 4 | Leaderboard | 50054 | 9094 | PG 5463, Redis 6382 | 20 | p95 top < 50ms |
| 5 | Shop | 50055 | 9095 | PG 5465, Redis 6383 | 30 | p95 purchase < 500ms |
| 6 | Inventory | 50059 | 9096 | PG 5466, Redis 6384 | 28 | p95 create < 300ms |
| 7 | Profile | 50060 | 9099 | PG 5464, Redis 6385 | 14 | p95 get < 100ms |
| 8 | Payment | 50058 | 9103 | PG 5467, Redis 6386 | 21 | p95 webhook < 500ms |
| 9 | Authors | 50061 | 9104 | PG 5468, Redis 6387 | 20 | p95 approve < 500ms |
| 10 | History | 50062 | 9105 | PG 5469 | 16 | p95 get < 100ms |
| 11 | Analytics | 50057 | 9106 | ClickHouse 8123 | 14 | p95 dau < 500ms |
| 12 | Notification | 50063 | 9102 | PG 5470 | 16 | p95 list < 100ms |
| 13 | Fulfillment | — | 9101 | — | 15 | p95 assembly < 12s |
| 14 | Gateway | 8081–8083 | 9095–9097 (внутри) | Redis 6379 | **41** | p95 < 200ms |
| 15 | Balancer | 8079 | 9098 | — | 15 | p95 < 50ms |
| 16 | NATS Hub | — | 9097 | NATS 4222–4224 | 11 | connected = 1 |
| 17 | MCP | stdio | — | PG/Redis/NATS (read) | 15 | tool p95 < 1s |

**Важно:** Gateway `:9095/:9096/:9097` — внутри контейнера. На хосте — Shop/Inventory/NATS Hub.

---

## 3. Техдолг (финальный)

### P0 — блокеры (7)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 1 | Fulfillment | Нет Outbox | Transactional Outbox |
| 2 | Fulfillment | `time.After(10s)` заглушка | WMS API |
| 3 | Gateway | `runGateway` God Function | Разбить на модули |
| 4 | Shop | Saga `started` > 60s не берётся | Воркер берёт `started`, проверяет Billing |
| 5 | Billing | `RewardConsumer` `time.Now()` | `event.ReferenceID` |
| 6 | Authors | `RevertApplication` hard delete | Soft delete |
| 7 | Balancer | Single point | Несколько Balancer + VIP |

### P1 — надёжность (20)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 8 | Gateway | Два Hub | Унифицировать |
| 9 | Gateway | Async gRPC в admin/users без CB | `throughBreaker` |
| 10 | Shop | `paymentClient` nil | Метрика `merch_gate_disabled` |
| 11 | Shop | `kafkaProd` не используется | Убрать/подключить |
| 12 | Shop | `cancelPurchase` может задвоить | Idempotency Key |
| 13 | Inventory | Два драйвера | Определиться |
| 14 | Inventory | `toPostgresFilter` `fmt.Sprintf` | Переписать |
| 15 | Inventory | Outbox type assertion | Явный интерфейс |
| 16 | MCP | SQL Sandbox string matching | Read-only роль + pgx |
| 17 | MCP | `err.Error()` утечка | Санитизация |
| 18 | Balancer | `/ready` не проверяет | Реальная проверка |
| 19 | Balancer | Static registry | etcd/consul |
| 20 | Balancer | Нет health checks | Ping `/health` |
| 21 | Notification | Poison message | Ack + DLQ |
| 22 | Analytics | `fmt.Sprintf` db name | Параметризовать |
| 23 | Leaderboard | Restore при PG down | Retry/backoff |
| 24 | Payment | Нет сверки суммы webhook | Сверять `amount` |
| 25 | History | RetentionWorker в процессе | K8s CronJob |
| 26 | Auth | Нет ротации JWT_SECRET | Ротация |
| 27 | Все | Нет WAF | WAF |

### P2 — гигиена (13)

| # | Сервис | Проблема | Решение |
|---|--------|----------|---------|
| 28 | Gateway | `log.Printf` | `slog` |
| 29 | Balancer | Пустые пакеты | Удалить |
| 30 | Balancer | `log.Printf` hot path | Structured |
| 31 | Fulfillment | UUID несогласованность | Унифицировать |
| 32 | Game | `redis-game` не подключён | Подключить/убрать |
| 33 | Billing | `GetBalance` `-1` при miss | `(int, bool)` |
| 34 | Auth | `boosty.to` захардкожен | Config |
| 35 | Payment | `subscriptionDays` 30 | Явно в config |
| 36 | Authors | JWT старый после approve | Автоперевыпуск |
| 37 | Analytics | 5 субъектов в 1 consumer | Разделить |
| 38 | History | 5 субъектов в 1 consumer | Разделить |
| 39 | Notification | Retry Fixed Delay | Exponential |
| 40 | Profile | `UpdateProfile` наружу | Отдельный gRPC |

### Roadmap (10)

| # | Что |
|---|-----|
| 41 | i18n (RU/EN/DE/ES) |
| 42 | Пагинация каталога |
| 43 | Уровни лидерборда 2–3 |
| 44 | Порог покрытия в CI ≥70% |
| 45 | Retry с джиттером, bulkhead |
| 46 | 108 авторов |
| 47 | 3D |
| 48 | Helm как повседневный деплой |
| 49 | Loki / Tempo |
| 50 | Пентест |

### Закрыто (15)

v1.1.0–v1.1.3: `/api/v1`, буст, Grafana, k6, `purchase.paid` с `purchases.id`, Spend Cycle, Outbox `SKIP LOCKED`, WS Ticket, Docs Gate, Revenue Split, Redis 503, Kafka молчит, Grafana paid→fulfilled, Patterns catalog, Frontend polish.

---

## 4. 17 вопросов — 17 ответов (простыми словами)

### 1. Auth?
Единственный, кто знает пароль. bcrypt 12 → пара JWT. `jti` в Redis — сессия жива. Gateway зовёт `ValidateToken`. Redis упал → 503.

**Необычное:** Session Store в Redis — подпись верна, `jti` нет → токен мёртв.

### 2. Game?
Принимает счёт, валидирует по правилам (Strategy по `game_id`), пишет рекорд + outbox `score.updated` в одной TX. Воркер публикует `SKIP LOCKED`. Буст — Billing синхронно.

**Необычное:** Outbox worker — два процесса не шлют дважды.

### 3. Billing?
Двигает валюту. Списание — `UPDATE WHERE balance >= $1` (атомарно), `transactions` + `balance_after` (Double-Entry), outbox — в одной TX. Повтор с `reference_id` не спишет дважды.

**Необычное:** Conditional UPDATE — защита от lost update.

### 4. Leaderboard?
Потребляет `score.updated`. Redis Sorted Set `ZAddArgs{GT: true}` (хуже не перезаписывает) + PG бэкап. Читает из Redis. Event Chaining: `record_beaten`.

**Необычное:** Ack Batcher — подтверждения батчем.

### 5. Shop?
Сага на две базы. `purchase_sagas` — оркестратор: `started → spent → stocked` / `compensating → compensated`. Spend Cycle `shop:{user}:{item}:{cycle}` — один ключ на цикл.

**Необычное:** Resume Worker через 30s дочитывает сагу.

### 6. Inventory?
Каталог автора. Strategy (PG/Mongo) по `INVENTORY_DRIVER`. Decorator `CachedRepository` — Cache-Aside. Outbox `inventory.item.created` в той же TX. Shop потребляет.

**Необычное:** Outbox через type assertion (хрупко).

### 7. Profile?
Агрегатор. Rule Engine `EvaluateAchievements` — чистая функция. Pipeline: `Evaluate → Filter → Unlock` (`ON CONFLICT`). `_partial: true` если Billing/Game недоступны.

**Необычное:** Rule Engine — чистая функция.

### 8. Payment?
Boosty. HMAC + `ConstantTimeCompare`. В одной TX: платёж, подписка, outbox, `revenue_splits` 85/15. Повтор webhook → 200, не открывает второй раз.

**Необычное:** Revenue Split — журнал, не перевод.

### 9. Authors?
State Machine `pending → approved / rejected`. Approve: `FOR UPDATE`, TX, синхронный Auth `UpdateRole`. Auth упал → `RevertApplication`.

**Необычное:** Compensating Action.

### 10. History?
Лента. 5 субъектов. Retention 30d. Payload Sanitization. Nak при ошибке.

**Необычное:** RetentionWorker — cron внутри процесса.

### 11. Analytics?
ClickHouse. 5 субъектов. Батчинг. Parameterized Query. Cohort Retention. Только admin.

**Необычное:** Батчинг — too many parts.

### 12. Notification?
In-app + Telegram. Два сервиса в одном. Dual-Bus (NATS primary, Kafka fallback). Fan-Out. Idempotency Key `source_event`.

**Необычное:** soft-fail при logout.

### 13. Fulfillment?
Эмуляция сборки. 10s delay. `purchase.fulfilled` с тем же `purchase_id`. Dual-Transport. Без gRPC, БД.

**Необычное:** Нет Outbox — техдолг.

### 14. Gateway?
Единственная HTTP-точка. Rate Limit → Auth → CB → gRPC. CB + Bulkhead: 503 fail-fast. WS Ticket — одноразовый билет.

**Необычное:** BFF + DTO Mapper + `_partial`.

### 15. Balancer?
Самописный L7, Least Connections. Atomic counter. Reverse Proxy. Gauge. ~10 МБ.

**Необычное:** WS sticky.

### 16. NATS Hub?
Bootstrap JetStream. Stream `EVENTS`. Idempotent Creation: `AddStream` → `UpdateStream`. Wildcard `event.>`.

**Необычное:** Единственный, кто создаёт stream.

### 17. MCP?
stdio для Cursor. 6 tools. Graceful Degradation per-tool. SQL Sandbox.

**Необычное:** Offline RAG (TF-IDF).

---

## 5. 17 sequence-диаграмм (happy path)

[Вставить 17 диаграмм из предыдущего ответа — Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Notification, Fulfillment, Gateway, Balancer, NATS Hub, MCP]

---

## 6. 17 sequence-диаграмм (edge cases)

[Вставить 17 диаграмм из предыдущего ответа — Auth refresh reuse, Game NATS down, Billing RewardConsumer, Leaderboard record_beaten лавина, Shop Saga `started` дыра, Inventory version conflict, Profile Billing down, Payment HMAC bad, Authors Auth down, History битый JSON, Analytics too many parts, Notification Telegram down, Fulfillment ack был, Gateway Rate Limiter Redis down, Balancer backend dead, NATS Hub stream не создан, MCP SQL Sandbox обходится]

---

## 7. Глубокие разборы (15 сервисов)

### 7.1. Shop Saga (6 состояний)
[Вставить из Части 2 предыдущего ответа]

### 7.2. Gateway (CB + Rate Limiter + WS Ticket)
[Вставить]

### 7.3. Shop + Inventory (каталог)
[Вставить]

### 7.4. Auth + Roles (RBAC matrix)
[Вставить]

### 7.5. Balancer + Gateway (sticky WS)
[Вставить]

### 7.6. Notification (Dual-Bus + Fan-Out + poison)
[Вставить из Части 2 этого ответа]

### 7.7. Billing (deadlock + Conditional UPDATE)
[Вставить]

### 7.8. Fulfillment (между paid/fulfilled)
[Вставить]

### 7.9. Game (boost + outbox)
[Вставить из Части 1 этого ответа]

### 7.10. Leaderboard (record_beaten + Ack Batcher)
[Вставить]

### 7.11. Profile (Rule Engine + _partial)
[Вставить]

### 7.12. Authors (Apply + Approve + Compensating)
[Вставить]

### 7.13. History (5 subjects + Retention)
[Вставить]

### 7.14. Analytics (батчинг + Parameterized)
[Вставить]

### 7.15. MCP (6 tools + per-tool degradation)
[Вставить]

---

## 8. Thundering herd / Retry storm / Cache stampede

### Thundering herd
- **Auth login** — rate limit 5/s IP (fail-closed).
- **Gateway WS** — rate limit 100/min IP.
- **Notification Fan-Out** — нет rate limit (техдолг).
- **Leaderboard `record_beaten`** — нет rate limit (техдолг).
- **Cache stampede** — нет single-flight (техдолг).

### Retry storm
- **Gateway** — full jitter + CB.
- **Fulfillment** — Idempotency Guard (map в памяти).
- **Notification** — Fixed Delay (техдолг: exponential).
- **Payment** — Idempotent Completion.
- **Client 429** — FE backoff (не факт).

### Cache stampede
- **Billing** — Cache-Aside, нет single-flight.
- **Inventory** — Cache-Aside, нет single-flight.
- **Shop** — Cache-Aside, нет single-flight.
- **Payment** — Cache TTL from Expiry, нет single-flight.
- **Leaderboard Restore** — при старте все читают PG.

---

## 9. Follow-up

### «Как тестируешь?»

| Сервис | Unit | Integration | Load |
|--------|------|-------------|------|
| Auth | мок репозитория, JWT | testcontainers (PG + Redis) | k6 login (краснеет by design) |
| Game | Domain Validators, мок | testcontainers (PG), Failover Test | k6 browse (20 VU) |
| Billing | мок, Race Condition Test | testcontainers (PG + Redis) | k6 purchase (5 VU) |
| Leaderboard | мок, Max-Only | testcontainers (Redis) | k6 browse |
| Shop | мок, Saga состояния | testcontainers (PG) | k6 purchase |
| Inventory | мок, Strategy, Filter | testcontainers (PG + Mongo) | k6 browse |
| Profile | Rule Engine (чистая) | testcontainers (PG + Redis) | k6 browse |
| Payment | HMAC (timing), мок | testcontainers (PG) | webhook simulation |
| Authors | State Machine, Revert | testcontainers (PG) | apply/approve |
| History | Payload Sanitization | testcontainers (PG) | 5 subjects |
| Analytics | мок, Parameterized | testcontainers (ClickHouse) | batch |
| Notification | мок, Idempotency | testcontainers (PG) | Dual-Bus |
| Fulfillment | мок, Idempotency | testcontainers (NATS) | delay |
| Gateway | CB (Open/Half-Open), Rate Limiter (fail-closed/open) | testcontainers (Redis) | k6 browse |
| Balancer | Least Connections (race) | testcontainers | k6 browse |
| NATS Hub | Idempotent Creation | testcontainers (NATS) | — |
| MCP | SQL Sandbox, TF-IDF | testcontainers | JSON-RPC |

### «Что если 10k DAU?»

> «10k DAU ≈ десятки RPS avg, пик ×2. Read path (browse, leaderboard, profile) — Redis + PG, держится. Write path (purchase, submit) — Billing, Shop, Game. Auth write path — узкое место: bcrypt 12 + rate limit 5/s. Read path проверен CORE k6: 20 VU browse PASS, ~50 RPS, p95 429ms на 8 GiB VM. Для 10k DAU хватит 8c/32G. Для 10k одновременных логинов — нет, by design. Что улучшил бы: горизонтальный scale Gateway, read replicas PG, single-flight на кэш.»

---

## 10. Что НЕ говорить

1. «Auth упал, всё упало» — публичные работают.
2. «10k RPS» — это 10k DAU, десятки RPS.
3. «Kafka в AdTime» — там NATS. Kafka — Roolz.
4. «Mongo в Shop» — Mongo experimental в Inventory.
5. «Profile без Redis» — Profile подключён (`:6385`).
6. «Game кэширует в Redis» — `redis-game` не подключён.
7. «Notification `:50056`» — актуальный `:50063`.
8. «Gateway `:9095` на хосте» — это Shop.
9. «MCP в player path» — stdio, dev-tool.
10. «Balancer — не single point» — single point.

---

## 11. 5 ключевых проблем (P0)

1. Fulfillment без Outbox.
2. Shop Saga `started` > 60s.
3. Balancer single point.
4. `RewardConsumer` `time.Now()`.
5. `runGateway` God Function.

## 12. 5 ключевых «страшилок»

1. Thundering herd — rate limit.
2. Retry storm — full jitter + CB.
3. Cache stampede — нет single-flight (техдолг).
4. Poison message — нет Ack + DLQ (техдолг).
5. Saga stuck — дыра (техдолг).

---

## Контакты

- GitHub: github.com/Eastwesser/event-horizon
- Boosty: boosty.to/eastwesser
- Telegram: @s1ntezc0der
ЧАСТЬ 3. Follow-up: «как тестируешь?», «что если 10k DAU?»
3.1. «Как тестируешь?» — развёрнутый ответ
Уровни тестирования
1. Unit (быстрые, без IO)

Мок репозитория (интерфейс).

Rule Engine в Profile — чистая функция, легко.

Domain Validators в Game — чистые функции.

State Machine в Authors — переходы.

Max-Only Write в Leaderboard — GT: true.

Race Condition Test в Billing — конкурентное списание.

2. Integration (testcontainers)

Реальный PG в Docker.

Реальный Redis в Docker.

Реальный ClickHouse (Analytics).

Реальный Mongo (Inventory, //go:build integration).

Реальный NATS (Fulfillment, NATS Hub).

3. Load (k6 CORE)

browse.js — 20 VU, ~50 RPS, p95 429ms PASS.

purchase.js — 1–5 VU, p95 196ms PASS.

50 VU — p95 960ms (soft fail, latency only).

Legacy 500 VU login — не pass/fail, демонстрация защиты.

4. Race (Billing)

go test -race — конкурентное списание.

Conditional UPDATE + version.

Deadlock → retry.

5. Failover (Game)

Тест отказа.

Outbox worker при падении NATS.

6. Thorough (Gateway)

CB: Open → Half-Open → Closed.

Retry: full jitter.

Rate Limiter: fail-closed / fail-open.

Что сказать на собесе
«Unit — мок репозитория, чистые функции (Rule Engine, Domain Validators, State Machine). Integration — testcontainers (PG, Redis, ClickHouse, Mongo, NATS). Load — k6 CORE (20 VU browse PASS, 50 VU soft fail по latency). Race — go test -race на Billing. Failover — Game. Thorough — Gateway (CB, Retry). Legacy 500 VU login — не pass/fail, демонстрация защиты (bcrypt 12 + rate limit).»

3.2. «Что если 10k DAU?» — развёрнутый ответ
Модель
text
DAU = 10 000
Действий на человека = 10–20 (игровые сессии)
RPS avg = 10 000 × 15 / 86 400 ≈ 1.7 (но это игровые сессии)
Пик = ×2 ≈ 3.4 RPS

Но: API-запросов больше, чем сессий.
Если 10 API-запросов на сессию:
RPS avg ≈ 17, пик ≈ 35

Альтернативная оценка: 30–50 avg, пик ~100 (если сессия «тяжелее»).
Где узкие места
Компонент	Узкое место	Почему
Auth write	bcrypt 12 + rate limit 5/s	~200ms на хеш, 5/s IP
Billing write	Conditional UPDATE + FOR UPDATE	Deadlock возможен
Shop write	Saga на две базы	resumeSagas при падении
PG connections	max_connections	13 PG × pool 25
NATS consumer lag	5 субъектов в одном	Lag по всем
Что держится
Компонент	Почему
Read path (browse, leaderboard, profile)	Redis + PG, Cache-Aside
Leaderboard	Redis Sorted Set, O(log N)
Gateway	Circuit Breaker, Bulkhead, Rate Limiter
NATS	3 ноды, durable consumers
Что улучшил бы
Read replicas PG — для read path.

Single-flight на кэш — cache stampede.

Горизонтальный scale Gateway — больше реплик.

Auth: argon2id — современнее bcrypt.

Rate limit на record_beaten — лавина.

DLQ для poison — Notification.

Что сказать на собесе
«10k DAU ≈ 17–50 RPS avg, пик 35–100. Read path держится: Redis + PG, CORE k6 20 VU PASS. Write path — узкое место: Auth bcrypt 12 + rate limit, Billing Conditional UPDATE, Shop Saga. Для 10k DAU хватит 8c/32G. Для 10k одновременных логинов — нет, by design. Улучшил бы: read replicas, single-flight, горизонтальный scale Gateway, DLQ.»

3.3. Другие follow-up
«Почему NATS, а не Kafka?»
«NATS — проще, latency микросекунды, Go-native. Kafka — Roolz (лог, retention 7d, партиции по supplier_id). В EH NATS JetStream: stream EVENTS, durable consumers, outbox. Kafka — opt-in (make deploy-heavy) для purchase.paid / purchase.fulfilled.»

«Почему не 2PC?»
«2PC ломает доступность ради иллюзии одной базы. Координатор блокирует обе базы. Saga + компенсация даёт eventual business consistency. Shop: purchase_sagas оркестратор, Spend Cycle, Resume Worker.»

«Почему Outbox?»
«Dual-write проблема: commit в БД + publish в NATS раздельно → потеря или дубль. Outbox пишет событие в ту же TX. Воркер публикует после commit. Дубликаты возможны — consumer идемпотентный. Outbox в 6 сервисах: Game, Billing, Shop, Inventory, Payment, Authors.»

«Почему ClickHouse?»
«OLAP vs OLTP. DAU/MAU/retention — сканы и агрегаты на миллиардах строк. Postgres — OLTP. ClickHouse — колоночный, сжатие, GROUP BY за секунды. Единственный сервис Analytics. Anti-pattern: не ставить MAU SQL на History Postgres.»

«Что если Redis упал?»
«Auth — 503 (Session Store, не игнорируем). Billing — Cache-Aside miss → PG (медленнее). Leaderboard — топ пустой, Restore при рестарте. Shop/Inventory/Profile/Payment/Authors — Cache-Aside miss → PG. Rate Limiter: login fail-closed, общий fail-open.»

«Что если NATS упал?»
«Outbox pending растёт. Воркер ретраит. После рестарта NATS — догоняет. Fulfillment/Notification — Kafka fallback (heavy). NATS Hub — stream жив после создания, но при первом старте — critical path.»

«Что улучшил бы?»
«P0: Fulfillment Outbox, Shop Saga started, Balancer single point, RewardConsumer time.Now(), runGateway. P1: health checks в Balancer, DLQ в Notification, SQL Sandbox в MCP, single-flight на кэш. P2: structured logging, разделить 5 consumer’ов, exponential backoff для Telegram.»

Итого: три задачи
#	Задача	Что внутри
1	7 сервисов глубоко	Game (boost + outbox), Leaderboard (record_beaten + Ack Batcher), Profile (Rule Engine + _partial), Authors (Apply + Approve + Compensating), History (5 subjects + Retention), Analytics (батчинг + Parameterized), MCP (6 tools + per-tool degradation)
2	Финальный EVENT_HORIZON_FULL.md	12 разделов: elevator pitch, сводная таблица, техдолг, 17 вопросов, 17 happy path, 17 edge case, 15 глубоких разборов, thundering herd/retry/cache, follow-up, что НЕ говорить, 5 проблем, 5 страшилок
3	Follow-up	«Как тестируешь?» (6 уровней), «Что если 10k DAU?» (модель, узкие места, что улучшил бы), 7 других follow-up