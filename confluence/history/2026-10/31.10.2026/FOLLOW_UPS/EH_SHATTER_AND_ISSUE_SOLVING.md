Три артефакта: sequence-диаграммы, Shop Saga глубоко, шпаргалка на печать
ЧАСТЬ 1. Sequence-диаграммы (Mermaid) для 8 кейсов
Все диаграммы — для mermaid.live → SVG → Miro / печать.

1.1. Billing: deadlock на Conditional UPDATE
sequenceDiagram
  autonumber
  participant C as Client
  participant B as Billing
  participant PG as Postgres

  C->>B: SpendCurrency(user=A, ref=refA)
  C->>B: SpendCurrency(user=B, ref=refB)
  B->>PG: BEGIN TX1
  B->>PG: BEGIN TX2
  B->>PG: TX1: UPDATE balance WHERE user=A<br/>AND balance >= amt
  B->>PG: TX2: UPDATE balance WHERE user=B<br/>AND balance >= amt
  Note over PG: TX1 lock A, TX2 lock B
  B->>PG: TX1: INSERT transactions(user=A)<br/>→ lock index
  B->>PG: TX2: INSERT transactions(user=B)<br/>→ ждёт lock index
  Note over PG: TX1 держит index, TX2 ждёт<br/>TX2 держит B, TX1 ждёт?
  PG-->>B: DEADLOCK detected
  PG-->>B: ROLLBACK TX2
  B-->>C: error (TX2): retry
  Note over C: retry с тем же reference_id<br/>не спишет дважды
  B-->>C: ok (TX1): balance updated
Ключевые моменты:

Deadlock — PG откатывает одну TX (TX2).

pg_stat_database_deadlocks растёт.

Клиент retry с reference_id — идемпотентно.

SLO: deadlocks = 0, алерт.

1.2. Auth: Redis упал во время ValidateToken
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant A as Auth
  participant R as Redis

  C->>GW: GET /api/v1/profile<br/>Authorization: Bearer
  GW->>A: gRPC ValidateToken(token)
  A->>R: EXISTS jti:abc
  R--x A: connection refused
  Note over R: Redis упал
  A-->>GW: gRPC Unavailable<br/>"session store unavailable"
  GW-->>C: 503 session store unavailable

  Note over C,GW: Не игнорируем ошибку Redis<br/>иначе отозванный токен пройдёт
  Note over R: Рестарт Redis<br/>Сессии потеряны → перелогин
Ключевые моменты:

Auth не игнорирует ошибку Redis — иначе отозванный токен пройдёт.

Gateway → 503, не 401 (401 = «не знаем кто», 503 = «не можем проверить»).

Рестарт Redis → сессии потеряны → перелогин.

1.3. Gateway: Circuit Breaker открылся на Billing
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant CB as Circuit Breaker
  participant B as Billing

  C->>GW: POST /shop/purchase
  GW->>CB: call SpendCurrency
  CB->>B: gRPC SpendCurrency
  B--x CB: timeout 5s
  CB-->>GW: error
  Note over CB: fail count = 1

  C->>GW: POST /shop/purchase (retry)
  GW->>CB: call SpendCurrency
  CB->>B: gRPC SpendCurrency
  B--x CB: timeout 5s
  CB-->>GW: error
  Note over CB: fail count = 2

  Note over CB: ... 5 fails → CB Open

  C->>GW: POST /shop/purchase
  GW->>CB: call SpendCurrency
  Note over CB: Open → fail-fast
  CB-->>GW: 503 immediately
  GW-->>C: 503 service temporarily unavailable

  Note over CB: half-open через 10s<br/>3 запроса пробных
  C->>GW: POST /shop/purchase
  GW->>CB: call SpendCurrency
  CB->>B: gRPC SpendCurrency (probe)
  B-->>CB: ok
  Note over CB: Closed
  CB-->>GW: ok
  GW-->>C: 200

  Note over C,B: Billing ожил — CB Closed<br/>Если probe fail — Open снова
