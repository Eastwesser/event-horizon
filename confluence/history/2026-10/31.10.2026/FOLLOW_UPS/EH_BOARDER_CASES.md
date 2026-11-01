Краевые случаи: глубокий разбор + шпаргалка + thundering herd/retry storm/cache stampede + ключевые сервисы
Три части: (1) глубокий разбор Shop Saga, (2) шпаргалка на 1 страницу, (3) thundering herd / retry storm / cache stampede отдельно. Плюс — глубокие кейсы по ключевым сервисам.

ЧАСТЬ 1. Shop: «упал после Spend, до TX» — sequence
1.1. Нормальный путь
text
Client      Gateway     Shop                    Billing            PG Shop          NATS
  │            │         │                        │                  │               │
  │ POST /purchase        │                        │                  │               │
  ├───────────►│         │                        │                  │               │
  │            │ gRPC PurchaseItem                │                  │               │
  │            ├────────►│                        │                  │               │
  │            │         │ 1. INSERT purchase_sagas (started)          │               │
  │            │         ├───────────────────────►│                  │               │
  │            │         │                        │ INSERT saga row  │               │
  │            │         │                        ├─────────────────►│               │
  │            │         │                        │◄─────────────────┤               │
  │            │         │ 2. SpendCurrency(spend_ref)                │               │
  │            │         ├───────────────────────►│                  │               │
  │            │         │                        │ UPDATE balance   │               │
  │            │         │                        │  WHERE >= amount │               │
  │            │         │                        ├─────────────────►│               │
  │            │         │                        │ INSERT transactions              │
  │            │         │                        │ INSERT outbox(balance.updated)   │
  │            │         │                        │ COMMIT           │               │
  │            │         │                        ├─────────────────►│               │
  │            │         │◄───────────────────────┤ ok               │               │
  │            │         │ 3. UPDATE saga → spent                     │               │
  │            │         ├───────────────────────►│                  │               │
  │            │         │ 4. TX: stock-1, inventory, purchases, outbox               │
  │            │         ├──────────────────────────────────────────►│               │
  │            │         │                        │                  │ COMMIT        │
  │            │         │◄──────────────────────────────────────────┤ ok            │
  │            │         │ 5. UPDATE saga → stocked                   │               │
  │            │         ├───────────────────────►│                  │               │
  │            │◄────────┤ ok                     │                  │               │
  │◄───────────┤         │                        │                  │               │
  │            │         │ 6. outbox worker: publish purchase.paid   │               │
  │            │         ├──────────────────────────────────────────────────────────►│
  │            │         │                        │                  │               │
1.2. Кейс: Shop упал после Spend, до TX
Что случилось: шаг 2 прошёл (деньги списаны, balance.updated в outbox Billing). Шаг 3 (UPDATE saga → spent) — не прошёл (Shop упал). Сага осталась в started. purchases не записана.

Состояние на диске:

Billing PG: баланс уменьшен, transactions с spend_ref, outbox balance.updated (published or not).

Shop PG: purchase_sagas строка started, spend_ref заполнен. purchases — нет. Склад — не тронут.

NATS: balance.updated (если воркер успел).

Sequence: восстановление через Resume Worker

text
Shop (рестарт)     PG Shop                          Billing              NATS
  │                 │                                │                    │
  │ resumeSagas()   │                                │                    │
  ├────────────────►│                                │                    │
  │  SELECT * FROM purchase_sagas                    │                    │
  │  WHERE status IN ('spent','compensating')        │                    │
  │    AND created_at < now() - 30s                  │                    │
  │  FOR UPDATE SKIP LOCKED                          │                    │
  │◄────────────────┤ rows                           │                    │
  │                 │                                │                    │
  │ Для каждой строки:                                │                    │
  │ 1. Проверить purchases по purchase_id            │                    │
  ├────────────────►│                                │                    │
  │◄────────────────┤ не найдено                     │                    │
  │                 │                                │                    │
  │ 2. Нужна компенсация:                            │                    │
  │    INSERT outbox(billing.refund, refund:...)     │                    │
  ├────────────────►│                                │                    │
  │◄────────────────┤ ok                             │                    │
  │                 │                                │                    │
  │ 3. AddCurrency(refund_ref)                       │                    │
  ├─────────────────────────────────────────────────►│                    │
  │                 │                                │ UPDATE balance += │                    │
  │                 │                                │ INSERT transactions│                    │
  │                 │                                │ (refund_ref unique)│                    │
  │                 │                                │ COMMIT             │                    │
  │◄─────────────────────────────────────────────────┤ ok                 │                    │
  │                 │                                │                    │                    │
  │ 4. UPDATE saga → compensated                     │                    │
  │    cycle += 1                                    │                    │
  ├────────────────►│                                │                    │
  │◄────────────────┤ ok                             │                    │
  │                 │                                │                    │
  │ 5. Outbox worker публикует billing.refund        │                    │
  ├──────────────────────────────────────────────────────────────────────►│
  │                 │                                │                    │
