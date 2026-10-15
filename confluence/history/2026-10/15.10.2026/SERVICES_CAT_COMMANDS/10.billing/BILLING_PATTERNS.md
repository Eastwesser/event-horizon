Разбор billing
Что делает сервис
billing — это управление валютами (лампы, билетики):

Балансы — GetBalance, GetAllBalances, AddCurrency, SpendCurrency.

Транзакции — полная история операций (transactions таблица) с balance_after.

Reward Consumer — слушает score.updated и начисляет лампы/билетики за игровые очки.

Outbox — публикует balance.updated.

Revenue Splits — учёт долей (хотя детали в payment).

Optimistic Locking — через version в user_currencies.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg       config.ConfigProvider
    dbpool    *pgxpool.Pool
    redisRepo *repository.RedisBillingRepo
    nc        *nats.Conn
    js        nats.JetStreamContext
    pgRepo    *repository.PostgresBillingRepo
    svc       service.BillingService
    api       pb.BillingServiceServer
}
Стандартный DI-контейнер.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer,
    a.initPostgres, a.initRedis, a.initNATS,
    a.initDomain, a.initOutbox, a.initGRPC,
    a.initMetricsHTTP, a.initScoreSubscribe,
}
Уникальная деталь: initScoreSubscribe — последний шаг. Подписка на score.updated стартует после gRPC и metrics. Это логично: сервис сначала готов принимать запросы, потом подписывается на события.

3. Port (Hexagonal) — два порта
Файл: internal/service/billing_service.go.

Код:

go
type BillingStore interface {
    GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
    AddBalance(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
    SpendBalance(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
    GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error)
    BatchInsertTransactions(ctx context.Context, transactions []repository.Transaction) error
}

type BillingCache interface {
    GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
    SetBalance(ctx context.Context, userID string, currency repository.CurrencyType, balance int, ttl time.Duration) error
    DeleteBalance(ctx context.Context, userID string, currency repository.CurrencyType) error
}
Два порта — BillingStore (Postgres) и BillingCache (Redis).

4. Facade — BillingService
Файл: internal/service/billing_service.go.

Код:

go
type billingService struct {
    pgRepo    BillingStore
    redisRepo BillingCache
}
Тонкий фасад над двумя портами.

5. Atomic Upsert-Increment (AddBalance)
Файл: internal/repository/postgres_repo.go, AddBalance.

Почему: Вместо классического SELECT + UPDATE (который имеет lost-update race), используется INSERT ... ON CONFLICT DO UPDATE SET balance = balance + amount. Это атомарная операция на уровне Postgres.

Код:

go
upsertQuery := `
    INSERT INTO user_currencies (user_id, currency_type, balance, version)
    VALUES ($1, $2, $3, 1)
    ON CONFLICT (user_id, currency_type)
    DO UPDATE SET
        balance = user_currencies.balance + EXCLUDED.balance,
        version = user_currencies.version + 1,
        updated_at = NOW()
    RETURNING balance
`
Комментарий из кода (ты сам написал, молодец):

AddBalance credits currency atomically via an upsert-increment (INSERT ... ON CONFLICT DO UPDATE SET balance = balance + amount), which avoids the classic SELECT-then-UPDATE lost-update race that the old implementation had.

Это Atomic Upsert-Increment — критичный паттерн для финансовых систем.

6. Conditional UPDATE (SpendBalance)
Файл: internal/repository/postgres_repo.go, SpendBalance.

Почему: Списание делается через условный UPDATE:

sql
UPDATE user_currencies
SET balance = balance - $1, version = version + 1, updated_at = NOW()
WHERE user_id = $2 AND currency_type = $3 AND balance >= $1
RETURNING balance
Postgres при UPDATE повторно проверяет WHERE на актуальном значении строки. Это гарантирует, что баланс никогда не уйдёт в минус.

Код:

go
updateQuery := `
    UPDATE user_currencies
    SET balance = balance - $1, version = version + 1, updated_at = NOW()
    WHERE user_id = $2 AND currency_type = $3 AND balance >= $1
    RETURNING balance
`
err = tx.QueryRow(ctx, updateQuery, amount, userID, currency).Scan(&newBalance)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        // ... проверить текущий баланс и вернуть "insufficient balance"
    }
    return 0, err
}
Комментарий из кода:

Postgres re-checks the WHERE clause against the row's current committed value when acquiring the row lock for the UPDATE, so this can never drive the balance negative or lose a concurrent update.

Conditional UPDATE — это самый надёжный способ для финансовых операций.

7. Double-Entry Ledger (через transactions + balance_after)
Файл: internal/repository/postgres_repo.go.

Почему: Каждая операция пишет транзакцию с полем balance_after — балансом после операции. Это Double-Entry Ledger в упрощённом виде.

Код:

go
txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
            VALUES ($1, $2, $3, $4, $5, $6)`
if _, err = tx.Exec(ctx, txQuery, userID, currency, amount, newBalance, reason, referenceID); err != nil {
    return 0, err
}
Плюс: amount хранится со знаком: +amount при пополнении, -amount при списании. Это позволяет восстанавливать историю.

Double-Entry Ledger — основа финансовых систем.

8. Idempotency через reference_id
Файл: internal/repository/postgres_repo.go + internal/service/billing_service.go.

Почему: Каждая операция имеет уникальный reference_id (передаётся извне). Postgres-таблица transactions имеет unique constraint на reference_id. Повторный вызов с тем же ID не спишет/начислит дважды.

Код:

go
func (s *billingService) AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error) {
    // ...
    newBalance, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
    // ...
}
Idempotency Key — критично для распределённых систем, особенно когда shop и payment вызывают billing.

9. Transactional Outbox (как в inventory, shop, payment)
Файл: internal/repository/postgres_repo.go, insertOutboxEvent.

Код:

go
func insertOutboxEvent(ctx context.Context, tx pgx.Tx, eventType string, event any) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal outbox payload: %w", err)
    }
    _, err = tx.Exec(ctx, `INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`, eventType, payload)
    return err
}
Плюс: событие balance.updated пишется в той же транзакции, что и изменение баланса. Это гарантирует at-least-once delivery.

Комментарий из кода:

The balance change, transaction row, and outbox event are all written in one transaction so an event is never lost or emitted without the balance actually changing.

Transactional Outbox — золотой стандарт для event-driven систем.

10. Outbox Worker (Polling Publisher)
Файл: internal/worker/outbox_worker.go.

Код:

go
rows, err := w.db.Query(ctx, `
    SELECT id, event_type, payload FROM outbox
    WHERE processed = false
    ORDER BY created_at
    LIMIT 100