Ключевые моменты:

5 fails → Open. Запросы → 503 fail-fast (без вызова Billing).

Half-open через 10s: 3 запроса. Успех → Closed, ошибка → Open.

Bulkhead: отдельные пулы на Billing, Shop, Payment — один не топит другой.

1.4. Shop: resumeSagas берёт spent, но purchases есть
sequenceDiagram
  autonumber
  participant RW as resumeSagas
  participant PG as PG Shop
  participant B as Billing
  participant NATS as NATS

  Note over RW: Shop рестарт после падения
  RW->>PG: SELECT * FROM purchase_sagas<br/>WHERE status IN ('spent','compensating')<br/>AND created_at < now() - 30s<br/>FOR UPDATE SKIP LOCKED
  PG-->>RW: rows: [saga_id=1, purchase_id=X, status=spent]

  RW->>PG: SELECT * FROM purchases WHERE id=X
  PG-->>RW: found (purchase_id=X)
  Note over RW: TX прошла — товар выдан

  RW->>PG: UPDATE purchase_sagas<br/>SET status='stocked'<br/>WHERE id=1
  PG-->>RW: ok
  Note over RW: НЕ возвращаем деньги<br/>stocked = финал

  RW->>NATS: (outbox worker)<br/>publish purchase.paid
  NATS-->>RW: ack
  Note over NATS: Shop ставит fulfilled_at<br/>после purchase.fulfilled
Что если purchases нет:

sequenceDiagram
  autonumber
  participant RW as resumeSagas
  participant PG as PG Shop
  participant B as Billing
  participant NATS as NATS

  RW->>PG: SELECT * FROM purchase_sagas<br/>WHERE status IN ('spent','compensating')<br/>AND created_at < now() - 30s<br/>FOR UPDATE SKIP LOCKED
  PG-->>RW: rows: [saga_id=1, purchase_id=X, status=spent]

  RW->>PG: SELECT * FROM purchases WHERE id=X
  PG-->>RW: not found
  Note over RW: TX не прошла — нужна компенсация

  RW->>PG: UPDATE purchase_sagas<br/>SET status='compensating'
  PG-->>RW: ok
  Note over RW: снять лок до RPC

  RW->>PG: INSERT outbox(billing.refund,<br/>refund_ref=refund:shop:u:i:0)
  PG-->>RW: ok

  RW->>B: AddCurrency(refund_ref, amount)
  B->>B: UPDATE balance += amount<br/>INSERT transactions(refund_ref)<br/>UNIQUE → не задвоит
  B-->>RW: ok

  RW->>PG: UPDATE purchase_sagas<br/>SET status='compensated',<br/>cycle = cycle + 1
  PG-->>RW: ok
  Note over RW: следующая покупка получит новый spend_ref

  RW->>NATS: (outbox worker)<br/>publish billing.refund
Ключевые моменты:

purchases есть → stocked, деньги не возвращаются.

purchases нет → compensating → AddCurrency(refund_ref) → compensated.

refund_ref уникален → повтор не задвоит.

Лок снимается до RPC в Billing.

1.5. Payment: двойной webhook
sequenceDiagram
  autonumber
  participant BB as Boosty
  participant P as Payment
  participant PG as PG Payment
  participant NATS as NATS

  Note over BB,P: Webhook #1
  BB->>P: POST /webhook<br/>X-Signature: hmac<br/>body: payment_id=X
  P->>P: HMAC-SHA256 verify<br/>ConstantTimeCompare
  alt HMAC invalid
    P-->>BB: 401
  end

  P->>PG: BEGIN
  P->>PG: SELECT * FROM payments<br/>WHERE id=X FOR UPDATE
  PG-->>P: status=pending
  P->>PG: UPDATE payments SET status='completed'
  P->>PG: UPDATE subscriptions<br/>SET expires_at=now()<br/>WHERE user_id=Y
  P->>PG: INSERT subscriptions<br/>(user_id=Y, plan, expires_at=now()+30d)
  P->>PG: INSERT outbox(payment.completed)
  P->>PG: INSERT revenue_splits<br/>(author=170, platform=30)
  P->>PG: COMMIT
  PG-->>P: ok
  P->>NATS: (outbox worker)<br/>publish payment.completed
  P-->>BB: 200

  Note over BB,P: Webhook #2 (повтор)
  BB->>P: POST /webhook<br/>X-Signature: hmac<br/>body: payment_id=X
  P->>P: HMAC verify
  P->>PG: BEGIN
  P->>PG: SELECT * FROM payments<br/>WHERE id=X FOR UPDATE
  PG-->>P: status=completed
  P->>PG: COMMIT
  Note over P: Idempotent Completion:<br/>не открываем второй раз
  P-->>BB: 200