Ключевые моменты:

Resume Worker берёт spent / compensating старше 30s. Если строка started — её не берёт (значит, Spend не прошёл, или прошёл, но saga не обновилась — проблема!).

Важно: в коде воркер берёт spent и compensating. Строка started — если Spend прошёл, но UPDATE saga → spent не прошёл — воркер её не увидит. Это дыра. В FINISHERS.md: «Если процесс умер в spent или compensating и строка старше 30 секунд, воркер shop (resumeSagas) берёт её». Строка started — не берётся.

Что делать: либо воркер берёт и started старше N, либо SpendCurrency и UPDATE saga → spent в одной транзакции (но они в разных базах — нельзя). Решение: started старше 60s → проверить Billing GetTransactions(spend_ref) → если транзакция есть → spent → компенсация. Это в техдолге.

Компенсация идемпотентна. refund:shop:{user}:{item}:{cycle} — уникальный ключ. Повтор не задвоит.

Cycle сдвигается. После компенсации cycle += 1. Следующая покупка получит новый spend_ref.

Лок снимается до RPC. FOR UPDATE SKIP LOCKED → прочитали → сняли лок → AddCurrency. Долгий RPC не держит лок.

Что если Billing упал во время компенсации:

AddCurrency → ошибка. Строка outbox billing.refund осталась. shop_purchase_refund_failures_total растёт.

Воркер повторит на следующем цикле.

Если долго — алерт.

Что если Shop упал во время компенсации:

Outbox billing.refund записан, но не опубликован. При рестарте — воркер опубликует.

Сага compensating — воркер снова возьмёт, проверит purchases (не найдено), повторит компенсацию (идемпотентно).

Что если NATS упал:

Outbox billing.refund pending. Компенсация не идёт. shop_purchase_refund_failures_total растёт.

При рестарте NATS — воркер публикует.

1.3. Sequence: Shop упал после TX, до publish
Что случилось: шаг 4 (TX) прошёл. purchases записана, stock-1, outbox shop.purchased + purchase.paid — в outbox. Шаг 5 (UPDATE saga → stocked) — не прошёл. Сага spent, но purchases есть.

Восстановление:

resumeSagas берёт строку spent старше 30s.

Проверяет purchases по purchase_id → найдено.

Ставит stocked, ничего не возвращает (деньги остаются списанными, товар выдан).

Outbox worker публикует shop.purchased + purchase.paid.

Fulfillment → purchase.fulfilled → Shop ставит fulfilled_at.

Ключевой момент: resumeSagas различает «TX прошла» и «TX не прошла» по наличию purchases.id. Это идемпотентный шаг.

