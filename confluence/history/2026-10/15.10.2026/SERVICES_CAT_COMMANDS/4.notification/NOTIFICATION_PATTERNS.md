Разбор notification
Что делает сервис
notification — это два разных механизма уведомлений в одном сервисе:

InboxService — «внутренний ящик» уведомлений. Хранит уведомления в Postgres (notifications таблица), показывает пользователю список, позволяет отметить как прочитанное. Реагирует на события через NATS JetStream (InboxConsumer).

Notifier — внешние уведомления в Telegram-бот. Реагирует на события через Kafka (PurchasePaid, PurchaseFulfilled).

Плюс gRPC API для чтения inbox.

Ключевая особенность: сервис работает с двумя очередями одновременно (NATS + Kafka) — с fallback-логикой (если js == nil, используем Kafka).

Найденные паттерны
1. DI Container (как всегда)
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg      *config.Config
    dbpool   *pgxpool.Pool
    js       nats.JetStreamContext
    nc       *nats.Conn
    repo     *repository.PostgresRepo
    inbox    *service.InboxService
    notifier *service.Notifier
    api      pb.NotificationServiceServer
    consumer kafka.Consumer
}
Обрати внимание: тут два сервиса (inbox и notifier) и два консьюмера (js для NATS, consumer для Kafka).

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
for _, step := range []func(context.Context) error{
    a.initLogger, a.initCloser, a.initPostgres, a.initNATS,
    a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
} {
    // ...
}
Порядок: NATS → Domain → Workers → gRPC → Metrics. Workers стартуют до gRPC — это правильно.

3. Два независимых сервиса в одном процессе (Separation of Concerns)
Файлы: internal/service/inbox.go, internal/service/notify.go.

Почему это интересно: у тебя два независимых сервиса в одном приложении:

InboxService — внутренние уведомления (БД, gRPC).

Notifier — внешние уведомления (Telegram HTTP).

У каждого — свой интерфейс, свои зависимости, свой потребитель событий. Это Separation of Concerns — каждый сервис отвечает за одну задачу.

Код:

go
// InboxService — внутренний ящик
type InboxService struct {
    repo InboxStore
    log  *slog.Logger
}

// Notifier — внешние уведомления (Telegram)
type Notifier struct {
    log    *slog.Logger
    token  string
    chatID string
    client *http.Client
}
Это не паттерн из GoF, но важный архитектурный приём: два bounded context в одном процессе.

4. Port (Hexagonal) — InboxStore
Файл: internal/service/inbox.go, интерфейс InboxStore.

Код:

go
type InboxStore interface {
    Insert(ctx context.Context, n *model.Notification) (string, error)
    List(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*model.Notification, int64, int64, error)
    MarkRead(ctx context.Context, userID, notificationID string) (int64, error)
}
Тот же приём, что и в history — сервис общается с внешним миром через порт.

5. Facade (×2) — InboxService и Notifier
Файлы: internal/service/inbox.go, internal/service/notify.go.

Почему: InboxService — фасад над InboxStore + бизнес-логика обработки событий (HandleSubmitted, HandleApproved, HandleRecordBeaten). Notifier — фасад над http.Client + Telegram API.

Код (InboxService):

go
func (s *InboxService) HandleSubmitted(ctx context.Context, applicationID, applicantID, displayName string, adminIDs []string) error {
    // ... формирует title, body, link, src, вызывает s.repo.Insert
}
Код (Notifier):

go
func (n *Notifier) Handle(ctx context.Context, msg kafka.Message) error {
    switch msg.Topic {
    case kafka.TopicPurchasePaid:
        return n.notify(ctx, ...)
    case kafka.TopicPurchaseFulfilled:
        return n.notify(ctx, ...)
    }
}
6. Event-Driven Consumer (×2 — NATS и Kafka)
Файлы:

internal/worker/inbox_consumer.go — NATS.

internal/app/di.go, consumePurchaseNATS — NATS для Notifier.

internal/app/di.go, initDomain — Kafka-консьюмер.

Почему: У тебя два способа получать события:

NATS JetStream — для внутренних событий (author.application.submitted, event.leaderboard.record_beaten).

Kafka — для внешних событий (purchase.paid, purchase.fulfilled).

Код (NATS):

go
_, err = c.js.Subscribe(s, func(msg *nats.Msg) {
    if err := c.handle(ctx, msg.Subject, msg.Data); err != nil {
        _ = msg.Nak()
        return
    }
    _ = msg.Ack()
}, nats.Durable(durable), nats.ManualAck())
Код (Kafka):

