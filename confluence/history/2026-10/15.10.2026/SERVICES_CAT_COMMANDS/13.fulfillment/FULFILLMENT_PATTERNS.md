Разбор fulfillment
Что делает сервис
fulfillment — это эмуляция сборки заказа:

Слушает purchase.paid (из NATS или Kafka).

Ждёт заданное время (AssembleDelaySec, по умолчанию 10 секунд) — симуляция сборки.

Публикует purchase.fulfilled.

Metrics — ObserveAssembly, PurchasesFulfilledTotal.

Нет БД, нет gRPC, нет handler, нет repository. Только NATS/Kafka + service.

Найденные паттерны
1. Service без DI-контейнера
Файл: internal/app/app.go.

Почему: В fulfillment нет отдельного diContainer. Всё собирается прямо в App.New:

Kafka producer/consumer.

NATS connection.

Service.

Это упрощение: сервис настолько мал, что DI-контейнер не нужен. Но паттерн сборки зависимостей остаётся.

Код:

go
func New(ctx context.Context) (*App, error) {
    cfg := config.Load()
    log := logger.NewFrom(getenv("LOG_LEVEL", "info"), getenv("LOG_FORMAT", "text"))
    a := &App{cfg: cfg, log: log, closer: closer.New(log)}

    kcfg := kafka.LoadConfig()
    a.producer = kafka.NewProducer(kcfg, kafka.TopicPurchaseFulfilled, log)
    a.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupFulfillment, []string{kafka.TopicPurchasePaid}, log)

    a.svc = service.New(a.producer, log, time.Duration(cfg.AssembleDelaySec)*time.Second)
    // ...
}
Manual DI вместо DI-контейнера.

2. Dual-Transport Consumer (NATS + Kafka fallback)
Файл: internal/app/app.go, RunUntilSignal.

Почему: fulfillment работает на двух транспортах:

Primary — NATS JetStream.

Fallback — Kafka (только если NATS недоступен).

Код:

go
// Primary path: NATS JetStream (thin deploy).
if a.js != nil {
    go func() {
        _, err := a.js.Subscribe(kafka.TopicPurchasePaid, func(msg *nats.Msg) {
            err := a.svc.HandlePurchasePaid(runCtx, kafka.Message{
                Topic: msg.Subject,
                Value: msg.Data,
            })
            if err != nil {
                _ = msg.Nak()
                return
            }
            _ = msg.Ack()
        }, nats.Durable("fulfillment-purchase-paid"), nats.ManualAck())
    }()
}

// Kafka only when JetStream is down. Both transports would fulfill the same order twice.
if a.js == nil && a.cfg.KafkaBrokers != "" {
    go func() {
        if err := a.consumer.Consume(runCtx, a.svc.HandlePurchasePaid); err != nil { /* ... */ }
    }()
}
Dual-Transport Consumer — с гарантией, что оба транспорта не работают одновременно (иначе двойной fulfillment). Уникально для fulfillment.

3. Idempotency Guard (Transport Level)
Файл: internal/app/app.go.

Почему: В комментарии явно сказано:

Kafka only when JetStream is down. Both transports would fulfill the same order twice.

Это защита от двойного fulfillment: если NATS работает, Kafka не подключается. Это Idempotency Guard на уровне транспорта.

4. Durable Consumer + Manual Ack
Файл: internal/app/app.go.

Код:

go
a.js.Subscribe(kafka.TopicPurchasePaid, func(msg *nats.Msg) {
    // ...
}, nats.Durable("fulfillment-purchase-paid"), nats.ManualAck())
Durable Consumer + Manual Ack — как в других сервисах.

5. Delay-Based Simulation
Файл: internal/service/fulfillment.go, HandlePurchasePaid.

Почему: Сервис симулирует сборку заказа через time.After(s.delay). По умолчанию 10 секунд.

Код:

go
func (s *FulfillmentService) HandlePurchasePaid(ctx context.Context, msg kafka.Message) error {
    paid, err := events.UnmarshalPurchasePaid(msg.Value)
    if err != nil {
        return err
    }
    s.log.Info("purchase paid received; assembling", "purchase", paid.PurchaseUUID, "delay", s.delay)
    start := time.Now()
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(s.delay):
    }
    metrics.ObserveAssembly(time.Since(start).Seconds())
    // ... publish purchase.fulfilled ...
}
Delay-Based Simulation — простое, но эффективное решение. Уникально для fulfillment.