`)
// ...
for _, p := range batch {
    if _, err := w.js.Publish(p.eventType, p.payload); err != nil { /* ... */ }
    if _, err := w.db.Exec(ctx, `UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1`, p.id); err != nil { /* ... */ }
}
Polling Publisher — без FOR UPDATE SKIP LOCKED (в отличие от shop), но с batch в 100 записей.

11. Cache-Aside с суррогатным значением
Файл: internal/repository/redis_repo.go, GetBalance.

Почему: При промахе кеша возвращается -1, а не redis.Nil. Это позволяет отличить промах от реального нулевого баланса (0 — валидный баланс).

Код:

go
func (r *RedisBillingRepo) GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error) {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return -1, nil // not found in cache
    }
    // ...
}
И в сервисе:

go
if cached >= 0 {
    return cached, nil
}
Sentinel Value — приём для отличия «нет данных» от «данные = 0».

12. Cache Invalidation on Write
Файл: internal/service/billing_service.go.

Код:

go
func (s *billingService) AddCurrency(...) (int, error) {
    // ...
    newBalance, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
    if err != nil {
        return 0, err
    }
    _ = s.redisRepo.DeleteBalance(ctx, userID, currency)
    return newBalance, nil
}
Cache Invalidation — после записи кеш удаляется. Следующий GetBalance подтянет свежее значение.

13. Check-Only Mode (SpendCurrency)
Файл: internal/service/billing_service.go, SpendCurrency.

Почему: SpendCurrency может работать в двух режимах:

checkOnly = true — только проверка, хватает ли баланса, без списания.

checkOnly = false — реальное списание.

Код:

go
if checkOnly {
    balance, err := s.pgRepo.GetBalance(ctx, userID, currency)
    if err != nil {
        return 0, err
    }
    if balance < amount {
        return 0, fmt.Errorf("insufficient balance: have %d, need %d", balance, amount)
    }
    return balance, nil
}
newBalance, err := s.pgRepo.SpendBalance(ctx, userID, currency, amount, reason, referenceID)
Check-Only Mode — важен для shop, где нужно сначала проверить баланс, потом списать.

14. Reward Consumer (Event-Driven)
Файл: internal/app/di.go, initScoreSubscribe.

Почему: Billing слушает score.updated и начисляет лампы и билетики за игровые очки.

Код:

go
_, err := a.di.js.Subscribe("score.updated", func(msg *nats.Msg) {
    var event ScoreEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        log.Error("failed to unmarshal score event", "err", err)
        return
    }
    lampsRefID := fmt.Sprintf("%s-lamps-%d", event.UserID, time.Now().UnixNano())
    ticketsRefID := fmt.Sprintf("%s-tickets-%d", event.UserID, time.Now().UnixNano())

    if event.LampsEarned > 0 {
        _, err := billingService.AddCurrency(context.Background(), event.UserID,
            repository.Lamps, event.LampsEarned, "game_reward", lampsRefID)
        // ...
    }
    if event.TicketsEarned > 0 {
        _, err := billingService.AddCurrency(context.Background(), event.UserID,
            repository.Tickets, event.TicketsEarned, "game_reward", ticketsRefID)
        // ...
    }
    _ = msg.Ack()
}, nats.Durable("billing-durable"), nats.ManualAck())
Reward Consumer — начисление валюты за игровые очки. Уникально для billing.

Замечание: reference_id генерируется с time.Now().UnixNano(). Это не идеально для идемпотентности: если событие придёт дважды, ID будет разный, и валюта начислится дважды. Но nats.Durable + ManualAck даёт at-least-once, а event.ReferenceID из события не используется. Это потенциальный баг — но он не в фокусе.

15. Batch INSERT через COPY
Файл: internal/repository/postgres_repo.go, BatchInsertTransactions.

Почему: Для массовой вставки транзакций используется pgx.CopyFrom — это в разы быстрее, чем INSERT в цикле.

Код:

go
copyFrom := pgx.CopyFromSlice(len(transactions), func(i int) ([]any, error) {
    t := transactions[i]
    return []any{
        t.ID, t.UserID, t.Currency, t.Amount,
        t.BalanceAfter, t.Reason, t.ReferenceID, t.CreatedAt,
    }, nil
})

_, err = tx.CopyFrom(ctx, pgx.Identifier{"transactions"},
    []string{"id", "user_id", "currency_type", "amount", "balance_after", "reason", "reference_id", "created_at"},
    copyFrom)
COPY Protocol — Postgres-specific быстрый bulk insert. Уникально для billing.

16. Transaction Batch (in-memory)
Файл: internal/service/billing_service.go, тип TransactionBatch.

Почему: TransactionBatch — in-memory буфер для транзакций. Он накапливает транзакции и flush'ит их пачкой через BatchInsertTransactions. Потокобезопасен (sync.Mutex).

Код:

go
type TransactionBatch struct {
    transactions []repository.Transaction
    mu           sync.Mutex
}

func (b *TransactionBatch) Add(tx repository.Transaction) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.transactions = append(b.transactions, tx)
}

func (b *TransactionBatch) Flush(ctx context.Context, pgRepo BillingStore) {
    b.mu.Lock()
    defer b.mu.Unlock()
    if len(b.transactions) == 0 {
        return
    }
    if err := pgRepo.BatchInsertTransactions(ctx, b.transactions); err != nil {
        log.Printf("Failed to batch insert transactions: %v", err)
    }
    b.transactions = b.transactions[:0]
}
Batch Buffer — накопление + flush. Уникально для billing.

17. Optimistic Locking (version)
Файл: internal/repository/postgres_repo.go.

Код:

go
upsertQuery := `
    INSERT INTO user_currencies (user_id, currency_type, balance, version)
    VALUES ($1, $2, $3, 1)
    ON CONFLICT (user_id, currency_type)
    DO UPDATE SET
        balance = user_currencies.balance + EXCLUDED.balance,
        version = user_currencies.version + 1,
        updated_at = NOW()
    RETURNING balance