go
a.di.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupNotify,
    []string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}, a.log)
Это два Event-Driven Consumer для разных шин.

7. Durable Consumer (как в profile, history)
Файл: internal/worker/inbox_consumer.go.

Код:

go
durable := "notification-inbox-" + strings.ReplaceAll(s, ".", "-")
nats.Durable(durable)
8. Manual Ack + Nak (как в history)
Файл: internal/worker/inbox_consumer.go, consumePurchaseNATS.

Код:

go
if err := c.handle(ctx, msg.Subject, msg.Data); err != nil {
    c.log.Error("inbox handle", "subject", msg.Subject, "err", err)
    _ = msg.Nak()
    return
}
_ = msg.Ack()
Всё та же пара: Ack при успехе, Nak при ошибке.

9. Retry with Exponential Backoff (или просто Retry)
Файл: internal/worker/inbox_consumer.go, цикл for attempt := 0; attempt < 30; attempt++.

Почему: При подключении к NATS retry делается 30 раз с задержкой в секунду. Это Retry Pattern.

Код:

go
for attempt := 0; attempt < 30; attempt++ {
    if ctx.Err() != nil {
        return
    }
    _, err = c.js.Subscribe(s, ..., nats.Durable(durable), nats.ManualAck())
    if err == nil {
        c.log.Info("inbox consuming NATS", "subject", s, "durable", durable)
        break
    }
    c.log.Warn("inbox subscribe retry", "subject", s, "attempt", attempt+1, "err", err)
    select {
    case <-ctx.Done():
        return
    case <-time.After(time.Second):
    }
}
Правда, здесь фиксированный интервал (1 сек), а не экспоненциальный. Это просто Retry, не Backoff. Назвать «Retry with Fixed Delay».

10. Idempotent Insert — ON CONFLICT DO NOTHING
Файл: internal/repository/repo.go, Insert.

Почему: Повторная вставка уведомления с тем же source_event для того же user_id не создаёт дубликат:

Код:

go
INSERT INTO notifications (id, user_id, title, body, link, source_event, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id, source_event) WHERE source_event <> '' DO NOTHING
Это Idempotent Write — важно для event-driven систем (события могут прийти дважды).

Бонус: здесь partial index (WHERE source_event <> '') — конфликт срабатывает только если source_event не пустой.

11. Idempotency Key — source_event
Файл: internal/service/inbox.go, HandleSubmitted / HandleApproved / HandleRecordBeaten.

Почему: Каждое уведомление имеет source_event — уникальный ключ, по которому можно отследить, откуда пришло уведомление:

author.application.submitted:app1:admin1

author.application.approved:app1

leaderboard.record_beaten:flappy:1:user123:5000

Это Idempotency Key — стандартный приём в event-driven системах. Если событие придёт дважды — ON CONFLICT не даст дубликата.

Код:

go
src := fmt.Sprintf("author.application.submitted:%s:%s", applicationID, adminID)
12. Fan-Out — HandleSubmitted
Файл: internal/service/inbox.go, HandleSubmitted.

Почему: Одно событие (author.application.submitted) размножается на N уведомлений — по одному для каждого админа.

Код:

go
for _, adminID := range adminIDs {
    adminID = strings.TrimSpace(adminID)
    if adminID == "" {
        continue
    }
    src := fmt.Sprintf("author.application.submitted:%s:%s", applicationID, adminID)
    if _, err := s.repo.Insert(ctx, &model.Notification{...}); err != nil {
        return err
    }
}
Это Fan-Out — одно событие, много получателей.

13. Multi-Bus Consumer с Fallback
Файл: internal/app/di.go, initWorkers.

Почему: У тебя логика выбора шины:

go
go worker.NewInboxConsumer(a.di.js, a.di.inbox, a.log).Start(runCtx)
if a.di.js != nil {
    go a.consumePurchaseNATS(runCtx)
}
if a.di.js == nil && a.cfg.KafkaBrokers != "" {
    go func() {
        // ... Kafka
    }()
}
Это Fallback Pattern: если NATS недоступен — используется Kafka, и наоборот. Такой гибкости нет в других твоих сервисах.

14. Graceful Shutdown / Closer (как всегда)
Файл: internal/app/di.go.

15. Mock-based Unit Tests (как в history)
Файл: internal/service/inbox_test.go, memStore.

Почему: memStore реализует InboxStore, но хранит данные в памяти. Это In-Memory Mock.

Код:

