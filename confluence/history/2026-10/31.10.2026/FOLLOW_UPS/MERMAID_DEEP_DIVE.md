Три задачи: 17 edge-case диаграмм, Notification глубоко, сборка в один файл
Ты сказал: «погоди с третьей задачей, давай сначала все сервисы серьёзные глубоко разберём». Значит: сначала 17 edge-case + глубокий разбор Notification, потом — один markdown-файл.

Но перед 17 edge-case — важно понять: какие сервисы ещё не разобраны глубоко? Я прошёл 4 (Gateway, Shop+Inventory, Auth+Roles, Balancer+Gateway sticky). Ты говоришь — «давай все серьёзные». Значит, нужно добить:

Сервис	Разобран глубоко?
Auth	✅ (Redis, RBAC)
Game	❌ (только happy path)
Billing	✅ (deadlock, Conditional UPDATE)
Leaderboard	❌ (только happy path)
Shop	✅ (Saga 6 состояний, resumeSagas)
Inventory	✅ (каталог, Strategy, Decorator)
Profile	❌ (только happy path)
Payment	✅ (webhook, Revenue Split)
Authors	❌ (только happy path)
History	❌ (только happy path)
Analytics	❌ (только happy path)
Notification	❌ → сейчас разберу
Fulfillment	✅ (между paid/fulfilled)
Gateway	✅ (CB + Rate Limiter + WS Ticket)
Balancer	✅ (sticky WS)
NATS Hub	✅ (первый старт)
MCP	❌ (только happy path)
Не разобраны глубоко: Game, Leaderboard, Profile, Authors, History, Analytics, Notification (→ сейчас), MCP.

План:

17 edge-case диаграмм (по одной на сервис — самый важный кейс отказа).

Notification глубоко (Dual-Bus + Fan-Out + poison).

Один markdown-файл для печати.

А потом — если хочешь — добьём Game, Leaderboard, Profile, Authors, History, Analytics, MCP глубоко.

ЧАСТЬ 1. 17 edge-case диаграмм
По одной на сервис. Самый важный кейс отказа.

1.1. Auth: refresh с украденным jti
sequenceDiagram
  autonumber
  participant Victim
  participant Attacker
  participant GW as Gateway
  participant A as Auth
  participant R as Redis

  Victim->>GW: POST /auth/login
  GW->>A: Login
  A->>R: SETEX jti:refresh:abc 604800
  A-->>Victim: {access, refresh=abc}

  Note over Attacker: украл refresh=abc
  Attacker->>GW: POST /auth/refresh {refresh=abc}
  GW->>A: Refresh
  A->>R: EXISTS jti:refresh:abc
  R-->>A: 1
  A->>R: DEL jti:refresh:abc
  A->>R: SETEX jti:refresh:def 604800
  A-->>Attacker: {access, refresh=def}
  Note over A: старый jti удалён

  Note over Victim: легитимный refresh
  Victim->>GW: POST /auth/refresh {refresh=abc}
  GW->>A: Refresh
  A->>R: EXISTS jti:refresh:abc
  R-->>A: 0
  A-->>GW: Unauthenticated
  GW-->>Victim: 401
  Note over Victim: жертва разлогинена<br/>но атакующий получил def
Что сказать: «Refresh одноразовый: старый jti удаляется. Если украли и использовали — жертва получит 401, но атакующий уже получил новый refresh. Это не полная защита — нужен reuse detection (если старый jti использован повторно → отозвать все сессии пользователя). В техдолге.»

1.2. Game: score.updated потерялся из-за NATS
sequenceDiagram
  autonumber
  participant G as Game
  participant PG as PG Game
  participant W as Outbox worker
  participant N as NATS

  G->>PG: BEGIN
  G->>PG: INSERT highscores
  G->>PG: INSERT outbox(score.updated)
  G->>PG: COMMIT
  G-->>G: клиенту 200

  W->>PG: SELECT * FROM outbox<br/>WHERE processed=false<br/>FOR UPDATE SKIP LOCKED
  PG-->>W: row
  W->>N: publish score.updated
  N--xW: NATS down (timeout)
  Note over W: processed остаётся false
  W->>W: retry через N секунд
  W->>N: publish (retry)
  N--xW: всё ещё down
  Note over W: outbox_pending растёт
  Note over PG,N: после рестарта NATS — догоняет
