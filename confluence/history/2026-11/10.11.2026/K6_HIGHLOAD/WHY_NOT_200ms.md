Почему не 200ms сейчас
Три причины:

applyOwnedFlags — N+1 SQL на shop/items (2 запроса вместо 1). Фикс — batch owned flags (один SQL на все items), или кэш owned в Redis. −30–50% на shop/items.

gateway 3 реплики × sync.Map — кэш ValidateToken не делится. Под P=200 30 cacheSet/s на реплику — не bottleneck, но не идеал.

PG на 8 GiB — shared_buffers дефолт (128MB) мало. inventory_items 293 строки — влезает, но outbox, purchases, inventory — растут. Под нагрузкой — buffer miss → disk → latency.


4.3. Что делать для 200ms
Не сейчас. Но план:

applyOwnedFlags — batch owned flags (один SQL). P=200 shop/items с 3051ms до ~2000ms.

pgbouncer или больше PG max_connections — убрать pool bottleneck.

shared_buffers=1GB — прогреть PG buffer.

pg_prewarm — cold start исчезнет.

Горизонтальный scale — не 3 Gateway, а 6 + больше Shop реплик. Но это не 8 GiB.

go pprof — найти hot spots в Gateway/Shop. go быстрый, но не магия — если N+1 SQL, Go не поможет.

4.4. Что такое 200ms на самом деле
По индустрии:

Google — 200ms p95 для search — цель.

Amazon — 100ms дополнительной latency = −1% продаж (классическое исследование).

Stripe — p99 < 500ms для API.

Shopify — p95 < 500ms для storefront.

Для нашего кейса (all-in-one 8 GiB, P=200 = 4× CORE) — 200ms нереально. P=50 — ~100–200ms реально (с batch owned). P=200 — ~500ms–1s реально (на 8 GiB). Для 200ms на P=200 — нужно 8 vCPU/32 GiB + batch + pgbouncer + 6 Gateway.