Разбор authors
Что делает сервис
authors — это управление авторами:

Author profile — UpsertProfile, GetAuthor, ListAuthors.

Application workflow — SubmitApplication → ApproveApplication / RejectApplication / RevertApplication.

State machine — заявка имеет состояния pending → approved/rejected, и может быть reverted обратно в pending.

Transactional outbox — события author.upserted пишутся в той же транзакции, что и изменения.

Cache-aside — кеширование авторов в Redis.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg    *config.Config
    dbpool *pgxpool.Pool
    cache  *repository.RedisRepo
    js     nats.JetStreamContext
    nc     *nats.Conn
    repo   *repository.PostgresRepo
    svc    *service.AuthorsService
    api    pb.AuthorsServiceServer
}
Стандартный DI.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
for _, step := range []func(context.Context) error{
    a.initLogger, a.initCloser, a.initPostgres, a.initRedis, a.initNATS,
    a.initDomain, a.initOutbox, a.initGRPC, a.initMetricsHTTP,
} {
    if err := step(ctx); err != nil {
        return err
    }
}
Стандартный pipeline.

3. Port (Hexagonal) — два порта
Файл: internal/service/authors_service.go.

Код:

go
type AuthorStore interface {
    Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error
    GetByUserID(ctx context.Context, userID string) (*model.Author, error)
    List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error)
    GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error)
    GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error)
    InsertApplication(ctx context.Context, a *model.AuthorApplication) error
    ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error)
    ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error)
    RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error)
    RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error)
}

type AuthorCache interface {
    Get(ctx context.Context, userID string) (*model.Author, error)
    Set(ctx context.Context, a *model.Author) error
    Delete(ctx context.Context, userID string) error
}
Два порта — AuthorStore (Postgres) + AuthorCache (Redis).

4. Facade — AuthorsService
Файл: internal/service/authors_service.go.

Код:

go
type AuthorsService struct {
    repo  AuthorStore
    cache AuthorCache
}
Тонкий фасад.

5. State Machine (Application Lifecycle)
Файл: internal/repository/repo.go + internal/service/authors_service.go.

Почему: Заявка имеет state machine:

pending → при создании.

approved → при approve (создаётся author).

rejected → при reject.

pending (обратно) → при revert.

Код (Approve):

go
func (r *PostgresRepo) ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
    tx, err := r.db.Begin(ctx)
    // ...
    row := tx.QueryRow(ctx, `
        SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
        FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
    app, err := scanApplication(row)
    if app.Status != model.ApplicationPending {
        return nil, nil, model.ErrAlreadyReviewed
    }
    // ...  UPDATE author_applications SET status = 'approved' ...
    // ...  INSERT/UPDATE authors ...
    // ...  INSERT INTO outbox ...
}
State Machine с transitions через бизнес-правила. Ключевой паттерн authors.

6. Pessimistic Lock (FOR UPDATE)
Файл: internal/repository/repo.go, ApproveApplicationInTx + RejectApplication + RevertApplication.

Код:

go
row := tx.QueryRow(ctx, `
    SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
    FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
Pessimistic Lock — заявка блокируется на время транзакции. Это защищает от race condition, когда два админа одновременно approve'ят одну заявку.

Уникально для authors — в других сервисах используется либо FOR UPDATE SKIP LOCKED (outbox), либо conditional UPDATE.

7. Idempotent State Transition
Файл: internal/repository/repo.go, ApproveApplicationInTx.

Код:

go
if app.Status != model.ApplicationPending {
    return nil, nil, model.ErrAlreadyReviewed
}
Idempotent State Transition — повторный approve возвращает ErrAlreadyReviewed (не создаёт дубликат).

8. Compensating Action (Revert)
Файл: internal/repository/repo.go, RevertApplication.

Почему: RevertApplication — компенсация для approve. Он:

Возвращает заявку в pending.

Удаляет автора (DELETE FROM authors WHERE user_id = ...).

Код:

go
func (r *PostgresRepo) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
    // ...
    if app.Status != model.ApplicationApproved {
        return nil, model.ErrInvalidInput
    }
    _, err = tx.Exec(ctx, `
        UPDATE author_applications
        SET status = 'pending', reviewed_at = NULL, reviewed_by = '', reviewer_note = ''
        WHERE id = $1`, applicationID)
    // ...
    _, err = tx.Exec(ctx, `DELETE FROM authors WHERE user_id = $1`, app.UserID)
    // ...
}
Compensating Action — обратное действие. Как в Saga.

9. Transactional Outbox
Файл: internal/repository/repo.go, Upsert + ApproveApplicationInTx.

Код (Upsert):

go
body, err := json.Marshal(eventPayload)
if err != nil {
    return err
}
_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
    uuid.NewString(), eventType, body)