1.4. Sequence: двойной клик «купить»
text
Client A   Client B   Gateway   Shop                 Billing            PG Shop
  │          │          │        │                    │                  │
  │ click    │          │        │                    │                  │
  ├─────────►│          │        │                    │                  │
  │          │ click    │        │                    │                  │
  │          ├─────────►│        │                    │                  │
  │          │          │ gRPC   │                    │                  │
  │          ├──────────┼───────►│                    │                  │
  │          │          │        │ 1. Spend Cycle: cycle=0, spend_ref=shop:u:i:0
  │          │          │        ├───────────────────►│                  │
  │          │          │        │                    │ INSERT transactions (spend_ref) │
  │          │          │        │                    │ COMMIT           │                  │
  │          │          │        │◄───────────────────┤ ok               │                  │
  │          │          │        │ 2. TX: stock, purchases, outbox       │                  │
  │          │          │        ├──────────────────────────────────────►│                  │
  │          │          │        │◄──────────────────────────────────────┤ ok               │
  │          │          │        │ 3. UPDATE saga → stocked              │                  │
  │          │          │◄───────┤ ok                 │                  │                  │
  │◄─────────┼──────────┤        │                    │                  │                  │
  │          │          │        │                    │                  │                  │
  │          │          │ gRPC   │                    │                  │                  │
  │          │          ├───────►│                    │                  │                  │
  │          │          │        │ 4. Spend Cycle: cycle=0 (тот же!), spend_ref=shop:u:i:0
  │          │          │        ├───────────────────►│                  │                  │
  │          │          │        │                    │ INSERT transactions (spend_ref) │
  │          │          │        │                    │ → UNIQUE VIOLATION → AlreadyExists
  │          │          │        │◄───────────────────┤ AlreadyExists    │                  │
  │          │          │        │ 5. «такой ключ уже был» → сага spent   │                  │
  │          │          │        │ 6. Проверить purchases по purchase_id? │                  │
  │          │          │        │    (нет purchase_id — второй клик)     │                  │
  │          │          │        │    → создать новый purchase_id         │                  │
  │          │          │        │ 7. TX: stock, purchases (новый), outbox│                  │
  │          │          │        ├──────────────────────────────────────►│                  │
  │          │          │        │◄──────────────────────────────────────┤ 0 rows (stock?)  │
  │          │          │        │    → если stock=0 → compensating       │                  │
  │          │          │◄───────┤ ok / fail          │                  │                  │
  │          │◄─────────┤        │                    │                  │                  │
Ключевой момент: второй клик с тем же cycle → spend_ref тот же → Billing AlreadyExists. Shop не списывает второй раз. Но создаёт новую покупку (если stock есть) — или compensating, если stock=0.

Это by design: один cycle — одно списание. Но покупка может быть создана дважды (если stock есть). Это баг? Нет: purchases primary key — id (UUID). Второй клик создаёт новую покупку с новым id, но деньги списаны один раз. Если stock=0 — компенсация, деньги возвращаются.

Техдолг: cancelPurchase в creditCancel — reference_id = cancel-{purchaseID}. Если покупка дважды отменена, вернёт AlreadyRefunded, но AddCurrency может начислить дважды (пункт 20).

