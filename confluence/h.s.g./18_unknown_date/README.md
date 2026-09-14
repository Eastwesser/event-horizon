# Урок 18 — HTTP `Do()` / proxy transform + concurrency

## Тема
Livecoding вокруг HTTP-клиента/сервера: проксирование upstream JSON с фильтрацией PII, приоритетный worker pool, weather-cache под `RWMutex`, шардированный TTL-кэш.

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_http_users_transform` | Upstream → DTO: убрать password/staticData; PII только если `amount <= 50000` |
| `02_priority_worker_pool` | N воркеров, `Add(fn, prio)`, heap/FIFO, `Stop()` без гонок |
| `03_weather_cache` | Фон-апдейтер + `Get` под `RLock` |
| `04_sharded_ttl_cache` | `[]*shard` + janitor, без `sync.Map` |

## Как решать / что говорить
- **Transform:** два типа (original/public), `omitempty`, timeout на `http.Client`, `defer Body.Close()`.
- **Pool:** «приоритет = max-heap + seq для FIFO; воркеры спят на `sync.Cond`; Stop закрывает приём и ждёт опустошения».
- **Cache:** write path `Lock`, read `RLock`; TTL — либо lazy expire на Get, либо фон-janitor по шардам, чтобы снизить contention.

## Собес-советы
- Не ходи на реальный IP с собеса — подними `httptest` или мок.
- `go test -race` на pool обязателен в голове: mutex на heap, не на fn().
- Бонус: `PendingCount`, идемпотентный `Stop`, `StopWithTimeout`.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

