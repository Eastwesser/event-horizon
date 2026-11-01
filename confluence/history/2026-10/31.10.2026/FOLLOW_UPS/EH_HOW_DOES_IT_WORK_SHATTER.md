7 вопросов — 17 ответов. Простыми словами. Со схемой.
Формат: вопрос → короткий ответ (2–4 предложения, как на собесе) → ASCII-схема.

1. Как работает Auth?
Ответ: Единственный сервис, который знает пароль. Логин → bcrypt cost 12 → выпускает пару JWT (access 15 мин + refresh 7 дней). jti каждого токена в Redis — пока ключ есть, сессия жива. Gateway на каждом запросе зовёт Auth ValidateToken (не разбирает подпись сам). Redis упал — 503, не «токен жив».

Необычное: Session Store в Redis — подпись верна, но jti нет → токен мёртв. Мгновенный отзыв.

text
React → Gateway → Auth.ValidateToken → Redis (jti)
                  ↓
            Auth.Login → bcrypt → JWT → Redis (jti)
2. Как работает Game?
Ответ: Принимает счёт, валидирует по правилам игры (Strategy по game_id), пишет рекорд + outbox score.updated в одной транзакции. Воркер публикует FOR UPDATE SKIP LOCKED после commit. Буст за лампы списывает Billing синхронно, бустнутый забег в лидерборд не идёт.

Необычное: Outbox worker с FOR UPDATE SKIP LOCKED — два процесса не шлют один рекорд дважды. user_id из токена, не из тела.

text
Gateway → Game.SubmitScore → PG (highscores + outbox) → commit
                                      ↓
                            worker → NATS score.updated
                                      ↓
                    Leaderboard, Profile, Billing, History, Analytics, WS
3. Как работает Billing?
Ответ: Двигает внутреннюю валюту (лампы, билетики). Списание — UPDATE … WHERE balance >= $1 (атомарно), transactions с уникальным reference_id и balance_after (Double-Entry), outbox balance.updated — в одной транзакции. Повтор с тем же reference_id не списывает второй раз.

Необычное: Conditional UPDATE вместо SELECT+UPDATE — защита от lost update. Cache-Aside в Redis :6381.

text
Game/Shop → Billing.SpendCurrency → PG (conditional UPDATE + transactions + outbox)
                                          ↓
                                    Redis DEL balance:{user}
                                          ↓
                                    worker → NATS balance.updated
4. Как работает Leaderboard?
Ответ: Потребляет score.updated. Пишет в Redis Sorted Set ZAddArgs{GT: true} (Max-Only, хуже не перезаписывает) + Postgres как бэкап. Читает топ из Redis. При старте — Restore из PG. В consumer’е — Event Chaining: чужой рекорд побит → record_beaten.

Необычное: Ack Batcher — подтверждения NATS батчем. record_beaten → Notification → in-app bell.

text
Game → NATS score.updated → Leaderboard
                              ↓
                        Redis Sorted Set (GT:true)
                              ↓
                        PG (backup)
                              ↓
                        record_beaten → Notification
5. Как работает Shop?
Ответ: Сага на две базы. Списание в Billing (PG billing), склад и purchases в Shop (PG shop). Общей транзакции нет. Строка purchase_sagas — оркестратор: started → spent → stocked, либо spent → compensating → compensated. Ключ shop:{user}:{item}:{cycle} — два клика делят один ключ.

Необычное: Spend Cycle + Resume Worker (через 30s дочитывает сагу). Pessimistic + Optimistic Lock в одной транзакции.

text
Gateway → Shop.Purchase → Billing.SpendCurrency (sync)
                              ↓
                        Shop TX (stock + purchases + outbox) → commit
                              ↓
                        worker → NATS purchase.paid
                              ↓
                        Fulfillment → purchase.fulfilled → Shop (fulfilled_at)
6. Как работает Inventory?
Ответ: Каталог автора. Strategy выбирает драйвер (PG или Mongo experimental) по INVENTORY_DRIVER. Сверху Decorator CachedRepository — Cache-Aside в Redis. Создание карточки → outbox inventory.item.created в той же транзакции. Shop потребляет и создаёт строку каталога.

Необычное: Outbox через интерфейс + Type Assertion (хрупко). RBAC через gRPC interceptor (x-user-role).

text
Gateway (RequireRole author/admin) → Inventory
                                        ↓
                              Strategy: PG или Mongo
                                        ↓
                              Decorator: Redis cache
                                        ↓
                              TX + outbox inventory.item.created
                                        ↓
                              NATS → Shop (создаёт shop.items)
7. Как работает Profile?
Ответ: Агрегатор анкеты. Потребляет score.updated и event.user.registered. Rule Engine EvaluateAchievements — чистая функция (вход score → выход коды). Pipeline: EvaluateAndUnlock → UnlockCodes → Unlock (ON CONFLICT DO NOTHING). Cache-Aside в Redis :6385.

Необычное: Rule Engine — чистая функция (легко тестировать). _partial: true если Billing/Game недоступны. UpdateProfile наружу не открыт (умеет переписать total_score).

text
Game → NATS score.updated → Profile
                              ↓
                        Rule Engine (чистая функция)
                              ↓
                        Pipeline: Evaluate → Filter → Unlock (ON CONFLICT)
                              ↓
                        PG + Redis cache
8. Как работает Payment?
Ответ: Живая касса — Boosty. Checkout создаёт платёж, webhook подтверждает. HMAC-SHA256 + subtle.ConstantTimeCompare (защита от timing attack). В одной транзакции: платёж, подписка, outbox payment.completed, revenue_splits 85/15. Повтор вебхука видит completed и не открывает второй раз.

