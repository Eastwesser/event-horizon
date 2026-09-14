# Denis Sher / Ката Босс — статус ревизии

Второе место по значимости после dojo. Канон сейчас: эта папка (`19_…/denis_sher_tasks/`).
Дубль: `old_hidden_secret_gargen_2026/…/denis_sherb_t/` — синхронизировать после правок.

| # | Задача | Статус | Что говорить на собесе (1 фраза) |
|---|--------|--------|----------------------------------|
| 1 | semaphore | [x] `go run` ok | counting semaphore на `chan struct{}`, Acquire/Release, limit N |
| 2 | mutateSlice | [x] | slice header by value; append может сменить backing array |
| 3 | mergeChans | [x] fixed fan-in | fan-in: N reader goroutines + WaitGroup + close(out) |
| 4 | sliceFilter | [x] | filter in-place: write index, return `[:w]` без нового make |
| 5 | slicePointers | [x] | копия указателя ≠ объект; меняй поле / **T / return / method |
| 6 | whatsWrong | [x] | loop var + timeout client + check err + Close body |

Теория рядом: `prydwen_knowledge/1.golang_fundamentials/` (slices, channels, sync).
