Разбор analytics
Что делает сервис
analytics — это продуктовая аналитика. Две функции:

Ingest — слушает события из NATS (score.updated, user.registered и др.) и пишет их в ClickHouse (analytics_events таблица).

Query — считает метрики: DAU (Daily Active Users), MAU (Monthly Active Users), Retention (удержание по когортам).

gRPC API предоставляет и запись (RecordEvent), и чтение (GetDAU, GetMAU, GetRetention).

Ключевое отличие: это первый сервис с ClickHouse (аналитическая БД), и первый с SQL-запросами, собранными через fmt.Sprintf.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg    *config.Config
    ch     *ch.Client
    js     nats.JetStreamContext
    nc     *nats.Conn
    repo   *repository.AnalyticsRepo
    svc    *service.AnalyticsService
    api    pb.AnalyticsServiceServer
}
Всё так же — зависимости в одном месте.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
for _, step := range []func(context.Context) error{
    a.initLogger, a.initCloser, a.initClickHouse, a.initNATS,
    a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
} {
    // ...
}
Порядок: ClickHouse → NATS → Domain → Workers → gRPC. Логично.

3. Port (Hexagonal) — AnalyticsStore
Файл: internal/service/analytics_service.go, интерфейс AnalyticsStore.

Код:

go
// AnalyticsStore is the persistence port for AnalyticsService.
type AnalyticsStore interface {
    Record(ctx context.Context, userID, eventType, payload string) error
    DAU(ctx context.Context, days int) ([]model.DayCount, error)
    MAU(ctx context.Context, days int) (int64, error)
    Retention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error)
}
Тот же приём, что в history и notification: сервис общается с внешним миром через порт. Правда, здесь порт — аналитический, с методами DAU/MAU/Retention, а не CRUD.

4. Facade (тонкий) — AnalyticsService
Файл: internal/service/analytics_service.go.

Код:

go
type AnalyticsService struct {
    repo AnalyticsStore
}
Сервис — тонкий фасад над AnalyticsStore. Логика:

Валидация payload ({} если пусто).

Клэмп days для MAU (по умолчанию 30).

5. Event-Driven Consumer (как в history, notification)
Файл: internal/worker/ingest.go, IngestWorker.Start.

Код:

go
var subjects = []string{
    "payment.completed",
    "author.upserted",
    "shop.purchased",
    "score.updated",
    "user.registered",
}
Подписки на 5 событий. Каждое записывается в ClickHouse через RecordEvent.

6. Durable Consumer (как всегда)
Файл: internal/worker/ingest.go.

Код:

go
durable := "analytics-" + strings.ReplaceAll(s, ".", "-")
nats.Durable(durable)
7. Manual Ack + Nak
Файл: internal/worker/ingest.go.

Код:

go
if err := w.svc.RecordEvent(context.Background(), userID, s, string(msg.Data)); err != nil {
    log.Printf("analytics ingest %s: %v", s, err)
    _ = msg.Nak()
    return
}
_ = msg.Ack()
8. Adapter (×2) — над ClickHouse
Файлы:

internal/repository/clickhouse/client.go — обёртка над ClickHouse-клиентом.

internal/repository/repo.go — обёртка над ch.Client для аналитики.

Почему: ch.Client оборачивает HTTP-запросы к ClickHouse, AnalyticsRepo — доменные методы (DAU, MAU, Retention) над ch.Client. Это два уровня адаптера.

Код (ch.Client):

go
type Client struct {
    url string
    db  string
    http *http.Client
}

func (c *Client) InsertEvent(ctx context.Context, userID, eventType, payload string, at time.Time) error {
    // HTTP POST в ClickHouse
}

func (c *Client) QueryTSV(ctx context.Context, sql string, params map[string]string) (string, error) {
    // HTTP GET в ClickHouse с параметрами
}
9. Repository — AnalyticsRepo
Файл: internal/repository/repo.go.

Почему: AnalyticsRepo скрывает ClickHouse за доменными методами (DAU, MAU, Retention). Это классический Repository, только с ClickHouse вместо Postgres.

10. Query Builder через fmt.Sprintf (анти-паттерн, но рабочий)
Файл: internal/repository/repo.go, DAU, MAU, Retention.

Почему: SQL собирается через fmt.Sprintf, а параметры передаются через ClickHouse-параметры ({days:Int32}). Это безопаснее, чем чистая конкатенация, но всё равно хрупко.

Код:

go
sql := fmt.Sprintf(`
    SELECT toString(event_date), uniqExact(user_id)
    FROM %s.analytics_events
    WHERE event_date >= today() - {days:Int32} AND user_id != ''
    GROUP BY event_date
    ORDER BY event_date`, r.db)
%s подставляется r.db (имя базы). Это не user input, поэтому безопасно. Но если r.db вдруг станет user input — будет SQL injection.

Название паттерна: Parameterized Query (через ClickHouse-параметры {days:Int32}) + Template SQL (через fmt.Sprintf).

11. Cohort Analysis / Retention Calculation
Файл: internal/repository/repo.go, Retention.