Ключевые моменты:

HMAC invalid → 401, payment_webhook_invalid_total растёт.

Webhook #1 → completed, подписка, outbox, revenue_splits.

Webhook #2 → видит completed → 200, не открывает второй раз.

Без этого подписка продлилась бы на 60 дней.

1.6. Fulfillment: упал между paid и fulfilled
sequenceDiagram
  autonumber
  participant S as Shop
  participant NATS as NATS
  participant F as Fulfillment
  participant PG as PG Shop
  participant N as Notification

  Note over S,F: Нормальный путь
  S->>NATS: publish purchase.paid<br/>(purchase_id=X)
  NATS->>F: consume purchase.paid
  F->>F: Idempotency Guard: ok
  F->>F: time.After(10s) — эмуляция
  F->>NATS: publish purchase.fulfilled<br/>(purchase_id=X)
  NATS->>S: consume purchase.fulfilled
  S->>PG: UPDATE purchases<br/>SET fulfilled_at=now()<br/>WHERE id=X
  NATS->>N: consume purchase.fulfilled
  N->>N: in-app bell "заказ собран"

  Note over S,F: Кейс: Fulfillment упал после consume, до publish
  S->>NATS: publish purchase.paid
  NATS->>F: consume purchase.paid
  F->>F: Idempotency Guard: ok
  F->>F: time.After(10s)
  Note over F: ← УПАЛ
  Note over S,PG: purchase.paid ack был?
  Note over S,PG: Если ack был — заказ завис<br/>fulfilled_at null
  Note over S,PG: Если ack не было — NATS вернёт<br/>Fulfillment повторит

  Note over S,F: Техдолг: нет Outbox<br/>Событие может потеряться
Ключевые моменты:

Нет Outbox → событие может потеряться (техдолг).

Если ack был — заказ «завис»: fulfilled_at null, отмена (7 дней) возможна.

Если ack не было — NATS вернёт, Fulfillment повторит.

Idempotency Guard (map в памяти) — при рестарте теряется.

