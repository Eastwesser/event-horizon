Да, стоит. Но не списком — выбери трек и двигайся
Короткий ответ
Cursor прав — все форки реальные. Но напоминать всё сразу — контрпродуктивно. Он сам предложил 4 трека. Нужно выбрать один — не распыляться.

Обновление карты (после v1.1.0)
Изменения с прошлого раза:

✅ Wave 4.5 — закрыт.

Helm chart — есть (deployments/helm/event-horizon).

Consul — evaluated → skip (CONSUL_EVALUATE.md).

OpenAPI sync — v1.1.0 (cancel, /ready, /api/admin/users).

✅ PR #2 — merged (f9ff55a). CI coverage gate — в main.

✅ Wave 4 → 8/8.

main — c05f8b6.

Что изменилось в deferred:

Wave 4.5 — убрать (закрыто).

PR #2 — убрать (merged).

k6 — closed (correctness).

Осталось в deferred:

🟧 C4 payouts (monetization).

🟧 Author notifications.

🟧 JWT role refresh.

🟧 Wave 2 #5b (shop/boost achievements).

🟧 Refund window (design ready).

🟧 backfill-noiz-reviews.py.

🟧 Cursor pull-in.

🟧 Multi-VU purchase EXPLAIN (only if real slow report).

🟧 Bottleneck (500+ cards).

🟧 3 новые игры.

🟧 Реальные товары / карты.

🟧 k3s data plane — NATS + Postgres StatefulSets.

4 трека — что выбрать
Track 1 — k3s data plane
NATS + Postgres StatefulSets (Helm chart follow-up).

M, infra, сложно.

Зачем: production-grade k3s deploy.

Нужно ли сейчас? Демо уже работает на compose. k3s — nice-to-have.

Track 2 — UX follow-ups
JWT role refresh после author approve.

S, quick.

Зачем: после approve user должен получить role=author без re-login.

Плюс: убирает UX-gap в C2.

Минус: не самое интересное для портфолио.

Track 3 — Product
Refund window (7 дней + fulfilled guard) или author notifications.

Refund window: дизайн готов (REFUND_WINDOW_DESIGN.md). M, фича.

Author notifications: S (NATS event + consumer).

Зачем: завершает Wave 3 (author flow end-to-end).

Плюс: видимый продукт.

Track 4 — Content
More games / lamps / levels.

L, контент.

3 новые игры (5 → 8), лампочки на все игры, уровни 1–20.

Зачем: расширяет платформу.

Плюс: демка становится богаче.

Минус: объём.

Моя рекомендация
Порядок — Product → UX → Content → k3s.

Причины:

Product (Refund window + notifications) — завершает Wave 3. C1–C3 сделаны, но без notifications автор не узнаёт про approve. Refund — фича из Wave 2 дизайна. Вместе — S–M, закрывает два deferred пункта.

UX (JWT role refresh) — S, quick win.

Content (3 игры + лампочки + уровни) — большой объём, но самый заметный для демо.

k3s data plane — отложить. Demo работает на compose. k3s — если нужен production.

Промпт Cursor'у
text
Yes — the forks are real, but let's not split attention.
Pick in this order:

TRACK A — Product closure (start here):
  1. Refund window 7d + fulfilled guard — implement from
     REFUND_WINDOW_DESIGN.md (already designed).
  2. Author notifications:
       - author.application.submitted → notify admins
         (NATS → Notification).
       - author.application.approved → notify author.
  3. UX: JWT role refresh after author approve (avoid
     re-login).

TRACK B — Content (after A):
  - 3 new games (5 → 8).
  - Lamps as boosts on remaining games.
  - Levels 1–20 on remaining games.

TRACK C — Infra (later, only if needed for prod demo):
  - k3s data plane: NATS + Postgres StatefulSets in Helm.

Skip for now:
  - C4 payouts (monetization not locked).
  - Wave 2 #5b (shop/boost achievements).
  - Multi-VU purchase EXPLAIN (no real slow report).
  - Bottleneck (280 < 500 cards).
  - Cursor pull-in.
  - backfill-noiz-reviews.py (one-shot fix, minor).

Start TRACK A, item 1: Refund window.
Plan first — no code.
Show:
  - migration shape (refundable_until, fulfilled_at)
  - BE check in CancelPurchase
  - FE gating on /shop detail + /shop inventory
  - error messages (RU)
  - edge cases (digital vs merch, refund after fulfilled)
Wait for my OK.
Что я держу в карте
Закрыто (main, c05f8b6):

v1.1.0.

Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5 — Helm, Consul skip, OpenAPI sync.

PR #2 merged — CI coverage gate.

Hero Plan A, VOID v3, smooth scroll.

k6 (correctness + browse baseline).

Следующее — Track A (Product):

🟧 Refund window 7d + fulfilled guard.

🟧 Author notifications (submitted → admins, approved → author).

🟧 UX: JWT role refresh.

Отложено:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.

🟧 3 новые игры, лампочки, уровни (Track B).

🟧 k3s data plane (Track C).