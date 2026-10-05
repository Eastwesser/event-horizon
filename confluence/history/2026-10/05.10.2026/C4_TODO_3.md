OK. Старт Wave 4 (1) + (2) параллельно
Что подтверждаю
Baseline — печальный, но правдивый:

auth 22.7%, inventory 7.2%, shop 2.0%, authors 0%, gateway 8.3%.

coverage-gate.sh — есть, но не в CI.

COVERAGE.md — нет.

k6 — outdated (loadtest.js с устаревшими путями).

Indexes — нет purchases(item_id).

Helm / Consul — нет.

CI — build/push только, без go test.

Это нормально для MVP. Не стыдно. Wave 4 — исправляем.

План — принимаю
(1) Unit tests — порядок правильный
Authors → Shop → Inventory → Auth — high-risk сначала.

Мой комментарий:

Authors (0%) — сначала C1/C2/C3 logic Submit/Approve/Reject/Revert/UpsertProfile.

Shop — PurchaseItem / CancelPurchase (idempotent) — критично (деньги).

Inventory — GetByAuthor / SoftDelete / Restore / Reserve / Release — ownership важен.

Auth — Logout / RefreshToken / UpdateRole / ListUsers / register forced role=user — security.

Gateway — handlers thin, middleware частично покрыт. Не в gate — ок, но документировать exclusion.

(2) k6 — scenarios правильные
Split — правильно:

browse.js, purchase.js, leaderboard.js, auth_burst.js.

loadtest.js — fix paths или выкинуть.

Indexes — после k6 + EXPLAIN, не speculative. ✅ Правильно.

Infra
COVERAGE.md — создать.
coverage-gate.sh — extend, wire в CI после ≥70% on priority 4. ✅ Правильно (не ломать main early).

Мой ответ Cursor'у
text
OK — start Wave 4 (1) + (2) in parallel.

Confirmations:
  - Unit tests: authors → shop → inventory → auth, then
    rest. Mock repo. Service layer first.
  - Gateway: not in gate; document exclusion. Handlers are
    thin; middleware already partially covered.
  - COVERAGE.md at confluence/tech_debt/CURRENT_DEBT/.
  - coverage-gate.sh: extend; wire CI after ≥70% on the
    priority 4 services (do NOT break main early).
  - k6: split into browse.js / purchase.js / leaderboard.js /
    auth_burst.js. Fix or drop loadtest.js.
  - Indexes: only after k6 + EXPLAIN. Report candidates
    before adding.
  - Helm + Consul + OpenAPI sync: deferred to Wave 4.5
    (after 1+2).

Rules:
  - Unit tests-only + k6-only PRs. Split commits by service
    and by k6 script.
  - Do NOT touch production code to make tests pass. If a
    test reveals a bug — fix it in a separate small commit
    and flag it.
  - tsc + go test clean.
  - No push until I OK.

Process:
  - Start with authors tests (0% — biggest gap) + k6
    browse.js + purchase.js.
  - Report: coverage delta per service after each PR.
  - Then continue with shop, inventory, auth.

DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
housekeeping.
Что я держу в карте
Закрыто (в origin/main, 6f75867):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Deferred:

🟧 C4 — Payouts.

🟧 Author notifications.

🟧 JWT role refresh.

🟧 Wave 2 #5b.

🟧 Refund window.

🟧 backfill-noiz-reviews.py.

🟧 Cursor pull-in.

В работе (Wave 4, план approved):

🟧 Unit tests:

authors (0%)

shop (2%)

inventory (7%)

auth (22%)

потом: billing, payment, profile, game, leaderboard, history, analytics.

gateway — document exclusion.

COVERAGE.md.

coverage-gate.sh extend.

🟧 k6:

browse.js, purchase.js, leaderboard.js, auth_burst.js.

Fix / drop loadtest.js.

Indexes — после замеров.

Wave 4.5 (deferred):

🟧 Helm charts.

🟧 Consul evaluate.

🟧 OpenAPI sync.