Разбор history
Что делает сервис
history — это журнал событий пользователя. Он:

Принимает события по gRPC (RecordEvent) или через NATS-подписки (IngestWorker).

Хранит их в Postgres (events таблица).

Позволяет читать (ListEvents) с фильтрами по userID и eventType.

Чистит старые записи (RetentionWorker) по retention policy (30 дней по умолчанию).

Найденные паттерны
1. DI Container (как в auth и profile)
Файл: internal/app/di.go, тип diContainer.

Код:

go
type diContainer struct {
    cfg    *config.Config
    dbpool *pgxpool.Pool
    js     nats.JetStreamContext
    nc     *nats.Conn
    repo   *repository.PostgresRepo
    svc    *service.HistoryService
    api    pb.HistoryServiceServer
}
Всё та же идея: зависимости в одном месте.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
for _, step := range []func(context.Context) error{
    a.initLogger, a.initCloser, a.initPostgres, a.initNATS,
    a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
} {
    if err := step(ctx); err != nil {
        return err
    }
}
Обрати внимание: initWorkers идёт перед initGRPC. Это важно — воркеры стартуют раньше, чтобы не потерять сообщения.

3. Repository (через Port / Interface)
Файл: internal/service/history_service.go, интерфейс HistoryStore.

Почему это интересно: ты назвал интерфейс HistoryStore, а не Repository. Это ближе к гексагональной архитектуре — Port (порт), через который сервис общается с внешним миром.

Код:

go
// HistoryStore is the persistence port for HistoryService.
type HistoryStore interface {
    Insert(ctx context.Context, userID, eventType, payload string) (string, error)
    List(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error)
    DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
Это Port (в терминах Hexagonal Architecture / Clean Architecture). Ты интуитивно назвал его Store, но по сути это тот же Repository, только с точки зрения порта.

4. Фасад (Facade) — упрощённый
Файл: internal/service/history_service.go, тип HistoryService.

Почему: HistoryService — это фасад над HistoryStore + retention policy. В отличие от auth/profile, здесь только одна зависимость, так что фасад тоньше, но всё равно есть.

Код:

go
type HistoryService struct {
    repo          HistoryStore
    retentionDays int
}
5. Worker (Background Worker)
Файл: internal/worker/workers.go, типы IngestWorker и RetentionWorker.

Почему: Два фоновых воркера, каждый со своим жизненным циклом:

IngestWorker — подписывается на NATS, пишет события в БД.

RetentionWorker — по тикеру (раз в час) чистит старые записи.

Код:

go
type IngestWorker struct {
    js  nats.JetStreamContext
    svc *service.HistoryService
}

type RetentionWorker struct {
    svc *service.HistoryService
}
Это Background Worker — паттерн, когда долгоживущая горутина делает фоновую работу. У тебя их два, каждый с методом Start(ctx).

6. Event-Driven Consumer (как в profile)
Файл: internal/worker/workers.go, IngestWorker.Start.

Почему: IngestWorker подписывается на 6 NATS-субжектов:

go
var subjects = []string{
    "payment.completed",
    "author.upserted",
    "shop.purchased",
    "score.updated",
    "user.registered",
    "event.user.registered",
}
Каждое событие превращается в запись в журнале истории. Это Event-Driven Consumer.

7. Durable Consumer (как в profile)
Файл: internal/worker/workers.go.

Код:

go
nats.Durable("history-" + s)  // history-payment.completed, history-score.updated, ...
У каждой подписки свой durable name. Это позволяет не потерять сообщения при перезапуске.

8. Manual Ack + Nak (как в profile, но с Nak)
Файл: internal/worker/workers.go, IngestWorker.

Почему: Ты используешь и Ack, и Nak:

go
if _, err := w.svc.RecordEvent(context.Background(), userID, eventType, string(msg.Data)); err != nil {
    log.Printf("history ingest %s: %v", s, err)
    _ = msg.Nak()   // ← явный отказ, сообщение вернётся
    return
}
_ = msg.Ack()       // ← успех
Nak — это Negative Acknowledgment, сигнал «сообщение не обработано, верни его в очередь». Это сильнее, чем просто «не Ack-ать», потому что NATS сразу узнаёт о провале и может переотправить сообщение быстрее.

Это Nak Pattern для обработки ошибок.

9. Scheduled Worker / Cron-like
Файл: internal/worker/workers.go, RetentionWorker.

Почему: RetentionWorker запускается по time.Ticker раз в час:

go
func (w *RetentionWorker) Start(ctx context.Context) {
    t := time.NewTicker(1 * time.Hour)
    defer t.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.C:
            n, err := w.svc.PurgeExpired(ctx)
            // ...
        }
    }
}
Это Scheduled Worker (или Cron Job внутри сервиса). Чистит старые события по retention policy.