ЧАСТЬ 2. Шпаргалка на 1 страницу: «если упал X → Y»
2.1. Таблица
#	Если упал	Что не работает	Что работает	Как чинить	Критично?
1	Auth	Все закрытые ручки	Публичные (/leaderboard, /authors), WS существующие	Рестарт Auth / Redis. Сессии потеряны → перелогин	Да
2	Game	Submit, score.updated	Leaderboard (из Redis), Billing (async), Shop	Рестарт. Outbox догоняет	Нет
3	Billing	Деньги (Spend/Add)	Shop (ошибка), Game (буст)	Рестарт. Outbox догоняет	Да
4	Leaderboard	Топ, WS обновления	Game (публикует), Billing (async)	Рестарт. Restore из PG	Нет
5	Shop	Покупки	Inventory (каталог), Payment (касса)	Рестарт. resumeSagas догоняет	Да
6	Inventory	Создание/редактирование	Shop (существующие карточки)	Рестарт. Outbox догоняет	Нет
7	Profile	Анкета	Auth, Game, Billing	Рестарт. Consumer догоняет	Нет
8	Payment	Подписки, мерч-гейт	Shop (без мерча)	Рестарт. Boosty ретраит webhook	Да
9	Authors	Заявки, approve	Публичные /authors	Рестарт. Compensating Action	Нет
10	History	Лента	Другие сервисы	Рестарт. Consumer догоняет	Нет
11	Analytics	DAU/MAU	Другие сервисы	Рестарт. Consumer догоняет	Нет
12	Notification	In-app, Telegram	Другие сервисы	Рестарт. Consumer догоняет	Нет
13	Fulfillment	Сборка	Shop (отмена 7 дней)	Рестарт. Нет Outbox — заказ завис	Да
14	Gateway	Все HTTP	Другие Gateway-реплики	Рестарт. Balancer перенаправит	Да
15	Balancer	Всё	—	Рестарт. Single point	Да
16	NATS Hub	Stream (первый старт)	Stream жив после	Рестарт. Idempotent Creation	Да (первый)
17	MCP	Tools в Cursor	Игроки, 16 сервисов	Рестарт в Cursor	Нет
2.2. Что сказать на собесе (одна фраза на сервис)
Сервис	Одна фраза
Auth	Redis упал → 503 на всё закрытое (Session Store)
Game	Outbox FOR UPDATE SKIP LOCKED → событие не теряется
Billing	Conditional UPDATE + Double-Entry → деньги не задвоятся
Leaderboard	Redis Sorted Set GT: true → топ не перезаписывается хуже
Shop	Saga purchase_sagas + Resume Worker → компенсация через 30s
Inventory	Strategy PG/Mongo + Decorator Cache-Aside → Redis упал → PG
Profile	Rule Engine (чистая функция) + _partial → Billing/Game упал → отдаём что есть
Payment	HMAC + ConstantTime + Idempotent Completion → повтор не задвоит
Authors	State Machine + Compensating Action → Auth упал → Revert
History	Retention 30d + Nak Pattern → БД упала → повтор
Analytics	ClickHouse + батчинг → too many parts → merge
Notification	Dual-Bus + Fan-Out + Idempotency Key → повтор не задвоит
Fulfillment	Dual-Transport + Delay 10s → нет Outbox (техдолг)
Gateway	Circuit Breaker + Bulkhead → dependency упал → 503 fail-fast
Balancer	Least Connections → single point (техдолг)
NATS Hub	Bootstrap stream + Idempotent Creation → первый старт → stream
MCP	stdio + 6 tools + Graceful Degradation per-tool → один упал → остальные работают
2.3. Что НЕ говорить
«Auth упал, всё упало» — публичные работают.

«Fulfillment упал, заказ потерян» — нет Outbox, но purchase.paid в NATS (durable), при рестарте догонит. Заказ «завис», но не потерян.

«Balancer упал, ничего не работает» — да, но это single point (техдолг).

«NATS Hub упал, всё упало» — только первый старт. После создания stream — жив.

«MCP в player path» — нет, stdio, dev-tool.

2.4. Ключевые «страшилки» (одна строка каждая)
Страшилка	Где	Решение
Thundering herd	Auth login, Gateway WS, Notification Fan-Out	Rate limit (5/s IP login, 100/min WS)
Retry storm	Gateway Retry	Full jitter + Circuit Breaker
Cache stampede	Redis miss на горячем ключе	Single-flight (сейчас нет — техдолг)
Poison message	Notification Kafka	Ack + лог (техдолг)
Split-brain	NATS кластер	Кворум (3 ноды)
Deadlock	Billing Conditional UPDATE + FOR UPDATE	Retry (PG rollback)
Saga stuck	Shop started старше 30s	resumeSagas берёт spent/compensating, started — дыра (техдолг)
ЧАСТЬ 3. Thundering herd / Retry storm / Cache stampede — отдельно
3.1. Thundering herd
Что это: много клиентов одновременно бьют в один сервис после события (рестарт, сбой, массовое переподключение). Лавина запросов.

Где в Event Horizon
1. Auth login после рестарта

Все клиенты потеряли сессию → бьют /login.

bcrypt cost 12 (~200ms) — CPU-bound. 1000 логинов = 1000 × 200ms = 200s CPU.

Защита: rate limit 5/s IP (fail-closed). Большинство получает 429.

Цена: 99% login fail — by design. Лучше 429, чем положить Auth.

2. Gateway WS после рестарта

Все WS-клиенты переподключаются.

Защита: WS 100/min IP. Reconnect storm → часть 429.

Цена: часть клиентов не переподключится сразу.

3. Notification Fan-Out

Одно событие score.updated → N уведомлений.

Если N большое (например, record_beaten для всех) — нагрузка.

Защита: нет явной. notification_fanout_size (histogram) — контроль.

Техдолг: rate limit на Fan-Out.