6. Transport Selection в Service
Файл: internal/service/fulfillment.go, HandlePurchasePaid.

Почему: Сервис сам выбирает, куда публиковать:

Если js != nil → NATS.

Иначе → Kafka producer.

Код:

go
if s.js != nil {
    if _, err := s.js.Publish(kafka.TopicPurchaseFulfilled, body); err != nil {
        s.log.Error("nats publish purchase.fulfilled", "err", err)
        return err
    }
    metrics.PurchasesFulfilledTotal.Inc()
    s.log.Info("purchase fulfilled published", "purchase", out.PurchaseUUID)
    return nil
}
if s.out != nil {
    if err := s.out.Send(ctx, []byte(out.PurchaseUUID), body); err != nil {
        return err
    }
}
Transport Selection — сервис решает, куда публиковать.

7. Metrics Recording (Assembly + Fulfilled)
Файл: internal/service/fulfillment.go.

Код:

go
metrics.ObserveAssembly(time.Since(start).Seconds())
// ...
metrics.PurchasesFulfilledTotal.Inc()
Prometheus Metrics — два счётчика: время сборки + количество fulfilled.

8. UUID Generation (crypto/rand)
Файл: internal/service/fulfillment.go, newUUID.

Почему: UUID генерируется через crypto/rand (не math/rand). Это криптографически стойкий UUID.

Код:

go
func newUUID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}
Crypto-Secure UUID — правильный подход.

9. Error Types (Contracts)
Файл: internal/service/fulfillment.go + contracts/events.

Код:

go
paid, err := events.UnmarshalPurchasePaid(msg.Value)
if err != nil {
    return err
}
out := events.PurchaseFulfilled{
    EventUUID:    newUUID(),
    PurchaseUUID: paid.PurchaseUUID,
    UserUUID:     paid.UserUUID,
    ItemUUID:     paid.ItemUUID,
}
body, err := out.Marshal()
Contract Events — типизированные события из пакета contracts/events. Это Shared Contract Package между сервисами.

10. Closer + Graceful Shutdown
Файл: internal/app/app.go.

Код:

go
a.closer.AddNamed("kafka producer", func(context.Context) error { return a.producer.Close() })
a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.consumer.Close() })
a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
a.closer.AddNamed("metrics http", func(ctx context.Context) error { return a.http.Shutdown(ctx) })
Closer Registry — как в других сервисах.

11. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal.

Код:

go
runCtx, cancel := context.WithCancel(ctx)
defer cancel()
// ... запуск горутин с runCtx ...
if err := a.closer.WaitSignal(ctx); err != nil {
    cancel()
    return fmt.Errorf("shutdown: %w", err)
}
cancel()
Context Cancellation для остановки горутин.

12. Metrics HTTP Server (без gRPC)
Файл: internal/app/app.go.

Почему: fulfillment не имеет gRPC-сервера, только HTTP для metrics и health.

Код:

go
mux := http.NewServeMux()
mux.Handle("/metrics", promhttp.Handler())
mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "fulfillment"})
})
mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "fulfillment"})
})
Metrics HTTP Server без gRPC. Уникально для fulfillment.

13. Graceful Degradation (NATS)
Файл: internal/app/app.go.

Код:

go
if cfg.NATSURL != "" {
    nc, err := nats.Connect(cfg.NATSURL)
    if err != nil {
        a.log.Warn("nats connect failed; Kafka-only path", "err", err)
    } else {
        // ...
    }
}
Graceful Degradation — если NATS недоступен, сервис работает через Kafka.

14. Testable Producer Interface
Файл: internal/service/fulfillment_test.go.

Код:

go
type stubProducer struct {
    sent [][]byte
}

func (s *stubProducer) Send(_ context.Context, _, value []byte) error {
    s.sent = append(s.sent, append([]byte(nil), value...))
    return nil
}

func (s *stubProducer) Close() error { return nil }
Testable Producer — kafka.Producer заменяется стабом.

15. Config Test
Файл: internal/config/config_test.go.

Простой тест конфига.

