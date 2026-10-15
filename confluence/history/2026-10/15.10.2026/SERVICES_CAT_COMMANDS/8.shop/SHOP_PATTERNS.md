Разбор shop
Что делает сервис
shop — это магазин: каталог товаров, покупка, отмена, инвентарь, продажи для авторов. Особенности:

Покупка затрагивает два сервиса — shop (сток, инвентарь) и billing (списание билетиков).

Saga Pattern — покупка это сага с состояниями started → spent → stocked → compensated.

Компенсация — если списание прошло, а сток не удалось обновить, билетики возвращаются.

Idempotency через reference_id — shop:user:item:cycle. Цикл бампается после refund.

Outbox + Polling Publisher — для публикации shop.purchased и purchase.paid.

Saga Resume Worker — доводит «зависшие» саги до конца.

Merch Gate — проверка подписки через payment.CanPurchaseMerch.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg       config.ConfigProvider
    db        *sql.DB
    redisRepo *repository.RedisShopRepo
    nc        *nats.Conn
    js        nats.JetStreamContext
    pgRepo    *repository.PostgresShopRepo
    billing   billingPb.BillingServiceClient
    svc       service.ShopService
    api       pb.ShopServiceServer
    kafkaProd kafka.Producer
}
Тут gRPC-клиенты к billing и payment — уникальная деталь.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer,
    a.initPostgres, a.initRedis, a.initNATS, a.initKafka,
    a.initDomain, a.initOutbox, a.initGRPC, a.initMetricsHTTP,
}
initOutbox — отдельный шаг после initDomain, потому что outbox worker зависит от billing-клиента.

3. Port (Hexagonal) — три порта
Файл: internal/service/shop_service.go.

Почему: shopService общается с внешним миром через три порта:

ShopStore — Postgres.

ShopCache — Redis.

billingPb.BillingServiceClient + paymentPb.PaymentServiceClient — gRPC.

Код:

go
type ShopStore interface {
    GetItems(ctx context.Context, category, gameID string) ([]repository.Item, error)
    GetItemByID(ctx context.Context, itemID string) (*repository.Item, error)
    // ...
    PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []repository.OutboxRecord) error
    SavePurchaseSaga(ctx context.Context, saga repository.PurchaseSaga) (string, string, error)
    MarkPurchaseSaga(ctx context.Context, id, state string) error
    // ...
}

type ShopCache interface {
    GetItems(ctx context.Context, key string) ([]repository.Item, error)
    SetItems(ctx context.Context, key string, items []repository.Item, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
Три порта — рекорд для shop.

4. Facade — ShopService
Файл: internal/service/shop_service.go.

Код:

go
type shopService struct {
    pgRepo    ShopStore
    redisRepo ShopCache
    js        nats.JetStreamContext
    billing   billingPb.BillingServiceClient
    payment   paymentPb.PaymentServiceClient
    kafkaProd kafka.Producer
}
Фасад над шестью зависимостями — самый сложный в проекте.

5. Saga Pattern (Orchestration)
Файл: internal/service/shop_service.go, PurchaseItem.

Почему: Покупка — это сага с оркестратором (purchase_sagas таблица). Состояния:

started → сага создана.

spent → билетики списаны в billing.

stocked → сток обновлён и запись создана.

compensating → откат.

compensated → откат завершён.

aborted → сага прервана до списания.

Код:

go
sagaID, purchaseID, err := s.pgRepo.SavePurchaseSaga(ctx, repository.PurchaseSaga{
    ID: uuid.NewString(), UserID: userID, ItemID: itemID, PurchaseID: uuid.NewString(),
    SpendRef: spendRef, RefundRef: refundRef, Price: item.Price, State: "started",
})
// ...
_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "spent")
// ...
_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
// или
_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensating")
// ...
_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensated")
Saga Pattern (Orchestration) — сага управляется централизованно через таблицу состояний. Ключевой паттерн shop.

6. Compensation Action (компенсация)
Файл: internal/service/shop_service.go, PurchaseItem (ветка ошибки).

Почему: Если списание прошло, а запись в БД упала — компенсация: возвращаем билетики через billing.AddCurrency.

Код:

go
if err := s.pgRepo.PurchaseItemWithStock(...); err != nil {
    if isDuplicateRef(err) {
        _ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
        return 0, model.ErrAlreadyOwned
    }
    _ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensating")
    log.Printf("CRITICAL: tickets spent but purchase failed for user %s, item %s: %v — refunding", userID, itemID, err)
    outboxID := s.enqueueRefund(ctx, userID, itemID, item.Price, refundRef, err)
    _ = s.pgRepo.BumpSpendCycle(ctx, userID, itemID)
    if _, refundErr := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
        UserId:      userID,
        Currency:    billingPb.CurrencyType_TICKETS,
        Amount:      int32(item.Price),
        Reason:      "shop_purchase_refund",
        ReferenceId: refundRef,
    }); refundErr != nil && !isDuplicateRef(refundErr) {
        metrics.PurchaseRefundFailures.Inc()
        log.Printf("CRITICAL: refund also failed for user %s, item %s: %v", userID, itemID, refundErr)
        return 0, fmt.Errorf("failed to record purchase: %w", err)
    }
    if outboxID != "" {
        _ = s.pgRepo.MarkOutboxProcessed(ctx, outboxID)
    }
    _ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensated")
    return 0, fmt.Errorf("failed to record purchase: %w", err)
}
Compensation Action — обратное действие при провале. Ключевой паттерн Saga.