4. Leaderboard record_beaten

Много рекордов → лавина record_beaten → Notification завален.

Защита: нет. «Один рекорд — одно уведомление» — в roadmap.

5. Cache stampede на Redis miss

Много параллельных запросов на горячий ключ → все идут в PG.

Защита: нет single-flight. Техдолг: single-flight / SETNX lock.

Как решать (общие принципы)
Принцип	Где в EH
Rate limit	Auth login (5/s), Gateway WS (100/min)
Full jitter	Gateway Retry
Single-flight	Нет (техдолг)
Circuit Breaker	Gateway (fail-fast)
Backpressure	NATS durable consumer (lag)
Bulkhead	Gateway (отдельные пулы)
Что сказать на собесе
«Thundering herd — это лавина после события. В EH защита: rate limit на login (5/s IP, fail-closed), WS (100/min IP), Retry full jitter в Gateway, Circuit Breaker. Чего нет: single-flight на cache stampede — техдолг.»

3.2. Retry storm
Что это: клиент / сервис ретраит запросы, которые не проходят. Ретраи накладываются, нагрузка растёт.

Где в Event Horizon
1. Gateway Retry full jitter

На transient-ошибки (Unavailable, timeout) — retry.

Задержка = random(0, base * 2^attempt). Full jitter размазывает.

Проблема: при массовом отказе dependency — лавина. CB защищает: после серии ошибок — Open, fail-fast.

2. Fulfillment retry

purchase.paid → purchase.fulfilled. При ошибке — Nak → повтор.

Idempotency Guard (map в памяти) защищает от повторной сборки.

Проблема: при рестарте map теряется → повторная сборка возможна. Техдолг.

3. Notification retry Fixed Delay

Telegram API упал → retry с фиксированной задержкой.

Проблема: при долгом сбое — постоянная нагрузка. Exponential backoff был бы лучше.

4. Boosty webhook retry

Payment → Boosty. При 5xx — Boosty ретраит.

Idempotent Completion защищает: повтор видит completed, не открывает второй раз.

5. Client retry на 429

Клиент получает 429 → retry. Если без backoff — лавина.

Защита: FE должен использовать backoff. Сейчас не факт.

Как решать
Принцип	Где в EH
Exponential backoff	Telegram — Fixed (техдолг: exponential)
Full jitter	Gateway Retry
Circuit Breaker	Gateway (fail-fast)
Idempotency	Payment (Completion), Fulfillment (Guard), Billing (reference_id)
Dead-letter queue	Нет (техдолг: DLQ для poison)
Что сказать на собесе
«Retry storm — когда ретраи накладываются. В EH: Gateway Retry с full jitter, Circuit Breaker fail-fast. Idempotency защищает от дублей. Чего нет: DLQ для poison message (техдолг), exponential backoff для Telegram (сейчас fixed).»

3.3. Cache stampede
Что это: много параллельных запросов на один ключ кэша, которого нет (или истёк). Все идут в БД одновременно.

Где в Event Horizon
1. Billing GetBalance

Cache-Aside: GET balance:{user}:{currency} → miss → PG.

Если 1000 параллельных запросов на один баланс → 1000 SELECT в PG.

Защита: нет single-flight. Техдолг.

2. Leaderboard RestoreFromPostgres при старте

При рестарте — все читают из PG одновременно.

Защита: нет. Restore — один раз, но если PG тормозит — лавина.

3. Inventory CachedRepository

Cache-Aside: GET item:{id} → miss → PG/Mongo.

Защита: нет single-flight.

4. Shop витрина

Cache-Aside: GET shop:items:{hash} → miss → PG.

Защита: нет single-flight.

5. Auth сессии

EXISTS jti:{id} — если ключ истёк, всегда miss. Но это не кэш, а source of truth. Stampede нет.

6. Payment подписка (Cache TTL based on Expiry)

TTL из expires_at. Когда истёк — miss → PG.

Защита: нет single-flight.

