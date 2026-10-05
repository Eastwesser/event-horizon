Push. Wave 4 (1) закрыт полностью
Что сделано — 6 коммитов локально
Coverage deltas — все ≥70%:

Service	Service layer
billing	85.7%
payment	91.1%
profile	78.8%
game	77.6%
leaderboard	100%
history	100%
analytics	100%
Плюс priority 4 ранее (authors 93.2%, shop 83.9%, inventory 98.5%, auth 86.7%).

CI:

coverage-gate.sh — service layer only (не package-wide). Правильно — package totals низкие по design (main + service).

New coverage job in main.yml blocks build / image push.

Gateway — excluded.

Local — OK.

COVERAGE.md — обновлён.

k6 — твой ход (после make deploy).

Один комментарий
accce06 — ports refactor.

Уточни — что за refactor? Это prod-код или test-related?

Если test-related (mocks / ports для DI) — ок.

Если prod (изменение архитектуры) — должно быть отдельным коммитом с явным message.

Скинь git log --oneline -10 — проверю.

Мой ответ Cursor'у
text
Wave 4 (1) reviewed — coverage across all gated services
≥70%, CI wired. Excellent.

Two before push:

1. Confirm accce06 (ports refactor):
   - test-only (DI / mocks) or prod code?
   - If prod → should be a standalone commit with a clear
     message, not bundled with tests.
   Report the diff scope.

2. Update the map:
   - Wave 4 (1) → done.
   - CI coverage gate → active.
   - Note: k6 smoke pending live stack.
   - Note: indexes pending EXPLAIN.

After confirmation:
  - Push the batch (6 commits).
  - Report new HEAD.

Then:
  - Wave 4 (2) k6 — I'll run browse.js / purchase.js after
    make deploy. Report RPS / p95 / error rate.
  - Indexes — separate small PR after EXPLAIN.
  - Wave 4.5 — Helm charts / Consul evaluate / OpenAPI sync.

DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
housekeeping.
Что я держу в карте
Закрыто (в origin/main, 80a6b63):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — priority 4 (authors, shop, inventory, auth).

Локально (пуш после подтверждения):

6 коммитов:

accce06 — ports refactor (⚠️ confirm)

4e3c24e — lb/history/analytics tests

2733f93 — billing/payment

ba5465a — profile/game

3c3610f — CI gate

b82330f — COVERAGE.md

Wave 4 (1) — практически закрыт:

✅ Priority 4 (authors/shop/inventory/auth).

✅ Remaining services (billing/payment/profile/game/leaderboard/history/analytics).

✅ coverage-gate в CI.

✅ COVERAGE.md.

Wave 4 (2) — next:

🟧 k6 smoke — твой ход после make deploy.

🟧 Indexes — после EXPLAIN.

Wave 4.5 (deferred):

🟧 Helm charts, Consul, OpenAPI sync.

Deferred (не сейчас):

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.