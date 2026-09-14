# Урок 24 — HFLabs / Sber / company tasks (Go extract)

## Тема
Смесь задач из `24_tasks.md`, `24_hflabs_task.md`, `24_sber_task.md`: алгоритмы, HTTP filter, платежные лимиты, ATM, LRU, RLE.

## Задачи (`code/`)

| Pack | Источник |
|------|----------|
| `01_find_target` | подотрезок с суммой |
| `02_xy_distance` | X/Y/O расстояние |
| `03_employees_filter` | `?status=` |
| `04_payments_checker` | daily + max single |
| `05_atm` | купюры RUB/EUR + mutex |
| `06_group_by_length` | groupBy len |
| `07_lru_cache` | map + DLL O(1) |
| `08_rle` | run-length A-Z |
| `09_unique_names` | first unique Name |

## Как решать / что говорить
- **findTarget:** prefix sum + map first index.
- **Payments:** контракты History/Limits отдельно от проведения платежа.
- **ATM:** greedy по номиналам под mutex; отказ если left≠0.
- **LRU:** «HashMap + двусвязный список, get двигает в head».

## Собес-советы
- HFLabs Java-разбор — в README теории не дублируем; фокус на Go-паках.
- RLE: валидация A-Z, count=1 без цифры.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