10. Retention Policy / TTL (архитектурный приём)
Файл: internal/service/history_service.go, PurgeExpired.

Почему: Сервис сам себя ограничивает по объёму данных: удаляет записи старше retentionDays (по умолчанию 30). Это Data Retention Policy — важный приём для event-логов.

Код:

go
func (s *HistoryService) PurgeExpired(ctx context.Context) (int64, error) {
    before := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
    return s.repo.DeleteOlderThan(ctx, before)
}
11. Payload Normalization (архитектурный приём)
Файл: internal/service/history_service.go, RecordEvent.

Почему: Сервис нормализует payload перед записью:

Если пусто — {}.

Если невалидный JSON — оборачивает в {"raw": "..."}.

Код:

go
if payload == "" {
    payload = "{}"
}
if !json.Valid([]byte(payload)) {
    b, _ := json.Marshal(map[string]string{"raw": payload})
    payload = string(b)
}
Это Payload Sanitization — важно для event-логов, чтобы никогда не падать на кривом JSON.

12. Graceful Shutdown через Closer (как в auth)
Файл: internal/app/di.go.

Код:

go
a.closer.AddNamed("postgres", func(context.Context) error { db.Close(); return nil })
a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
a.closer.AddNamed("workers", func(context.Context) error { cancel(); return nil })
a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
Всё так же — реестр закрывашек. Важно: workers регистрируется до grpc, чтобы воркеры остановились первыми.

13. Context Cancellation для воркеров
Файл: internal/app/di.go, initWorkers.

Почему: Каждый воркер получает ctx, который отменяется через closer:

go
func (a *App) initWorkers(_ context.Context) error {
    ctx, cancel := context.WithCancel(context.Background())
    go worker.NewIngestWorker(a.di.js, a.di.svc).Start(ctx)
    go worker.NewRetentionWorker(a.di.svc).Start(ctx)
    a.closer.AddNamed("workers", func(context.Context) error { cancel(); return nil })
    return nil
}
Это Context Cancellation — стандартный Go-паттерн для остановки горутин.

14. Mock-based Unit Tests (паттерн тестирования)
Файл: internal/service/history_service_test.go.

Почему: У тебя есть mockHistory, который реализует HistoryStore. Это Mock Object — позволяет тестировать сервис без Postgres.

Код:

go
type mockHistory struct {
    lastPayload string
    // ...
}

func (m *mockHistory) Insert(_ context.Context, userID, eventType, payload string) (string, error) {
    m.lastUser, m.lastType, m.lastPayload = userID, eventType, payload
    // ...
}
15. Testcontainers-based Integration Tests
Файл: internal/repository/postgres_integration_test.go.

Почему: Используешь testcontainers-go — поднимаешь настоящий Postgres в докере для интеграционных тестов. Это Testcontainers Pattern.

Код:

go
req := testcontainers.ContainerRequest{
    Image:        "postgres:16-alpine",
    // ...
}
c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{...})
16. Build Tags для Integration Tests
Файл: internal/repository/postgres_integration_test.go.

Код:

go
//go:build integration
Это Build Tags — интеграционные тесты не запускаются при обычном go test ./.... Их запускают только с go test -tags=integration. Паттерн Test Separation.