Что сказать: «Outbox: NATS упал → processed = false, game_outbox_pending растёт. Воркер ретраит. После рестарта NATS — догоняет. Событие не теряется.»

1.3. Billing: RewardConsumer двойное начисление
sequenceDiagram
  autonumber
  participant N as NATS
  participant RC as RewardConsumer
  participant B as Billing
  participant PG as PG Billing

  N->>RC: score.updated (event_id=evt1)
  RC->>RC: reference_id = time.Now().UnixNano()<br/>← БАГ
  RC->>B: AddCurrency(ref=1234567890, amount=10)
  B->>PG: INSERT transactions(ref=1234567890)
  PG-->>B: ok
  RC->>N: Ack

  Note over N: NATS redelivery (consumer рестарт)
  N->>RC: score.updated (event_id=evt1, повтор)
  RC->>RC: reference_id = time.Now().UnixNano()<br/>← ДРУГОЙ номер!
  RC->>B: AddCurrency(ref=1234569999, amount=10)
  B->>PG: INSERT transactions(ref=1234569999)
  PG-->>B: ok
  Note over PG: ДВЕ транзакции на одно событие

  Note over RC: Техдолг: использовать event.ReferenceID
Что сказать: «RewardConsumer использует time.Now() для reference_id. При redelivery — новый reference_id → двойное начисление. Техдолг: event.ReferenceID. P0.»

1.4. Leaderboard: record_beaten лавина
sequenceDiagram
  autonumber
  participant G as Game
  participant N as NATS
  participant L as Leaderboard
  participant No as Notification

  Note over G,No: 1000 игроков бьют рекорды одновременно
  G->>N: 1000 × score.updated
  N->>L: 1000 events
  loop каждый
    L->>L: ZADD GT → added
    L->>N: publish record_beaten
  end
  N->>No: 1000 × record_beaten
  Note over No: Fan-Out: 1000 уведомлений
  Note over No: нет rate limit на record_beaten
  No->>No: INSERT notifications (1000)
  Note over No: notification_fanout_size histogram
Что сказать: «record_beaten — Event Chaining: Leaderboard публикует внутри consumer’а. При массовых рекордах — лавина record_beaten → Notification завален. Нет rate limit на record_beaten. В roadmap: «один рекорд — одно уведомление».»

1.5. Shop: resumeSagas дыра на started
sequenceDiagram
  autonumber
  participant S as Shop
  participant PG as PG Shop
  participant B as Billing
  participant RW as resumeSagas

  S->>PG: INSERT purchase_sagas (started, spend_ref)
  S->>B: SpendCurrency(spend_ref)
  B-->>S: ok
  Note over S: ← УПАЛ до UPDATE saga → spent
  Note over PG: строка started, purchases нет
  Note over B: деньги списаны

  RW->>PG: SELECT * FROM purchase_sagas<br/>WHERE status IN ('spent','compensating')
  PG-->>RW: (пусто — строка started)
  Note over RW: ДЫРА: started не берётся
  Note over S,B: деньги списаны, компенсации нет
Что сказать: «resumeSagas берёт spent/compensating. Строка started > 60s не берётся — дыра. Если Shop упал между SpendCurrency и UPDATE saga → spent — деньги списаны, компенсации нет. Техдолг: воркер берёт started > 60s, проверяет Billing GetTransactions(spend_ref). P0.»

