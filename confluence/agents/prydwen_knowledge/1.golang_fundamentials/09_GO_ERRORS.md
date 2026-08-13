# Go: ошибки

Идиомы Go 1.13+: wrapping, `errors.Is` / `errors.As`, sentinel vs custom types. Без «просто строка в лог».

## Базовый контракт

- Ошибки — значения типа `error` (`Error() string`).
- Возвращайте `(T, error)`, проверяйте сразу; не глотайте без причины.
- Не используйте `panic` для ожидаемых бизнес-ошибок (panic — для багов; на границе gRPC — Recovery interceptor).

## Sentinel errors

```go
var ErrNotFound = errors.New("not found")
```

- Простые фиксированные случаи: not found, conflict, unauthorized.
- Сравнение: **только** через `errors.Is(err, ErrNotFound)`, не `err == ErrNotFound` после wrap.
- Пакет `io.EOF`, `sql.ErrNoRows` — классика; мапьте в domain на уровне repository.

## Wrap: `%w` и `%v`

```go
fmt.Errorf("get user %s: %w", id, err) // цепочка для Is/As
fmt.Errorf("get user %s: %v", id, err) // только текст, цепочка рвётся
```

- `%w` — сохраняет unwrap (`Unwrap() error`).
- Несколько wrap — ok; `errors.Join` (1.20+) — несколько причин.
- Логируйте с `%+v` / slog attrs на границе, внутри — wrap с контекстом операции.

## errors.Is / errors.As

- **`errors.Is(err, target)`** — идёт по цепочке Unwrap, сравнивает sentinel / Equal.
- **`errors.As(err, &target)`** — находит первое подходящее **concrete type** в цепочке, кладёт в `*target`.

```go
var ne *NotFoundError
if errors.As(err, &ne) {
    // ne.ID, HTTP 404 / gRPC NotFound
}
```

Никогда не делайте type assert только к верхнему err после wrap.

## Когда custom types

Пишите свой тип, если нужны **поля** или поведение:

- код/HTTP/gRPC status;
- retryable flag;
- validation details (поля формы);
- domain ID в сообщении.

```go
type ValidationError struct {
    Field, Msg string
}
func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }
```

Реализуйте `Is` / `Unwrap` по необходимости. Не плодите типы на каждую строку — для «просто not found» хватит sentinel.

## Слои в Clean Architecture (EH)

- **Repository:** оборачивает драйвер (`pgx`, mongo) → domain/sentinel (`ErrNotFound`).
- **Service:** бизнес-правила, wrap с use-case контекстом.
- **Handler:** маппинг в gRPC codes / HTTP; не светить internal SQL в клиент.
- Interceptor `Validate()` — входные ошибки до бизнес-логики.

## Питфолы

- `if err != nil { return nil }` — потеряли ошибку.
- Сравнение строк `err.Error() == "…"` — хрупко.
- `err == sql.ErrNoRows` после `fmt.Errorf("%w")` — сломается без `errors.Is`.
- Возврат `(*T)(nil)` как `error` — не-nil interface (см. basic types).

## Типичные вопросы на собесе

- Чем `%w` отличается от `%v` в `fmt.Errorf`?
- Зачем `errors.Is`, если есть `==`?
- Когда sentinel, когда struct с полями?
- Как смапить `sql.ErrNoRows` на gRPC `NotFound` в Clean Architecture?
- Что не так с `var err error = (*MyErr)(nil)`?
- Как логировать wrap-цепочку, не потеряв cause?