Как решать (общие принципы)
Принцип	Где в EH
Single-flight	Нет (техдолг)
SETNX lock	Нет (техдолг)
TTL jitter	Нет (техдолг)
Probabilistic early expiration	Нет (техдолг)
Circuit Breaker	Gateway (но не на кэш)
Что сказать на собесе
«Cache stampede — когда 1000 запросов на один ключ, все идут в БД. В EH Cache-Aside везде, но single-flight нет — техдолг. Решение: SETNX lock на ключ, первый идёт в БД, остальные ждут. Или TTL jitter — размазать истечение.»

3.4. Сводная таблица
Проблема	Где	Защита есть	Защита нет (техдолг)
Thundering herd	Auth login	Rate limit 5/s IP	—
Thundering herd	Gateway WS	Rate limit 100/min IP	—
Thundering herd	Notification Fan-Out	—	Rate limit (техдолг)
Thundering herd	Leaderboard record_beaten	—	Rate limit (техдолг)
Retry storm	Gateway	Full jitter + CB	—
Retry storm	Fulfillment	Idempotency Guard	Persistent Guard (техдолг)
Retry storm	Notification	Fixed Delay	Exponential (техдолг)
Retry storm	Payment	Idempotent Completion	—
Retry storm	Client 429	FE backoff (не факт)	—
Cache stampede	Billing	—	Single-flight (техдолг)
Cache stampede	Inventory	—	Single-flight (техдолг)
Cache stampede	Shop	—	Single-flight (техдолг)
Cache stampede	Payment	—	Single-flight (техдолг)
Poison message	Notification Kafka	—	Ack + DLQ (техдолг)
Split-brain	NATS	3 ноды, кворум	—
Deadlock	Billing	PG rollback + retry	—
Saga stuck	Shop started	resumeSagas берёт spent/compensating	started старше 60s (техдолг)
ЧАСТЬ 4. Глубокие кейсы по ключевым сервисам
4.1. Billing: deadlock на Conditional UPDATE
Что случилось: два параллельных SpendCurrency на разные счета, но transactions вставка даёт deadlock на индексе.

text
TX1: UPDATE balance WHERE user=A → lock A
TX2: UPDATE balance WHERE user=B → lock B
TX1: INSERT transactions (user=A) → lock index
TX2: INSERT transactions (user=B) → lock index
TX1: ждёт lock index (TX2 держит)
TX2: ждёт lock index (TX1 держит)
→ DEADLOCK → PG откатывает одну TX
Что происходит:

PG откатывает TX2 (deadlock detected).

TX2 возвращает ошибку клиенту.

pg_stat_database_deadlocks растёт.

Клиент должен retry.

Как чинить:

Retry с idempotency key (reference_id). Повтор не спишет дважды.

Порядок блокировок: всегда user_id ASC.

FOR UPDATE на user_currencies вместо UPDATE — но это медленнее.

Что сказать: «Billing может дать deadlock на Conditional UPDATE + INSERT transactions. PG откатывает одну TX, клиент retry с reference_id. deadlocks = 0 — SLO, алерт.»

4.2. Auth: Redis упал во время ValidateToken
text
Gateway     Auth                 Redis
  │          │                    │
  │ ValidateToken                 │
  ├─────────►│                    │
  │          │ EXISTS jti:abc     │
  │          ├───────────────────►│
  │          │                    │ connection refused
  │          │◄───────────────────┤ error
  │          │ return Unavailable │
  │◄─────────┤                    │
  │ 503      │                    │
Что происходит:

Auth получает ошибку Redis → Unavailable.

Gateway → 503 session store unavailable.

Все закрытые ручки мертвы.

Что НЕ делать: не «игнорировать ошибку Redis и пропустить токен». Это дыра: отозванный токен пройдёт.

Как чинить:

Redis рестарт. Сессии потеряны → перелогин.

Auth остаётся живым (Redis — отдельный контейнер).

Что сказать: «Auth: Redis упал → Unavailable → Gateway 503. Не игнорируем ошибку — иначе отозванный токен пройдёт. Лучше 503, чем дыра.»

4.3. Gateway: Circuit Breaker открылся на Billing
text
Gateway     Billing
  │          │
  │ gRPC SpendCurrency
  ├─────────►│ timeout (5s)
  │◄─────────┤ error
  │ CB: 1 fail
  │          │
  │ gRPC (retry)
  ├─────────►│ timeout
  │◄─────────┤ error
  │ CB: 2 fail
  │          │
  │ ... 5 fail → CB Open
  │          │
  │ gRPC → CB Open → 503 fail-fast
  │ (без вызова Billing)