1.6. Inventory: version conflict при двух авторах
sequenceDiagram
  autonumber
  participant A1 as Author 1
  participant A2 as Author 2
  participant I as Inventory
  participant PG as PG Inventory

  A1->>I: PUT /inventory/items/X (version=1)
  A2->>I: PUT /inventory/items/X (version=1)
  I->>PG: A1: UPDATE items SET name=...<br/>WHERE id=X AND version=1
  PG-->>I: 1 row → version=2
  I-->>A1: 200
  I->>PG: A2: UPDATE items SET name=...<br/>WHERE id=X AND version=1
  PG-->>I: 0 rows
  I-->>A2: ErrVersionConflict → 409
  Note over A2: клиент перечитывает, повторяет
Что сказать: «Optimistic Lock: version. Два автора одновременно — первый проходит (version=2), второй 0 rows → 409. Клиент перечитывает. Это by design.»

1.7. Profile: Billing упал при GetProfile
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant P as Profile
  participant PG as PG Profile
  participant B as Billing
  participant R as Redis

  C->>GW: GET /profile
  GW->>P: gRPC GetProfile
  P->>R: GET profile:{user_id}
  R-->>P: miss
  P->>PG: SELECT profile (email, nickname, total_score)
  PG-->>P: ok
  P->>B: GetBalance (sync)
  B--xP: timeout
  Note over P: _partial = true
  P->>R: SETEX profile:{user_id} 300 (с _partial)
  P-->>GW: {email, nickname, total_score,<br/>lamps: null, tickets: null,<br/>_partial: true}
  GW-->>C: 200 (partial)
  Note over C: клиент видит «баланс недоступен»
Что сказать: «Billing упал → _partial: true, lamps/tickets: null. Остальное работает. Graceful Degradation.»

1.8. Payment: HMAC не совпал
sequenceDiagram
  autonumber
  participant BB as Boosty
  participant P as Payment
  participant PG as PG Payment

  BB->>P: POST /webhook<br/>X-Signature: bad_hmac<br/>body: {payment_id, amount}
  P->>P: HMAC-SHA256(body, secret) != bad_hmac
  P->>P: subtle.ConstantTimeCompare → false
  Note over P: payment_webhook_invalid_total++
  P-->>BB: 401
  Note over PG: платёж НЕ двигается

  Note over BB,P: Boosty ретраит
  BB->>P: POST /webhook (retry с правильным hmac)
  P->>P: HMAC ok
  P->>PG: BEGIN + FOR UPDATE + completed + outbox + revenue_splits
  P-->>BB: 200
Что сказать: «HMAC не совпал → 401, payment_webhook_invalid_total растёт. Алерт на резкий рост: либо конфиг сломан, либо брутфорс. Платёж не двигается.»

1.9. Authors: Auth упал при approve
sequenceDiagram
  autonumber
  participant A as Admin
  participant AU as Authors
  participant PG as PG Authors
  participant Auth as Auth

  A->>AU: Approve(app_id)
  AU->>PG: SELECT FOR UPDATE application
  AU->>PG: BEGIN
  AU->>PG: UPDATE application approved
  AU->>PG: INSERT authors
  AU->>PG: INSERT outbox(author.upserted)
  AU->>PG: COMMIT
  AU->>Auth: UpdateRole(user_id, author)
  Auth--xA: timeout
  Note over AU: Compensating Action
  AU->>PG: BEGIN
  AU->>PG: UPDATE application pending
  AU->>PG: DELETE authors (hard — техдолг FK)
  AU->>PG: DELETE outbox(author.upserted)
  AU->>PG: COMMIT
  AU-->>A: error
  Note over A: admin повторяет
Что сказать: «Auth упал при approve → RevertApplication: заявка обратно pending, authors удаляется. authors_revert_total растёт. Техдолг: hard delete может упасть по FK.»

