# Module 2 — Load Balancers — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_2_load_balancers/`

| # | Папка | Статус | Что сказать на собесе |
|---|--------|--------|------------------------|
| 1 | `1.task_warmup` | [x] | range map unordered; concurrent write → Mutex/RWMutex |
| 2 | `2.task_leetcode` | [x] | Valid Parentheses: stack + pairs map |
| 3 | `3.task_concurrency` | [x] | worker pool: close(jobs) + WaitGroup; cancel через context |
| 4a | `1.first` | [x] | Round Robin: `(i+1)%n`, atomic или Mutex |
| 4b | `2.second` | [x] | Weighted RR / Least Conn / IP Hash — когда какой |
| 5 | `5.task_sql` | [~] | range sharding по user_id; риск hot shard |

**Trade-offs:** RR — просто/равномерно при одинаковых нодах; Weighted — разная мощность; Least Conn — долгие запросы; IP Hash — sticky без Redis.

Запуск: `export GOWORK=off && go run main.go` в папке задачи.
