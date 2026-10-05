Отлично. Wave 4 (1) закрыт наполовину. Продолжаем
Что сделано
Flags — resolved
1. merch_gate:

prod correct — ErrSubscriptionRequired («subscription_required» since Aug 2026).

Тесты stale — fix тестов правильный.

gRPC PermissionDenied → 403 + subscription_required — документирован в OpenAPI / STATUS_CODES.

Ок. Закрыто.

2. Refactor split:

ad23f7b — refactor(shop): dial + NATS sync → app.

15c8c03 — test(shop).

Pushed 6f75867..80a6b63.

Coverage — сильный скачок
Service	Service layer coverage
authors	93.2%
shop	83.9%
inventory	98.5%
auth	86.7% (miniredis sessions)
Priority 4 — все ≥70%.

Coverage gate — всё ещё не в CI (по плану).
k6 — ready, needs live stack.

Что осталось в Wave 4 (1)
Priority 4 закрыт. Следующее:

Billing / payment / profile / game / leaderboard / history / analytics — не покрыты в сервисном слое.

Gateway — documented exclusion (handlers thin).

Что делать
Вариант A — продолжить Wave 4 (1) до полного покрытия
Порядок:

billing — лампочки / билетики operations.

payment — subscription checks (тревожный путь — merch_gate).

profile — read user data.

game — submit score (жизненно важный путь).

leaderboard — hot read.

history, analytics — read-only.

Плюс — wire coverage-gate в CI (сейчас уже priority 4 ≥70%).

Вариант B — закрыть Wave 4 (1) и перейти к (2) k6 + indexes
k6 — smoke сам после make deploy.
Indexes — после EXPLAIN.

Вариант C — смешать A + B
k6 параллельно с оставшимися тестами.

Моя рекомендация — A + wire coverage-gate
Причины:

Priority 4 — критичные сервисы. Остальные — тоже важны, но менее рискованны.

Wire coverage-gate в CI сейчас — не даст coverage упасть в будущем.

billing + payment — деньги. Тесты обязательны.

game — submit score — если сломается, все рекорды потеряются.

k6 — можно параллельно, но требует live stack (твой ход).

Промпт Cursor'у
text
Priority 4 done — all ≥70% service. Coverage gate not in CI
yet, per plan.

Next steps (in this order):

1. Continue Wave 4 (1) — remaining services:
   - billing (lamps / tickets ops)
   - payment (subscription / merch_gate path)
   - profile (user data reads)
   - game (submit score — critical)
   - leaderboard (hot read)
   - history, analytics (read-only)
   Same approach: mock repo, service layer first, target
   ≥70% per service. Report per-service deltas.

2. Wire coverage-gate into CI (now that priority 4 is ≥70%):
   - scripts/coverage-gate.sh already exists (MIN=70).
   - Add a CI job that runs `go test ./... -cover` for the
     services in the gate list.
   - Fail on <70% per package (or set threshold to current
     value + margin).
   - Do NOT include gateway in the gate (documented
     exclusion).
   - Do NOT break main early — merge only when local run
     passes.

3. k6 — I'll run browse.js + purchase.js myself after
   `make deploy`. Indexes — after EXPLAIN, in a separate
   small PR.

Rules:
  - Tests-only PRs (unless a real bug is found — flag it
    separately).
  - tsc + go test clean.
  - No push until I OK.

DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund.
Что я держу в карте
Закрыто (в origin/main, 80a6b63):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — priority 4 (authors 93.2%, shop 83.9%, inventory 98.5%, auth 86.7%).

Wave 4 — в работе:

🟧 Wave 4 (1) — remaining services (billing, payment, profile, game, leaderboard, history, analytics).

🟧 coverage-gate — wire в CI (теперь можно).

🟧 Wave 4 (2) k6 — ready, needs live stack.

🟧 Indexes — после EXPLAIN.

Wave 4.5 (deferred):

🟧 Helm charts.

🟧 Consul evaluate.

🟧 OpenAPI sync.

Deferred (не сейчас):

🟧 C4 (payouts), author notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.