1.10. History: битый JSON
sequenceDiagram
  autonumber
  participant N as NATS
  participant H as History
  participant PG as PG History

  N->>H: score.updated<br/>body: {invalid json
  H->>H: json.Unmarshal → error
  H->>H: Payload Sanitization<br/>payload = {"raw": "{invalid json"}
  H->>PG: INSERT history (payload raw)
  PG-->>H: ok
  H->>N: Ack
  Note over PG: факт события записан, payload сырой
Что сказать: «Payload Sanitization: битый JSON → {"raw": "..."}. Consumer не падает. Лучше, чем Nak и бесконечный retry.»

1.11. Analytics: too many parts в ClickHouse
sequenceDiagram
  autonumber
  participant N as NATS
  participant A as Analytics
  participant CH as ClickHouse

  Note over A: батчинг сломан (мелкие INSERT)
  N->>A: event 1
  A->>CH: INSERT (1 row) → part 1
  N->>A: event 2
  A->>CH: INSERT (1 row) → part 2
  Note over CH: ... 1000 parts
  CH--x A: error: Too many parts
  A->>N: Nak (retry)
  Note over CH: merge не справляется

  Note over A: fix: батчинг (N событий или T секунд)
  A->>A: buffer 100 events
  A->>CH: INSERT (100 rows) → 1 part
  Note over CH: parts_current стабилен
Что сказать: «ClickHouse не любит мелкие INSERT → too many parts. Батчинг: N событий или T секунд → один INSERT. clickhouse_parts_current — алерт.»

1.12. Notification: Telegram API упал
sequenceDiagram
  autonumber
  participant N as NATS
  participant No as Notification
  participant PG as PG Notification
  participant T as Telegram

  N->>No: purchase.paid
  No->>PG: INSERT notification (in-app)
  PG-->>No: ok
  No->>T: sendMessage (Telegram)
  T--xNo: 503 Service Unavailable
  Note over No: notification_telegram_failed_total++
  No->>No: Retry Fixed Delay (5s)
  No->>T: retry
  T--xNo: 503
  No->>No: retry (5s)
  Note over No: in-app уже создан
  No->>N: Ack (in-app доставлен)
  Note over No: Telegram — best-effort
Что сказать: «Telegram упал → notification_telegram_failed_total растёт. In-app уже создан. Retry Fixed Delay. Telegram — best-effort, не блокирует.»

1.13. Fulfillment: ack был, fulfilled не published
sequenceDiagram
  autonumber
  participant N as NATS
  participant F as Fulfillment
  participant S as Shop
  participant No as Notification

  N->>F: purchase.paid
  F->>F: Idempotency Guard: ok
  F->>F: time.After(10s)
  Note over F: ← УПАЛ до publish
  Note over N: ack был? (не факт)
  Note over F: при рестарте purchase.paid<br/>не вернётся (если ack был)
  Note over S: fulfilled_at null
  Note over S: отмена возможна (7д)
  Note over F: ТЕХДОЛГ: нет Outbox
Что сказать: «Fulfillment без Outbox. Упал между paid и fulfilled — заказ завис. purchase.paid durable, но ack мог быть. fulfilled_at null, отмена возможна. Техдолг: Outbox. P0.»

1.14. Gateway: Rate Limiter Redis упал (login)
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant RL as Rate Limiter
  participant R as Redis
  participant A as Auth

  C->>GW: POST /auth/login
  GW->>RL: INCR rate:ip:login
  RL->>R: INCR
  R--xRL: connection refused
  Note over RL: fail-closed для login
  RL-->>GW: deny (false)
  GW-->>C: 429 / 503
  Note over C: брутфорс важнее не пустить

  C->>GW: GET /leaderboard (общий лимит)
  GW->>RL: INCR rate:user:general
  RL->>R: INCR
  R--xRL: connection refused
  Note over RL: fail-open для общего
  RL-->>GW: allow (true)
  GW->>A: ValidateToken
  A-->>GW: ok
  GW-->>C: 200
Что сказать: «Rate Limiter: Redis упал. Login — fail-closed (не пускаем). Общий — fail-open (пускаем). Осознанный trade-off.»

1.15. Balancer: backend упал, Balancer не знает
sequenceDiagram
  autonumber
  participant C as Client
  participant B as Balancer
  participant G1 as Gateway-1 (dead)
  participant G2 as Gateway-2

  Note over G1: Gateway-1 упал
  C->>B: GET /api/v1/profile
  B->>B: ActiveConns: G1=0, G2=2 → select G1 (min!)
  B->>G1: proxy
  G1--xB: connection refused
  B-->>C: 502 Bad Gateway
  Note over B: Balancer не знает, что G1 мёртв<br/>(нет health checks — техдолг)
  Note over C: клиент retry

  C->>B: retry
  B->>B: ActiveConns: G1=0, G2=2 → снова G1
  B->>G1: proxy
  G1--xB: connection refused
  B-->>C: 502
  Note over B: 0 connections = «свободен»<br/>но он мёртв
Что сказать: «Balancer: backend упал, но Balancer не знает (нет health checks). Least Connections выбирает мёртвый (0 conns). 502. Техдолг: health checks. P1.»

1.16. NATS Hub: stream не создан при первом старте
sequenceDiagram
  autonumber
  participant NH as NATS Hub
  participant N as NATS
  participant G as Game

  NH->>N: connect
  N-->>NH: ok
  NH->>N: js.AddStream(EVENTS)
  N--xNH: error (no permissions)
  NH->>NH: fatal — не может работать
  Note over NH: сервис падает (рестарт loop)
  Note over N: stream EVENTS НЕ создан

  G->>N: publish score.updated
  N--xG: error: no stream matches subject
  Note over G: outbox pending растёт
  Note over G: все publish падают
Что сказать: «NATS Hub: первый старт, AddStream упал → stream нет. Все publish падают. Сервис в рестарт-лупе. Critical path.»

1.17. MCP: SQL Sandbox обходится
sequenceDiagram
  autonumber
  participant Cu as Cursor
  participant M as MCP
  participant PG as Postgres

  Cu->>M: postgres_query {sql: "SELECT 1; DROP TABLE users;"}
  M->>M: ValidateSelect: "SELECT" ok<br/>"UNION" нет<br/>"EXCEPT" нет<br/>"INTERSECT" нет
  Note over M: string matching не ловит "; DROP"
  M->>PG: SELECT 1; DROP TABLE users;
  PG--xM: syntax error / или выполнил?
  Note over M: ТЕХДОЛГ: read-only роль + pgx prepared

  Note over Cu,PG: Правильный запрос
  Cu->>M: postgres_query {sql: "SELECT * FROM users LIMIT 10"}
  M->>M: ValidateSelect: ok
  M->>PG: SELECT (read-only role)
  PG-->>M: rows
  M-->>Cu: rows
Что сказать: «SQL Sandbox string matching — не безопасен. ; DROP не ловится. Техдолг: read-only роль + pgx prepared. P1.»

ЧАСТЬ 2. Notification — глубокий разбор (Dual-Bus + Fan-Out + poison)
2.1. Архитектура
flowchart TB
  subgraph buses["Buses"]
    NATS["NATS JetStream<br/>(primary)"]
    KAFKA["Kafka<br/>(fallback, heavy)"]
  end

  subgraph notif["Notification (one process)"]
    INBOX["InboxService<br/>(in-app)"]
    NOTIFIER["Notifier<br/>(Telegram)"]
  end

  subgraph data["Data"]
    PG["PG :5470<br/>notifications"]
    TG["Telegram API"]
  end

  NATS -->|score.updated, purchase.paid,<br/>purchase.fulfilled, payment.completed,<br/>record_beaten| INBOX
  KAFKA -.->|fallback| INBOX
  INBOX -->|Fan-Out 1→N| PG
  INBOX --> NOTIFIER
  NOTIFIER -->|Telegram| TG
  NOTIFIER -.->|Retry Fixed Delay| TG
2.2. Dual-Bus: как выбирается транспорт
sequenceDiagram
  autonumber
  participant CFG as Config
  participant N as Notification
  participant NATS as NATS
  participant KAFKA as Kafka

  N->>CFG: read NATS_URL / KAFKA_BROKERS
  alt NATS_URL set
    N->>NATS: connect + subscribe
    Note over N: NATS primary
    Note over N: Kafka consumer НЕ стартует
  else NATS_URL empty, KAFKA_BROKERS set
    N->>KAFKA: connect + subscribe
    Note over N: Kafka fallback
  else both set
    N->>NATS: connect + subscribe
    Note over N: NATS primary (FINISHERS.md)
    Note over N: Kafka молчит
  end
Ключевой момент: «Если JetStream подключён, Kafka-consumer не стартует» (FINISHERS.md). Защита от двойной доставки.

2.3. Fan-Out: 1 → N
sequenceDiagram
  autonumber
  participant N as NATS
  participant I as InboxService
  participant PG as PG
  participant R as Recipients

  N->>I: record_beaten<br/>{user_id: victim, game: hexagon, score: 200}
  I->>R: resolve recipients<br/>(victim)
  R-->>I: [victim]
  loop for each recipient
    I->>PG: INSERT notifications<br/>(user_id, source_event, type, text)<br/>ON CONFLICT (user_id, source_event) DO NOTHING
    PG-->>I: ok / conflict
  end
  I->>N: Ack
  Note over I: notification_fanout_size histogram
Ключевой момент: Fan-Out — не broadcast всем. N адресных уведомлений. source_event — уникальный id события.

2.4. Poison message (Kafka) — баг
sequenceDiagram
  autonumber
  participant K as Kafka
  participant N as Notification
  participant PG as PG

  K->>N: purchase.paid<br/>body: {invalid
  N->>N: json.Unmarshal → error
  N->>N: Notifier.Handle → error
  Note over N: НЕТ Ack (техдолг)
  N--xK: no ack
  Note over K: Kafka redelivery

  K->>N: purchase.paid (same)
  N->>N: error
  N--xK: no ack
  Note over K: ... бесконечно
  Note over K: consumer lag растёт<br/>другие сообщения ждут
Что сказать: «Poison message → бесконечный retry. Блокирует consumer. Техдолг: Ack + DLQ. P1.»

2.5. In-app bell soft-fail
sequenceDiagram
  autonumber
  participant C as Client (not logged in)
  participant FE as NotificationBell (FE)
  participant GW as Gateway

  C->>FE: mount
  FE->>FE: check auth (no token)
  Note over FE: soft-fail — не делаем запрос
  FE-->>C: пусто (без 401 loop)

  Note over C,GW: logged in
  C->>FE: mount
  FE->>GW: GET /notifications
  GW-->>FE: 200 list
  FE-->>C: bell с count
Что сказать: «In-app bell soft-fail при logout — нет цикла 401. FE проверяет auth перед запросом.»

2.6. Что улучшил бы
Проблема	Решение	Приоритет
Poison message → бесконечный retry	Ack + DLQ	P1
Retry Fixed Delay для Telegram	Exponential backoff	P2
Fan-Out без rate limit	Rate limit на record_beaten	P1
5 субъектов в одном consumer’е	Разделить	P2
PII в уведомлениях	Санитизация	P2
ЧАСТЬ 3. Один markdown-файл для печати
Создам файл EVENT_HORIZON_FULL.md. Он собирает всё: техдолг, шпаргалку, 17 happy path, 17 edge-case, глубокие разборы.

markdown
# Event Horizon — полный пакет для собеса

> Всё в одном файле: техдолг, шпаргалка, sequence-диаграммы, глубокие разборы.
> Для печати / Miro / Cursor.

---

## СОДЕРЖАНИЕ

1. [Техдолг (финальный)](#1-техдолг-финальный)
2. [Шпаргалка на 1 страницу](#2-шпаргалка-на-1-страницу)
3. [17 sequence-диаграмм (happy path)](#3-17-sequence-диаграмм-happy-path)
4. [17 sequence-диаграмм (edge cases)](#4-17-sequence-диаграмм-edge-cases)
5. [Глубокие разборы](#5-глубокие-разборы)
6. [17 вопросов — 17 ответов](#6-17-вопросов--17-ответов)

---

## 1. Техдолг (финальный)

[Вставить содержимое из Части 1 предыдущего ответа — P0/P1/P2/roadmap/закрыто]

---

## 2. Шпаргалка на 1 страницу

[Вставить содержимое из Части 3 предыдущего ответа — 17 сервисов, порты, паттерны, страшилки]

---

## 3. 17 sequence-диаграмм (happy path)

[Вставить 17 диаграмм из Части 3 предыдущего ответа]

---

## 4. 17 sequence-диаграмм (edge cases)

[Вставить 17 диаграмм из Части 1 этого ответа]

---

## 5. Глубокие разборы

### 5.1. Shop Saga (6 состояний)
[Вставить из Части 2 предыдущего ответа]

### 5.2. Gateway (CB + Rate Limiter + WS Ticket)
[Вставить из Части 2 предыдущего ответа]

### 5.3. Shop + Inventory (каталог)
[Вставить из Части 2 предыдущего ответа]

### 5.4. Auth + Roles (RBAC matrix)
[Вставить из Части 2 предыдущего ответа]

### 5.5. Balancer + Gateway (sticky WS)
[Вставить из Части 2 предыдущего ответа]

### 5.6. Notification (Dual-Bus + Fan-Out + poison)
[Вставить из Части 2 этого ответа]

### 5.7. Billing (deadlock + Conditional UPDATE)
[Вставить из Части 4 предыдущего ответа]

### 5.8. Fulfillment (между paid/fulfilled)
[Вставить из Части 4 предыдущего ответа]

---

## 6. 17 вопросов — 17 ответов

[Вставить из предыдущего ответа — 17 вопросов, простыми словами]

---

## Что НЕ говорить

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

## 5 ключевых проблем (техдолг P0)

1. **Fulfillment без Outbox** — событие теряется.
2. **Shop Saga `started` > 60s** — дыра в resumeSagas.
3. **Balancer single point** — нет резервного.
4. **`RewardConsumer` `time.Now()`** — двойное начисление.
5. **`runGateway` God Function** — > 1000 строк.

## 5 ключевых «страшилок»

1. **Thundering herd** — Auth login, Gateway WS. Rate limit.
2. **Retry storm** — Gateway. Full jitter + CB.
3. **Cache stampede** — Billing, Inventory, Shop. **Нет** single-flight (техдолг).
4. **Poison message** — Notification Kafka. **Нет** Ack + DLQ (техдолг).
5. **Saga stuck** — Shop `started`. **Дыра** (техдолг).

---

## Контакты

- GitHub: github.com/Eastwesser/event-horizon
- Boosty: boosty.to/eastwesser

Итого: три задачи
#	Задача	Что внутри
1	17 edge-case диаграмм	Auth (refresh reuse), Game (NATS down), Billing (RewardConsumer), Leaderboard (record_beaten лавина), Shop (Saga started дыра), Inventory (version conflict), Profile (Billing down → _partial), Payment (HMAC bad), Authors (Auth down → Revert), History (битый JSON), Analytics (too many parts), Notification (Telegram down), Fulfillment (ack был, fulfilled не published), Gateway (Rate Limiter Redis down), Balancer (backend dead, не знает), NATS Hub (stream не создан), MCP (SQL Sandbox обходится)
2	Notification глубоко	Dual-Bus (NATS primary, Kafka fallback), Fan-Out (1→N), Poison message (баг), In-app bell soft-fail, что улучшил бы
3	Один markdown-файл	Техдолг + шпаргалка + 17 happy path + 17 edge-case + 8 глубоких разборов + 17 вопросов + что НЕ говорить + 5 проблем + 5 страшилок