# Go: базовые типы

Краткий cheat-sheet для senior backend: что спрашивают про типы, zero values и алиасы — и где это бьёт в прод.

## Базовые типы

- **Числа:** `int`, `int8`…`int64`, `uint…`, `float32`/`float64`, `complex64`/`complex128`.
- **Булев:** `bool` — только `true`/`false`, без «truthy».
- **Строка:** `string` — неизменяемая последовательность байт (не рун!).
- **Байт / руна:** `byte` = `uint8`, `rune` = `int32` (Unicode code point).
- **Указатели, массивы, слайсы, map, chan, func, interface, struct** — составные типы.

## Zero values (нулевые значения)

Каждый тип имеет zero value — переменная без явного init уже «валидна»:

| Тип | Zero |
|-----|------|
| числа | `0` |
| bool | `false` |
| string | `""` |
| указатель / slice / map / chan / func / interface | `nil` |
| struct | zero каждого поля |

**Питфол:** `var m map[string]int` — `m == nil`, чтение ок (`m["x"]` → 0), запись → panic. Пишите `make` или литерал `{}`.

## string vs []byte vs rune

- `string` — immutable; конверсия `string` ↔ `[]byte` копирует данные (до Go 1.22+ есть оптимизации, но не полагайтесь на «zero-copy»).
- Итерация `for i, r := range s` даёт **индекс байта** и **руну**; `len(s)` — длина в байтах, не в символах.
- `[]rune(s)` — для подсчёта/нарезки по символам; дорого на больших строках.
- JSON/UTF-8 в gRPC/HTTP: битые байты в `string` допустимы в Go, но сломают protobuf/JSON — валидируйте на границе (как `Validate()` в EH proto).

## int vs int64 (и размерность)

- `int` / `uint` — **platform-dependent**: 32 или 64 бита. На amd64/arm64 в EH — 64.
- В API, БД, protobuf, Kafka payload, деньгах/ID — явные `int64` / `int32` / `string` (UUID). Не тащите `int` через wire.
- Overflow при `int`↔`int64` на 32-bit — редкий, но классический вопрос на собесе.
- В Postgres `BIGINT` ↔ Go `int64`; `INTEGER` ↔ `int32`. Несоответствие — silent truncate в драйверах, если неаккуратно сканить.

**Event Horizon:** ID сущностей, цены, счётчики в billing/inventory — фиксированная ширина (`int64` / decimal-строка), не «голый» `int`.

## Type definition vs type alias

```go
type UserID int64        // новый тип: нужна конверсия
type UserID = int64      // alias (Go 1.9+): тот же тип
```

- Новый тип: отдельный method set, нельзя передать `UserID` туда, где ждёт `int64`, без каста — полезно для domain (Clean Architecture `internal/model`).
- Alias: удобен для миграции/переименования без ломания API.
- Не путать с `type MyString string` — это не alias, методы можно вешать.

## Указатели и интерфейсы (коротко)

- Передача struct по значению копирует; большой struct / мутация → pointer.
- `nil` pointer ≠ «пустой» domain object — в EH handlers лучше явные ошибки, чем тихий zero.
- Interface value = (type, value); `(*T)(nil)` в interface ≠ `nil` interface — классический баг в error returns.

## Структуры и теги

- Exported поля с тегами `json:"…"`, `db:"…"` — контракт сериализации.
- Неэкспортированные поля в JSON не видны — для DTO vs domain разделяйте слои (handler ≠ model).

## Типичные вопросы на собесе

- Чем `string` отличается от `[]byte`? Когда конверсия копирует память?
- Что такое zero value у map/slice/channel? Можно ли писать в nil map?
- Почему в protobuf/БД не используют `int`, а `int64`/`int32`?
- Разница `type A int` и `type A = int`?
- Сколько байт в `len("привет")` и сколько рун в `range`?
- Когда `var err error = (*MyError)(nil)` даст `err != nil`?