Что происходит:

После 5 ошибок — CB Open.

Запросы не идут в Billing, сразу 503.

Half-open через 10s: пропускает 3 запроса.

Если успех — Closed. Если ошибка — Open снова.

Что НЕ работает: Shop не может купить. Game не может бустить. Merсh — нет.

Что работает: публичные ручки, Auth, другие gRPC-сервисы (у каждого свой CB).

Как чинить:

Billing рестарт. CB half-open → Closed.

Если Billing долго — алерт.

Что сказать: «Gateway CB: после 5 ошибок Billing — Open, fail-fast 503. Half-open через 10s, 3 запроса. Если Billing ожил — Closed. Это защита от каскадного отказа.»

4.4. Shop: resumeSagas берёт spent, но purchases есть
text
resumeSagas      PG Shop              Billing
  │               │                    │
  │ SELECT * FROM purchase_sagas       │
  │  WHERE status='spent'              │
  │    AND created_at < now()-30s      │
  │  FOR UPDATE SKIP LOCKED            │
  ├──────────────►│                    │
  │◄──────────────┤ rows (purchase_id=X)
  │               │                    │
  │ SELECT * FROM purchases WHERE id=X │
  ├──────────────►│                    │
  │◄──────────────┤ найдено            │
  │               │                    │
  │ UPDATE saga → stocked              │
  ├──────────────►│                    │
  │◄──────────────┤ ok                 │
  │               │                    │
  │ (ничего не возвращаем)             │
Что происходит:

purchases есть → TX прошла → stocked.

Деньги не возвращаются (товар выдан).

Outbox purchase.paid уже в outbox → воркер опубликует.

Что если purchases нет:

TX не прошла → компенсация.

AddCurrency(refund:...) → compensated.

cycle += 1.

Что сказать: «resumeSagas различает "TX прошла" и "не прошла" по наличию purchases.id. Есть — stocked, нет — компенсация. Идемпотентно.»

4.5. Payment: двойной webhook
text
Boosty      Payment                    PG Payment
  │          │                          │
  │ webhook #1 (payment_id=X)          │
  ├─────────►│                          │
  │          │ SELECT FOR UPDATE payment │
  │          ├─────────────────────────►│
  │          │◄─────────────────────────┤ status=pending
  │          │ UPDATE status=completed  │
  │          │ INSERT subscription      │
  │          │ INSERT outbox(payment.completed)
  │          │ INSERT revenue_splits    │
  │          │ COMMIT                   │
  │          ├─────────────────────────►│
  │◄─────────┤ 200                      │
  │          │                          │
  │ webhook #2 (payment_id=X)          │
  ├─────────►│                          │
  │          │ SELECT FOR UPDATE payment │
  │          ├─────────────────────────►│
  │          │◄─────────────────────────┤ status=completed
  │          │ return 200 (не открываем)│
  │◄─────────┤ 200                      │
Что происходит:

Webhook #1 → платёж completed, подписка создана, outbox, revenue_splits.

Webhook #2 → видит completed → 200, не открывает второй раз.

Что если Payment упал между #1 и #2:

#1 прошёл → completed.

Payment упал.

Boosty ретраит #2 → Payment рестарт → видит completed → 200.

Что сказать: «Payment: Idempotent Completion. Повтор webhook видит completed, возвращает 200, не открывает второй раз. Без этого подписка продлилась бы на 60 дней.»

4.6. Fulfillment: упал между paid и fulfilled
text
Shop     NATS           Fulfillment         NATS
  │       │                │                 │
  │ purchase.paid          │                 │
  ├──────►│                │                 │
  │       ├───────────────►│                 │
  │       │                │ consume         │
  │       │                │ Idempotency Guard: ok
  │       │                │ time.After(10s) │
  │       │                │ ← УПАЛ          │
  │       │                │                 │
  │       │ purchase.paid ещё в NATS?        │
  │       │ (durable consumer — ack не было) │
  │       │                │                 │
Что происходит:

