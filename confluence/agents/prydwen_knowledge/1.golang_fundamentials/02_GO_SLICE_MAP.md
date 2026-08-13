# Go: slice и map

Senior-фокус: header слайса, capacity, aliasing при append, concurrent map — самые частые баги в проде.

## Slice: что внутри

Слайс — дескриптор из трёх полей:

- **ptr** — указатель на underlying array
- **len** — видимая длина
- **cap** — ёмкость до реаллокации

Передача слайса в функцию копирует **header**, не массив. Мутация `s[i] = …` видна вызывающему; `append`, уехавший за cap, может выделить новый массив — вызывающий старый header не увидит.

```go
func appendOne(s []int) { s = append(s, 1) } // локальный header!
```

Возвращайте новый слайс или передавайте `*[]T`, если нужно изменить len у caller.

## make, nil vs empty

| | nil slice | empty non-nil |
|--|-----------|---------------|
| объявление | `var s []T` | `s := []T{}` / `make([]T, 0)` |
| `s == nil` | true | false |
| `len`/`cap` | 0 | 0 / ≥0 |
| JSON | `null` | `[]` |

**EH / API:** для списков в OpenAPI/JSON обычно хотят `[]`, не `null` — инициализируйте `make([]Item, 0)` в handler/service, если элементов нет.

## append и capacity growth

- Если `len < cap` — append пишет в тот же массив (возможен **aliasing** с другими слайсами того же backing store).
- Если не хватает — новый массив, обычно ~2× (для больших — меньший рост). Не полагайтесь на точный множитель.
- `append(s[:0:0], s...)` / `slices.Clone` — явная копия, чтобы отвязаться от чужого массива.
- `s = s[:0]` — reuse буфера (как в hot path парсеров); не отдавайте такой слайс наружу без копии.

**Питфол:** `a := b[:1]; a = append(a, x)` может перезаписать `b[1]`, если cap общий.

## Subslice и memory leak

`big := make([]byte, 1<<20); small := big[:10]` держит весь megabyte живым через ptr. Для долгоживущего `small` — `clone := append([]byte(nil), big[:10]...)`.

## Map

- Hash table; порядок итерации **рандомизирован** (с Go 1.0+ не стабилен).
- Zero: `nil` map — read OK, write panic.
- `delete(m, k)` на отсутствующем ключе — no-op.
- Lookup: `v, ok := m[k]` — различайте missing и zero value.

### nil vs empty map

| | nil | `map[K]V{}` |
|--|-----|-------------|
| write | panic | ok |
| JSON | `null` | `{}` |
| `len` | 0 | 0 |

## Map и concurrency

**Правило:** concurrent read+write на map → data race, часто fatal `concurrent map writes`.

Варианты:

1. `sync.RWMutex` вокруг map (часто в кэшах).
2. `sync.Map` — для особого паттерна (ключ пишется раз, читается часто; или disjoint keys). Не «дефолт вместо mutex».
3. Шардирование map’ов / channel ownership (одна горутина владеет map).
4. Immutable copy-on-write под RLock.

В EH: сессии auth в Redis, а не в process-local map — избегаем shared mutable state между инстансами. In-memory кэши (если есть) — только с mutex или per-request.

## Tip: ключи map

- Ключ должен быть comparable: нельзя slice, map, func.
- Struct-ключи сравниваются по полям — осторожно с float/`[]byte` внутри.
- Для `[]byte` ключей — `string(b)` (копия) или `m[string(b)]` с пониманием lifetime.

## Типичные вопросы на собесе

- Что хранится в slice header? Почему append «не видно» снаружи?
- Чем опасен append в общий capacity с другим слайсом?
- nil slice vs empty: JSON, `== nil`, append?
- Почему нельзя конкурентно писать в map? Чем заменить?
- Когда выбирать `sync.Map`, а когда mutex + map?
- Как subslice может «держать» гигабайтный буфер?