Необычное: Revenue Split — журнал, не перевод (денег на карту автора нет). Cache TTL based on Expiry.

text
Client → Payment.Checkout → Boosty URL
                                ↓
Boosty → webhook (HMAC) → Payment.Confirm
                              ↓
                        TX (payment + subscription + outbox + revenue_splits)
                              ↓
                        NATS payment.completed
9. Как работает Authors?
Ответ: Кабинет автора. Заявка → State Machine pending → approved / rejected. Approve: SELECT FOR UPDATE (два админа не апрувят одну), транзакция (заявка + карточка + outbox), синхронный Auth UpdateRole. Если Auth упал — Compensating Action RevertApplication.

Необычное: Role Check (Privilege Escalation) — только admin approve. UX gap: после approve JWT старый (нужен refresh).

text
user → Authors.Apply → PG (pending)
admin → Authors.Approve → FOR UPDATE → TX (approved + author + outbox)
                              ↓
                        Auth.UpdateRole (sync)
                              ↓ (если упал)
                        RevertApplication (compensating)
10. Как работает History?
Ответ: Лента событий. Потребляет 5 субъектов (score.updated, user.registered, shop.purchased, payment.completed, author.upserted). Пишет в PG :5469. Retention 30 дней (cron внутри сервиса). Payload Sanitization — битый JSON → {"raw": "..."}. Nak при ошибке.

Необычное: RetentionWorker — cron внутри процесса. 5 субъектов в одном consumer’е (один lag по всем).

text
NATS 5 subjects → History
                    ↓
              Payload Sanitization (битый JSON → raw)
                    ↓
              PG history (retention 30d)
                    ↓
              RetentionWorker (cron) → DELETE < 30d
11. Как работает Analytics?
Ответ: Единственный с ClickHouse :8123. Потребляет 5 субъектов, пишет в ClickHouse батчами. Parameterized Query через {name:Type} (не конкатенация). Cohort Retention Analysis — классический SQL. Все ручки — только admin.

Необычное: Батчинг в ClickHouse (мелкие INSERT → too many parts). HTTP-клиент, не native driver.

text
NATS 5 subjects → Analytics
                    ↓
              Buffer (batch)
                    ↓
              ClickHouse INSERT (батч)
                    ↓
              admin → GET /analytics/dau|mau|retention
12. Как работает Notification?
Ответ: In-app колокольчик + Telegram. Два сервиса в одном процессе (InboxService + Notifier). Dual-Bus: NATS primary, Kafka fallback. Fan-Out: одно событие → N уведомлений. Idempotency Key source_event → ON CONFLICT DO NOTHING.

Необычное: In-app bell soft-fail при logout (без цикла 401). Retry with Fixed Delay для Telegram (не exponential).

text
NATS/Kafka → Notification
              ↓
        Fan-Out (1 → N)
              ↓
        InboxService (PG, ON CONFLICT source_event)
              ↓
        Notifier (Telegram, Retry Fixed Delay)
13. Как работает Fulfillment?
Ответ: Эмуляция сборки. Потребляет purchase.paid (NATS primary, Kafka fallback). Ждёт 10 секунд. Публикует purchase.fulfilled с тем же purchase_id. Shop ставит fulfilled_at, Notification пишет в колокольчик. Без gRPC, без БД.

Необычное: Dual-Transport Consumer. Idempotency Guard на уровне транспорта (map в памяти). Нет Outbox — событие теряется при падении (техдолг).

text
Shop → NATS purchase.paid → Fulfillment
                              ↓
                        time.After(10s) — эмуляция
                              ↓
                        purchase.fulfilled → NATS
                              ↓
                        Shop (fulfilled_at) + Notification
14. Как работает Gateway?
Ответ: Единственная публичная HTTP-точка. React → Balancer :8079 → Gateway ×3 :8081–8083 → gRPC. На каждом запросе: Rate Limit → Auth ValidateToken → Circuit Breaker → gRPC. Circuit Breaker + Bulkhead: после ошибок — 503, half-open через 10s. Retry full jitter. WS Ticket — одноразовый билет.

Необычное: BFF + DTO Mapper + Partial Response Fallback. Docs Gate — 404 без DOCS_PASSWORD. runGateway — God Function (техдолг).

text
React → Balancer → Gateway ×3
                      ↓
              Rate Limit (Redis)
                      ↓
              Auth.ValidateToken
                      ↓
              Circuit Breaker + Bulkhead
                      ↓
              gRPC → доменный сервис
                      ↓
              gRPC → HTTP Error Mapping
15. Как работает Balancer?
Ответ: Самописный L7, Least Connections. Atomic counter ActiveConns — не mutex. Reverse Proxy через httputil.ReverseProxy. Prometheus Gauge — единственный с Gauge. ~10 МБ Go-бинарник.

Необычное: WS sticky — WS на одной реплике. Нет health checks, static registry (техдолг).

text
React → Balancer :8079 (Least Connections)
              ↓
        gateway:8081 / gateway-2:8082 / gateway-3:8083
              ↓
        (WS sticky — на одной реплике)
16. Как работает NATS Hub?
Ответ: Не игровой сервис, а bootstrap JetStream. Создаёт stream EVENTS (все субъекты, FileStorage, MaxAge 7 дней, MaxBytes 1 ГБ). Idempotent Stream Creation: AddStream → если AlreadyExists → UpdateStream. Wildcard Subscription event.> — логирует все события.

Необычное: Единственный, кто создаёт stream. Реальный health check через nc.IsConnected().