Итог по fulfillment
Паттерн	Где
Manual DI (без контейнера)	app/app.go, App.New
Dual-Transport Consumer	app/app.go, RunUntilSignal
Idempotency Guard (Transport Level)	app/app.go (NATS или Kafka, не оба)
Durable Consumer + Manual Ack	app/app.go
Delay-Based Simulation	service/fulfillment.go, HandlePurchasePaid
Transport Selection в Service	service/fulfillment.go, HandlePurchasePaid
Metrics Recording (Assembly + Fulfilled)	service/fulfillment.go
Crypto-Secure UUID	service/fulfillment.go, newUUID
Contract Events (Shared Package)	contracts/events
Closer + Graceful Shutdown	app/app.go
Context Cancellation	app/app.go, RunUntilSignal
Metrics HTTP Server (без gRPC)	app/app.go
Graceful Degradation (NATS)	app/app.go
Testable Producer Interface	service/fulfillment_test.go
Config Test	internal/config/config_test.go
Итого 15 паттернов и приёмов — для такого маленького сервиса это много.

Уникальные фичи fulfillment:

Dual-Transport Consumer — NATS primary, Kafka fallback.

Idempotency Guard на уровне транспорта.

Delay-Based Simulation — эмуляция сборки заказа.

Transport Selection в Service — сам решает, куда публиковать.

Нет gRPC — только NATS/Kafka + HTTP.

Нет DI-контейнера — manual DI.

Готовый markdown для FULFILLMENT_PATTERNS.md
md
# fulfillment — Programming Patterns

> Сервис: эмуляция сборки заказа (purchase.paid → purchase.fulfilled).
> NATS JetStream + Kafka (fallback) + Prometheus metrics.

## Что делает сервис

Слушает `purchase.paid` (NATS primary, Kafka fallback), ждёт `AssembleDelaySec` (по умолчанию 10 секунд), публикует `purchase.fulfilled`. Не имеет БД, gRPC, handler, repository.

## Паттерны

### 1. Manual DI (без контейнера)
- **Файл:** `internal/app/app.go`, `App.New`
- **Суть:** все зависимости собираются прямо в конструкторе `App`.

