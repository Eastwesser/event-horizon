# Go: context

`context.Context` — стандартный сигнал отмены, дедлайна и request-scoped значений. В Event Horizon **запрещено** хранить context в долгоживущих struct’ах.

## Зачем нужен

- Отмена дерева операций (gRPC/HTTP request → service → repo → SQL).
- Deadline/timeout на внешних вызовах.
- Request values: trace id, редко — auth (лучше явный параметр; Value — last resort).

Корень: `context.Background()` (main, init) или `context.TODO()` (временно, когда непонятно). В handlers приходит ctx от фреймворка.

## WithCancel / WithTimeout / WithDeadline / WithValue

```go
ctx, cancel := context.WithCancel(parent)
defer cancel() // всегда, иначе leak таймера/ребёнка до parent cancel

ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel()

ctx := context.WithValue(parent, key, val) // key — свой unexported type!
```

- **WithCancel:** ручная отмена (shutdown worker’ов, abort fan-out).
- **WithTimeout / WithDeadline:** автоотмена по времени; `ctx.Err()` → `context.DeadlineExceeded`.
- **WithValue:** только request-scoped данные, не обязательные зависимости. Ключ — приватный тип (`type ctxKey struct{}`), не `string`, чтобы не конфликтовать.

## Правила отмены

1. Кто создал `WithCancel`/`WithTimeout` — тот вызывает `cancel` (обычно `defer`).
2. Дочерние горутины слушают `ctx.Done()` или передают ctx в API, которые его уважают (`QueryContext`, gRPC, http с ctx).
3. После отмены работа должна **быстро** остановиться; игнор ctx = leak + превышение SLA.
4. Не глушите `ctx.Err()` без причины: наверх — `return fmt.Errorf("…: %w", ctx.Err())` или маппинг в gRPC `codes.Canceled` / `DeadlineExceeded`.

## НЕ хранить context в struct

Антипаттерн:

```go
type Service struct {
    ctx context.Context // плохо: lifetime struct ≠ lifetime request
}
```

Почему:

- Context привязан к запросу/операции; сервис живёт весь процесс.
- Легко использовать уже отменённый ctx или наоборот — не отменять никогда.
- Нарушает тестируемость и EH rules (`Do not store context.Context in structs`).

Правильно: `func (s *Service) Do(ctx context.Context, …)`. Долгоживущие поля — клиенты, pool, config; не ctx.

Исключение-исключение: узкий helper с коротким lifetime, создаваемый на один request — всё равно чаще передают явно.

## Value: осторожно

- Не для обязательных параметров (logger, db) — DI в конструкторе (как `internal/app/di.go` в EH).
- Ок: `trace_id`, correlation id для логов (см. `platform/pkg/tracing`).
- Извлечение без ok-check → скрытые nil/wrong type.

## Связь с gRPC / HTTP

- Incoming RPC уже несёт ctx с deadline клиента.
- Исходящие вызовы: тот же ctx или child с меньшим timeout.
- Interceptors (Logger, Recovery, Validate) получают ctx — не подменяйте Background без нужды.

## Типичные вопросы на собесе

- Разница Background / TODO / WithCancel / WithTimeout?
- Почему всегда нужен `defer cancel()` даже после успешного завершения?
- Почему нельзя класть context в struct сервиса?
- Что вернёт `ctx.Err()` после timeout vs cancel?
- Когда допустим `WithValue`, а когда лучше явный аргумент?
- Как отмена ctx должна дойти до SQL-запроса?
