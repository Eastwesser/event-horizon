# BACKEND TECH DEBT 06.11.2026

AGREGATED TECH DEBT FROM 15th Oct 2026 - current

### Legend:

Легенда:

    🟥 P0 — блокер

    🟧 P1 — надёжность/безопасность

    🟨 P2 — гигиена/UX

    🟦 P3 — roadmap

    🟩 P4 — долгосрок/парковка

    ✅ Закрыто — не тащим

    - [дубль] — слито из нескольких источников
    - [противоречие] — расхождение источников
    - [упущено] — не в README, но есть в бреже
    - [проверить] — статус неясен
    - [проверки] — из аудита, отдельной секцией

🟥 P0 — Блокеры (нельзя в прод)
Итого P0: 9 задач (было 6 в одном файле, 7 в другом — после слияния и добавления IDOR + paymentClient).

🟧 P1 — Надёжность / безопасность
Итого P1: 23 задачи (было 20 в одном файле, 12 в другом — после слияния и добавления single-flight, rate limit, reuse detection).

🟨 P2 — Гигиена / UX (бэкенд)
Итого P2: 15 задач (было 12 в одном, 13 в другом — после слияния + Gin + seed-v110).

## BACKEND PRIMARY TASKS:

### Задача  Домен   Источник    Пометка

### ========================= FULFILLMENT =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-1	        Fulfillment: внедрить Transactional Outbox — событие теряется между purchase.paid и purchase.fulfilled	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README (не отражено)	[упущено] [дубль]
P0-2	        Fulfillment: заменить time.After(10s) на реальный WMS API	Бэкенд	TECH_DEBT_15.10, README («заменится на WMS»)	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-4	Fulfillment: newUUID (crypto/rand) vs uuid.NewString() — унифицировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= GATEWAY =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-3    Gateway: разбить runGateway God Function (>1000 строк) на bootstrap/clients.go, routes.go, middleware.go    Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟧 P1 — Надёжность / безопасность
P1-1	Gateway: два Hub (app/gateway.go + websocket/hub.go) — унифицировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-2	Gateway: async gRPC в admin/users без CB — обернуть в throughBreaker	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-17	Все: нет WAF — WAF перед Gateway	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-1	Gateway: log.Printf везде — перейти на slog	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= SHOP =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-4	Shop: Saga started > 60s не берётся resumeSagas — деньги списаны, компенсации нет. Воркер должен брать started, проверять Billing.GetTransactions(spend_ref)	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, FINAL_AUDIT	[дубль] [упущено] ✅ Закрыто — не тащим
P0-9	Shop: paymentClient может быть nil — мерч-гейт деградирует. Метрика merch_gate_disabled	Бэкенд	TECH_DEBT_15.10, EVENT_HORIZON_PATTERNS	[упущено]

🟧 P1 — Надёжность / безопасность
P1-18	Shop: kafkaProd создаётся, не используется — убрать или подключить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] Need to test
P1-19	Shop: cancelPurchase reference_id = cancel-{id} — может задвоить. Idempotency Key	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-20	Shop: не хватает single-flight на cache stampede (Billing, Inventory, Shop, Payment)	Бэкенд	EH_BOARDER_CASES, FINAL_AUDIT	[упущено] Need to write

### ========================= BILLING =========================

🟥 P0 — Блокеры (нельзя в прод) ✅ Закрыто — не тащим
P0-5	Billing: RewardConsumer использует time.Now().UnixNano() для reference_id → двойное начисление при redelivery. Использовать event.ReferenceID	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] [упущено]

🟨 P2 — Гигиена / UX (бэкенд)
P2-6	Billing: GetBalance возвращает -1 при cache miss — (int, bool)	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= AUTHORS =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-6	Authors: RevertApplication hard delete — FK падает. Soft delete + check FKs	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] [упущено]

🟨 P2 — Гигиена / UX (бэкенд)
P2-9	Authors: UX gap — JWT старый после approve — автоперевыпуск JWT или инструкция	Бэкенд/Фронт	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= BALANCER =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-7	Balancer: single point of failure — нет резервного. Несколько Balancer + VIP/DNS	Девопс	BACKEND_BIG_TECH_DEBT, EH_BOARDER_CASES	[дубль] [упущено]

🟧 P1 — Надёжность / безопасность
P1-8	Balancer: /ready не проверяет бэкенды — реальная проверка хотя бы одного	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-9	Balancer: static registry — нет service discovery. etcd/consul/k8s	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-10	Balancer: нет health checks бэкендов — периодический ping /health	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-2	Balancer: пустые пакеты backend/, proxy/ — удалить или наполнить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-3	Balancer: log.Printf в hot path — structured logging	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= INVENTORY =========================

