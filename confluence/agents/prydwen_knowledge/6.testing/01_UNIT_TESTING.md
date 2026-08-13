# Юнит-тесты — шпаргалка senior (Go / Event Horizon)

## Зачем на собесе

Юнит-тесты проверяют **чистую бизнес-логику** без сети, БД и брокера. Цель — быстрый фидбек (миллисекунды), детерминизм, понятные кейсы. На собесе ждут: table-driven, моки интерфейсов, реалистичное покрытие, знание что *не* юнитить.

## Table-driven (идиома Go)

Один тест-функция, срез кейсов: `name`, `in`, `want`, `wantErr`. Цикл `for _, tt := range tests { t.Run(tt.name, ...) }`. Плюсы: меньше копипасты, легко добавить кейс, отчёт по имени кейса.

```go
tests := []struct {
    name    string
    in      int64
    want    string
    wantErr bool
}{
    {"zero", 0, "0", false},
    {"negative", -1, "", true},
}
```

Правила: независимые кейсы, без shared mutable state, без `t.Parallel()` если кейсы делят мок с общим ожиданием вызовов.

## Моки и границы

- Мокаем **зависимости через интерфейс** (repository, publisher), не конкретный Postgres.
- Стек EH: `testify/assert|require`, иногда `gomock` / hand-rolled stubs.
- Handler почти пустой → юнитить **service**; converter/JWT — отдельные чистые пакеты.
- Не мокайте то, что дешевле вызвать реально (чистые функции, парсеры).

Антипаттерн: мок всего мира + assertion только «вызвали X» без проверки результата бизнес-логики.

## Coverage — реализм

- 100% строк ≠ качество. Цель: критичные ветки (ошибки, границы, роли, идемпотентность).
- В EH: critical paths покрыты; «везде ≥70%» — ориентир пайплайна, не догма.
- `go test -cover` + фокус на `internal/service`, converters, auth JWT/bcrypt.
- Не гонитесь за coverage на DI/`main`/сгенерированном proto.

## Event Horizon: что реально юнитят

| Зона | Пример | Что проверять |
|------|--------|----------------|
| Converters | `billing/internal/converter` | маппинг model↔proto/DTO, nil, нули, валюты |
| Auth JWT | `auth` service | выпуск/парсинг токена, claims `user\|author\|admin`, expiry |
| bcrypt | cost **12** | `GenerateFromPassword` / `CompareHashAndPassword` на stub-хеше |
| Cached repo | Inventory decorator | hit/miss/invalidate без реального Redis (интерфейс) |
| Circuit | Gateway `internal/circuit` | open → ErrOpen, half-open после timeout |

Конвертеры — идеальный table-driven: много комбинаций полей, нет I/O. JWT — фиксированный секрет из env в тестах (не прод-ключ).

## Пирамида и границы юнита

```
много юнитов → меньше интеграционных → мало e2e/smoke/k6
```

Юнит: service + mock repo. Интеграция: testcontainers PG/Redis. E2E: compose + gateway. Не пихайте Docker в каждый `*_test.go`.

## Чеклист перед MR

- Имена кейсов читаемые (`purchase_insufficient_balance`).
- `require` для фатальных precondition, `assert` для ожиданий.
- Нет flake: нет `time.Sleep`, нет порядка map iteration без фиксации.
- Ошибки через `errors.Is` / доменные типы, не сравнение строк логов.

## Типичные вопросы на собесе

1. Чем table-driven лучше копипасты трёх `TestXxx`?
2. Что мокаете в Clean Architecture, а что нет?
3. Почему 100% coverage — плохая KPI?
4. Как протестировать JWT и bcrypt cost 12 без поднятия Auth в Docker?
5. Где граница юнит vs интеграционный тест для Outbox-worker?
6. Как тестировать converter proto↔model на граничных значениях?
7. `t.Parallel()` — когда опасно с моками?
8. Чем `require` отличается от `assert` в testify?
