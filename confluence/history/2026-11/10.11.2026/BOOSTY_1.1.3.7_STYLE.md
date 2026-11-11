🎮 Event Horizon v1.1.3.7 — highload, deadlock и 8 GiB

Привет, друзья!

Последние две недели я не писал посты — копался в кишках Event Horizon.
Под капотом 17 сервисов, NATS JetStream, gRPC mesh, Gateway ×3, Balancer ×2, Nginx.
И всё это на виртуалке 8 GiB. Спойлер: 8 GiB — это потолок.

Что было:
— k6 highload: P=50 → 76% failed (rate limiter на одном admin-токене).
  Написал seed-users.js + пул токенов → P=50: 0.67% failed.
— P=200: deadlock 40P01 в Shop (purchase и cancel брали локи в разном порядке).
  Фикс: items → inventory → purchases, lock_timeout 2s. Deadlock исчез.
— N+1 в Shop: applyOwnedFlags делал 2 SQL на shop/items. Batch owned → 1 SQL.
  Circuit shop исчез, purchase_owned ×11 лучше.
— Cold start: после ребилда PG buffer пустой → circuit inventory.
  Фикс: pg_prewarm + warmup.
— Pool starvation: SetMaxOpenConns(25) на 11 сервисов. 100/64/32 → inventory/items ×3.4.
— Nginx: proxy_read_timeout 30s → p99 36s. 5s → p99 8s.
— Bulkhead: MaxConcurrent 32 → 256/128/64/32.
— ValidateToken кэш 1s → ×10 быстрее.

Результат:
P=100 — 6044ms browse, 0% errors, 0% 5xx. Baseline 8 GiB.
P=150 — 7359ms, 0% errors, 0% 5xx. Стресс.
P=200 — 14027ms, 4% errors (contention на 9 items, не код), 0.94% 5xx.
Для 200ms на P=200 — нужно 32 GiB. Roadmap v1.2.

Что дальше:
— ORDER-сервис (корзина, заказы, резерв на время корзины) — v1.2.
— Batch owned, pgbouncer, 6 Gateway — v1.2.
— 6 карточек в Shop — наконец-то.

Спасибо, что вы со мной. Event Horizon растёт. 🚀

P.S. Скайриму 15 лет.