Почему: Это не паттерн из GoF, но важный аналитический приём: Cohort Analysis. Ты считаешь retention:

Берёшь когорту пользователей, зарегистрировавшихся cohortDaysAgo назад.

Считаешь cohortSize — сколько их было.

Смотришь, сколько из них активны на day_n (0..windowDays).

Считаешь rate = active / cohortSize.

Код:

go
sizeSQL := fmt.Sprintf(`
    SELECT uniqExact(user_id) FROM %s.analytics_events
    WHERE event_date = toDate({cohort_day:String}) AND user_id != ''`, r.db)
// ...
pointsSQL := fmt.Sprintf(`
    SELECT dateDiff('day', toDate({cohort_day:String}), event_date) AS day_n, uniqExact(user_id)
    FROM %s.analytics_events
    WHERE user_id IN (
        SELECT user_id FROM %s.analytics_events
        WHERE event_date = toDate({cohort_day:String}) AND user_id != ''
    )
    AND event_date BETWEEN toDate({cohort_day:String}) AND toDate({cohort_day:String}) + {window_days:Int32}
    GROUP BY day_n
    ORDER BY day_n`, r.db, r.db)
Это Cohort Retention Analysis — классика продуктовой аналитики.

12. Graceful Shutdown / Closer
Файл: internal/app/di.go.

13. Mock Object (Tests)
Файл: internal/service/analytics_service_test.go, mockAnalytics.

Код:

go
type mockAnalytics struct {
    lastPayload string
    lastType    string
    lastDays    int
    dau         []model.DayCount
    mau         int64
    ret         *model.Retention
    err         error
}
14. Testcontainers (для ClickHouse)
Файл: internal/repository/clickhouse/client_integration_test.go — судя по имени, там Testcontainers для ClickHouse.

Паттерн: Testcontainers.

Итог по analytics
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go, App.init
Port (Hexagonal)	service/analytics_service.go, AnalyticsStore
Facade (тонкий)	service/analytics_service.go, AnalyticsService
Event-Driven Consumer	worker/ingest.go, IngestWorker
Durable Consumer	worker/ingest.go
Manual Ack + Nak	worker/ingest.go
Adapter (×2) над ClickHouse	repository/clickhouse/client.go, repository/repo.go
Repository (ClickHouse)	repository/repo.go, AnalyticsRepo
Parameterized Query + Template SQL	repository/repo.go
Cohort Retention Analysis	repository/repo.go, Retention
Graceful Shutdown	app/di.go
Mock Object	service/analytics_service_test.go
Testcontainers	repository/clickhouse/client_integration_test.go
Итого 14 паттернов и приёмов.

Ключевые отличия от history / notification:

ClickHouse вместо Postgres (аналитическая БД).

Cohort Retention Analysis — аналитический приём.

Parameterized Query через ClickHouse-параметры.

Два уровня адаптера над ClickHouse.

Готовый markdown для ANALYTICS_PATTERNS.md
md
# analytics — Programming Patterns

> Сервис: продуктовая аналитика (DAU / MAU / Retention).
> ClickHouse + NATS JetStream (ingest worker) + gRPC.

## Что делает сервис

1. **Ingest** — слушает события NATS и пишет их в ClickHouse.
2. **Query** — считает метрики: DAU, MAU, Retention (cohort analysis).

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Port (Hexagonal)
- **Файл:** `internal/service/analytics_service.go`, `AnalyticsStore`
- **Код:**
  ```go
  type AnalyticsStore interface {
      Record(ctx context.Context, userID, eventType, payload string) error
      DAU(ctx context.Context, days int) ([]model.DayCount, error)
      MAU(ctx context.Context, days int) (int64, error)
      Retention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error)
  }
4. Facade (тонкий)
Файл: internal/service/analytics_service.go, AnalyticsService

5. Event-Driven Consumer
Файл: internal/worker/ingest.go, IngestWorker

6. Durable Consumer
Файл: internal/worker/ingest.go

Код: nats.Durable("analytics-" + strings.ReplaceAll(s, ".", "-"))

7. Manual Ack + Nak
Файл: internal/worker/ingest.go

8. Adapter (×2) над ClickHouse
Файлы: internal/repository/clickhouse/client.go, internal/repository/repo.go

9. Repository (ClickHouse)
Файл: internal/repository/repo.go, AnalyticsRepo

10. Parameterized Query + Template SQL
Файл: internal/repository/repo.go

Код:

go
sql := fmt.Sprintf(`
    SELECT toString(event_date), uniqExact(user_id)
    FROM %s.analytics_events
    WHERE event_date >= today() - {days:Int32} AND user_id != ''`, r.db)
11. Cohort Retention Analysis
Файл: internal/repository/repo.go, Retention

Суть: когорта → size → активные на day_n → rate.

12. Graceful Shutdown
Файл: internal/app/di.go

13. Mock Object
Файл: internal/service/analytics_service_test.go, mockAnalytics

14. Testcontainers
Файл: internal/repository/clickhouse/client_integration_test.go