go
type memStore struct {
    items []*model.Notification
}
16. Payload Handling с Unmarshal
Файл: internal/service/notify.go, Handle.

Почему: Notifier.Handle разбирает msg.Topic и вызывает events.UnmarshalPurchasePaid / UnmarshalPurchaseFulfilled. Это Strategy по теме сообщения (switch по topic).

Код:

go
switch msg.Topic {
case kafka.TopicPurchasePaid:
    e, err := events.UnmarshalPurchasePaid(msg.Value)
    if err != nil {
        return err
    }
    return n.notify(ctx, ...)
case kafka.TopicPurchaseFulfilled:
    // ...
}
Это можно назвать Strategy или Dispatcher.

Итог по notification
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go, App.init
Separation of Concerns (2 сервиса)	service/inbox.go, service/notify.go
Port (Hexagonal)	service/inbox.go, InboxStore
Facade (×2)	InboxService, Notifier
Event-Driven Consumer (×2 — NATS, Kafka)	worker/inbox_consumer.go, app/di.go
Durable Consumer	worker/inbox_consumer.go
Manual Ack + Nak	worker/inbox_consumer.go, consumePurchaseNATS
Retry with Fixed Delay	worker/inbox_consumer.go
Idempotent Insert	repository/repo.go, ON CONFLICT DO NOTHING
Idempotency Key	source_event в service/inbox.go
Fan-Out	service/inbox.go, HandleSubmitted
Multi-Bus Fallback	app/di.go, initWorkers
Graceful Shutdown	app/di.go
Mock Object	service/inbox_test.go, memStore
Strategy / Dispatcher	service/notify.go, Handle
Итого 16 паттернов и приёмов. Рекорд вместе с history.

Ключевые отличия от history:

Два независимых сервиса в одном процессе (InboxService + Notifier).

Две шины одновременно (NATS + Kafka) с fallback.

Fan-Out — одно событие, много уведомлений.

Idempotency Key (source_event) — важно для event-driven.

Готовый markdown для NOTIFICATION_PATTERNS.md
md
# notification — Programming Patterns

> Сервис: внутренний inbox уведомлений + внешние уведомления в Telegram.
> Postgres + NATS JetStream + Kafka + gRPC.

## Что делает сервис

Два независимых механизма уведомлений:
1. **InboxService** — внутренний ящик уведомлений (БД + gRPC), слушает NATS.
2. **Notifier** — внешние уведомления в Telegram, слушает Kafka.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Separation of Concerns (2 сервиса)
- **Файлы:** `internal/service/inbox.go`, `internal/service/notify.go`
- **Суть:** `InboxService` (внутренние) и `Notifier` (внешние) — независимые bounded context.

### 4. Port (Hexagonal)
- **Файл:** `internal/service/inbox.go`, `InboxStore`

### 5. Facade (×2)
- **Файлы:** `InboxService`, `Notifier`

### 6. Event-Driven Consumer (×2 — NATS, Kafka)
- **Файлы:** `internal/worker/inbox_consumer.go` (NATS), `app/di.go` (Kafka)

### 7. Durable Consumer
- **Файл:** `internal/worker/inbox_consumer.go`
- **Код:** `nats.Durable("notification-inbox-" + strings.ReplaceAll(s, ".", "-"))`

### 8. Manual Ack + Nak
- **Файлы:** `internal/worker/inbox_consumer.go`, `app/di.go`

### 9. Retry with Fixed Delay
- **Файл:** `internal/worker/inbox_consumer.go`, цикл `for attempt := 0; attempt < 30; attempt++`

### 10. Idempotent Insert
- **Файл:** `internal/repository/repo.go`, `Insert`
- **Код:**
  ```sql
  INSERT INTO notifications (...) VALUES (...)
  ON CONFLICT (user_id, source_event) WHERE source_event <> '' DO NOTHING
11. Idempotency Key
Файл: internal/service/inbox.go

Суть: source_event — уникальный ключ, author.application.submitted:app1:admin1.

12. Fan-Out
Файл: internal/service/inbox.go, HandleSubmitted

Суть: одно событие → N уведомлений (по одному на админа).

13. Multi-Bus Fallback
Файл: internal/app/di.go, initWorkers

Суть: если NATS недоступен — используется Kafka.

14. Graceful Shutdown
Файл: internal/app/di.go

15. Mock Object
Файл: internal/service/inbox_test.go, memStore

16. Strategy / Dispatcher
Файл: internal/service/notify.go, Handle

Суть: switch msg.Topic → разный обработчик.