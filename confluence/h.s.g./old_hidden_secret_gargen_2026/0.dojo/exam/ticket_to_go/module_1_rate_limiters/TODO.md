# Module 1 — Rate Limiters — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_1_rate_limiters/`

| # | Папка | Статус | Что сказать на собесе (1 фраза) |
|---|--------|--------|--------------------------------|
| 1 | `1.task_warmup` | [x] | замыкание держит переменную цикла → передавай `i` аргументом |
| 2 | `2.task_leetcode` | [x] | Two Sum: map complement, O(n) |
| 3 | `3.task_concurrency` | [x] | это **semaphore/worker pool** (параллельность), не rate limit (частота) |
| 4a | `4…/1.fixed_window` | [x] | просто, но burst на границе окна ≈ 2N |
| 4b | `4…/2.sliding_window` | [x] | log timestamps — точно, память O(лимит) |
| 4c | `4…/3.token_bucket` | [x] | capacity + refill rate + `Allow()`; burst после простоя |
| 4d | `4…/4.leaky_bucket` | [x] | ровный выходной rate, сглаживает burst |
| 5 | `5.task_sql` | [~] | sql файл есть — прогнать глазами перед собесом |
| 6 | `6.task_legend` | [~] | легенда AdTime/Roolz — связать с Prydwen при необходимости |

**Trade-off шпаргалка:** Fixed → просто/burst; Sliding log → точно/память; Token Bucket → API default; Leaky → равномерный drain.

Запуск:
```bash
export GOWORK=off
cd …/module_1_rate_limiters/<task>
go run main.go
```