if err != nil {
    return err
}
return tx.Commit(ctx)
Код (Approve):

go
event := map[string]any{
    "event":          "author.upserted",
    "user_id":        author.UserID,
    "display_name":   author.DisplayName,
    "source":         "application_approved",
    "application_id": applicationID,
    "timestamp":      now.Unix(),
}
body, err := json.Marshal(event)
// ...
_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
    uuid.NewString(), "author.upserted", body)
Transactional Outbox — событие в той же транзакции. Как в других сервисах.

10. Outbox Worker
Файл: internal/worker/outbox_worker.go.

Код:

go
func (w *OutboxWorker) processBatch(ctx context.Context) {
    rows, err := w.db.Query(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false ORDER BY created_at ASC LIMIT 50`)
    // ...
    for rows.Next() {
        // Публикуем в NATS
        w.js.Publish(eventType, payload)
        // Помечаем как обработанное
        w.db.Exec(ctx, `UPDATE outbox SET processed=true, processed_at=NOW() WHERE id=$1`, id)
    }
}
Простой Polling Publisher (без FOR UPDATE SKIP LOCKED).

11. Cache-Aside
Файл: internal/service/authors_service.go, GetAuthor + UpsertProfile.

Код (GetAuthor):

go
func (s *AuthorsService) GetAuthor(ctx context.Context, userID string) (*model.Author, error) {
    if s.cache != nil {
        if a, err := s.cache.Get(ctx, userID); err == nil {
            return a, nil
        }
    }
    a, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if s.cache != nil {
        _ = s.cache.Set(ctx, a)
    }
    return a, nil
}
Cache-Aside — стандартный приём.

Плюс: при RevertApplication кеш инвалидируется:

go
if s.cache != nil {
    _ = s.cache.Delete(ctx, app.UserID)
}
12. Idempotent Application Submission
Файл: internal/service/authors_service.go, SubmitApplication.

Почему: Если у пользователя уже есть pending заявка, SubmitApplication возвращает её (не создаёт дубликат).

Код:

go
if pending, err := s.repo.GetPendingApplicationByUserID(ctx, userID); err == nil {
    return pending, nil
} else if !errors.Is(err, model.ErrApplicationMissing) {
    return nil, err
}
Плюс: обработка race condition при insert:

go
if err := s.repo.InsertApplication(ctx, app); err != nil {
    // Race: another pending insert — return existing.
    if pending, getErr := s.repo.GetPendingApplicationByUserID(ctx, userID); getErr == nil {
        return pending, nil
    }
    return nil, err
}
Idempotent Application Submission — с обработкой race condition. Уникально для authors.

13. Role Check (Privilege Escalation Prevention)
Файл: internal/service/authors_service.go, SubmitApplication.

Код:

go
role := strings.ToLower(strings.TrimSpace(callerRole))
if role == "author" || role == "admin" {
    return nil, model.ErrAlreadyPrivileged
}
Role Check — автор или админ не может подать заявку на автора. Защита от privilege escalation.

14. Validator (proto Validate)
Файл: internal/validate/validate_test.go + proto/validate.go.

Почему: Для каждой proto-структуры есть Validate(). Он вызывается через interceptor.

Код:

go
func TestUpsertProfileRequest_Validate(t *testing.T) {
    // ...
    err := tc.req.Validate()
    // ...
}
Proto Validator — стандартный приём.

15. Converter (inline в handler)
Файл: internal/handler/grpc_handler.go.

Почему: В authors нет internal/converter/. Конвертеры toPB и applicationToPB inline в handler.

Код:

go
func toPB(a *model.Author) *pb.Author { /* ... */ }
func applicationToPB(a *model.AuthorApplication) *pb.AuthorApplication { /* ... */ }
Inline Converter — упрощение.

16. Error Mapping
Файл: internal/handler/grpc_handler.go, mapErr.

Код:

go
func mapErr(err error) error {
    switch {
    case errors.Is(err, model.ErrNotFound), errors.Is(err, model.ErrApplicationMissing):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, model.ErrInvalidInput),
        errors.Is(err, model.ErrAlreadyPrivileged),
        errors.Is(err, model.ErrAlreadyReviewed):
        return status.Error(codes.InvalidArgument, err.Error())
    default:
        return status.Errorf(codes.Internal, "%v", err)
    }
}
Error Mapping — стандартный приём.

17. Mock Object (Tests)
Файл: internal/service/authors_service_test.go.

Три мока:

mockStore — для AuthorStore.

mockCache — для AuthorCache.

Функциональные моки через *Fn — интересный приём.

18. Testcontainers
Файл: internal/repository/postgres_integration_test.go.

19. Build Tags
Файл: internal/repository/postgres_integration_test.go, //go:build integration.

20. Graceful Shutdown / Closer
Файл: internal/app/di.go.

Итог по authors
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go
Port (Hexagonal) — 2 порта	service/authors_service.go
Facade	service/authors_service.go
State Machine (Application Lifecycle)	repository/repo.go + service/authors_service.go
Pessimistic Lock (FOR UPDATE)	repository/repo.go, ApproveApplicationInTx
Idempotent State Transition	repository/repo.go, ApproveApplicationInTx
Compensating Action (Revert)	repository/repo.go, RevertApplication
Transactional Outbox	repository/repo.go, Upsert + ApproveApplicationInTx
Outbox Worker	worker/outbox_worker.go
Cache-Aside	service/authors_service.go, GetAuthor
Idempotent Application Submission	service/authors_service.go, SubmitApplication
Role Check (Privilege Escalation Prevention)	service/authors_service.go, SubmitApplication
Proto Validator	validate/validate_test.go + proto/validate.go
Inline Converter	handler/grpc_handler.go, toPB, applicationToPB
Error Mapping	handler/grpc_handler.go, mapErr
Mock Object (×2)	service/authors_service_test.go
Testcontainers	repository/postgres_integration_test.go
Build Tags	repository/postgres_integration_test.go
Graceful Shutdown	app/di.go
Итого 20 паттернов и приёмов.

Уникальные фичи authors:

State Machine для заявок (pending → approved/rejected → pending).

Pessimistic Lock через FOR UPDATE — единственный сервис, где это используется в бизнес-логике.

Compensating Action (Revert) — обратное действие к approve.

Idempotent Application Submission с обработкой race condition.

Role Check для предотвращения privilege escalation.

Готовый markdown для AUTHORS_PATTERNS.md
md
# authors — Programming Patterns

> Сервис: управление авторами + заявки на роль автора.
> Postgres + Redis + NATS + gRPC.

## Что делает сервис

Author profile (Upsert, Get, List). Approval workflow для заявок на автора (submit → approve/reject/revert). Транзакционный outbox для `author.upserted`.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Port (Hexagonal) — 2 порта
- **Файл:** `internal/service/authors_service.go`
- **Порты:** `AuthorStore`, `AuthorCache`.

### 4. Facade
- **Файл:** `internal/service/authors_service.go`, `AuthorsService`

### 5. State Machine (Application Lifecycle)
- **Файлы:** `internal/repository/repo.go` + `internal/service/authors_service.go`
- **Состояния:** `pending → approved/rejected → pending`.

### 6. Pessimistic Lock (`FOR UPDATE`)
- **Файл:** `internal/repository/repo.go`, `ApproveApplicationInTx`
- **Код:** `SELECT ... FROM author_applications WHERE id = $1 FOR UPDATE`

### 7. Idempotent State Transition
- **Файл:** `internal/repository/repo.go`, `ApproveApplicationInTx`
- **Код:** `if app.Status != model.ApplicationPending { return ErrAlreadyReviewed }`

### 8. Compensating Action (Revert)
- **Файл:** `internal/repository/repo.go`, `RevertApplication`
- **Суть:** `approved → pending` + `DELETE FROM authors`.

### 9. Transactional Outbox
- **Файл:** `internal/repository/repo.go`, `Upsert` + `ApproveApplicationInTx`

### 10. Outbox Worker
- **Файл:** `internal/worker/outbox_worker.go`

### 11. Cache-Aside
- **Файл:** `internal/service/authors_service.go`, `GetAuthor`

### 12. Idempotent Application Submission
- **Файл:** `internal/service/authors_service.go`, `SubmitApplication`
- **Суть:** возвращает существующую pending-заявку; обрабатывает race condition.

### 13. Role Check (Privilege Escalation Prevention)
- **Файл:** `internal/service/authors_service.go`, `SubmitApplication`
- **Код:** `if role == "author" || role == "admin" { return ErrAlreadyPrivileged }`

### 14. Proto Validator
- **Файл:** `internal/validate/validate_test.go` + `proto/validate.go`

### 15. Inline Converter
- **Файл:** `internal/handler/grpc_handler.go`, `toPB`, `applicationToPB`

### 16. Error Mapping
- **Файл:** `internal/handler/grpc_handler.go`, `mapErr`

### 17. Mock Object (×2)
- **Файл:** `internal/service/authors_service_test.go`, `mockStore`, `mockCache`

### 18. Testcontainers
- **Файл:** `internal/repository/postgres_integration_test.go`

### 19. Build Tags
- **Файл:** `internal/repository/postgres_integration_test.go`, `//go:build integration`

### 20. Graceful Shutdown
- **Файл:** `internal/app/di.go`