### 2. Dual-Transport Consumer
- **Файл:** `internal/app/app.go`, `RunUntilSignal`
- **Суть:** NATS primary, Kafka fallback.
- **Код:**
  ```go
  if a.js != nil {
      // NATS consumer
  }
  if a.js == nil && a.cfg.KafkaBrokers != "" {
      // Kafka consumer
  }
3. Idempotency Guard (Transport Level)
Файл: internal/app/app.go

Комментарий: Both transports would fulfill the same order twice.

4. Durable Consumer + Manual Ack
Файл: internal/app/app.go

Код: nats.Durable("fulfillment-purchase-paid"), nats.ManualAck()

5. Delay-Based Simulation
Файл: internal/service/fulfillment.go, HandlePurchasePaid

Код:

go
select {
case <-ctx.Done():
    return ctx.Err()
case <-time.After(s.delay):
}
6. Transport Selection в Service
Файл: internal/service/fulfillment.go, HandlePurchasePaid

Суть: если js != nil → NATS; иначе → Kafka.

7. Metrics Recording
Файл: internal/service/fulfillment.go

Код: metrics.ObserveAssembly(...), metrics.PurchasesFulfilledTotal.Inc()

8. Crypto-Secure UUID
Файл: internal/service/fulfillment.go, newUUID

Код: rand.Read(b) из crypto/rand.

9. Contract Events (Shared Package)
Файл: contracts/events

10. Closer + Graceful Shutdown
Файл: internal/app/app.go

11. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal

12. Metrics HTTP Server (без gRPC)
Файл: internal/app/app.go

13. Graceful Degradation (NATS)
Файл: internal/app/app.go

Код: if err != nil { a.log.Warn("nats connect failed; Kafka-only path") }

14. Testable Producer Interface
Файл: internal/service/fulfillment_test.go, stubProducer

15. Config Test
Файл: internal/config/config_test.go

Зачем нужен fulfillment
Короткий ответ
fulfillment — это сервис доставки/сборки заказа. В реальной системе он бы:

Принял purchase.paid.

Отправил бы заказ (например, фулфилмент-центру — тому, кто пакует и отправляет физический мерч).

Получил бы трек-номер.

Опубликовал purchase.fulfilled.

Сейчас он симулирует этот процесс через time.After(10 * time.Second). Это не тест — это заглушка для разработки, которая позже заменяется на реальный вызов внешнего API.

Что было бы в реальной системе
go
func (s *FulfillmentService) HandlePurchasePaid(ctx context.Context, msg kafka.Message) error {
    paid, _ := events.UnmarshalPurchasePaid(msg.Value)

    // 1. Запросить склад
    order, err := s.warehouse.Reserve(ctx, paid.ItemUUID, paid.UserUUID)
    // 2. Создать задачу сборки
    task, err := s.wms.CreateAssemblyTask(ctx, order)
    // 3. Ждать, пока соберут
    status, err := s.wms.WaitForAssembly(ctx, task.ID)
    // 4. Получить трек-номер
    tracking, err := s.delivery.CreateShipment(ctx, order)
    // 5. Опубликовать результат
    out := events.PurchaseFulfilled{
        PurchaseUUID: paid.PurchaseUUID,
        TrackingNumber: tracking.Number,
        ...
    }
    return s.js.Publish(kafka.TopicPurchaseFulfilled, out.Marshal())
}
Сейчас time.After(s.delay) — это placeholder. Когда появится реальный WMS (Warehouse Management System), его заменят на вызовы.

Почему это не тест
Признак теста	Что у fulfillment
Запускается только в CI	Работает в prod (в docker-compose, в k8s)
Проверяет код	Обрабатывает реальные события purchase.paid
Не имеет побочек	Публикует реальные purchase.fulfilled, которые ловят shop и notification
Не входит в main	Входит в основной пайплайн
Это рабочий сервис с заглушкой вместо внешнего API.

Что он делает прямо сейчас в твоей системе
Смотри цепочку:

text
shop.PurchaseItem
  → billing.SpendCurrency (списали билетики)
  → shop.PurchaseItemWithStock (записали в БД + outbox)
  → outbox_worker: publish purchase.paid

fulfillment.HandlePurchasePaid
  → ждёт 10 секунд
  → publish purchase.fulfilled

shop.subscribePurchaseFulfilled
  → MarkPurchaseFulfilled (записали fulfilled_at)

notification.consumePurchaseNATS
  → send Telegram
То есть fulfillment — это звено в цепочке «купил → отправили → уведомили». Без него:

shop.purchases.fulfilled_at никогда не заполнится.

Refund window для физических товаров не закроется (см. canCancelInventoryItem — он проверяет FulfilledAt != nil).

Пользователь не получит уведомление, что заказ отправлен.

Это критично для бизнес-логики возврата физических товаров.

Аналогия
Представь, что у тебя есть интернет-магазин:

shop — принимает заказ.

billing — списывает деньги.

payment — принимает платёж.

fulfillment — собирает и отправляет заказ (сейчас эмулирует).

notification — говорит клиенту «ваш заказ отправлен».

Без fulfillment клиент заплатил, но заказ не отправлен. Это баг бизнес-логики, а не тестовый сервис.

Что делать с fulfillment в будущем
Есть три варианта:

Вариант 1: Оставить как заглушку
Пока проект — игра, а не реальный магазин, fulfillment может просто эмулировать отправку. Это нормально. Главное — событие purchase.fulfilled публикуется, и вся цепочка работает.

Вариант 2: Заменить на реальный WMS
Когда появится реальный партнёр (например, склад для отправки мерча), fulfillment будет вызывать его API:

POST /api/orders — создать заказ.

GET /api/orders/{id} — проверить статус.

Webhook от WMS → publish purchase.fulfilled.

Вариант 3: Убрать сервис совсем
Если физического мерча нет — shop может сам публиковать purchase.fulfilled сразу после purchase.paid. Но тогда теряется возможность тестировать всю цепочку.

Мой совет: оставить. Он маленький, работает, и держит всю цепочку.

Что можно улучшить в fulfillment (когда будет время)
Задержку сделать конфигурируемой (сейчас AssembleDelaySec).

Добавить retry — если purchase.fulfilled не опубликовался, повторить.

Добавить outbox — если сервис упадёт между purchase.paid и purchase.fulfilled, событие потеряется. Outbox бы это решил.

Разделить fulfillment и delivery — если появится реальная логистика.

Короткий ответ
fulfillment — не тест. Это рабочий сервис с заглушкой вместо внешнего WMS. Он держит цепочку purchase.paid → purchase.fulfilled, от которой зависит:

Закрытие refund window для физических товаров.

Уведомления пользователю.

Статус заказа в shop.

Когда появится реальная логистика — заглушка заменится на API-вызовы, но структура сервиса останется.