Итог по history
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go, App.init
Port (Hexagonal)	service/history_service.go, HistoryStore
Facade (тонкий)	service/history_service.go, HistoryService
Background Worker	worker/workers.go, IngestWorker, RetentionWorker
Event-Driven Consumer	worker/workers.go, IngestWorker.Start
Durable Consumer	nats.Durable("history-" + s)
Manual Ack + Nak	worker/workers.go
Scheduled Worker / Cron	worker/workers.go, RetentionWorker
Data Retention Policy	service/history_service.go, PurgeExpired
Payload Sanitization	service/history_service.go, RecordEvent
Graceful Shutdown / Closer	app/di.go
Context Cancellation	app/di.go, initWorkers
Mock Object (Tests)	service/history_service_test.go
Testcontainers	repository/postgres_integration_test.go
Build Tags	repository/postgres_integration_test.go, //go:build integration
Итого 16 паттернов и приёмов. Тут больше, чем в auth и profile — за счёт воркеров, тестов и retention policy.

Ключевые отличия от profile:

Port вместо Repository (гексагональная архитектура).

Nak вместо простого «не Ack».

Scheduled Worker (RetentionWorker) — фоновый cron.

Retention Policy — самоочистка данных.

Testcontainers для интеграционных тестов.

Готовый markdown для HISTORY_PATTERNS.md
md
# history — Programming Patterns

> Сервис: журнал событий пользователя + retention policy.
> Postgres + NATS JetStream (ingest worker) + gRPC.

## Что делает сервис

Хранит события пользователя (event log). Принимает их по gRPC (`RecordEvent`) или через NATS-подписки (`IngestWorker`). Позволяет читать с фильтрами (`ListEvents`) и чистит старые записи (`RetentionWorker`) по retention policy.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`
- **Суть:** `initWorkers` до `initGRPC` — воркеры стартуют раньше.

### 3. Port (Hexagonal / Clean Architecture)
- **Файл:** `internal/service/history_service.go`, интерфейс `HistoryStore`
- **Суть:** сервис общается с внешним миром через порт, а не через репозиторий.
- **Код:**
  ```go
  // HistoryStore is the persistence port for HistoryService.
  type HistoryStore interface {
      Insert(ctx context.Context, userID, eventType, payload string) (string, error)
      List(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error)
      DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
  }
4. Facade (тонкий)
Файл: internal/service/history_service.go, HistoryService

5. Background Worker
Файл: internal/worker/workers.go, IngestWorker, RetentionWorker

Суть: два фоновых воркера с методом Start(ctx).

6. Event-Driven Consumer
Файл: internal/worker/workers.go, IngestWorker.Start

Суть: подписки на 6 NATS-субжектов.

7. Durable Consumer
Файл: internal/worker/workers.go

Суть: nats.Durable("history-" + s).

8. Manual Ack + Nak
Файл: internal/worker/workers.go

Суть: msg.Ack() при успехе, msg.Nak() при ошибке.

Код:

go
if _, err := w.svc.RecordEvent(...); err != nil {
    _ = msg.Nak()
    return
}
_ = msg.Ack()
9. Scheduled Worker / Cron
Файл: internal/worker/workers.go, RetentionWorker

Суть: time.NewTicker(1 * time.Hour).

10. Data Retention Policy
Файл: internal/service/history_service.go, PurgeExpired

Суть: удаление записей старше retentionDays.

11. Payload Sanitization
Файл: internal/service/history_service.go, RecordEvent

Суть: пусто → {}, невалидный JSON → {"raw": "..."}.

12. Graceful Shutdown / Closer
Файл: internal/app/di.go

13. Context Cancellation
Файл: internal/app/di.go, initWorkers

14. Mock Object (Tests)
Файл: internal/service/history_service_test.go, mockHistory

15. Testcontainers
Файл: internal/repository/postgres_integration_test.go

16. Build Tags
Файл: internal/repository/postgres_integration_test.go

Код: //go:build integration