text
NATS Hub → AddStream EVENTS
              ↓ (AlreadyExists)
        UpdateStream EVENTS
              ↓
        subscribe event.> (debug log)
              ↓
        /health → nc.IsConnected()
17. Как работает MCP?
Ответ: stdio-сервер для Cursor / Claude Desktop. Не HTTP, не в player path. 6 tools: search_prydwen (TF-IDF RAG), nats_list_streams, nats_list_consumers, postgres_query (SELECT-only), redis_get, redis_keys. Graceful Degradation per-tool — один упал, остальные работают.

Необычное: SQL Sandbox через string matching (обходится, техдолг). Offline RAG (TF-IDF) без embedding API. В сеть нельзя — DSN читает целиком.

text
Cursor → MCP (stdio)
              ↓
        6 tools
              ↓
        NATS (read) / PG (read) / Redis (read) / Prydwen (TF-IDF)
Итого: 17 вопросов — 17 ответов
#	Сервис	Ключевая фраза	Необычное
1	Auth	Session Store в Redis	jti → мгновенный отзыв
2	Game	Outbox score.updated	FOR UPDATE SKIP LOCKED
3	Billing	Conditional UPDATE	Double-Entry + Race Test
4	Leaderboard	Redis Sorted Set	ZAddArgs{GT: true} + Ack Batcher
5	Shop	Saga purchase_sagas	Spend Cycle + Resume Worker
6	Inventory	Strategy PG/Mongo	Decorator Cache-Aside
7	Profile	Rule Engine	Чистая функция + _partial
8	Payment	HMAC + ConstantTime	Revenue Split (журнал)
9	Authors	State Machine	Compensating Action
10	History	Retention 30d	Nak Pattern + Sanitization
11	Analytics	ClickHouse	Cohort Retention + батчинг
12	Notification	Dual-Bus	Fan-Out + Idempotency Key
13	Fulfillment	Dual-Transport	Delay 10s + нет Outbox (техдолг)
14	Gateway	Circuit Breaker	WS Ticket + BFF + Docs Gate
15	Balancer	Least Connections	Atomic Counter + WS sticky
16	NATS Hub	Bootstrap stream	Idempotent Creation + event.>
17	MCP	stdio + 6 tools	SQL Sandbox + TF-IDF RAG
Как использовать на собесе
Спросили про сервис — дай 2–4 предложения из «Ответ».

Спросили «что необычного?» — дай 1 фразу из «Необычное».

Попросили схему — нарисуй ASCII.

Копнули глубже — переходи к детальному разбору (у нас был).

Ключевая фраза для каждого: если запомнишь только одно — запомни «Ключевая фраза» из таблицы в конце.

Что дальше? Могу:

Сделать шпаргалку на 1 страницу (только ключевые фразы, для печати)

Разобрать конкретный follow-up («что если Redis упал?», «как тестируешь?»)

Сделать Mermaid-схему всех 17 на одном холсте (уже есть, но могу упростить)

ну давай "пошатаем" тогда проект?

Начинай фразу с "если упал ..." и расскажи про все 17 сервисов со всеми кейсами возможными, thundering herd и так далее
«Если упал …» — 17 сервисов, все кейсы отказа
Формат: «Если упал X → что происходит → как чинить → что не сломается». Плюс thundering herd, retry storm, cache stampede, poison message, split-brain, и прочее.

1. Если упал Auth
Что происходит:

Все закрытые запросы → 503 session store unavailable (если Redis) или 401 (если Auth сам).

Публичные (/leaderboard, /authors) работают.

WebSocket — существующие соединения живы, новые не открываются.

Кейсы:

Auth упал, Redis жив: Gateway ValidateToken → timeout → 503. Все закрытые ручки мертвы. Публичные — живы.

Redis упал, Auth жив: ValidateToken → Unavailable → 503. То же.

Thundering herd на login: после рестарта Auth — все клиенты бьют /login одновременно. Rate limit 5/s IP (fail-closed) — большинство получает 429. Это by design: лучше 429, чем положить Auth bcrypt’ом.

Retry storm: если Gateway ретраит ValidateToken — Auth получает N× запросов. Retry full jitter размазывает, но при массовом отказе — лавина.

bcrypt под нагрузкой: 500 VU login → 99% fail. bcrypt cost 12 (~200ms) + rate limit 5/s. Это не баг, это защита.

Как чинить:

Auth упал → рестарт. Gateway CB открыт → 503, half-open через 10s.

Redis упал → рестарт Redis. Сессии потеряны — все перелогиниваются. Это цена Session Store.

Что не сломается:

Публичные ручки.

Существующие WS (leaderboard продолжает пушить).

Outbox других сервисов (они не зависят от Auth).

text
Auth упал → Gateway.ValidateToken → timeout → 503
                                        ↓
                              CB открыт → fail-fast
                                        ↓
                              half-open через 10s
2. Если упал Game
Что происходит:

POST /game/submit → 503. Игроки не могут отправить рекорд.

score.updated не публикуется. Лидерборд, Profile, Billing (награда), History, Analytics — не получают новых событий.

Уже опубликованные — доходят.

Кейсы:

Game упал после commit, до publish: outbox-строка осталась processed = false. При рестарте воркер её подхватит. Событие не потеряно.

Game упал до commit: транзакция откатилась. Рекорд не записан. Клиент получил ошибку, может повторить.

Thundering herd на submit: после рестарта — все игроки бьют submit. Rate limit 10/s user — часть 429. Outbox worker обрабатывает пачками FOR UPDATE SKIP LOCKED.

