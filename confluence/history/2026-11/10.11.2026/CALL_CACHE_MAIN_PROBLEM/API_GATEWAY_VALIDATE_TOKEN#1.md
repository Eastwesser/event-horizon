## Результат (10.11.2026, кэш ValidateToken 1s)

### Ручной тест
- 1-й запрос (slow): 680ms
- 2-й запрос (fast): 66ms
- **×10 быстрее** — кэш работает.

### P=50 — прогрев подтверждён
| Прогон | browse p95 | purchase p95 | Failed | 5xx |
|---|---|---|---|---|
| Cold (первый) | 5212ms | 7332ms | 0.79% | 0.49% (inventory CB) |
| Warm (второй) | 3283ms | 3769ms | 0.64% | 0.00% |
| Warm (третий) | **3137ms** | 3838ms | 0.51% | 0.00% |

**Вернулось** к baseline (3166ms). Cold start был проблемой, **не** кэш.

### Метрики Auth (cumulative, не rate)
- `grpc_request_duration_seconds_count{ValidateToken}` = 82180 (за всё время).
- `sum/count` = **24ms** средний ValidateToken (HMAC + Redis).

### Вывод
- **Кэш полезен** — ручной тест ×10, warm P=50 = baseline.
- **Cold start** — PG buffer + Redis + 3× sync.Map. **Не** баг, **физика**. Первые 30–60s медленно, потом baseline.
- **`circuit inventory`** открылся на cold start (5 timeout подряд), **закрылся** через 10s half-open.

### Что дальше
- P=100 / P=200 warm — проверить эффект кэша на больших профилях.
- `pg_prewarm` для inventory_items (не в этом спринте).

Что реально bottleneck (по breakdown)
text
P=200 browse breakdown:
  shop/items      : 8503ms  ← ×1.8 хуже (было 4328ms после индексов)
  inventory/items : 7052ms  ← ×1.7 хуже (было 4173ms)
  leaderboard     : 3286ms  ← ×1.5 хуже (было 2218ms)
  profile         : 3152ms  ← ×1.5 хуже (было 2119ms)
Все 4 ручки выросли одновременно. Не одна. Не ValidateToken (он был бы только в browse, но все выросли). Значит — общая задержка стека: Gateway → Nginx → Balancer → Gateway → gRPC → сервис → PG/Redis.

Гипотеза: Nginx или Balancer стали bottleneck. Nginx worker_connections 4096 — хватает. Balancer — least-connections, 2 реплики, должны тянуть. Но — keepalive 32 в Nginx upstream — только 32 соединения на Balancer. 2 Balancer × 32 = 64 keepalive. 560 rps через 64 keepalive — 8.75 rps на соединение. Должно хватать (keepalive reuse). Но — если запрос долгий (10s), соединение занято 10s, 64 соединений × 10s = 640 rps макс. 560 < 640 — на грани.

Это гипотеза. Нужно проверить: balancer_active_connections{backend="..."} — сколько активных на Gateway? Если 256 (max) — bottleneck в bulkhead. Если 32 — в keepalive.