🟥 P0 — Блокеры (нельзя в прод)
P0-8	Inventory: IDOR в UpdateItem/DeleteItem — на бэке нет проверки author_id. Фронт скрывает кнопки, но API открыт	Бэкенд	PATCH_NOTES_1.1.3.1, AUTHORS_AUDIT, README (упомянут в патче)	[дубль] [проверить] — бэк не починен

🟧 P1 — Надёжность / безопасность
P1-3	Inventory: два драйвера (PG + Mongo) — определиться с одним	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] POSTGRES FOR ALL (MONGO JUST TESTING)
P1-4	Inventory: toPostgresFilter через fmt.Sprintf — переписать без fmt.Sprintf для имён колонок	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-5	Inventory: Outbox через type assertion — явный интерфейс TransactionalRepository	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] CHECK

# BACKEND SECONDARY TASKS:

### ========================= MCP =========================

🟧 P1 — Надёжность / безопасность
P1-6	MCP: SQL Sandbox string matching — обходится. Read-only роль + pgx prepared	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-7	MCP: err.Error() — утечка внутренних деталей — санитизировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)

### ========================= NOTIFICATION =========================

🟧 P1 — Надёжность / безопасность
P1-11	Notification: poison message из Kafka — бесконечный retry. Ack + DLQ	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-22	Notification: Fan-Out без rate limit	Бэкенд	EH_BOARDER_CASES	[упущено]

🟨 P2 — Гигиена / UX (бэкенд)
P2-12	Notification: Retry Fixed Delay для Telegram — exponential backoff	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= ANALYTICS =========================

🟧 P1 — Надёжность / безопасность
P1-12	Analytics: fmt.Sprintf для db name — параметризовать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-10	Analytics: 5 субъектов в одном consumer'е — разделить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= LEADERBOARD =========================

🟧 P1 — Надёжность / безопасность
P1-13	Leaderboard: RestoreFromPostgres при падении PG — пустой leaderboard. Retry/backoff	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-21	Leaderboard: нет rate limit на record_beaten — лавина при массовых рекордах	Бэкенд	EH_BOARDER_CASES, FINAL_AUDIT	[упущено]

### ========================= PAYMENT =========================

🟧 P1 — Надёжность / безопасность
P1-14	Payment: нет сверки суммы webhook с payments.amount — сверять amount	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-8	Payment: subscriptionDays default 30 — явно в config	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= AUTH =========================

🟧 P1 — Надёжность / безопасность
P1-16	Auth: нет ротации JWT_SECRET — ротация	Бэкенд/Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README	[дубль]
P1-23	Auth: reuse detection для refresh — если старый jti использован повторно → отозвать все сессии	Бекенд	EH_SHATTER_AND_ISSUE_SOLVING	[упущено]

🟨 P2 — Гигиена / UX (бэкенд)
P2-7	Auth: boosty.to URL захардкожен — Config	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= HISTORY =========================

🟧 P1 — Надёжность / безопасность
P1-15	History: RetentionWorker в процессе — при рестарте сброс. K8s CronJob	Бэкенд/Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

🟨 P2 — Гигиена / UX (бэкенд)
P2-11	History: 5 субъектов в одном consumer'е — разделить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

### ========================= GAME =========================

🟨 P2 — Гигиена / UX (бэкенд)
P2-5	Game: redis-game :6380 не подключён — реализовать anti-cheat по GAME_REDIS.md (INCR submit, SETNX dedup, GET score cache, fail-open). Оставляем.	Бэкенд	GAME_REDIS.md, GAME_REDIS_ANTICHEAT.md, README	[противоречие] README vs GAME_REDIS.md — решено: подключаем

### ========================= PROFILE =========================

🟨 P2 — Гигиена / UX (бэкенд)
P2-13	Profile: UpdateProfile наружу не открыт (умеет переписать total_score) — отдельный gRPC без total_score	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]

## ========== OTHERS ==========

P2-14	FUTURE_TASKS: Gin matches routes before middleware v1 — legacy /api/* rewrite хрупок. Проверить в новых Gin, добавить интеграционный тест	Бэкенд	FUTURE_TASKS.md, BACKEND_BIG_TECH_DEBT (критичные)	[упущено] — в позднем P0/P1 нет
P2-15	FUTURE_TASKS: make seed-v110 заметка — описать в DAILY_OPS.md / README.md, что засеивает	Девопс/Док	FUTURE_TASKS.md, BACKEND_BIG_TECH_DEBT (мелкие)	[упущено]
P2-16	BACKEND_BIG_TECH_DEBT: poison message в Notification → Ack + лог — уже в P1-11, но пометить как [дубль]	—	—	[дубль]