7. Idempotency через reference_id
Файл: internal/service/shop_service.go, PurchaseItem + creditCancel.

Почему: Каждая операция имеет уникальный reference_id:

shop:{user}:{item}:{cycle} — для списания.

refund:shop:{user}:{item}:{cycle} — для возврата при компенсации.

cancel-{purchaseID} — для возврата при отмене.

Billing идемпотентен по reference_id: повторный вызов с тем же ID не спишет дважды.

Код:

go
cycle, err := s.pgRepo.NextSpendCycle(ctx, userID, itemID)
spendRef := fmt.Sprintf("shop:%s:%s:%d", userID, itemID, cycle)
refundRef := fmt.Sprintf("refund:shop:%s:%s:%d", userID, itemID, cycle)
Idempotency Key — критично для распределённых систем.

8. Spend Cycle (уникальный паттерн)
Файл: internal/repository/postgres_repo.go, NextSpendCycle / BumpSpendCycle.

Почему: purchase_claims таблица хранит текущий цикл для пары (user_id, item_id). Это гарантирует, что параллельные клики используют тот же reference_id (не спишут дважды), а после refund cycle инкрементится, и следующий вызов получает новый reference_id.

Код:

go
func (r *PostgresShopRepo) NextSpendCycle(ctx context.Context, userID, itemID string) (int, error) {
    var cycle int
    err := r.db.QueryRowContext(ctx, `
        INSERT INTO purchase_claims (user_id, item_id, cycle)
        VALUES ($1, $2, 1)
        ON CONFLICT (user_id, item_id) DO UPDATE SET cycle = purchase_claims.cycle
        RETURNING cycle
    `, userID, itemID).Scan(&cycle)
    return cycle, err
}
Это Spend Cycle Pattern — уникальный приём для идемпотентности покупок. Рекорд по оригинальности.

9. Outbox Pattern (как в inventory)
Файл: internal/repository/postgres_repo.go, PurchaseItemWithStock.

Код:

go
for _, rec := range records {
    if _, err := tx.ExecContext(ctx, `
        INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
        rec.EventType, rec.Payload); err != nil {
        return err
    }
}
Событие shop.purchased и purchase.paid пишутся в той же транзакции, что и покупка.

10. Outbox Worker + Compensation Handler
Файл: internal/worker/outbox_worker.go.

Почему: Outbox worker не только публикует события в NATS, но и обрабатывает компенсации через CompensationEvent = "billing.refund". Это уникально для shop.

Код:

go
const CompensationEvent = "billing.refund"

type CompensateFunc func(ctx context.Context, payload []byte) error

func (w *OutboxWorker) deliver(ctx context.Context, row outboxRow) error {
    if row.eventType == CompensationEvent {
        if w.compensate == nil {
            return errNoCompensator
        }
        err := w.compensate(ctx, row.payload)
        if err != nil && !isDuplicateRef(err) {
            return err
        }
        return nil
    }
    _, err := w.js.Publish(row.eventType, row.payload)
    return err
}
Compensation Handler — outbox worker не только публикует, но и выполняет компенсации.

11. Saga Resume Worker (уникальный паттерн)
Файл: internal/worker/outbox_worker.go, resumeSagas.

Почему: Если сага «зависла» в состоянии spent или compensating дольше 30 секунд, воркер доводит её до конца: либо помечает stocked (если запись появилась), либо делает refund.

Код:

go
func (w *OutboxWorker) resumeSagas(ctx context.Context) {
    // ...
    rows, err := tx.QueryContext(ctx, `
        SELECT id::text, user_id, item_id, purchase_id::text, refund_ref, price
        FROM purchase_sagas
        WHERE state IN ('spent', 'compensating')
          AND updated_at < NOW() - INTERVAL '30 seconds'
        ORDER BY updated_at
        LIMIT 10
        FOR UPDATE SKIP LOCKED
    `)
    // ...
    // для каждой саги:
    //   если purchase существует → state = 'stocked'
    //   иначе → state = 'compensating' → refund → state = 'compensated'
}
Это Saga Resume Worker — фоновый воркер, доводящий «зависшие» саги до конца. Уникально для shop.

12. Pessimistic Lock + Optimistic Lock (в одной транзакции)
Файл: internal/repository/postgres_repo.go, PurchaseItemWithStock.

Почему: Сток проверяется через pessimistic lock (SELECT ... FOR UPDATE), а обновление — через optimistic lock (version = $2).

Код:

go
var stock, version int
err = tx.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&stock, &version)
// ...
res, err := tx.ExecContext(ctx, `
    UPDATE items SET stock = stock - 1, version = version + 1 WHERE id = $1 AND version = $2`, itemID, version)
if n, _ := res.RowsAffected(); n == 0 {
    return fmt.Errorf("optimistic lock conflict on item %s", itemID)
}
Гибрид двух стратегий блокировок — редкость. Уникально для shop.

13. Refund Window / Soft Constraint
Файл: internal/repository/postgres_repo.go, RefundPurchase.

Почему: Refund разрешён только в течение 7 дней (refundable_until). Плюс блокировка для физических товаров после fulfilled_at.

Код:

go
now := time.Now().UTC()
if now.After(refundableUntil) {
    return nil, model.ErrRefundWindowExpired
}
if fulfilledAt.Valid && model.IsPhysicalItem(category) {
    return nil, model.ErrAlreadyFulfilled
}
Refund Window + Physical Fulfillment Lock — бизнес-правила.

14. Idempotent Refund
Файл: internal/repository/postgres_repo.go, RefundPurchase.

Код:

go
if err == sql.ErrNoRows {
    var alreadyID string
    var alreadyPrice int
    qerr := tx.QueryRowContext(ctx, `
        SELECT id, price FROM purchases
        WHERE user_id = $1 AND item_id = $2
          AND (status = 'REFUNDED' OR refunded_at IS NOT NULL)
        ORDER BY refunded_at DESC NULLS LAST
        LIMIT 1
    `, userID, itemID).Scan(&alreadyID, &alreadyPrice)
    if qerr == nil {
        return &RefundResult{PurchaseID: alreadyID, Price: alreadyPrice, AlreadyRefunded: true}, nil
    }
    // ...
}
Idempotent Refund — повторный cancel не вернёт билетики дважды.

15. FOR UPDATE SKIP LOCKED
Файл: internal/worker/outbox_worker.go + postgres_repo.go.

Код:

go
rows, err := tx.QueryContext(ctx, `
    SELECT id, event_type, payload FROM outbox
    WHERE processed = false
    ORDER BY created_at
    LIMIT 20
    FOR UPDATE SKIP LOCKED
`)
FOR UPDATE SKIP LOCKED — паттерн для многопоточных воркеров: каждый воркер берёт свою порцию, не блокируя другие.

16. gRPC Client Dial + Graceful Close
Файл: internal/app/di.go, initDomain.

Код:

go
billingConn, err := grpc.Dial(a.cfg.BillingConfig().Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
a.closer.AddNamed("billing grpc", func(context.Context) error { return billingConn.Close() })
billingClient := billingPb.NewBillingServiceClient(billingConn)
gRPC Client Dial — с graceful close. Уникально для shop — другие сервисы не зовут gRPC-соседей.

17. Optional Dependency (Payment)
Файл: internal/app/di.go, initDomain.

Почему: paymentClient может быть nil, если addr пустой. Это грациозная деградация: мерч-гейт отключается, но обычные покупки работают.

Код:

go
var paymentClient paymentPb.PaymentServiceClient
if addr := a.cfg.PaymentConfig().Addr(); addr != "" {
    pconn, perr := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if perr != nil {
        a.log.Warn("payment dial failed (merch gate degraded)", "err", perr)
    } else {
        a.closer.AddNamed("payment grpc", func(context.Context) error { return pconn.Close() })
        paymentClient = paymentPb.NewPaymentServiceClient(pconn)
    }
}
Optional Dependency — сервис работает, даже если payment недоступен.

18. Cache-Aside (как в других сервисах)
Файл: internal/service/shop_service.go, GetItems.

Код:

go
items, err := s.redisRepo.GetItems(ctx, cacheKey)
if err == nil {
    s.applyOwnedFlags(ctx, userID, items)
    return items, nil
}
items, err = s.pgRepo.GetItems(ctx, category, gameID)
// ...
_ = s.redisRepo.SetItems(ctx, cacheKey, items, 5*time.Minute)
Cache-Aside — стандартный приём.

19. Batch Owned Flags (оптимизация N+1)
Файл: internal/service/shop_service.go, applyOwnedFlags.

Почему: Вместо N запросов IsItemOwned делается один запрос ListOwnedItemIDs, и все Owned выставляются из map.

Код:

go
func (s *shopService) applyOwnedFlags(ctx context.Context, userID string, items []repository.Item) {
    ids, err := s.pgRepo.ListOwnedItemIDs(ctx, userID)
    if err != nil {
        return
    }
    owned := make(map[string]struct{}, len(ids))
    for _, id := range ids {
        owned[id] = struct{}{}
    }
    for i := range items {
        _, items[i].Owned = owned[items[i].ID]
    }
}
N+1 Fix / Batch Query — важная оптимизация.

20. Cache Invalidation (множественная)
Файл: internal/service/shop_service.go, PurchaseItem + CancelPurchase.

Код:

go
_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
_ = s.redisRepo.Delete(ctx, "shop:items:all:")
_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))
Множественная инвалидация — удаляются все связанные ключи.

21. Metrics Recording
Файл: internal/service/shop_service.go.

Код:

go
metrics.PurchaseRefundFailures.Inc()   // при провале refund
metrics.RecordOrder(float64(item.Price))  // при успешной покупке
Prometheus Metrics — метрики бизнес-логики.

22. Multi-Topic Kafka Producer
Файл: internal/app/di.go, initKafka.

Код:

go
a.di.kafkaProd = kafka.NewProducer(cfg, kafka.TopicPurchasePaid, a.log)
Kafka Producer — публикация в Kafka. Уникально для shop — другие сервисы используют только NATS.

23. Outbox → Kafka Bridge
Файл: internal/app/di.go, initOutbox.

Код:

go
outboxWorker := worker.NewOutboxWorker(a.di.db, a.di.js, func(ctx context.Context, payload []byte) error {
    // ...
    _, err := a.di.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{...})
    return err
})
Outbox → Compensate Func — outbox worker вызывает компенсацию через callback, а не публикует в NATS. Это Callback Pattern внутри Outbox.

24. Converter (односторонний)
Файл: internal/converter/item.go.

Код:

go
func ItemToProto(item model.Item) *pb.Item { /* ... */ }
func ItemsToProto(items []model.Item) []*pb.Item { /* ... */ }
25. Mock Object (Tests)
Файлы: shop_service_test.go, merch_gate_test.go.

Три мока:

mockShopStore — для ShopStore.

mockShopCache — для ShopCache.

stubBilling — для billingPb.BillingServiceClient.

stubPayment — для paymentPb.PaymentServiceClient.

Четыре мока — рекорд для shop.

26. Testcontainers
Файл: internal/repository/postgres_integration_test.go.

27. Build Tags
Файл: internal/repository/postgres_integration_test.go, //go:build integration.

28. Graceful Shutdown / Closer
Файл: internal/app/di.go.

29. NATS Consumer для синхронизации с inventory
Файл: internal/app/di.go, subscribeInventorySync.

Почему: Shop слушает inventory.item.created и создаёт зеркальную запись в shop.items. Это Event-Driven Sync.

Код:

go
_, err := js.Subscribe("inventory.item.created", func(msg *nats.Msg) {
    // ...
    if err := pg.CreateItemFromInventory(context.Background(), itemID, name, description, price, stock); err != nil {
        _ = msg.Nak()
        return
    }
    _ = msg.Ack()
}, nats.Durable("shop-inventory-sync"))
Event-Driven Sync — одна из архитектурных особенностей shop.

30. NATS Consumer для fulfillment
Файл: internal/app/di.go, subscribePurchaseFulfilled.

Код:

go
_, err := js.Subscribe(kafka.TopicPurchaseFulfilled, func(msg *nats.Msg) {
    fulfilled, err := events.UnmarshalPurchaseFulfilled(msg.Data)
    if err := pg.MarkPurchaseFulfilled(context.Background(), fulfilled.PurchaseUUID); err != nil {
        _ = msg.Nak()
        return
    }
    _ = msg.Ack()
}, nats.Durable("shop-purchase-fulfilled"))
Event-Driven Sync — shop узнаёт о fulfilled через NATS.

Итог по shop
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go
Port (Hexagonal) — 3 порта	service/shop_service.go
Facade	service/shop_service.go, shopService
Saga Pattern (Orchestration)	service/shop_service.go + postgres_repo.go
Compensation Action	service/shop_service.go
Idempotency через reference_id	service/shop_service.go
Spend Cycle	postgres_repo.go, NextSpendCycle
Outbox Pattern	postgres_repo.go, PurchaseItemWithStock
Outbox Worker + Compensation Handler	worker/outbox_worker.go
Saga Resume Worker	worker/outbox_worker.go, resumeSagas
Pessimistic + Optimistic Lock	postgres_repo.go, PurchaseItemWithStock
Refund Window / Soft Constraint	postgres_repo.go, RefundPurchase
Idempotent Refund	postgres_repo.go, RefundPurchase
FOR UPDATE SKIP LOCKED	worker/outbox_worker.go + postgres_repo.go
gRPC Client Dial + Graceful Close	app/di.go, initDomain
Optional Dependency (Payment)	app/di.go, initDomain
Cache-Aside	service/shop_service.go, GetItems
Batch Owned Flags (N+1 Fix)	service/shop_service.go, applyOwnedFlags
Cache Invalidation (множественная)	service/shop_service.go
Metrics Recording	service/shop_service.go
Kafka Producer	app/di.go, initKafka
Callback Pattern (Outbox → Compensate)	app/di.go, initOutbox
Converter	converter/item.go
Mock Object (×4)	service/*_test.go
Testcontainers	repository/postgres_integration_test.go
Build Tags	repository/postgres_integration_test.go
Graceful Shutdown	app/di.go
Event-Driven Sync (inventory)	app/di.go, subscribeInventorySync
Event-Driven Sync (fulfillment)	app/di.go, subscribePurchaseFulfilled
Итого 30 паттернов и приёмов — абсолютный рекорд среди всех сервисов.

Уникальные фичи shop:

Saga Pattern (Orchestration) — единственный сервис с явной сагой.

Spend Cycle — оригинальный приём для идемпотентности.

Saga Resume Worker — фоновое доведение саг.

Pessimistic + Optimistic Lock в одной транзакции.

Outbox → Compensate Callback.

Optional gRPC-клиент (Payment).

Готовый markdown для SHOP_PATTERNS.md
Допиши в 8.shop/SHOP_PATTERNS.md:

md
# shop — Programming Patterns

> Сервис: магазин (каталог, покупка, отмена, инвентарь).
> Postgres + Redis + NATS + Kafka + gRPC (billing, payment) + OTel.

## Что делает сервис

Каталог товаров, покупка за билетики, отмена с возвратом, инвентарь, продажи авторов. Покупка — это **Saga** между `shop` и `billing` с компенсацией при провале.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Port (Hexagonal) — 3 порта
- **Файл:** `internal/service/shop_service.go`
- **Порты:** `ShopStore`, `ShopCache`, gRPC-клиенты.

### 4. Facade — ShopService
- **Файл:** `internal/service/shop_service.go`

### 5. Saga Pattern (Orchestration)
- **Файлы:** `internal/service/shop_service.go`, `internal/repository/postgres_repo.go`
- **Суть:** сага управляется через таблицу `purchase_sagas`.
- **Состояния:** `started → spent → stocked → compensating → compensated`.

### 6. Compensation Action
- **Файл:** `internal/service/shop_service.go`, `PurchaseItem`
- **Суть:** при провале после списания — возврат билетиков.

### 7. Idempotency через reference_id
- **Файл:** `internal/service/shop_service.go`
- **Ключи:** `shop:u:i:c`, `refund:shop:u:i:c`, `cancel-{purchaseID}`.

### 8. Spend Cycle
- **Файл:** `internal/repository/postgres_repo.go`, `NextSpendCycle`
- **Суть:** таблица `purchase_claims` хранит цикл для пары `(user, item)`.

### 9. Outbox Pattern
- **Файл:** `internal/repository/postgres_repo.go`, `PurchaseItemWithStock`

### 10. Outbox Worker + Compensation Handler
- **Файл:** `internal/worker/outbox_worker.go`
- **Код:** `const CompensationEvent = "billing.refund"`

### 11. Saga Resume Worker
- **Файл:** `internal/worker/outbox_worker.go`, `resumeSagas`
- **Суть:** доводит зависшие саги через 30 секунд.

### 12. Pessimistic + Optimistic Lock
- **Файл:** `internal/repository/postgres_repo.go`, `PurchaseItemWithStock`
- **Код:** `SELECT ... FOR UPDATE` + `version = $2`.

### 13. Refund Window / Soft Constraint
- **Файл:** `internal/repository/postgres_repo.go`, `RefundPurchase`
- **Суть:** 7 дней + `fulfilled_at` для физических товаров.

### 14. Idempotent Refund
- **Файл:** `internal/repository/postgres_repo.go`, `RefundPurchase`

### 15. FOR UPDATE SKIP LOCKED
- **Файлы:** `internal/worker/outbox_worker.go`, `internal/repository/postgres_repo.go`

### 16. gRPC Client Dial + Graceful Close
- **Файл:** `internal/app/di.go`, `initDomain`

### 17. Optional Dependency (Payment)
- **Файл:** `internal/app/di.go`, `initDomain`
- **Суть:** payment-клиент может быть nil.

### 18. Cache-Aside
- **Файл:** `internal/service/shop_service.go`, `GetItems`

### 19. Batch Owned Flags (N+1 Fix)
- **Файл:** `internal/service/shop_service.go`, `applyOwnedFlags`

### 20. Cache Invalidation (множественная)
- **Файл:** `internal/service/shop_service.go`

### 21. Metrics Recording
- **Файл:** `internal/service/shop_service.go`

### 22. Kafka Producer
- **Файл:** `internal/app/di.go`, `initKafka`

### 23. Callback Pattern (Outbox → Compensate)
- **Файл:** `internal/app/di.go`, `initOutbox`

### 24. Converter
- **Файл:** `internal/converter/item.go`

### 25. Mock Object (×4)
- **Файлы:** `internal/service/*_test.go`

### 26. Testcontainers
- **Файл:** `internal/repository/postgres_integration_test.go`

### 27. Build Tags
- **Файл:** `internal/repository/postgres_integration_test.go`, `//go:build integration`

### 28. Graceful Shutdown
- **Файл:** `internal/app/di.go`

### 29. Event-Driven Sync (inventory)
- **Файл:** `internal/app/di.go`, `subscribeInventorySync`

### 30. Event-Driven Sync (fulfillment)
- **Файл:** `internal/app/di.go`, `subscribePurchaseFulfilled`