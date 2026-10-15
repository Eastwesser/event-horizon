Разбор payment
Что делает сервис
payment — это подписки и платежи:

Создание checkout — генерация ссылки на Boosty.

Подтверждение платежа — через webhook (HMAC-SHA256) или gRPC.

Активация подписки — при успешном платеже.

Проверка подписки — GetSubscription, CanPurchaseMerch.

Revenue Splits — разделение дохода (85% автору, 15% платформе).

Outbox — публикация payment.completed.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg    *config.Config
    dbpool *pgxpool.Pool
    cache  *repository.RedisRepo
    nc     *nats.Conn
    js     nats.JetStreamContext
    repo   *repository.PostgresRepo
    svc    *service.PaymentService
    api    pb.PaymentServiceServer
}
Стандартный DI-контейнер.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initPostgres, a.initRedis,
    a.initNATS, a.initDomain, a.initOutbox, a.initGRPC, a.initMetricsHTTP,
}
initOutbox после initDomain — воркер стартует с готовым repo.

3. Port (Hexagonal) — два порта
Файл: internal/service/payment_service.go.

Код:

go
type PaymentStore interface {
    CreatePayment(ctx context.Context, p *model.Payment) error
    GetPayment(ctx context.Context, id string) (*model.Payment, error)
    CompletePaymentAndActivateSubscription(ctx context.Context, paymentID, providerRef string, sub *model.Subscription, eventType string, eventPayload map[string]any) error
    GetActiveSubscription(ctx context.Context, userID string) (*model.Subscription, error)
}

type SubscriptionCache interface {
    GetSubscription(ctx context.Context, userID string) (*model.Subscription, error)
    SetSubscription(ctx context.Context, sub *model.Subscription) error
}
Два порта — PaymentStore (Postgres) и SubscriptionCache (Redis).

4. Facade — PaymentService
Файл: internal/service/payment_service.go.

Код:

go
type PaymentService struct {
    repo             PaymentStore
    cache            SubscriptionCache
    boostyURL        string
    webhookSecret    string
    subscriptionDays int
}
Фасад над двумя портами + конфиг.

5. HMAC-SHA256 Verification (Webhook Security)
Файл: internal/webhook/verify.go.

Почему: Webhook-подпись проверяется через HMAC-SHA256 с constant-time comparison (subtle.ConstantTimeCompare). Это стандарт для webhook security.

Код:

go
func VerifyHMACSHA256(secret string, rawBody []byte, signatureHeader string) bool {
    sig := strings.TrimSpace(signatureHeader)
    if i := strings.Index(sig, "="); i >= 0 {
        sig = sig[i+1:]
    }
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(rawBody)
    sum := mac.Sum(nil)
    got, err := hex.DecodeString(sig)
    if err != nil {
        return false
    }
    return subtle.ConstantTimeCompare(sum, got) == 1
}
Плюс: subtle.ConstantTimeCompare защищает от timing attacks. Ключевой паттерн безопасности.

6. Dual Auth Path (HMAC + Shared Secret)
Файл: internal/webhook/verify.go, Authorize.

Почему: Webhook можно авторизовать двумя способами:

HMAC (preferred) — подпись в заголовке.

Shared Secret (fallback) — секрет в теле.

Код:

go
func Authorize(configuredSecret, providedSecret string, rawBody []byte, signatureHeader string) error {
    if configuredSecret == "" {
        return nil // dev: no secret configured
    }
    if signatureHeader != "" {
        if VerifyHMACSHA256(configuredSecret, rawBody, signatureHeader) {
            return nil
        }
        return ErrInvalidSignature
    }
    if providedSecret != "" && subtle.ConstantTimeCompare([]byte(providedSecret), []byte(configuredSecret)) == 1 {
        return nil
    }
    return ErrUnauthorized
}
Dual Auth Path — гибкая авторизация. Уникально для payment.

7. Idempotent Payment Completion
Файл: internal/repository/postgres_repo.go, CompletePaymentAndActivateSubscription.

Почему: Payment completion идемпотентна:

Берём payment под FOR UPDATE.

Если статус уже COMPLETED — возвращаем ErrAlreadyPaid.