purchase.paid consumed, но purchase.fulfilled не published.

Нет Outbox (техдолг).

Idempotency Guard в памяти — при рестарте теряется.

При рестарте Fulfillment: purchase.paid не вернётся (ack был? Или нет?).

Ключевой вопрос: был ли ack?

Если ack был: purchase.paid consumed, но purchase.fulfilled не published. Заказ завис. fulfilled_at null, отмена возможна (7 дней). Shop не знает, что сборка не прошла. Ручное вмешательство.

Если ack не было: purchase.paid вернётся. Fulfillment повторит. Idempotency Guard (map) может сработать (если не рестарт) или нет (если рестарт).

Техдолг: «Нет Outbox — если сервис упадёт между purchase.paid и purchase.fulfilled, событие потеряется. Внедрить Transactional Outbox».

Что сказать: «Fulfillment: нет Outbox. Упал между paid и fulfilled — заказ завис. purchase.paid в NATS durable, но ack мог быть. Техдолг: Outbox.»

4.7. Notification: poison message из Kafka
text
Kafka       Notification
  │           │
  │ message (битый JSON)
  ├──────────►│
  │           │ unmarshal → error
  │           │ Notifier.Handle → error
  │           │ (нет Ack — техдолг)
  │           │ → retry
  │◄──────────┤ (не ack)
  │           │
  │ message (тот же битый)
  ├──────────►│
  │           │ unmarshal → error
  │           │ → retry
  │           │
  │ ... бесконечно
Что происходит:

Битое сообщение → unmarshal error → Notifier.Handle error.

Нет Ack → Kafka ретраит.

Бесконечный retry. notification_poison_messages_total (если есть) растёт.

Техдолг: «Notifier (Telegram) в Handle для Kafka падает без Ack при unmarshal error. Решение: Poison message → Ack».

Что сказать: «Notification: Kafka poison message — бесконечный retry. Техдолг: Ack + DLQ. Сейчас битое сообщение блокирует consumer.»

4.8. NATS Hub: первый старт, stream не создан
text
NATS Hub      NATS-кластер
  │             │
  │ connect     │
  ├────────────►│
  │◄────────────┤ ok
  │             │
  │ AddStream EVENTS
  ├────────────►│
  │◄────────────┤ ok (создан)
  │             │
  │ subscribe event.>
  ├────────────►│
  │◄────────────┤ ok
  │             │
  │ /health → nc.IsConnected() = true
Что если AddStream упал:

Stream не создан.

Все публикации → no stream matches subject.

Critical path. Сервис должен упасть (не может работать без stream).

Что если stream уже есть:

AddStream → AlreadyExists → UpdateStream.

Не падает. Конфиг обновлён.

Что сказать: «NATS Hub: первый старт — AddStream, если упал — stream нет, все publish падают. При рестарте — Idempotent Creation: AddStream → AlreadyExists → UpdateStream.»

Итого
Часть	Что разобрано
1	Shop Saga: упал после Spend, до TX (sequence), двойной клик
2	Шпаргалка: 17 сервисов, «если упал X → Y», одна фраза на сервис, страшилки
3	Thundering herd (Auth, WS, Fan-Out), Retry storm (Gateway, Fulfillment, Notification), Cache stampede (Billing, Inventory, Shop, Payment)
4	8 глубоких кейсов: Billing deadlock, Auth Redis, Gateway CB, Shop resumeSagas, Payment double webhook, Fulfillment между paid/fulfilled, Notification poison, NATS Hub первый старт
Что сказать на собесе, если спросят «краевые случаи»:

«Три класса: thundering herd (rate limit), retry storm (full jitter + CB + idempotency), cache stampede (single-flight — техдолг). Плюс конкретные: Shop Saga stuck на started (дыра в resumeSagas), Fulfillment без Outbox, Notification poison message, Billing deadlock, Auth Redis → 503. Каждый — с решением или честным "техдолг".»

Что дальше? Могу:

Сделать sequence-диаграмму для каждого из 8 кейсов (Mermaid)

Разобрать конкретный сервис ещё глубже (например, Shop Saga целиком с 6 состояниями)

Сделать шпаргалку на 1 страницу в markdown-файле (для печати)