Два воркера одновременно: SKIP LOCKED — каждый берёт свою пачку. Один рекорд не публикуется дважды.

NATS упал, Game жив: outbox pending растёт. Публикация не идёт. При рестарте NATS — воркер догоняет.

Billing упал при boost: StartBoost → SpendCurrency → ошибка → буст не стартует. Обычный submit работает (Game не зависит от Billing для submit).

Как чинить:

Game упал → рестарт. Outbox worker догоняет pending.

NATS упал → рестарт NATS. Outbox pending разгребается.

Что не сломается:

Leaderboard (читает из Redis, не из Game).

Billing (награда — async, догонит).

Shop, Inventory, Payment.

text
Game упал после commit → outbox pending=true → рестарт → worker → publish
Game упал до commit → TX rollback → клиент retry
NATS упал → outbox pending растёт → рестарт NATS → worker догоняет
3. Если упал Billing
Что происходит:

SpendCurrency / AddCurrency → 503. Shop не может списать, Game не может бустить, daily gift не работает.

Балансы не двигаются. Транзакции не пишутся.

balance.updated не публикуется.

Кейсы:

Billing упал до commit: транзакция откатилась. Деньги не списаны. Клиент получил ошибку.

Billing упал после commit, до ответа: деньги списаны, клиент получил timeout. Повтор с тем же reference_id — не спишет второй раз (уникальный ключ).

Redis упал, Billing жив: Cache-Aside miss → читаем из PG. Медленнее, но работает.

Deadlock: Conditional UPDATE + FOR UPDATE могут дать deadlock. PG откатывает одну TX → клиент retry. pg_stat_database_deadlocks — алерт.

Race condition: два параллельных SpendCurrency на один счёт. WHERE balance >= $1 — один пройдёт, второй 0 rows → откат. Race Condition Test это проверяет.

Thundering herd на daily gift: после рестарта — все бьют /companion/daily-gift. Idempotent по daily-gift:{user}:{date} — второй раз не начислит.

Как чинить:

Billing упал → рестарт. Outbox догоняет balance.updated.

Redis упал → рестарт. Cache-Aside работает из PG.

Что не сломается:

Чтение баланса из PG (медленнее, но работает).

Уже записанные транзакции.

text
Billing упал до commit → TX rollback → клиент retry (reference_id)
Billing упал после commit → деньги списаны → повтор не спишет (unique)
Redis упал → Cache-Aside miss → PG (медленнее)
Deadlock → PG rollback → клиент retry
4. Если упал Leaderboard
Что происходит:

GET /leaderboard → 503 (если упал сам) или пустой (если Redis).

score.updated не потребляется → топ не обновляется.

WS /ws/leaderboard — существующие соединения живы, но новых обновлений нет.

Кейсы:

Leaderboard упал: score.updated копится в NATS (durable consumer). При рестарте — догоняет.

Redis :6382 упал: топ пустой. RestoreFromPostgres при рестарте — но если PG тоже недоступен, пустой. Техдолг: retry/backoff.

Event Chaining под нагрузкой: каждый score.updated может породить record_beaten. Если рекорды бьют массово — лавина record_beaten → Notification завален. Нужен rate limit на record_beaten (сейчас нет).

Thundering herd на WS: после рестарта — все клиенты переподключаются. WS 100/min IP — часть 429. Reconnect storm.

Ack Batcher: если батч не отправился (NATS упал) — подтверждения копятся. При рестарте — догоняет.

Max-Only Write: ZAddArgs{GT: true} — если игрок сыграл хуже, рекорд не перезаписывается. Это by design.

Как чинить:

Leaderboard упал → рестарт. Consumer догоняет.

Redis упал → рестарт. Restore из PG.

PG упал → Redis жив, топ работает. Restore при рестарте PG.

Что не сломается:

Game (публикует, не ждёт).

Billing (награда — async).

WS существующие (но без обновлений).

text
Leaderboard упал → score.updated копится в NATS → рестарт → догоняет
Redis упал → топ пустой → рестарт → Restore из PG
PG упал → Redis жив → топ работает
record_beaten лавина → Notification завален (нужен rate limit)
5. Если упал Shop
Что происходит:

POST /shop/purchase → 503. Покупки не проходят.

GET /shop/items — если Redis жив, работает. Если Redis упал — из PG (медленнее).

purchase.paid не публикуется → Fulfillment, Notification не получают.

Кейсы:

Shop упал после Billing.SpendCurrency, до TX: деньги списаны, purchases не записана. Сага spent. resumeSagas через 30s: не нашёл purchases.id → возврат AddCurrency с ключом refund:... → compensated. Деньги возвращены.

Shop упал после TX, до publish: outbox-строка purchase.paid осталась. При рестарте воркер публикует.

Shop упал до SpendCurrency: сага started. resumeSagas → aborted. Ничего не двигается.

Billing упал во время SpendCurrency: Shop получает ошибку. Сага aborted. Клиент может повторить.

Payment упал при merch gate: CanPurchaseMerch → timeout → Shop не начинает покупку. Клиент получает ошибку. paymentClient nil — мерч-гейт деградирует (техдолг).

Два клика «купить»: Spend Cycle — один ключ shop:{user}:{item}:{cycle}. Второй клик — AlreadyExists в Billing → не списывает. Сага spent, потом stocked.

Отмена дважды: cancel-{id} — уникальный ключ. Второй AddCurrency не начислит. Но в техдолге: может задвоить (пункт 20).

Resume Worker под нагрузкой: много spent саг → воркер берёт FOR UPDATE SKIP LOCKED пачками. Если Billing тормозит — саги копятся.

