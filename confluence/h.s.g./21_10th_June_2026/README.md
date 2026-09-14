# Урок 21 — 9 algo + parallel sum mock

## Тема
Набор алгоритмических задач из `21_9_algotasks.md` + сумма 1000 NetworkRequest (mock Denis Sher).

## Задачи (`code/`)

| Pack | Тема |
|------|------|
| `01_find_triggers` | соседняя разница ≥ threshold |
| `02_unique_goods` | товары только у одного продавца |
| `03_top_k_categories` | top-k + lexicographic tie-break |
| `04_merge_intervals` | merge overlapping |
| `05_transaction_pair` | two-sum индексы |
| `06_config_brackets` | валидация ()[]{} |
| `07_sort_logs` | Dutch flag 0/1/2 |
| `08_unique_traffic` | longest unique window |
| `09_event_bus` | Subscribe/Publish non-blocking |
| `10_parallel_sum` | 1000 req → atomic sum |

## Как решать / что говорить
- Сначала сложность и edge cases вслух.
- Для top-k: count → sort by (-freq, name).
- EventBus: буфер + `select/default`, чтобы один подписчик не стопорил остальных.

## Собес-советы
- `08` — классический sliding window с last-seen индексом.
- `07` — три указателя, in-place без доп. массива.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

