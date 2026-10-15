Разбор nats-hub
Что делает сервис
nats-hub — это bootstrap + observability для NATS:

Подключается к NATS.

Создаёт/обновляет stream EVENTS со списком subjects.

Подписывается на event.> и логирует каждое событие.

Экспортирует /health и /ready (проверка nc.IsConnected()).

Нет БД, нет gRPC, нет бизнес-логики. Только NATS.

Найденные паттерны
1. Manual DI (без контейнера)
Файл: internal/app/app.go, App.New.

Почему: Как в fulfillment и balancer, здесь нет diContainer. Всё собирается прямо в New:

Logger.

Closer.

NATS connection.

Stream config.

HTTP-сервер.

Код:

go
func New(ctx context.Context) (*App, error) {
    log := logger.NewFrom(level, format)
    a := &App{log: log, closer: closer.New(log)}

    nc, err := nats.Connect(natsURL)
    a.nc = nc
    a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })

    js, err := nc.JetStream()
    // ...
}
Manual DI — сервис маленький, контейнер не нужен.

2. Bootstrap / Initializer (главный паттерн)
Файл: internal/app/app.go, New.

Почему: nats-hub — это bootstrap-сервис. Его задача — создать stream EVENTS со всеми subjects, которые используют другие сервисы. Это гарантирует, что когда game публикует score.updated, stream уже существует.

Код:

go
streamConfig := &nats.StreamConfig{
    Name: "EVENTS",
    Subjects: []string{
        "event.>",
        "score.updated",
        "user.registered",
        "shop.purchased",
        "payment.completed",
        "purchase.paid",
        "purchase.fulfilled",
        "boost.started",
        "author.upserted",
        "inventory.item.created",
        // ...
    },
    Storage:  nats.FileStorage,
    MaxAge:   7 * 24 * time.Hour,
    MaxBytes: 1024 * 1024 * 1024,
}

if info, err := js.AddStream(streamConfig); err != nil {
    a.log.Warn("stream create (may exist)", "err", err)
    if _, uerr := js.UpdateStream(streamConfig); uerr != nil {
        a.log.Warn("stream update failed", "err", uerr)
    } else {
        a.log.Info("stream EVENTS subjects updated")
    }
    // ...
} else {
    a.log.Info("stream EVENTS created", "name", info.Config.Name)
}
Bootstrap / Initializer — сервис создаёт инфраструктуру. Уникально для nats-hub.

3. Idempotent Stream Creation (AddStream → UpdateStream)
Файл: internal/app/app.go.

Почему: js.AddStream падает, если stream уже существует. Тогда вызывается UpdateStream, чтобы добавить новые subjects. Это идемпотентно — можно запускать много раз.

Код:

go
if info, err := js.AddStream(streamConfig); err != nil {
    a.log.Warn("stream create (may exist)", "err", err)
    if _, uerr := js.UpdateStream(streamConfig); uerr != nil {
        a.log.Warn("stream update failed", "err", uerr)
    } else {
        a.log.Info("stream EVENTS subjects updated")
    }
    // ...
}
Idempotent Stream Creation — правильная работа с инфраструктурой.

4. Stream Configuration as Code
Файл: internal/app/app.go.

Почему: Stream config задаётся в коде, а не в конфиге. Это IaC (Infrastructure as Code) в миниатюре.

Код:

go
streamConfig := &nats.StreamConfig{
    Name: "EVENTS",
    Subjects: []string{
        "event.>",
        "score.updated",
        // ...
    },
    Storage:  nats.FileStorage,
    MaxAge:   7 * 24 * time.Hour,
    MaxBytes: 1024 * 1024 * 1024,
}
Stream Config as Code — вся конфигурация в одном месте.

Плюс: MaxAge (7 дней) + MaxBytes (1 ГБ) — retention policy.

5. Wildcard Subscription (event.>)
Файл: internal/app/app.go.

Почему: Подписка на все события через wildcard event.>:

go
js.Subscribe("event.>", func(msg *nats.Msg) {
    a.log.Info("event", "subject", msg.Subject, "bytes", len(msg.Data))
    _ = msg.Ack()
}, nats.Durable("nats-hub-listener"))
Wildcard Subscription — слушает всё, логирует, ack-ает.

6. Durable Consumer (как везде)
Код:

go
nats.Durable("nats-hub-listener")
Durable Consumer — как в profile, history, notification, analytics, leaderboard, billing, game.

7. Health Check с IsConnected()
Файл: internal/app/app.go.

Почему: В отличие от balancer, /health и /ready реально проверяют nc.IsConnected().

Код:

go
mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
    if nc.IsConnected() {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ok","service":"nats-hub"}`))
        return
    }
    w.WriteHeader(http.StatusServiceUnavailable)
    _, _ = w.Write([]byte(`{"status":"degraded","service":"nats-hub"}`))
})
mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
    if nc.IsConnected() {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ready","service":"nats-hub"}`))
        return
    }
    w.WriteHeader(http.StatusServiceUnavailable)
    _, _ = w.Write([]byte(`{"status":"not_ready","service":"nats-hub"}`))
})
Health Check с реальной проверкой — правильно, а не как в balancer.

8. Graceful Shutdown / Closer
Файл: internal/app/app.go.

Код:

go
a.closer.AddNamed("nats", func(context.Context) error {
    return nc.Drain()
})
a.closer.AddNamed("metrics http", func(ctx context.Context) error {
    return a.http.Shutdown(ctx)
})
Closer Registry — как в других сервисах.

9. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal.

Код:

go
if err := a.closer.WaitSignal(ctx); err != nil {
    return fmt.Errorf("shutdown: %w", err)
}
Context Cancellation — через closer.

10. Env-Based Config
Файл: internal/app/app.go, getenv.

Код:

go
func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}
Env-Based Config — простая конфигурация.

11. Logger + Level + Format
Файл: internal/app/app.go.

Код:

go
level := getenv("LOG_LEVEL", "info")
format := getenv("LOG_FORMAT", "text")
log := logger.NewFrom(level, format)
Configurable Logger — уровень и формат из ENV.

Итог по nats-hub
Паттерн	Где
Manual DI	app/app.go, App.New
Bootstrap / Initializer	app/app.go, New
Idempotent Stream Creation	app/app.go, AddStream → UpdateStream
Stream Configuration as Code	app/app.go, streamConfig
Wildcard Subscription	app/app.go, js.Subscribe("event.>")
Durable Consumer	app/app.go, nats.Durable("nats-hub-listener")
Health Check с IsConnected()	app/app.go
Graceful Shutdown / Closer	app/app.go
Context Cancellation	app/app.go, RunUntilSignal
Env-Based Config	app/app.go, getenv
Configurable Logger	app/app.go
Итого 11 паттернов и приёмов.

Уникальные фичи nats-hub:

Bootstrap / Initializer — единственный сервис, который создаёт инфраструктуру для других.

Idempotent Stream Creation — AddStream → UpdateStream.

Stream Config as Code — вся конфигурация в одном месте.

Реальный health check — через nc.IsConnected().

По сути nats-hub — это «миграция» для NATS: как migrator.Up для Postgres, только для JetStream stream.

Готовый markdown для NATS_PATTERNS.md
md
# nats-hub — Programming Patterns

> Сервис: bootstrap JetStream stream `EVENTS` + event logger.
> Только NATS + health HTTP.

## Что делает сервис

Создаёт/обновляет stream `EVENTS` со списком subjects (идемпотентно). Подписывается на `event.>` и логирует все события. Экспортирует `/health`, `/ready`.

## Паттерны

### 1. Manual DI (без контейнера)
- **Файл:** `internal/app/app.go`, `App.New`

### 2. Bootstrap / Initializer
- **Файл:** `internal/app/app.go`, `New`
- **Суть:** сервис создаёт инфраструктуру (stream) для других.

### 3. Idempotent Stream Creation
- **Файл:** `internal/app/app.go`
- **Код:**
  ```go
  if info, err := js.AddStream(streamConfig); err != nil {
      js.UpdateStream(streamConfig) // add new subjects
  }
4. Stream Configuration as Code
Файл: internal/app/app.go, streamConfig

Код:

go
streamConfig := &nats.StreamConfig{
    Name: "EVENTS",
    Subjects: []string{ "event.>", "score.updated", /* ... */ },
    Storage:  nats.FileStorage,
    MaxAge:   7 * 24 * time.Hour,
    MaxBytes: 1024 * 1024 * 1024,
}
5. Wildcard Subscription
Файл: internal/app/app.go

Код: js.Subscribe("event.>", ...)

6. Durable Consumer
Файл: internal/app/app.go

Код: nats.Durable("nats-hub-listener")

7. Health Check с IsConnected()
Файл: internal/app/app.go

Код:

go
if nc.IsConnected() {
    w.WriteHeader(http.StatusOK)
} else {
    w.WriteHeader(http.StatusServiceUnavailable)
}
8. Graceful Shutdown / Closer
Файл: internal/app/app.go

9. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal

10. Env-Based Config
Файл: internal/app/app.go, getenv

11. Configurable Logger
Файл: internal/app/app.go

Код: logger.NewFrom(level, format)