`
Optimistic Locking через version — стандартный приём.

18. Tracer OTel
Файл: internal/app/di.go, initTracer.

Как в leaderboard и shop. OTel + Jaeger.

19. Converter (Currency)
Файл: internal/converter/currency.go.

Код:

go
func ProtoCurrencyToRepo(c pb.CurrencyType) repository.CurrencyType { /* ... */ }
func RepoCurrencyToProto(c repository.CurrencyType) pb.CurrencyType { /* ... */ }
func BalanceToProto(userID string, currency repository.CurrencyType, amount int32, updatedAt int64) *pb.GetBalanceResponse { /* ... */ }
func DomainBalanceToEntry(b model.Balance) *pb.BalanceEntry { /* ... */ }
Converter — четыре функции для конвертации валют.

20. Mock Object (Tests)
Файлы: billing_ops_test.go, billing_batch_test.go, billing_validation_test.go.

Три мока:

mockBillingStore — для BillingStore.

mockBillingCache — для BillingCache.

TransactionBatch — собственный тест.

Четыре тестовых файла — рекорд для тестов.

21. Testcontainers
Файл: internal/repository/postgres_integration_test.go.

Уникальный тест: TestBilling_ConcurrentSpend_NoNegative — проверка race condition на списании.

Код:

go
errCh := make(chan error, 2)
for i := 0; i < 2; i++ {
    i := i
    go func() {
        _, err := repo.SpendBalance(ctx, userID, repository.Tickets, 40, "race", fmt.Sprintf("ref-race-%d", i))
        errCh <- err
    }()
}
var fails int
for i := 0; i < 2; i++ {
    if err := <-errCh; err != nil {
        fails++
    }
}
if fails != 1 {
    t.Fatalf("exactly one spend should fail under race, fails=%d", fails)
}
Race Condition Test — тест, который проверяет, что при параллельных списаниях баланс не уйдёт в минус. Редко встречается в реальных проектах.

22. Build Tags
Файл: internal/repository/postgres_integration_test.go, //go:build integration.

23. Graceful Shutdown / Closer
Файл: internal/app/di.go.

Итог по billing
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go
Port (Hexagonal) — 2 порта	service/billing_service.go
Facade	service/billing_service.go
Atomic Upsert-Increment	repository/postgres_repo.go, AddBalance
Conditional UPDATE	repository/postgres_repo.go, SpendBalance
Double-Entry Ledger	repository/postgres_repo.go, transactions + balance_after
Idempotency через reference_id	service/billing_service.go + transactions.reference_id
Transactional Outbox	repository/postgres_repo.go, insertOutboxEvent
Outbox Worker (Polling Publisher)	worker/outbox_worker.go
Cache-Aside с суррогатным значением	repository/redis_repo.go, GetBalance
Cache Invalidation on Write	service/billing_service.go
Check-Only Mode	service/billing_service.go, SpendCurrency
Reward Consumer	app/di.go, initScoreSubscribe
Batch INSERT через COPY	repository/postgres_repo.go, BatchInsertTransactions
Transaction Batch (in-memory)	service/billing_service.go, TransactionBatch
Optimistic Locking (version)	repository/postgres_repo.go
Tracer OTel	app/di.go, initTracer
Converter (Currency)	converter/currency.go
Mock Object (×3)	service/*_test.go
Testcontainers	repository/postgres_integration_test.go
Race Condition Test	repository/postgres_integration_test.go, TestBilling_ConcurrentSpend_NoNegative
Build Tags	repository/postgres_integration_test.go
Graceful Shutdown	app/di.go
Итого 24 паттерна и приёма.

Уникальные фичи billing:

Atomic Upsert-Increment — единственный сервис с ON CONFLICT DO UPDATE SET balance = balance + amount.

Conditional UPDATE — списание через WHERE balance >= $1.

Double-Entry Ledger — транзакции с balance_after.

Batch INSERT через COPY — Postgres COPY protocol.

Transaction Batch (in-memory) — накопление + flush.

Check-Only Mode — проверка без списания.

Race Condition Test — тест на конкурентность.

Готовый markdown для BILLING_PATTERNS.md
md
# billing — Programming Patterns

> Сервис: управление валютами (лампы, билетики).
> Postgres + Redis + NATS + gRPC + OTel.

## Что делает сервис

Управляет балансами пользователей (лампы, билетики). Ведёт полную историю транзакций с `balance_after`. Слушает `score.updated` и начисляет валюту за игровые очки. Публикует `balance.updated` через Outbox.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`
- **Особенность:** `initScoreSubscribe` — последний шаг.

