# Module 4 — Health Checker — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_4_health_checkers/`  
EH: `/health` liveness, `/ready` dependency ping.

| # | Папка | Статус | Что сказать на собесе |
|---|--------|--------|------------------------|
| 1 | `1.task_warmup` | [x] | select: результат vs `time.After` / deadline |
| 2 | `2.task_leetcode` | [x] | Linked List Cycle: Floyd slow/fast |
| 3 | `3.task_concurrency` | [x] | N сервисов = N probe-loop + общий ctx |
| 4a | `1.first` | [x] | ticker + `WithTimeout` на каждый check + RWMutex |
| 4b | `2.second` | [x] | exponential backoff пока unhealthy |
| 5 | `5.task_sql` | [~] | uptime %; 99.9% ≈ 43 мин/мес |

**liveness vs readiness:** процесс жив ≠ готов принимать трафик (БД/Redis down → 503 ready).

Запуск: `export GOWORK=off && go run main.go`