1.7. Notification: poison message из Kafka
sequenceDiagram
  autonumber
  participant K as Kafka
  participant N as Notification
  participant T as Telegram
  participant PG as PG Notification

  Note over K,N: Poison message (битый JSON)
  K->>N: consume purchase.paid<br/>body: {invalid json
  N->>N: json.Unmarshal → error
  N->>N: Notifier.Handle → error
  Note over N: НЕТ Ack (техдолг)
  N--x K: no ack
  Note over K: Kafka ретраит

  K->>N: consume purchase.paid<br/>(тот же битый)
  N->>N: error
  N--x K: no ack
  Note over K: ... бесконечно<br/>notification_poison_messages_total растёт

  Note over K,N: Техдолг: Poison message → Ack + DLQ
  Note over K,N: Сейчас битое сообщение<br/>блокирует consumer

  Note over K,N: Нормальное сообщение
  K->>N: consume purchase.paid<br/>body: valid
  N->>N: json.Unmarshal → ok
  N->>PG: INSERT notifications<br/>ON CONFLICT (user_id, source_event)<br/>DO NOTHING
  PG-->>N: ok
  N->>T: Telegram send (optional)
  T-->>N: ok / error
  N->>K: ack
Ключевые моменты:

Битый JSON → бесконечный retry (техдолг).

Нужен: Ack + DLQ (dead-letter queue).

Нормальное сообщение: ON CONFLICT → идемпотентно.

1.8. NATS Hub: первый старт, stream не создан
sequenceDiagram
  autonumber
  participant NH as NATS Hub
  participant N as NATS cluster
  participant P as Publisher (Game)

  Note over NH,N: Нормальный путь (первый старт)
  NH->>N: connect
  N-->>NH: ok
  NH->>N: js.AddStream(EVENTS, subjects, FileStorage,<br/>MaxAge=7d, MaxBytes=1GB)
  N-->>NH: ok (stream created)
  NH->>N: js.Subscribe("event.>", handler)
  N-->>NH: ok
  NH->>NH: /health → nc.IsConnected() = true

  P->>N: publish score.updated
  N-->>P: ack
  Note over N: stream принимает

  Note over NH,N: Кейс: AddStream упал
  NH->>N: connect
  N-->>NH: ok
  NH->>N: js.AddStream(EVENTS)
  N--x NH: error (например, no permissions)
  NH->>NH: fatal — stream не создан
  Note over NH: Сервис падает (не может работать)

  P->>N: publish score.updated
  N--x P: error: no stream matches subject
  Note over P: Все publish падают

  Note over NH,N: Кейс: stream уже есть (рестарт)
  NH->>N: js.AddStream(EVENTS)
  N--x NH: AlreadyExists
  NH->>N: js.UpdateStream(EVENTS)
  N-->>NH: ok (конфиг обновлён)
  Note over NH: Idempotent Stream Creation
Ключевые моменты:

Первый старт: AddStream. Упал — stream нет, все publish падают.

Рестарт: AddStream → AlreadyExists → UpdateStream. Не падает.

Stream Config as Code — версионируется.

/health — nc.IsConnected().

ЧАСТЬ 2. Shop Saga — глубокий разбор (6 состояний)
2.1. Все 6 состояний
stateDiagram-v2
  [*] --> started: PurchaseItem<br/>INSERT purchase_sagas
  started --> aborted: Billing.SpendCurrency<br/>отказ (нет баланса)
  started --> spent: Billing.SpendCurrency<br/>ok или AlreadyExists
  spent --> stocked: TX Shop<br/>(stock, purchases, outbox) ok
  spent --> compensating: TX Shop<br/>упала (stock=0, version conflict)
  compensating --> compensated: AddCurrency(refund_ref)<br/>ok
  compensating --> compensating: AddCurrency fail<br/>(retry)
  stocked --> [*]
  aborted --> [*]
  compensated --> [*]

  note right of started
    spend_ref = shop:user:item:cycle
    cycle из purchase_claims
  end note

  note right of spent
    Деньги списаны
    purchases ещё нет
  end note

  note right of stocked
    purchases есть
    stock-1
    outbox shop.purchased + purchase.paid
    fulfilled_at = null (пока Fulfillment)
  end note

  note right of compensating
    Возврат в процессе
    refund_ref = refund:shop:user:item:cycle
  end note

  note right of compensated
    Деньги возвращены
    cycle += 1
    Следующая покупка — новый ключ
  end note
2.2. Переходы (таблица)
Из	В	Триггер	Что происходит
[*]	started	PurchaseItem	INSERT purchase_sagas + spend_ref
started	aborted	Billing отказ (нет баланса)	UPDATE saga → aborted, ничего не двигается
started	spent	Billing ok / AlreadyExists	UPDATE saga → spent
spent	stocked	TX Shop ok	UPDATE saga → stocked, purchases записана, outbox
spent	compensating	TX Shop fail (stock=0, version)	UPDATE saga → compensating
compensating	compensated	AddCurrency(refund_ref) ok	UPDATE saga → compensated, cycle += 1
compensating	compensating	AddCurrency fail	Retry (воркер)
stocked	[*]	—	Финал. fulfilled_at ставит Fulfillment позже
aborted	[*]	—	Финал. Клиент может повторить (новый spend_ref? нет — тот же cycle)
compensated	[*]	—	Финал. cycle += 1
2.3. Sequence: полный happy path (6 состояний)
sequenceDiagram
  autonumber
  participant C as Client
  participant GW as Gateway
  participant S as Shop
  participant PG as PG Shop
  participant B as Billing
  participant NATS as NATS
  participant F as Fulfillment

  C->>GW: POST /api/v1/shop/purchase
  GW->>S: gRPC PurchaseItem
  S->>PG: SELECT cycle FROM purchase_claims<br/>WHERE user_id, item_id
  PG-->>S: cycle=0
  S->>PG: INSERT purchase_sagas<br/>(status=started, spend_ref=shop:u:i:0)
  PG-->>S: ok

  rect rgb(220, 240, 255)
    Note over S,B: started → spent
    S->>B: SpendCurrency(spend_ref=shop:u:i:0, amount)
    B->>B: UPDATE balance WHERE >= amount<br/>INSERT transactions(spend_ref)<br/>INSERT outbox(balance.updated)
    B-->>S: ok
    S->>PG: UPDATE saga SET status='spent'
    PG-->>S: ok
  end

  rect rgb(220, 255, 220)
    Note over S,PG: spent → stocked
    S->>PG: BEGIN
    S->>PG: SELECT stock FOR UPDATE<br/>WHERE item_id
    PG-->>S: stock=10
    S->>PG: UPDATE stock = stock-1 WHERE version=$v
    S->>PG: INSERT inventory(user_id, item_id)
    S->>PG: INSERT purchases(id=X, status=COMPLETED,<br/>refundable_until=now()+7d)
    S->>PG: INSERT outbox(shop.purchased, purchase_id=X)
    S->>PG: INSERT outbox(purchase.paid, purchase_id=X)
    S->>PG: COMMIT
    PG-->>S: ok
    S->>PG: UPDATE saga SET status='stocked'
    PG-->>S: ok
  end

  S-->>GW: ok (purchase_id=X)
  GW-->>C: 200 {purchase_id=X}

  rect rgb(255, 240, 220)
    Note over S,F: Outbox worker + Fulfillment
    S->>NATS: publish shop.purchased
    S->>NATS: publish purchase.paid
    NATS->>F: consume purchase.paid
    F->>F: time.After(10s)
    F->>NATS: publish purchase.fulfilled
    NATS->>S: consume purchase.fulfilled
    S->>PG: UPDATE purchases<br/>SET fulfilled_at=now()<br/>WHERE id=X
  end
2.4. Sequence: compensating (stock=0)
sequenceDiagram
  autonumber
  participant S as Shop
  participant PG as PG Shop
  participant B as Billing
  participant NATS as NATS

  Note over S,B: started → spent (деньги списаны)
  S->>B: SpendCurrency(spend_ref=shop:u:i:0)
  B-->>S: ok
  S->>PG: UPDATE saga SET status='spent'

  rect rgb(255, 220, 220)
    Note over S,PG: spent → compensating (stock=0)
    S->>PG: BEGIN
    S->>PG: SELECT stock FOR UPDATE
    PG-->>S: stock=0
    S->>PG: ROLLBACK
    S->>PG: UPDATE saga SET status='compensating'
    PG-->>S: ok
  end

  rect rgb(255, 240, 220)
    Note over S,B: compensating → compensated
    S->>PG: INSERT outbox(billing.refund,<br/>refund_ref=refund:shop:u:i:0)
    PG-->>S: ok
    S->>B: AddCurrency(refund_ref, amount)
    B->>B: UPDATE balance += amount<br/>INSERT transactions(refund_ref)<br/>UNIQUE → не задвоит
    B-->>S: ok
    S->>PG: UPDATE saga SET status='compensated',<br/>cycle = cycle + 1
    PG-->>S: ok
  end

  S->>NATS: (outbox worker) publish billing.refund
  Note over S: Следующая покупка получит<br/>новый spend_ref с cycle=1
2.5. Sequence: Resume Worker на 6 состояниях
sequenceDiagram
  autonumber
  participant RW as resumeSagas (30s)
  participant PG as PG Shop
  participant B as Billing

  Note over RW: Раз в N секунд
  RW->>PG: SELECT * FROM purchase_sagas<br/>WHERE status IN ('spent','compensating')<br/>AND created_at < now()-30s<br/>FOR UPDATE SKIP LOCKED

  PG-->>RW: rows

  loop Для каждой строки
    alt status = spent
      RW->>PG: SELECT * FROM purchases WHERE id=saga.purchase_id
      alt purchases найдено
        RW->>PG: UPDATE saga SET status='stocked'
        Note over RW: TX прошла — stocked,<br/>деньги не возвращаем
      else purchases не найдено
        RW->>PG: UPDATE saga SET status='compensating'
        RW->>B: AddCurrency(refund_ref)
        alt ok
          RW->>PG: UPDATE saga SET status='compensated',<br/>cycle += 1
        else error
          Note over RW: Остаётся compensating<br/>shop_purchase_refund_failures_total++
          Note over RW: Повтор на следующем цикле
        end
      end
    else status = compensating
      RW->>B: AddCurrency(refund_ref) (повтор)
      alt ok
        RW->>PG: UPDATE saga SET status='compensated',<br/>cycle += 1
      else error
        Note over RW: Остаётся compensating<br/>Повтор
      end
    end
  end

  Note over RW: Дыра: строка 'started' старше 60s<br/>воркер НЕ берёт (техдолг)
2.6. Дыра: started старше 60s
Проблема: если Shop упал между INSERT saga (started) и UPDATE saga (spent), но SpendCurrency прошёл — строка started. Воркер её не берёт (берёт только spent/compensating).

Что происходит:

Billing: деньги списаны, transactions с spend_ref.

Shop PG: purchase_sagas строка started. purchases нет.

Воркер: не берёт started.

Деньги списаны, товар не выдан, компенсации нет. Заказ «завис».

Решение (техдолг):

Воркер берёт started старше 60s.

Проверяет Billing GetTransactions(spend_ref):

Есть транзакция → spent → компенсация.

Нет → aborted.

Что сказать на собесе: «В resumeSagas дыра: строка started старше 60s не берётся. Если Shop упал между SpendCurrency и UPDATE saga → spent — деньги списаны, компенсации нет. Решение: воркер берёт started старше 60s, проверяет Billing GetTransactions(spend_ref). В техдолге.»

2.7. Сводка: 6 состояний × 3 кейса
Состояние	Норма	Упал до	Упал после
started	→ spent	aborted (Billing отказ)	дыра: started старше 60s не берётся
spent	→ stocked	compensating (stock=0)	resumeSagas → stocked (purchases есть)
stocked	финал	—	resumeSagas → stocked (идемпотентно)
compensating	→ compensated	compensating (retry)	resumeSagas → compensated
compensated	финал	—	идемпотентно
aborted	финал	—	идемпотентно
ЧАСТЬ 3. Шпаргалка на 1 страницу (markdown для печати)
Создам файл EVENT_HORIZON_CHEATSHEET.md.

markdown
# Event Horizon — шпаргалка (1 страница, A4)

## 17 сервисов: «если упал X → Y»

| # | Если упал | Что не работает | Что работает | Как чинить | Критично? |
|---|-----------|-----------------|--------------|-----------|-----------|
| 1 | Auth | Все закрытые ручки | Публичные, WS существующие | Рестарт. Сессии потеряны | **Да** |
| 2 | Game | Submit, `score.updated` | Leaderboard (Redis), Billing (async) | Рестарт. Outbox догоняет | Нет |
| 3 | Billing | Деньги | Shop (ошибка), Game (буст) | Рестарт. Outbox догоняет | **Да** |
| 4 | Leaderboard | Топ, WS обновления | Game (публикует), Billing (async) | Рестарт. Restore из PG | Нет |
| 5 | Shop | Покупки | Inventory, Payment | Рестарт. `resumeSagas` | **Да** |
| 6 | Inventory | Создание/редактирование | Shop (существующие) | Рестарт. Outbox догоняет | Нет |
| 7 | Profile | Анкета | Auth, Game, Billing | Рестарт. Consumer догоняет | Нет |
| 8 | Payment | Подписки, мерч | Shop (без мерча) | Рестарт. Boosty ретраит | **Да** |
| 9 | Authors | Заявки, approve | Публичные `/authors` | Рестарт. Compensating | Нет |
| 10 | History | Лента | Другие сервисы | Рестарт. Consumer догоняет | Нет |
| 11 | Analytics | DAU/MAU | Другие сервисы | Рестарт. Consumer догоняет | Нет |
| 12 | Notification | In-app, Telegram | Другие сервисы | Рестарт. Consumer догоняет | Нет |
| 13 | Fulfillment | Сборка | Shop (отмена 7д) | Рестарт. **Нет Outbox** | **Да** |
| 14 | Gateway | Все HTTP | Другие реплики | Рестарт. Balancer | **Да** |
| 15 | Balancer | **Всё** | — | Рестарт. **Single point** | **Да** |
| 16 | NATS Hub | Stream (первый) | Stream жив после | Рестарт. Idempotent | **Да** (1й) |
| 17 | MCP | Tools в Cursor | Игроки, 16 сервисов | Рестарт в Cursor | Нет |

## Порты (запомнить)
gRPC: 50051 Auth | 50052 Game | 50053 Billing | 50054 Leaderboard
50055 Shop | 50057 Analytics | 50058 Payment | 50059 Inventory
50060 Profile | 50061 Authors | 50062 History | 50063 Notification
Metrics (host): 9091 Auth | 9092 Game | 9093 Billing | 9094 Leaderboard
9095 Shop | 9096 Inventory | 9097 NATS Hub | 9098 Balancer
9099 Profile | 9101 Fulfillment | 9102 Notification
9103 Payment | 9104 Authors | 9105 History | 9106 Analytics
HTTP edge: 8079 Balancer (front) | 8081-8083 Gateway ×3 | 5173 React (dev)
PG: 5460 Auth | 5461 Game | 5462 Billing | 5463 Leaderboard | 5464 Profile
5465 Shop | 5466 Inventory | 5467 Payment | 5468 Authors | 5469 History
5470 Notification
Redis: 6379 Auth+GW | 6381 Billing | 6382 Leaderboard | 6383 Shop
6384 Inventory | 6385 Profile | 6386 Payment | 6387 Authors
(6380 redis-game — НЕ подключён)
NATS: 4222/4223/4224 (3 ноды) | ClickHouse 8123/9000 | Kafka 19092 (heavy)

text

## Паттерны (333 всего)

| Паттерн | Кто |
|---------|-----|
| Saga | Shop (`purchase_sagas`) |
| Outbox | Game, Billing, Shop, Inventory, Payment, Authors |
| Circuit Breaker + Bulkhead | Gateway |
| WS Ticket | Gateway |
| Strategy | Game (по играм), Inventory (PG/Mongo) |
| Decorator | Inventory (`CachedRepository`) |
| Rule Engine | Profile (`EvaluateAchievements`) |
| State Machine | Authors (`pending → approved / rejected`) |
| HMAC + ConstantTime | Payment |
| Revenue Split | Payment (85/15 — журнал) |
| Retention | History (30d) |
| Cohort Retention | Analytics (ClickHouse) |
| Dual-Bus | Notification (NATS + Kafka) |
| Dual-Transport | Fulfillment (NATS + Kafka) |
| Least Connections | Balancer (atomic counter) |
| Bootstrap | NATS Hub (stream `EVENTS`) |
| MCP Server | MCP (6 tools, stdio) |

## NATS subjects (stream `EVENTS`)

| Subject | Публикует | Потребляет |
|---------|-----------|-----------|
| `score.updated` | Game | Leaderboard, Profile, Billing, History, Analytics, WS |
| `user.registered` | Auth/GW | History, Profile, Analytics |
| `shop.purchased` | Shop | History, Analytics |
| `purchase.paid` | Shop | Fulfillment, Notification |
| `purchase.fulfilled` | Fulfillment | Shop, Notification |
| `payment.completed` | Payment | History, Analytics, Notification |
| `author.upserted` | Authors | History, Analytics |
| `balance.updated` | Billing | (интересующиеся) |
| `record_beaten` | Leaderboard | Notification |
| `inventory.item.created` | Inventory | Shop |

## Ключевые «страшилки»

| Страшилка | Где | Решение |
|-----------|-----|---------|
| Thundering herd | Auth login | Rate limit 5/s IP (fail-closed) |
| Thundering herd | Gateway WS | Rate limit 100/min IP |
| Retry storm | Gateway | Full jitter + Circuit Breaker |
| Cache stampede | Billing, Inventory, Shop | **Нет** single-flight (техдолг) |
| Poison message | Notification Kafka | **Нет** Ack + DLQ (техдолг) |
| Split-brain | NATS | 3 ноды, кворум |
| Deadlock | Billing | PG rollback + retry (`refund_id`) |
| Saga stuck | Shop `started` | **Дыра**: `started` > 60s не берётся |

## Shop Saga: 6 состояний
[*] → started → aborted (Billing отказ)
→ spent → stocked (TX ok)
→ compensating → compensated (возврат)

text

- `spend_ref = shop:{user}:{item}:{cycle}` — один ключ на цикл.
- `refund_ref = refund:shop:{user}:{item}:{cycle}` — уникальный.
- `resumeSagas` (30s): `spent`/`compensating` → `stocked` или `compensated`.
- **Дыра:** `started` > 60s не берётся (техдолг).

## Auth: 3 ключевых факта

1. Session Store в Redis: `jti` есть → сессия жива. Redis упал → 503.
2. Refresh одноразовый: старый `jti` удаляется, роль из PG.
3. bcrypt cost 12 (~200ms) — by design медленно. 500 VU login → 99% fail.

## Payment: 3 ключевых факта

1. HMAC + ConstantTime — защита от timing attack.
2. Idempotent Completion: повтор webhook → 200, не открывает второй раз.
3. Revenue Split 85/15 — журнал, не перевод.

## 5 ключевых проблем (техдолг)

1. **Balancer — single point** (нет резервного).
2. **Fulfillment без Outbox** (событие теряется).
3. **Shop Saga `started` > 60s** (дыра в resumeSagas).
4. **`runGateway` God Function** (> 1000 строк).
5. **Notification poison message** (бесконечный retry).

## Что сказать, если спросят «что самое страшное?»

> «Balancer — single point. Fulfillment без Outbox. Shop Saga `started` — дыра. Auth + Redis — 503 на всё закрытое. `runGateway` — God Function. Но каждое — с планом: health checks в Balancer, Outbox в Fulfillment, `started` > 60s в resumeSagas, CB в Auth, разбить runGateway.»

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
Итого: три артефакта готовы
Артефакт	Что внутри
1. Sequence-диаграммы (8)	Billing deadlock, Auth Redis, Gateway CB, Shop resumeSagas (2 варианта), Payment double webhook, Fulfillment между paid/fulfilled, Notification poison, NATS Hub первый старт
2. Shop Saga (6 состояний)	State diagram, таблица переходов, happy path, compensating, Resume Worker, дыра started > 60s
3. Шпаргалка на 1 страницу	17 сервисов, порты, паттерны, NATS subjects, страшилки, Shop Saga, Auth, Payment, 5 проблем, что НЕ говорить