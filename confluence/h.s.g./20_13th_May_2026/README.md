# Урок 20 — async / mutex / monitor / champions

## Тема
Примеры параллелизма: seq vs WaitGroup+Mutex, graceful monitor, TTL user-cache, чемпионат по шагам.

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_seq_vs_parallel` | 10 задач seq vs parallel под mutex |
| `02_monitor` | ticker + done + WaitGroup (return, не break) |
| `03_ttl_user_cache` | Set/Get/Delete, Get не продлевает TTL |
| `04_steps_champions` | без пропусков дней, max steps, ничья |

## Как решать / что говорить
- **Mutex:** критическая секция минимальна; `defer Unlock`.
- **Monitor:** `break` в `select` не выходит из `for` — нужен `return` или label.
- **Champions:** day0 инициализация → дальше только непрерывные дни → фильтр `days==maxDays` → max steps (несколько userId).

## Собес-советы
- Спроси про soft: эпики, дежурства, flow выкатки (см. md).
- Для кэша уточни: продлеваем ли TTL на read (здесь — нет).

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

