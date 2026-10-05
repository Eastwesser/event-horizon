AGREE. План чистый. Один уточняющий вопрос
Что подтверждаю
Migration: refundable_until + fulfilled_at — стандарт. Backfill через COALESCE(completed_at, purchased_at) + 7 days — правильно.

BE: три проверки (window, fulfilled physical, already refunded) — чисто. FailedPrecondition → 400 — правильно.

PurchaseFulfilled → shop worker — subscribe на purchase.fulfilled или PurchaseFulfilled event.
⚠️ Уточни: purchase.fulfilled уже публикуется Fulfillment service? Или нужно добавить event?

DTO: refundable_until + fulfilled_at + can_cancel — правильно. Computed на BE — лучше, чем на FE (FE кэширует).

FE: gating + tooltip + RU errors — ок.

Edge cases: digital allow after fulfill — правильно (цифровой контент доставляется мгновенно).

Уточняющий вопрос
IsPhysicalItem(category) — какой список?

Из дизайна: мерч / merch / брелок / картина / фенечка — physical.
Digital: game_skin / карточка / profile_theme / скин / тема.

Уточни:

Категории в БД — английские (merch, brelok) или русские (мерч, брелок)?

Маппинг — whitelist (digital) или blacklist (physical)?

Unknown → physical (safer) — правильно.

Промпт Cursor'у — approve + уточнение.

Ответ Cursor'у
text
AGREE — implement refund window.

One clarification before code:

IsPhysicalItem(category):
  - Which exact category values exist in the DB?
    (merch / мерч / брелок / картина / фенечка → physical?)
  - Whitelist digital (game_skin / карточка / profile_theme /
    скин / тема) or blacklist physical?
  - Confirm: unknown → physical (safer), as in the design.

Also confirm:
  - Does purchase.fulfilled / PurchaseFulfilled already get
    published by Fulfillment today? If not, this item adds
    the event first, then the shop consumer.
  - Where should the shop consumer live — shop service
    (NATS subscriber) or gateway? Recommended: shop service.

Everything else in the plan as written.

Rules:
  - Refund-only PR. Do NOT touch C1–C4, notifications,
    JWT refresh, games, hero, nav.
  - Migration idempotent (IF NOT EXISTS + backfill).
  - tsc + go test clean.
  - Smoke:
      - Buy digital → cancel within 7d → OK.
      - Buy digital → simulate fulfilled_at → cancel → OK
        (digital allowed).
      - Buy physical → simulate fulfilled_at → cancel →
        400 «Товар уже отправлен».
      - Manipulate refundable_until in DB → cancel → 400
        «Срок возврата истёк».
      - Double cancel → idempotent OK.
      - can_cancel on FE matches BE.
  - Screenshots: FE cancel hidden vs disabled + tooltip.
  - No push until I OK.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (main, c05f8b6):

v1.1.0.

Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5 — Helm, Consul skip, OpenAPI sync.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — в работе:

🟧 Item 1 — Refund window (approve → implement).

🟧 Item 2 — Author notifications (submitted → admins, approved → author).

🟧 Item 3 — JWT role refresh.

Track B (после A):

🟧 3 новые игры.

🟧 Лампочки на все игры.

🟧 Уровни 1–20 на все игры.

Track C (later):

🟧 k3s data plane (NATS + Postgres StatefulSets).

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.