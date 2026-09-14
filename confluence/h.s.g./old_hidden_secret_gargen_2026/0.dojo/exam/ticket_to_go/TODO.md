# ticket_to_go — статус всего dojo

Все 6 модулей доведены до runnable haven-формата.

| M | Паттерн | Папка | TODO |
|---|---------|--------|------|
| 1 | Rate Limiter | `module_1_rate_limiters/` | есть |
| 2 | Load Balancer | `module_2_load_balancers/` | есть |
| 3 | Circuit Breaker | `module_3_circuit_breakers/` | есть |
| 4 | Health Checker | `module_4_health_checkers/` | есть |
| 5 | Retry + Backoff | `module_5_retryers/` | есть |
| 6 | LRU / Cache | `module_6_caches/` | есть |

**Связный ответ на middle+:**  
Rate limit на входе → LB между инстансами → CB+Retry на downstream → Health/ready → Cache на чтение.

Слои выше dojo уже собраны: twenty-strikes 20/20, company packs, master [`MANUSCRIPT.md`](../../../../MANUSCRIPT.md). Дальше — прогон вслух по ритуалу 90 мин.
