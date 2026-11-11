P=50 — вернулось к baseline
text
                    До кэша    Первый P=50 (cold)    Второй P=50 (warm)    Третий P=50
browse p95          3166ms     5212ms               3283ms                3137ms
purchase p95        3463ms     7332ms               3769ms                3838ms
Failed              0.30%      0.79%                0.64%                 0.51%
5xx                 0%         0.49%                0.00%                 0.00%
inventory errors    0%         2.26%                0.00%                 0.00%
Первый P=50 (cold) — circuit inventory открылся (2.26% errors), browse 5212ms. Второй/третий — 0% errors, browse 3137ms. Это ровно baseline (3166ms). Кэш работает, cold start был проблемой, не кэш.

Вывод: кэш полезен (по ручному тесту 680→66ms), не вреден (P=50 warm вернулся к baseline). ✅ Шаг 4 закрыт.

Что показала 9091/metrics — кэш работает
text
grpc_request_duration_seconds_count{method="ValidateToken"} = 82180
sum = 2012s
82180 вызовов ValidateToken за всё время (не за прогон). `sum/count = 2012/82180 = 24ms средний** ValidateToken. Это нормально (HMAC + Redis lookup).

Но — это cumulative (за всё время с рестарта Auth), не rate. Чтобы увидеть rate, нужно rate(...[1m]) в Prometheus. Твоя команда:

text
rate(grpc_client_handled_total{grpc_method="ValidateToken"}[1m])
вернула ничего — потому что метрика называется иначе. Правильно:

text
rate(grpc_requests_total{method="/auth.AuthService/ValidateToken"}[1m])
(это видно из твоего же grep — grpc_requests_total{code="OK",method="/auth.AuthService/ValidateToken"}).

Скинь:

bash
curl -s 'http://localhost:9090/api/v1/query?query=sum(rate(grpc_requests_total{method="/auth.AuthService/ValidateToken"}[1m]))' | jq -r '.data.result[0].value[1]'
Если < 100/s — кэш работает. Если > 500/s — не работает.

Про cold start — это нормально, но можно улучшить
Что было: после рестарта Inventory (миграция индексов) — PG cold buffer, Redis cold cache, Gateway 3 реплики × пустой sync.Map. Первые 30–60 секунд — медленно, circuit inventory открылся (5 подряд timeout), залип на 10s half-open, потом закрылся.

Это не баг — это физика. PG buffer пустой, каждый запрос — disk read. Redis cache пустой, каждый inventory/items — PG. Но через 30s — buffer прогрет, всё быстро.

Что можно улучшить (не сейчас):

pg_prewarm — расширение Postgres, прогревает buffer при старте. CREATE EXTENSION pg_prewarm; SELECT pg_prewarm('inventory_items');

Healthcheck в compose уже ждёт pg_isready, но не ждёт прогрева buffer. Можно добавить warmup в deploy.

Sync.Map — прогреть кэш нельзя без запросов. Но TTL 1s — быстро прогревается (1s × 10 токенов = 10 запросов). Не проблема.

