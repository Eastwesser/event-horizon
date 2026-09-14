# Module 3 — Circuit Breaker — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_3_circuit_breakers/`  
Теория рядом: `prydwen_knowledge/4.architecture_patterns/04_ARCH_NETWORK.md`, EH Gateway → 503.

| # | Папка | Статус | Что сказать на собесе |
|---|--------|--------|------------------------|
| 1 | `1.task_warmup` | [x] | recover только в defer той же горутины |
| 2 | `2.task_leetcode` | [x] | Merge Two Lists: dummy head |
| 3 | `3.task_concurrency` | [x] | graceful: `ctx.Done()` / NotifyContext → stop work |
| 4a | `1.first` | [x] | Closed→Open→HalfOpen; Mutex не держать на `fn()` |
| 4b | `2.second` | [x] | Open → **503** fail-fast; Retry ≠ CB |
| 5 | `5.task_sql` | [~] | error_rate по часу — постмортем, не hot-path |

**Связь с EH:** Gateway circuit на критичном gRPC → клиенту 503, не зависание.

Запуск: `export GOWORK=off && go run main.go`
