# Module 5 — Retry + Backoff — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_5_retryers/`

| # | Папка | Статус | Что сказать на собесе |
|---|--------|--------|------------------------|
| 1 | `1.task_warmup` | [x] | defer LIFO; аргументы считаются сразу |
| 2 | `2.task_leetcode` | [x] | LRU: map + DLL, Get/Put O(1) (ещё Module 6) |
| 3 | `3.task_concurrency` | [x] | retry HTTP 5xx с backoff + ctx cancel |
| 4a | `1.first` | [x] | exponential: `base * 2^i`, maxRetries обязателен |
| 4b | `2.with_backoff` | [x] | + jitter (anti thundering herd); permanent errors не ретраить |
| 5 | `5.task_sql` | [~] | MATERIALIZED VIEW = DB-side cache |

**Связка:** Retry внутри, Circuit Breaker снаружи (или политика команды).  
Идемпотентность: POST оплату без ключа — не ретраить вслепую.

Запуск: `export GOWORK=off && go run main.go`
