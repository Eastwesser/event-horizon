Промпт Cursor'у — короткий
text
Two things:

1. git / CI workflow branch:
   I'll handle the push myself.
   - If gh auth refresh works → push ci/coverage-gate-workflow
     + open PR into main.
   - If not → leave the branch local; we'll do it later.
   Do NOT push on your side. Commit messages only.

   NEXT_STEPS_1.md — I'll decide whether to commit; don't
   include in any future commit unless I ask.

2. k6 — I'll run after `make deploy`:
   - browse.js (RPS, p95, errors)
   - purchase.js (idempotency, p95, errors)
   I'll paste the end-of-run blocks.
   Wait for my numbers before suggesting any indexes.
   No speculative changes.

DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
housekeeping.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

Локально:

ci/coverage-gate-workflow — ветка с 6319083 (CI workflow, не запушена).

NEXT_STEPS_1.md — untracked.

Wave 4 — в работе:

✅ (1) Unit tests — done.

🟧 (2) k6 — твой ход.

🟧 Indexes — после EXPLAIN.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.