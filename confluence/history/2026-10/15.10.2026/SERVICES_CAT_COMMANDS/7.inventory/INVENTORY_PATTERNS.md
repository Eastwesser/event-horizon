Разбор inventory
Что делает сервис
inventory — это каталог товаров (мерч, скины, карточки). Особенности:

Два драйвера БД — Postgres и MongoDB (переключается через cfg.Driver).

Cache-Aside через декоратор — CachedRepository оборачивает любой InventoryRepository.

Outbox Pattern — для Postgres: событие и товар пишутся в одной транзакции.

Optimistic Locking — через version в Postgres.

Резервирование / возврат stock — ReserveItem, ReleaseItem.

Soft Delete / Restore — мягкое удаление.

Role-based Interceptors — RequireRoles для разных методов.

Tracer OTel — как в leaderboard.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg *config.Config
    db          *sql.DB
    mongoClient *mongo.Client
    nc          *nats.Conn
    js          nats.JetStreamContext
    baseRepo    repository.InventoryRepository
    cache       *repository.RedisCacheRepo
    repo        repository.InventoryRepository
    outbox      repository.ItemOutboxWriter
    svc         *service.InventoryService
    api         pb.InventoryServiceServer
    workerCancel context.CancelFunc
}
Тут две репозитория — baseRepo (сырой) и repo (с кешем). Это важно.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer, a.initRepository,
    a.initNATS, a.initDomain, a.initGRPC, a.initMetricsHTTP,
    a.initOutboxWorker,
}
initOutboxWorker — последний, после gRPC. Логично.

3. Strategy (Factory) — выбор драйвера БД
Файл: internal/app/di.go, initRepository.

Почему: Ты выбираешь драйвер через cfg.Driver — "postgres" или "mongo". Каждый драйвер возвращает свой InventoryRepository. Это Strategy (взаимозаменяемые алгоритмы) + Factory (выбор по имени).

Код:

go
switch a.cfg.Driver {
case "postgres":
    repo, db, err := a.connectPostgres()
    a.di.baseRepo = repo
    // ...
case "mongo":
    repo, client, err := a.connectMongo(ctx)
    a.di.baseRepo = repo
    // ...
}
Уникально для inventory — ни один другой сервис не поддерживает два драйвера.

4. Port (Hexagonal) — InventoryRepository
Файл: internal/repository/repository.go, интерфейс InventoryRepository.

Код:

go
type InventoryRepository interface {
    CreateItem(ctx context.Context, item *model.Item) error
    GetItem(ctx context.Context, id string) (*model.Item, error)
    UpdateItem(ctx context.Context, item *model.Item) error
    DeleteItem(ctx context.Context, id string) error
    SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error)
    // ...
}
Тот же приём, что в history/analytics.

5. Декоратор (Decorator) — CachedRepository
Файл: internal/repository/cached_repo.go.

Почему: CachedRepository оборачивает любой InventoryRepository и добавляет кеширование. Он реализует тот же интерфейс InventoryRepository.

Код:

go
type CachedRepository struct {
    next  InventoryRepository
    cache *RedisCacheRepo
}

func NewCachedRepository(next InventoryRepository, cache *RedisCacheRepo) InventoryRepository {
    if cache == nil || next == nil {
        return next
    }
    return &CachedRepository{next: next, cache: cache}
}
Это чистый Декоратор — добавляет поведение (кеш) без изменения интерфейса. Ключевой паттерн inventory.

6. Cache-Aside (через декоратор)
Файл: internal/repository/cached_repo.go, GetItem.

Код:

go
func (c *CachedRepository) GetItem(ctx context.Context, id string) (*model.Item, error) {
    if item, err := c.cache.GetItem(ctx, id); err == nil {
        return item, nil  // ← cache hit
    }
    item, err := c.next.GetItem(ctx, id)  // ← cache miss
    if err != nil {
        return nil, err
    }
    if item != nil {
        _ = c.cache.SetItem(ctx, item)  // ← warm cache
    }
    return item, nil
}
Тот же Cache-Aside, но реализованный через декоратор. Более чистая архитектура, чем в auth/profile.

7. Write-Through Cache (через декоратор)
Файл: internal/repository/cached_repo.go, CreateItem, UpdateItem.

Код:

go
func (c *CachedRepository) CreateItem(ctx context.Context, item *model.Item) error {
    if err := c.next.CreateItem(ctx, item); err != nil {
        return err
    }
    c.warmItem(ctx, item)      // ← write-through
    c.invalidateSearch(ctx)    // ← invalidation
    return nil
}
Write-Through Cache + Invalidation — после записи в БД сразу пишем в кеш и инвалидируем поиск.

8. Cache Invalidation (поиск)
Файл: internal/repository/cached_repo.go, invalidateSearch.

Код:

go
func (c *CachedRepository) invalidateSearch(ctx context.Context) {
    _ = c.cache.InvalidateSearchCache(ctx)
}
При любом изменении — все кеши поиска инвалидируются. Радикально, но надёжно.

9. Outbox Pattern
Файл: internal/repository/outbox.go + internal/repository/postgres_repo.go, CreateItemWithOutbox.

Почему: Товар и событие пишутся в одной транзакции в БД. Потом outbox worker публикует событие в NATS.

Код (интерфейс):

go
type ItemOutboxWriter interface {
    CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error
}
Код (реализация):

go
func (r *PostgresRepo) CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    // INSERT INTO inventory_items ...
    // INSERT INTO outbox ...
    return tx.Commit()
}
Outbox Pattern — гарантирует, что событие не потеряется, даже если NATS временно недоступен. Уникально для inventory (в других сервисах outbox был, но не всегда с интерфейсом).

10. Outbox Worker
Файл: internal/worker/outbox_worker.go.

Почему: Воркер раз в секунду читает необработанные события из таблицы outbox, публикует их в NATS и помечает как processed.

Код:

go
func (w *OutboxWorker) processBatch(ctx context.Context) {
    rows, err := w.db.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox 
        WHERE processed = false 
        ORDER BY created_at 
        LIMIT 100
    `)
    // ...
    for rows.Next() {
        // Публикуем в NATS
        w.js.Publish(eventType, payload)
        // Помечаем как обработанное
        w.db.ExecContext(ctx, `UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1`, id)
    }
}
Это Polling Publisher — один из двух вариантов реализации Outbox (второй — CDC через Debezium).

11. Optimistic Locking (version)
Файл: internal/repository/postgres_repo.go, UpdateItem.

Почему: Обновление товара проверяет version = $11 в WHERE. Если версия изменилась — конфликт.

Код:

go
query := `
    UPDATE inventory_items SET
        ..., version = version + 1
    WHERE id = $10 AND deleted_at IS NULL AND version = $11
`
res, err := r.db.ExecContext(ctx, query, ..., expected)
n, _ := res.RowsAffected()
if n == 0 {
    return model.ErrVersionConflict
}
Optimistic Locking — стандартный паттерн для конкурентных обновлений.

12. Atomic Conditional Update (ReserveItem)
Файл: internal/repository/postgres_repo.go, ReserveItem.

Почему: Уменьшение стока делается одним атомарным UPDATE с условием stock >= $1.

Код:

go
err := r.db.QueryRowContext(ctx, `
    UPDATE inventory_items
    SET stock = stock - $1, version = version + 1, updated_at = NOW()
    WHERE id = $2 AND deleted_at IS NULL AND stock >= $1
    RETURNING stock
`, quantity, id).Scan(&remaining)
Atomic Conditional Update — защита от race condition без явных транзакций.

13. Compensation Action (ReleaseItem)
Файл: internal/repository/postgres_repo.go, ReleaseItem.

Почему: ReleaseItem — обратная операция к ReserveItem. Это Compensation (компенсация) в терминах Saga. Если покупка не удалась — возвращаем сток.

Код:

go
UPDATE inventory_items
SET stock = stock + $1, version = version + 1, updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL
14. Soft Delete / Restore
Файл: internal/repository/postgres_repo.go, SoftDeleteItem и RestoreItem.

Почему: Вместо удаления — ставим deleted_at = NOW(). Restore — deleted_at = NULL.

Код:

go
func (r *PostgresRepo) SoftDeleteItem(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, `UPDATE inventory_items SET deleted_at = NOW() WHERE id = $1`, id)
    return err
}
func (r *PostgresRepo) RestoreItem(ctx context.Context, id string) error {
    result, err := r.db.ExecContext(ctx, `
        UPDATE inventory_items SET deleted_at = NULL, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NOT NULL
    `, id)
    // ...
}
15. Filter Builder (toPostgresFilter / toMongoFilter)
Файл: internal/repository/postgres_repo.go, toPostgresFilter.

Почему: Фильтры превращаются в SQL-условия с placeholder'ами. Никакой конкатенации user input.

Код:

go
func toPostgresFilter(filters map[string]interface{}) ([]string, []interface{}, int) {
    // ...
    if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
        conditions = append(conditions, fmt.Sprintf("author_id = $%d", argCount))
        args = append(args, authorID)
        argCount++
    }
    // ...
}
Это Query Builder + Parameterized Query. Безопасно.

16. Role-Based Interceptors (RequireRoles)
Файл: internal/interceptor/role.go (ты его не скинул, но он используется в initGRPC).

Код:

go
grpc.ChainUnaryInterceptor(
    interceptor.Recovery(),
    interceptor.Logger(),
    interceptor.Validate(),
    interceptor.RequireRoles(
        []string{"author", "admin"},
        []string{"/CreateItem", "/CreateItems", "/UpdateItem", "/DeleteItem", /* ... */},
    ),
    interceptor.RequireRoles(
        []string{"user", "author", "admin"},
        []string{"/ReserveItem", "/ReleaseItem"},
    ),
    metrics.UnaryServerInterceptor("inventory"),
)
Role-Based Access Control (RBAC) через интерсепторы. Уникально для inventory — самый сложный RBAC из всех сервисов.

17. Health Check + gRPC Health
Файл: internal/app/di.go, initGRPC.

Код:

go
healthServer := health.NewServer()
grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
gRPC Health Check — стандартный протокол для проверки здоровья. Уникально для inventory.

18. Multi-Driver Health Check
Файл: internal/app/di.go, initMetricsHTTP.

Код:

go
mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
    if cfg.Driver == "postgres" && db != nil {
        if err := db.Ping(); err != nil { /* degraded */ }
    }
    if cfg.Driver == "mongo" && mongoClient != nil {
        if err := mongoClient.Ping(pingCtx, nil); err != nil { /* degraded */ }
    }
    // ...
})
Multi-Driver Health Check — health check учитывает, какой драйвер используется.

19. MongoDB Change Stream (WatchChanges)
Файл: internal/repository/mongo_repo.go, WatchChanges.

Почему: Для MongoDB реализован альтернативный Outbox — через Change Streams. Слушает изменения коллекции и публикует их в NATS.

Код:

go
func (r *MongoRepo) WatchChanges(ctx context.Context, js nats.JetStreamContext) error {
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: bson.D{
            {Key: "operationType", Value: bson.D{{Key: "$in", Value: []string{"insert", "update"}}}},
        }}},
    }
    stream, err := r.collection.Watch(ctx, pipeline)
    // ...
    for stream.Next(ctx) {
        // Публикуем в NATS
        js.Publish("inventory.item.changed", payload)
    }
}
Это Change Data Capture (CDC) через Change Streams. Уникально для inventory.

20. MongoDB Transaction (ReserveItem)
Файл: internal/repository/mongo_repo.go, ReserveItem.

Почему: В MongoDB ReserveItem использует транзакцию через session.WithTransaction. Это атомарная операция с проверкой стока и записью в историю.

Код:

go
_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
    // 1. Проверяем наличие и текущий остаток
    // 2. Уменьшаем остаток
    // 3. Получаем новый остаток
    // 4. Создаём запись в истории
    return nil, nil
})
MongoDB Transaction — ACID-транзакции в MongoDB. Уникально для inventory.

21. MongoDB Aggregation Pipeline (GetStats)
Файл: internal/repository/mongo_repo.go, GetStats.

Почему: Статистика считается через Aggregation Pipeline MongoDB:

$match (фильтр)

$group (группировка)

$sort (сортировка)

$limit (лимит)

$project (проекция)

Код:

go
pipeline := mongo.Pipeline{
    {{Key: "$match", Value: bson.M{"deleted_at": nil}}},
    {{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$type"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
    }}},
}
cursor, err := r.collection.Aggregate(ctx, pipeline)
Это Aggregation Pipeline — стандартный приём MongoDB для аналитики.

22. MongoDB $lookup (JOIN) (GetItemWithAuthor)
Файл: internal/repository/mongo_repo.go, GetItemWithAuthor.

Код:

go
pipeline := mongo.Pipeline{
    {{Key: "$match", Value: bson.M{"id": itemID, "deleted_at": nil}}},
    {{Key: "$lookup", Value: bson.D{
        {Key: "from", Value: "authors"},
        {Key: "localField", Value: "author_id"},
        {Key: "foreignField", Value: "id"},
        {Key: "as", Value: "author"},
    }}},
    {{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$author"}}}},
    // ...
}
$lookup — это JOIN в MongoDB. Уникально для inventory.

23. Converter (двусторонний)
Файл: internal/converter/item.go.

Код:

go
func ItemFromCreateRequest(req *pb.CreateItemRequest) *model.Item { /* ... */ }
func ItemFromUpdateRequest(req *pb.UpdateItemRequest) *model.Item { /* ... */ }
func ItemToProto(item *model.Item) (*pb.Item, error) { /* ... */ }
Три функции — два направления.

24. Mock Object (Tests)
Файл: internal/service/inventory_service_test.go, memRepo + stubOutbox.

Код:

go
type memRepo struct {
    items        map[string]*model.Item
    reserveCalls int
    // ...
}
Два мока: memRepo для репозитория, stubOutbox для outbox.

25. Testcontainers
Файл: internal/repository/postgres_integration_test.go, mongo_repo_test.go.

Код:

go
req := testcontainers.ContainerRequest{
    Image:        "postgres:16-alpine",
    // ...
}
26. Build Tags
Файл: internal/repository/mongo_repo_test.go — //go:build integration.

27. Graceful Shutdown / Closer
Файл: internal/app/di.go.

28. Uniqueness — Outbox через интерфейс
Файл: internal/repository/outbox.go.

Почему: ItemOutboxWriter — отдельный интерфейс для outbox. Это позволяет:

Postgres-репозиторий реализует InventoryRepository + ItemOutboxWriter.

MongoDB-репозиторий реализует только InventoryRepository.

Код:

go
type ItemOutboxWriter interface {
    CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error
}
И потом в di.go:

go
if ow, ok := a.di.repo.(repository.ItemOutboxWriter); ok {
    a.di.outbox = ow
}
Это Interface Segregation + Type Assertion.

Итог по inventory
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go, App.init
Strategy (Factory) — выбор драйвера БД	app/di.go, initRepository
Port (Hexagonal)	repository/repository.go, InventoryRepository
Decorator	repository/cached_repo.go, CachedRepository
Cache-Aside	repository/cached_repo.go, GetItem
Write-Through Cache	repository/cached_repo.go, CreateItem / UpdateItem
Cache Invalidation	repository/cached_repo.go, invalidateSearch
Outbox Pattern	repository/outbox.go + postgres_repo.go, CreateItemWithOutbox
Outbox Worker (Polling Publisher)	worker/outbox_worker.go
Optimistic Locking (version)	repository/postgres_repo.go, UpdateItem
Atomic Conditional Update	repository/postgres_repo.go, ReserveItem
Compensation Action	repository/postgres_repo.go, ReleaseItem
Soft Delete / Restore	repository/postgres_repo.go
Filter Builder / Parameterized Query	repository/postgres_repo.go, toPostgresFilter
Role-Based Interceptors (RBAC)	app/di.go, RequireRoles
gRPC Health Check	app/di.go, initGRPC
Multi-Driver Health Check	app/di.go, initMetricsHTTP
MongoDB Change Stream (CDC)	repository/mongo_repo.go, WatchChanges
MongoDB Transaction	repository/mongo_repo.go, ReserveItem
MongoDB Aggregation Pipeline	repository/mongo_repo.go, GetStats
MongoDB $lookup (JOIN)	repository/mongo_repo.go, GetItemWithAuthor
Converter (двусторонний)	converter/item.go
Mock Object	service/inventory_service_test.go, memRepo, stubOutbox
Testcontainers	repository/postgres_integration_test.go
Build Tags	repository/mongo_repo_test.go
Graceful Shutdown	app/di.go
Interface Segregation + Type Assertion	repository/outbox.go + app/di.go
Итого 28 паттернов и приёмов — рекорд среди всех сервисов. Тут почти весь каталог GoF + микросервисные паттерны.

Уникальные фичи:

Два драйвера БД (Strategy).

Decorator для кеша — самый чистый Cache-Aside.

Outbox + Polling Publisher.

Change Streams (CDC) для MongoDB.

MongoDB Transaction + Aggregation + $lookup.

RBAC через интерсепторы.

Готовый markdown для INVENTORY_PATTERNS.md
Ты уже начал заполнять папку 7.inventory, так что допиши этот markdown туда. Я его подсократил — если нужно что-то подробнее, скажи.

md
# inventory — Programming Patterns

> Сервис: каталог товаров (мерч, скины, карточки).
> Postgres ИЛИ MongoDB + Redis + NATS + gRPC + OTel.

## Что делает сервис

Хранит товары с гибкими атрибутами (JSONB в Postgres, документы в MongoDB). Поддерживает CRUD, поиск, резервирование стока, soft delete. Два драйвера БД переключаются через `cfg.Driver`. Использует Outbox Pattern для публикации событий.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Strategy (Factory) — выбор драйвера БД
- **Файл:** `internal/app/di.go`, `initRepository`
- **Код:**
  ```go
  switch a.cfg.Driver {
  case "postgres": // ...
  case "mongo": // ...
  }
4. Port (Hexagonal)
Файл: internal/repository/repository.go, InventoryRepository

5. Decorator — CachedRepository
Файл: internal/repository/cached_repo.go

Суть: оборачивает любой InventoryRepository и добавляет кеширование.

Код:

go
func NewCachedRepository(next InventoryRepository, cache *RedisCacheRepo) InventoryRepository {
    if cache == nil || next == nil { return next }
    return &CachedRepository{next: next, cache: cache}
}
6. Cache-Aside
Файл: internal/repository/cached_repo.go, GetItem

7. Write-Through Cache
Файл: internal/repository/cached_repo.go, CreateItem / UpdateItem

8. Cache Invalidation
Файл: internal/repository/cached_repo.go, invalidateSearch

9. Outbox Pattern
Файлы: internal/repository/outbox.go, internal/repository/postgres_repo.go (CreateItemWithOutbox)

10. Outbox Worker (Polling Publisher)
Файл: internal/worker/outbox_worker.go

11. Optimistic Locking (version)
Файл: internal/repository/postgres_repo.go, UpdateItem

12. Atomic Conditional Update
Файл: internal/repository/postgres_repo.go, ReserveItem

13. Compensation Action
Файл: internal/repository/postgres_repo.go, ReleaseItem

14. Soft Delete / Restore
Файл: internal/repository/postgres_repo.go

15. Filter Builder / Parameterized Query
Файл: internal/repository/postgres_repo.go, toPostgresFilter

16. RBAC через интерсепторы
Файл: internal/app/di.go, interceptor.RequireRoles

17. gRPC Health Check
Файл: internal/app/di.go, initGRPC

18. Multi-Driver Health Check
Файл: internal/app/di.go, initMetricsHTTP

19. MongoDB Change Stream (CDC)
Файл: internal/repository/mongo_repo.go, WatchChanges

20. MongoDB Transaction
Файл: internal/repository/mongo_repo.go, ReserveItem

21. MongoDB Aggregation Pipeline
Файл: internal/repository/mongo_repo.go, GetStats

22. MongoDB $lookup (JOIN)
Файл: internal/repository/mongo_repo.go, GetItemWithAuthor

23. Converter (двусторонний)
Файл: internal/converter/item.go

24. Mock Object
Файл: internal/service/inventory_service_test.go, memRepo, stubOutbox

25. Testcontainers
Файлы: internal/repository/postgres_integration_test.go, mongo_repo_test.go

26. Build Tags
Файл: internal/repository/mongo_repo_test.go, //go:build integration

27. Graceful Shutdown
Файл: internal/app/di.go

28. Interface Segregation + Type Assertion
Файлы: internal/repository/outbox.go + internal/app/di.go

Код:

go
if ow, ok := a.di.repo.(repository.ItemOutboxWriter); ok {
    a.di.outbox = ow
}