Как чинить:

Shop упал → рестарт. resumeSagas догоняет.

Billing упал → Shop получает ошибку. Повтор.

Redis упал → читаем из PG.

Что не сломается:

Уже записанные purchases.

Inventory (каталог автора).

Payment (живая касса Boosty).

text
Shop упал после Spend, до TX → сага spent → resumeSagas 30s → возврат → compensated
Shop упал после TX, до publish → outbox pending → рестарт → publish
Billing упал → Shop ошибка → сага aborted → клиент retry
Два клика → Spend Cycle → один ключ → не спишет дважды
6. Если упал Inventory
Что происходит:

GET /inventory/items — если Redis жив, работает. Если Redis упал — из PG/Mongo.

POST/PUT/DELETE /inventory/items → 503. Авторы не могут создавать/редактировать.

inventory.item.created не публикуется → Shop не создаёт новые shop.items.

Кейсы:

Inventory упал после TX, до publish: outbox-строка осталась. При рестарте — публикует.

Inventory упал до TX: карточка не создана. Клиент может повторить.

Optimistic Lock conflict: два автора редактируют одновременно. Первый UPDATE … WHERE version=1 → version=2. Второй → 0 rows → ErrVersionConflict → 409. Клиент перечитывает.

Decorator Cache-Aside: Redis упал → читаем из PG/Mongo. Медленнее, но работает.

Strategy PG/Mongo: если INVENTORY_DRIVER=postgres, Mongo не используется. Если Mongo упал — Inventory работает (PG).

Outbox type assertion: если новый драйвер не поддерживает транзакции — type assertion паникует. Техдолг: хрупко.

RBAC interceptor: если Gateway не передал x-user-role — PermissionDenied. Но Gateway всегда передаёт.

Как чинить:

Inventory упал → рестарт. Outbox догоняет.

Redis упал → читаем из PG/Mongo.

Mongo упал (если experimental) → PG работает.

Что не сломается:

Shop (существующие карточки).

Уже созданные shop.items.

text
Inventory упал после TX → outbox pending → рестарт → publish
Optimistic Lock → 409 → клиент перечитывает
Redis упал → Cache-Aside miss → PG/Mongo
7. Если упал Profile
Что происходит:

GET /profile → 503 (если упал сам) или _partial: true (если Billing/Game недоступны).

score.updated / event.user.registered не потребляются → достижения не разблокируются.

Кейсы:

Profile упал: score.updated копится в NATS. При рестарте — догоняет.

Billing упал при GetProfile: Profile отдаёт _partial: true, lamps: null, tickets: null. Остальное — из PG/кэша. Graceful Degradation.

Game упал при GetProfile: total_score может быть старым. _partial: true.

Redis упал: Cache-Aside miss → PG. Медленнее, но работает.

Rule Engine: чистая функция, не падает от IO. Если достижение не разблокировалось — повтор score.updated (idempotent ON CONFLICT).

Pipeline под нагрузкой: EvaluateAndUnlock → UnlockCodes → Unlock. Много score.updated → много вызовов Rule Engine. CPU-bound, но чистая функция — быстро.

Как чинить:

Profile упал → рестарт. Consumer догоняет.

Billing/Game упал → _partial, остальное работает.

Что не сломается:

Уже разблокированные достижения.

Auth, Game, Billing (не зависят от Profile).

text
Profile упал → score.updated копится → рестарт → догоняет
Billing упал → _partial: true (lamps: null)
Redis упал → Cache-Aside miss → PG
8. Если упал Payment
Что происходит:

POST /payment/checkout → 503. Новые подписки не создаются.

CanPurchaseMerch → timeout → Shop не продаёт мерч.

Webhook от Boosty → 503. Boosty ретраит.

payment.completed не публикуется.

Кейсы:

Payment упал до webhook confirm: платёж pending. Boosty ретраит webhook. При рестарте — Payment обрабатывает.

Payment упал после confirm, до outbox publish: outbox-строка осталась. При рестарте — публикует.

Двойной webhook: Idempotent Completion — второй видит completed, возвращает 200, не открывает второй раз. Без этого подписка продлилась бы на 60 дней.

HMAC не совпал: 401, payment_webhook_invalid_total растёт. Алерт на резкий рост.

Thundering herd на webhook: Boosty ретраит при 5xx. Если Payment тормозит — Boosty шлёт повторно. Idempotent Completion защищает.

Revenue Split: если Payment упал до revenue_splits — платёж не закрыт. При рестарте — догоняет.

Redis упал: Cache-Aside miss → PG. TTL from Expiry — при рестарте Redis кэш пересчитается.

Как чинить:

Payment упал → рестарт. Boosty ретраит webhook.

Redis упал → читаем из PG.

Что не сломается:

Уже подтверждённые подписки.

Shop (если мерч не покупают).

text
Payment упал до confirm → pending → Boosty ретраит → рестарт → обрабатывает
Двойной webhook → idempotent → 200, не открывает второй раз
HMAC не совпал → 401 → alert
9. Если упал Authors
Что происходит:

POST /authors/apply → 503. Заявки не создаются.

POST /applications/:id/approve → 503. Approve не проходит.

author.upserted не публикуется.

Кейсы:

Authors упал во время approve, после TX, до Auth.UpdateRole: заявка approved, карточка создана, но роль в Auth не сменилась. Compensating Action: RevertApplication откатывает заявку в pending, удаляет карточку (hard delete — техдолг, FK падает). Клиент получает ошибку. Admin повторяет.