Если статус не PENDING — ErrInvalidStatus.

Код:

go
var status string
err = tx.QueryRow(ctx, `SELECT status FROM payments WHERE id = $1 FOR UPDATE`, paymentID).Scan(&status)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return model.ErrNotFound
    }
    return err
}
if status == model.StatusCompleted {
    return model.ErrAlreadyPaid
}
if status != model.StatusPending {
    return model.ErrInvalidStatus
}
Pessimistic Lock + Idempotency Check — защита от двойного webhook.

8. Idempotent Handler (в сервисе)
Файл: internal/service/payment_service.go, ConfirmPayment.

Код:

go
if err := s.repo.CompletePaymentAndActivateSubscription(ctx, paymentID, providerRef, sub, "payment.completed", event); err != nil {
    if errors.Is(err, model.ErrAlreadyPaid) {
        existing, getErr := s.repo.GetActiveSubscription(ctx, p.UserID)
        if getErr == nil {
            if s.cache != nil {
                _ = s.cache.SetSubscription(ctx, existing)
            }
            return existing, nil
        }
    }
    return nil, err
}
Idempotent Handler — повторный webhook возвращает уже существующую подписку. Это обработка race condition.

9. Transactional Outbox (как в inventory, shop)
Файл: internal/repository/postgres_repo.go, CompletePaymentAndActivateSubscription.

Почему: В одной транзакции:

Обновляется payment → COMPLETED.

Старые подписки → EXPIRED.

Создаётся новая подписка.

Пишется outbox event payment.completed.

Пишется revenue split.

Код:

go
tx, err := r.db.Begin(ctx)
// ...
// 1. UPDATE payments SET status = COMPLETED
// 2. UPDATE subscriptions SET status = EXPIRED WHERE user_id = ... AND status = ACTIVE
// 3. INSERT INTO subscriptions ...
// 4. INSERT INTO outbox (event_type, payload) VALUES ('payment.completed', ...)
// 5. INSERT INTO revenue_splits (payment_id, amount_rub, author_share_rub, platform_share_rub) ...
return tx.Commit(ctx)
Transactional Outbox + Subscription Expiry + Revenue Split в одной транзакции. Мощно.

10. Subscription Expiry on New Payment (бизнес-правило)
Файл: internal/repository/postgres_repo.go.

Код:

go
_, err = tx.Exec(ctx, `
    UPDATE subscriptions SET status = $2, updated_at = $3
    WHERE user_id = $1 AND status = $4`,
    sub.UserID, model.StatusExpired, now, model.StatusActive,
)
Старая активная подписка автоматически истекает при покупке новой. Это важное бизнес-правило: не может быть двух активных подписок.

11. Revenue Split (бизнес-логика)
Файл: internal/repository/postgres_repo.go.

Код:

go
authorShare := sub.AmountRub * 85 / 100
platformShare := sub.AmountRub - authorShare
_, err = tx.Exec(ctx, `
    INSERT INTO revenue_splits (payment_id, amount_rub, author_share_rub, platform_share_rub)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (payment_id) DO NOTHING`,
    paymentID, sub.AmountRub, authorShare, platformShare,
)
Revenue Split — 85% автору, 15% платформе. Идемпотентно через ON CONFLICT DO NOTHING.

12. Cache-Aside с TTL по Expiry
Файл: internal/repository/redis_repo.go, SetSubscription.

Код:

go
func (r *RedisRepo) SetSubscription(ctx context.Context, s *model.Subscription) error {
    if s == nil {
        return nil
    }
    data, err := json.Marshal(s)
    if err != nil {
        return err
    }
    ttl := r.ttl
    if rem := time.Until(s.ExpiresAt); rem > 0 && rem < ttl {
        ttl = rem
    }
    return r.client.Set(ctx, subKey(s.UserID), data, ttl).Err()
}
TTL based on Expiry — кеш живёт не дольше подписки. Если подписка истекает через час, а TTL = 30 минут, кеш будет жить час.

13. Cache-Aside с проверкой активности
Файл: internal/service/payment_service.go, GetSubscription.

Код:

go
func (s *PaymentService) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
    if s.cache != nil {
        if cached, err := s.cache.GetSubscription(ctx, userID); err == nil && cached.IsActive(time.Now().UTC()) {
            return cached, nil
        }
    }
    sub, err := s.repo.GetActiveSubscription(ctx, userID)
    if err != nil {
        return nil, err
    }
    if s.cache != nil {
        _ = s.cache.SetSubscription(ctx, sub)
    }
    return sub, nil
}
Cache-Aside с проверкой IsActive — если кеш говорит, что подписка неактивна, идём в БД. Это защита от протухшего кеша.

14. Outbox Worker (Polling Publisher)
Файл: internal/worker/outbox_worker.go.

Код:

go
func (w *OutboxWorker) processBatch(ctx context.Context) {
    rows, err := w.db.Query(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at ASC
        LIMIT 50`)
    // ...
    for rows.Next() {
        // Публикуем в NATS
        w.js.Publish(eventType, payload)
        // Помечаем как обработанное
        w.db.Exec(ctx, `UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1`, id)
    }
}
Простой Polling Publisher без FOR UPDATE SKIP LOCKED (в отличие от shop).

15. Optional Redis (Graceful Degradation)
Файл: internal/app/di.go, initRedis.

Код:

go
if err := cache.Ping(ctx); err != nil {
    a.log.Warn("redis unavailable; subscription cache degraded", "addr", a.cfg.RedisAddr, "err", err)
} else {
    a.log.Info("redis connected", "addr", a.cfg.RedisAddr)
}
a.di.cache = cache
Graceful Degradation — если Redis недоступен, сервис работает без кеша.

16. Converter inline (нет отдельного пакета)
Файл: internal/service/payment_service.go.

Почему: В payment нет internal/converter/. Конвертация model.Payment → pb.CreateCheckoutResponse делается inline в handler.

Код:

go
func (h *GRPCHandler) CreateCheckout(ctx context.Context, req *pb.CreateCheckoutRequest) (*pb.CreateCheckoutResponse, error) {
    p, err := h.svc.CreateCheckout(ctx, req.GetUserId(), req.GetPlan())
    if err != nil {
        return nil, mapErr(err)
    }
    return &pb.CreateCheckoutResponse{
        PaymentId:   p.ID,
        CheckoutUrl: p.CheckoutURL,
        AmountRub:   int32(p.AmountRub),
        Plan:        p.Plan,
        Status:      p.Status,
    }, nil
}
Inline Conversion — упрощение по сравнению с другими сервисами.

17. Error Mapping
Файл: internal/handler/grpc_handler.go, mapErr.

Код:

go
func mapErr(err error) error {
    switch {
    case errors.Is(err, model.ErrNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, model.ErrInvalidInput), errors.Is(err, model.ErrInvalidStatus):
        return status.Error(codes.InvalidArgument, err.Error())
    case errors.Is(err, model.ErrUnauthorized):
        return status.Error(codes.PermissionDenied, err.Error())
    case errors.Is(err, model.ErrAlreadyPaid):
        return status.Error(codes.AlreadyExists, err.Error())
    default:
        return status.Errorf(codes.Internal, "%v", err)
    }
}
Error Mapping — доменные ошибки → gRPC-коды.

18. Mock Object (Tests)
Файлы: payment_service_test.go, verify_test.go.

Два мока: mockPayStore + mockSubCache.

19. Testcontainers
Файл: internal/repository/postgres_integration_test.go.

20. Build Tags
Файл: internal/repository/postgres_integration_test.go, //go:build integration.

21. Graceful Shutdown / Closer
Файл: internal/app/di.go.

Итог по payment
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go
Port (Hexagonal) — 2 порта	service/payment_service.go
Facade	service/payment_service.go
HMAC-SHA256 Verification	webhook/verify.go
Dual Auth Path	webhook/verify.go, Authorize
Idempotent Payment Completion	repository/postgres_repo.go
Idempotent Handler	service/payment_service.go, ConfirmPayment
Transactional Outbox	repository/postgres_repo.go
Subscription Expiry on New Payment	repository/postgres_repo.go
Revenue Split	repository/postgres_repo.go
Cache-Aside с TTL по Expiry	repository/redis_repo.go
Cache-Aside с проверкой IsActive	service/payment_service.go
Outbox Worker (Polling Publisher)	worker/outbox_worker.go
Optional Redis (Graceful Degradation)	app/di.go, initRedis
Inline Conversion	handler/grpc_handler.go
Error Mapping	handler/grpc_handler.go, mapErr
Mock Object (×2)	service/*_test.go, webhook/verify_test.go
Testcontainers	repository/postgres_integration_test.go
Build Tags	repository/postgres_integration_test.go
Graceful Shutdown	app/di.go
Итого 21 паттерн и приём.

Уникальные фичи payment:

HMAC-SHA256 Verification + subtle.ConstantTimeCompare — единственный сервис с webhook-security.

Dual Auth Path (HMAC + Shared Secret).

Subscription Expiry при новом платеже.

Revenue Split (85/15).

Idempotent Payment Completion через FOR UPDATE.

Cache TTL based on Expiry — кеш живёт не дольше подписки.

Готовый markdown для PAYMENT_PATTERNS.md
md
# payment — Programming Patterns

> Сервис: подписки и платежи через Boosty.
> Postgres + Redis + NATS + gRPC + HMAC-SHA256 webhook.

## Что делает сервис

Создаёт checkout, подтверждает платежи (webhook + gRPC), активирует подписки, проверяет активность, считает revenue splits (85% автору / 15% платформе). Публикует `payment.completed` через Outbox.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Port (Hexagonal) — 2 порта
- **Файл:** `internal/service/payment_service.go`
- **Порты:** `PaymentStore`, `SubscriptionCache`.

### 4. Facade
- **Файл:** `internal/service/payment_service.go`, `PaymentService`

### 5. HMAC-SHA256 Verification
- **Файл:** `internal/webhook/verify.go`, `VerifyHMACSHA256`
- **Код:** `subtle.ConstantTimeCompare(sum, got) == 1`

### 6. Dual Auth Path
- **Файл:** `internal/webhook/verify.go`, `Authorize`
- **Суть:** HMAC (preferred) или Shared Secret (fallback).

### 7. Idempotent Payment Completion
- **Файл:** `internal/repository/postgres_repo.go`
- **Код:** `SELECT status FROM payments WHERE id = $1 FOR UPDATE`

### 8. Idempotent Handler
- **Файл:** `internal/service/payment_service.go`, `ConfirmPayment`
- **Суть:** при `ErrAlreadyPaid` возвращаем существующую подписку.

### 9. Transactional Outbox
- **Файл:** `internal/repository/postgres_repo.go`, `CompletePaymentAndActivateSubscription`

### 10. Subscription Expiry on New Payment
- **Файл:** `internal/repository/postgres_repo.go`
- **Суть:** старая активная подписка → EXPIRED.

### 11. Revenue Split
- **Файл:** `internal/repository/postgres_repo.go`
- **Формула:** `author_share = amount * 85 / 100`.

### 12. Cache-Aside с TTL по Expiry
- **Файл:** `internal/repository/redis_repo.go`, `SetSubscription`

### 13. Cache-Aside с проверкой IsActive
- **Файл:** `internal/service/payment_service.go`, `GetSubscription`

### 14. Outbox Worker (Polling Publisher)
- **Файл:** `internal/worker/outbox_worker.go`

### 15. Optional Redis (Graceful Degradation)
- **Файл:** `internal/app/di.go`, `initRedis`

### 16. Inline Conversion
- **Файл:** `internal/handler/grpc_handler.go`

### 17. Error Mapping
- **Файл:** `internal/handler/grpc_handler.go`, `mapErr`

### 18. Mock Object (×2)
- **Файлы:** `internal/service/payment_service_test.go`, `internal/webhook/verify_test.go`

### 19. Testcontainers
- **Файл:** `internal/repository/postgres_integration_test.go`

### 20. Build Tags
- **Файл:** `internal/repository/postgres_integration_test.go`, `//go:build integration`

### 21. Graceful Shutdown
- **Файл:** `internal/app/di.go`