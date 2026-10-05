Мой выбор — Option 1 (Wave 4 infra) + D (defer C4)
Обоснование
Q1–Q4 — D. Defer C4
Причины:

Нет пользователей — payouts не срочны.

Monetization — бизнес-решение, не техническое. Ты должен подумать спокойно, не в спешке.

Refund после payout'а — юридически сложно (вычитать из зарплаты? Клеймить долг?).

Kids-safe — этический вопрос (получать деньги за продажу детям). За тебя не решить.

C1–C3 — уже самостоятельная фича (авторы подают, админы approve'ят, авторы управляют). Payouts — отдельно.

Wave 3 закрыта на 3/4. C4 → deferred.

Option 1 — Wave 4 infra (мой выбор)
Почему не Option 2 или 3:

Option 2 (follow-up'ы):

Notifications (author.application.submitted → admins) — нужно тебе, если ты один модератор. Пока не критично.

JWT role refresh — UX-gap, не блокер.

Option 3 (housekeeping):

Wave 2 #5b (shop/boost achievements) — nice-to-have, не критично.

Refund window — дизайн готов, но не горит.

backfill-noiz — одноразовый фикс.

Wave 4 — что даёт:

Юнит-тесты ≥ 70% — защита от регрессий. У тебя 50+ сервисов + UI. Одна ошибка — поломка в прод.

k6 нагрузка + индексы БД — ты целишься в 10k RPS. Без замеров — не узнаешь, где узкое место.

Helm charts + Consul — production-grade деплой. Демка на собесе — усилит.

OpenAPI docs — gateway-only решено. Просто закрыть checkbox.

Для портфолио Wave 4 — сильный аргумент: «у меня не только фичи, но и инфраструктура, тесты, нагрузочное».

Промпт Cursor'у
text
Q1–Q4: lock D — defer C4 (payouts) until monetization is
locked. Wave 3 = C1–C3 done; C4 deferred.

Next: Option 1 — Wave 4 infra.

================================================================
WAVE 4 — INFRASTRUCTURE (parallel, M–L)
================================================================

Plan first, no code.

Items (in this order):

1. UNIT TESTS ≥ 70% (M, long-running, parallel)
   - Coverage per service:
       auth, game, billing, leaderboard, shop, inventory,
       profile, payment, authors, history, analytics,
       gateway.
   - Use Go stdlib testing + testify where needed.
   - Start with the highest-risk services:
       * auth (roles, JWT, register)
       * inventory (author_id, soft-delete, reserve/release)
       * shop (purchase, cancel, refund)
       * authors (C1/C2/C3 logic)
   - Target: ≥70% coverage on service layer.
   - Coverage gate: scripts/coverage-gate.sh exists —
     extend to fail CI below threshold.
   - Track per-service progress in a small doc:
     confluence/tech_debt/CURRENT_DEBT/COVERAGE.md.

2. k6 LOAD + DB INDEXES (M–L)
   - k6 scenarios in deployments/k6:
       - Browsing: GET /api/shop/items (list DTO),
         /api/inventory/items.
       - Purchase flow: shop → cancel (idempotency check).
       - Leaderboard read (hot path).
       - Auth: login burst.
   - Measure RPS / latency / error rate.
   - Report top slow queries + suggest indexes.
   - Apply indexes via migrations only.

3. HELM CHARTS FOR k3s (M, DevOps)
   - One chart per service + infra (postgres, redis, nats,
     clickhouse, grafana, prometheus, jaeger, alertmanager).
   - Reuse existing k3s manifests as base.
   - Values files: dev / prod.

4. SERVICE DISCOVERY — CONSUL (M)
   - Evaluate whether Consul is actually needed now
     (k3s already has DNS-based discovery via Services).
   - If yes — install, register services, wire gateway.
   - If no — document why and skip. Do not add infra for
     the sake of it.

5. OpenAPI DOCS (M, low)
   - Gateway-only already decided.
   - Ensure docs/openapi.yaml is up-to-date with all public
     routes (auth, game, billing, leaderboard, shop,
     inventory, profile, payment, authors, history,
     analytics).
   - No per-service Swagger UI.

Process:
  - Pick the highest-value item first — my recommendation:
    (1) unit tests, in parallel with (2) k6.
  - Show the plan for (1) + (2) first — which services,
    which scenarios, which queries.
  - Wait for my OK before code.

DO NOT TOUCH: C1/C2/C3 (Wave 3 done), C4 (deferred),
hero, nav, disk, games, seed, refund window, other
housekeeping.
Что я держу в карте
Закрыто (в origin/main, 6f75867):

v1.0.9.

Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Deferred:

🟧 C4 — Payouts (до monetization).

🟧 Author notifications.

🟧 JWT role refresh.

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window 7d.

🟧 backfill-noiz-reviews.py.

🟧 Cursor pull-in.

В работе (Wave 4):

🟧 Unit tests ≥ 70%.

🟧 k6 + индексы.

🟧 Helm charts.

🟧 Consul (evaluate).

🟧 OpenAPI sync.