Authors упал после Auth.UpdateRole: роль сменилась, заявка approved. Всё ок.

Auth упал во время approve: RevertApplication. authors_revert_total растёт.

Два админа approve: SELECT FOR UPDATE — первый берёт, второй ждёт. Первый approved. Второй видит approved → AlreadyExists.

Idempotent Submission: два Apply одновременно — unique constraint или FOR UPDATE на user_id. Второй видит pending → возвращает существующую.

RevertApplication hard delete: если у автора уже есть заказы — FK падает. Техдолг: soft delete + check FKs.

UX gap: после approve JWT старый. Inventory не пустит писать, пока refresh.

Как чинить:

Authors упал → рестарт. resumeSagas? Нет, у Authors нет саги. Admin повторяет approve.

Auth упал → RevertApplication. Admin повторяет.

Что не сломается:

Публичные /authors, /:user_id.

Уже approved авторы.

text
Authors упал после TX, до UpdateRole → RevertApplication → pending
Два админа approve → FOR UPDATE → первый approved, второй AlreadyExists
RevertApplication hard delete → FK падает (техдолг)
10. Если упал History
Что происходит:

GET /history → 503. Лента не читается.

score.updated / user.registered / shop.purchased / payment.completed / author.upserted не потребляются → лента не пополняется.

Кейсы:

History упал: события копятся в NATS (durable consumer history-ingest). При рестарте — догоняет.

NATS упал: ingest останавливается. При рестарте NATS — догоняет.

Payload Sanitization: битый JSON → {"raw": "..."}. Consumer не падает. Факт события записан.

Nak Pattern: БД недоступна → Nak → сообщение вернётся. history_nak_total растёт.

5 субъектов в одном consumer’е: если один субъект заваливает — lag по всем пяти. Техдолг: разделить.

RetentionWorker: cron внутри процесса. Если History рестартует раз в сутки — retention может не сработать. Техдолг: Kubernetes CronJob.

user.registered + event.user.registered: оба потребляются, дедуп по user_id + event_type. Один факт регистрации — одна строка.

Как чинить:

History упал → рестарт. Consumer догоняет.

БД упала → Nak → сообщение вернётся.

Что не сломается:

Уже записанные события.

Другие сервисы (History — только потребитель).

text
History упал → события копятся в NATS → рестарт → догоняет
БД упала → Nak → сообщение вернётся
Битый JSON → Sanitization → {"raw": "..."}
RetentionWorker рестарт → cron сброс (техдолг)
11. Если упал Analytics
Что происходит:

GET /analytics/dau|mau|retention → 503. Admin не видит метрики.

score.updated / user.registered / shop.purchased / payment.completed / author.upserted не потребляются → ClickHouse не пополняется.

Кейсы:

Analytics упал: события копятся в NATS. При рестарте — догоняет.

ClickHouse упал: Nak → сообщение вернётся. analytics_nak_total растёт. Батч в памяти — при рестарте теряется (at-least-once, повтор).

Too many parts: если батчинг сломан (мелкие INSERT) — ClickHouse выдаёт ошибку. clickhouse_parts_current растёт. Merge не справляется.

Parameterized Query: {name:Type} — защита от injection. Если кто-то передаст UNION — не пройдёт.

5 субъектов в одном consumer’е: lag по всем пяти. Техдолг: разделить.

Батч при рестарте: события в буфере теряются. При рестарте — повтор из NATS (at-least-once).

Как чинить:

Analytics упал → рестарт. Consumer догоняет.

ClickHouse упал → Nak → повтор.

Что не сломается:

Уже записанные события в ClickHouse.

Другие сервисы.

text
Analytics упал → события копятся в NATS → рестарт → догоняет
ClickHouse упал → Nak → повтор
Too many parts → батчинг сломан → clickhouse_parts_current растёт
12. Если упал Notification
Что происходит:

GET /notifications → 503. In-app bell не работает.

score.updated / purchase.paid / purchase.fulfilled / payment.completed / record_beaten не потребляются → уведомления не создаются.

Telegram не отправляется.

Кейсы:

Notification упал: события копятся в NATS. При рестарте — догоняет.

NATS упал: Kafka fallback (если heavy). Если оба — ingest останавливается.

Telegram API упал: notification_telegram_failed_total растёт. Retry with Fixed Delay — повторяем. In-app работает (не зависит от Telegram).

Poison message из Kafka: битое сообщение — бесконечный retry (техдолг). Нужен Ack + лог.

Fan-Out: одно событие → N уведомлений. Если N большое — нагрузка. notification_fanout_size (histogram) — контроль.

Idempotency Key source_event: повтор не создаёт дубль. ON CONFLICT (user_id, source_event) DO NOTHING.

In-app bell soft-fail: если пользователь не залогинен — нет цикла 401. FE просто не делает запрос.

Как чинить:

Notification упал → рестарт. Consumer догоняет.

Telegram API упал → retry. In-app работает.

Kafka poison → Ack (техдолг).

Что не сломается:

Уже созданные уведомления в PG.

In-app (если PG жив).

text
Notification упал → события копятся в NATS → рестарт → догоняет
Telegram упал → retry fixed delay → in-app работает
Kafka poison → бесконечный retry (техдолг)
13. Если упал Fulfillment
Что происходит:

purchase.paid не потребляется → заказы не собираются.

purchase.fulfilled не публикуется → Shop не ставит fulfilled_at, Notification не пишет.

Кейсы:

Fulfillment упал: purchase.paid копится в NATS/Kafka. При рестарте — догоняет.