### 3. Port (Hexagonal) — 2 порта
- **Файл:** `internal/service/billing_service.go`
- **Порты:** `BillingStore`, `BillingCache`.

### 4. Facade
- **Файл:** `internal/service/billing_service.go`

### 5. Atomic Upsert-Increment
- **Файл:** `internal/repository/postgres_repo.go`, `AddBalance`
- **Код:**
  ```sql
  INSERT INTO user_currencies (...) VALUES (...)
  ON CONFLICT (user_id, currency_type)
  DO UPDATE SET balance = user_currencies.balance + EXCLUDED.balance, version = version + 1
  RETURNING balance
6. Conditional UPDATE
Файл: internal/repository/postgres_repo.go, SpendBalance

Код:

sql
UPDATE user_currencies
SET balance = balance - $1, version = version + 1
WHERE user_id = $2 AND currency_type = $3 AND balance >= $1
RETURNING balance
7. Double-Entry Ledger
Файл: internal/repository/postgres_repo.go

Суть: таблица transactions с полем balance_after, amount со знаком.

8. Idempotency через reference_id
Файл: internal/service/billing_service.go + transactions.reference_id

9. Transactional Outbox
Файл: internal/repository/postgres_repo.go, insertOutboxEvent

10. Outbox Worker (Polling Publisher)
Файл: internal/worker/outbox_worker.go

11. Cache-Aside с суррогатным значением
Файл: internal/repository/redis_repo.go, GetBalance

Код: return -1, nil при промахе.

12. Cache Invalidation on Write
Файл: internal/service/billing_service.go

13. Check-Only Mode
Файл: internal/service/billing_service.go, SpendCurrency

14. Reward Consumer
Файл: internal/app/di.go, initScoreSubscribe

15. Batch INSERT через COPY
Файл: internal/repository/postgres_repo.go, BatchInsertTransactions

Код: tx.CopyFrom(ctx, pgx.Identifier{"transactions"}, ...)

16. Transaction Batch (in-memory)
Файл: internal/service/billing_service.go, TransactionBatch

17. Optimistic Locking (version)
Файл: internal/repository/postgres_repo.go

18. Tracer OTel
Файл: internal/app/di.go, initTracer

19. Converter (Currency)
Файл: internal/converter/currency.go

20. Mock Object (×3)
Файлы: internal/service/*_test.go

21. Testcontainers
Файл: internal/repository/postgres_integration_test.go

22. Race Condition Test
Файл: internal/repository/postgres_integration_test.go, TestBilling_ConcurrentSpend_NoNegative

Суть: два параллельных списания по 40 из баланса 50 → одно падает.

23. Build Tags
Файл: internal/repository/postgres_integration_test.go, //go:build integration

24. Graceful Shutdown
Файл: internal/app/di.go