Fulfillment упал между purchase.paid и purchase.fulfilled: нет Outbox (техдолг). Событие purchase.paid уже consumed, purchase.fulfilled не published. Заказ «завис»: fulfilled_at null, отмена возможна (7 дней). Shop не знает, что сборка не прошла. Техдолг: Outbox.

Idempotency Guard: map в памяти. При рестарте теряется → повторная сборка возможна. Техдолг: persistent storage.

NATS упал: Kafka fallback. Если оба — ingest останавливается.

Delay-Based Simulation: time.After(10s). При рестарте — таймер сброшен. Повторная сборка (если Idempotency Guard не сработал).

newUUID (crypto/rand): UUID непредсказуем, но несогласован с uuid.NewString() (техдолг).

Как чинить:

Fulfillment упал → рестарт. Consumer догоняет.

Между paid и fulfilled → заказ завис. Ручное вмешательство или повтор.

Что не сломается:

Shop (уже записанные purchases).

Отмена (7 дней) — работает.

text
Fulfillment упал между paid и fulfilled → заказ завис (нет Outbox)
Idempotency Guard рестарт → map теряется → повторная сборка
NATS упал → Kafka fallback
14. Если упал Gateway
Что происходит:

Все HTTP-ручки → 503 (если Balancer не перенаправит на живую реплику).

WS /ws/leaderboard — существующие соединения на этой реплике рвутся, клиенты переподключаются к другим.

Кейсы:

Gateway-1 упал: Balancer перенаправляет на Gateway-2/3. Клиенты не замечают (кроме WS на Gateway-1).

Все 3 Gateway упали: 503. Balancer не знает (нет health checks — техдолг), продолжает слать.

Circuit Breaker открыт: dependency болен → 503 fail-fast. Half-open через 10s.

Retry storm: Gateway ретраит gRPC с full jitter. При массовом отказе — лавина. CB защищает.

Thundering herd на WS после рестарта: все клиенты переподключаются. WS 100/min IP — часть 429.

Rate Limiter Redis упал: login fail-closed (не пускает), общий fail-open (пускает). Осознанный trade-off.

Auth упал: ValidateToken timeout → 503.

runGateway God Function: при рестарте — долгая инициализация (техдолг). Все gRPC-клиенты, routes, middleware в одной функции.

Как чинить:

Gateway упал → рестарт. Balancer перенаправит.

CB открыт → ждём half-open (10s).

Redis упал → login fail-closed, общий fail-open.

Что не сломается:

Другие Gateway-реплики.

WS на других репликах.

text
Gateway-1 упал → Balancer → Gateway-2/3 (клиенты не замечают)
Все 3 упали → 503 (Balancer не знает — нет health checks)
CB открыт → 503 fail-fast → half-open 10s
WS рестарт → thundering herd → 100/min IP → часть 429
15. Если упал Balancer
Что происходит:

Всё. Balancer — front door. Клиент не может достучаться.

Gateway ×3 живы, но клиент до них не доходит.

Кейсы:

Balancer упал: single point of failure. Нет резервного. Техдолг: несколько Balancer’ов + VIP/DNS.

Backend упал (Gateway-1): Balancer не знает (нет health checks — техдолг). Продолжает слать. Запрос → ошибка → 502.

/ready не проверяет бэкенды: Balancer говорит ready всегда. Даже если все Gateway упали.

Static registry: если Gateway-4 добавлен — Balancer не знает. Нужен рестарт.

log.Printf в hot path: при нагрузке — logging bottleneck. Техдолг.

Пустые пакеты backend/, proxy/: мёртвый код, не влияет на работу, но грязно.

Как чинить:

Balancer упал → рестарт. Клиенты переподключаются.

Backend упал → запросы в него → 502. Нужны health checks (техдолг).

Что не сломается:

Ничего — Balancer front door.

text
Balancer упал → single point of failure → всё недоступно
Backend упал → Balancer не знает → 502
/ready всегда ready → даже если все Gateway упали
16. Если упал NATS Hub
Что происходит:

Stream EVENTS не создаётся (если первый старт).

Если stream уже создан — NATS Hub упал, но stream жив. Публикация/потребление работают.

Wildcard Subscription event.> — не логирует (debug).

Кейсы:

NATS Hub упал при первом старте: stream EVENTS не создан. Все публикации падают (no stream matches subject). Critical path. При рестарте — создаёт.

NATS Hub упал после создания stream: stream жив. Публикация/потребление работают. Только debug-логи не идут.

Idempotent Stream Creation: AddStream → AlreadyExists → UpdateStream. При рестарте — не падает.

Stream Config as Code: если конфиг изменился (новый subject) — при рестарте UpdateStream обновит.

NATS-кластер упал: NATS Hub не может подключиться. /health → ошибка. Сервис падает. Graceful Degradation нет (техдолг).

MaxBytes 1 ГБ: если stream переполнен — старые сообщения удаляются. Публикация не падает.

Как чинить:

NATS Hub упал → рестарт. Stream создан/обновлён.

NATS-кластер упал → рестарт NATS. NATS Hub подключится.

Что не сломается:

Stream EVENTS (если уже создан).

Публикация/потребление (если stream жив).

text
NATS Hub упал при первом старте → stream не создан → все publish падают
NATS Hub упал после создания → stream жив → publish/consume работают
NATS-кластер упал → NATS Hub не подключается → падает (нет Graceful Degradation)
17. Если упал MCP
Что происходит:

Cursor / Claude Desktop не может вызвать tools. AI-агент теряет доступ к NATS/PG/Redis/Prydwen.

Игроки не замечают — MCP не в player path.

Кейсы:

MCP упал: Cursor видит ошибку. Можно перезапустить (Settings → MCP → Restart).

Один tool упал: Graceful Degradation per-tool — остальные работают.

NATS недоступен: nats_list_streams / nats_list_consumers → ошибка. Остальные tools работают.

Postgres недоступен: postgres_query → ошибка. Остальные работают.

Redis недоступен: redis_get / redis_keys → ошибка.

Prydwen индекс не загружен: search_prydwen → ошибка или пустой результат.

SQL Sandbox: postgres_query с UNION → отклонён. Легко обходится (техдолг).

Утечка ошибок: err.Error() — клиент видит DSN, SQL. Техдолг: санитизировать.

В сеть нельзя: DSN читает целиком. Только stdio, локально.

Как чинить:

MCP упал → рестарт в Cursor / Claude Desktop.

Один tool упал → остальные работают (Graceful Degradation per-tool).

Что не сломается:

Игроки (MCP не в player path).

Все 16 других сервисов.

text
MCP упал → Cursor видит ошибку → рестарт
Один tool упал → остальные работают (Graceful Degradation)
NATS/PG/Redis недоступен → этот tool ошибка, остальные работают
Сводная таблица отказов
#	Сервис	Что падает	Blast radius	Критично?	Graceful?
1	Auth	Закрытые ручки	Все закрытые	Да	Redis упал → 503
2	Game	Submit, score.updated	Лидерборд, Profile	Нет	Outbox догоняет
3	Billing	Деньги	Shop, Game boost	Да	Redis miss → PG
4	Leaderboard	Топ	WS	Нет	Restore из PG
5	Shop	Покупки	Fulfillment	Да	Saga + compensate
6	Inventory	Каталог	Shop (новые)	Нет	Cache miss → PG
7	Profile	Анкета	—	Нет	_partial
8	Payment	Подписки	Shop (мерч)	Да	Idempotent webhook
9	Authors	Заявки	—	Нет	Compensating Action
10	History	Лента	—	Нет	Nak → retry
11	Analytics	DAU/MAU	—	Нет	Nak → retry
12	Notification	Уведомления	—	Нет	In-app работает
13	Fulfillment	Сборка	Shop (fulfilled_at)	Да	Нет Outbox (техдолг)
14	Gateway	Все HTTP	Всё	Да	Balancer → другие реплики
15	Balancer	Всё	Всё	Да	Single point (техдолг)
16	NATS Hub	Stream (первый старт)	Всё async	Да (первый старт)	Stream жив после
17	MCP	Tools	Cursor	Нет	Per-tool
Ключевые «страшилки» на собесе
Thundering herd
Auth login после рестарта: все бьют → rate limit 5/s IP (fail-closed) → 429. By design.

Gateway WS после рестарта: все переподключаются → 100/min IP → часть 429.

Notification Fan-Out: одно событие → N уведомлений. Если N большое — нагрузка.

Leaderboard record_beaten: много рекордов → лавина record_beaten → Notification завален.

Retry storm
Gateway Retry full jitter: размазывает, но при массовом отказе — лавина. CB защищает.

Fulfillment retry: при ошибке — повтор. Idempotency Guard защищает.

Notification retry Fixed Delay: при долгом сбое Telegram — постоянная нагрузка.

Cache stampede
Redis miss на горячем ключе: много параллельных запросов → все идут в PG. Нужен single-flight (сейчас нет).

Leaderboard Restore: при старте — все читают из PG одновременно.

Poison message
Notification Kafka: битое сообщение → бесконечный retry (техдолг).

History: Payload Sanitization — битый JSON → {"raw": "..."}. Не poison.

Split-brain
NATS кластер: 3 ноды, кворум. При split-brain — меньшинство не может писать. NATS решает.

Postgres: нет реплик в текущем compose. Split-brain невозможен.

Deadlock
Billing: Conditional UPDATE + FOR UPDATE могут дать deadlock. PG откатывает одну TX → retry.

Shop: SELECT FOR UPDATE на складе + version. Порядок блокировок важен.

Итого: 17 сервисов, 17 отказов
Что важно запомнить	Сервис
Session Store в Redis	Auth
Outbox FOR UPDATE SKIP LOCKED	Game
Conditional UPDATE + Double-Entry	Billing
ZAddArgs{GT: true} + Ack Batcher	Leaderboard
Saga + Spend Cycle + Resume Worker	Shop
Strategy + Decorator	Inventory
Rule Engine (чистая функция)	Profile
HMAC + ConstantTime + Revenue Split	Payment
State Machine + Compensating Action	Authors
Retention + Nak Pattern	History
ClickHouse + Cohort Retention	Analytics
Dual-Bus + Fan-Out	Notification
Dual-Transport + нет Outbox (техдолг)	Fulfillment
Circuit Breaker + WS Ticket	Gateway
Least Connections + single point	Balancer
Bootstrap stream + Idempotent Creation	NATS Hub
stdio + 6 tools + SQL Sandbox	MCP
Что сказать, если спросят «что самое страшное?»:

Balancer — single point of failure. Нет резервного.

Fulfillment без Outbox. Событие теряется.

Auth + Redis. Redis упал → 503 на всё закрытое.

runGateway God Function. При рестарте — долгая инициализация.

NATS Hub при первом старте. Stream не создан → все publish падают.

Что дальше? Могу:

Разобрать конкретный кейс глубже (например, «Shop упал после Spend, до TX» с sequence)

Сделать шпаргалку на 1 страницу (только «если упал X → Y»)

Разобрать thundering herd / retry storm / cache stampede отдельно