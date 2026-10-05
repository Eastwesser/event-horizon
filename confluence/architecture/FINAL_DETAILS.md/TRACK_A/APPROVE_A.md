Ответ Cursor'у — approve + старт Track A #2
Что было — рекап
Track A #1 Refund window — реализован локально:

✅ Migration refundable_until + fulfilled_at.

✅ RefundPurchase — enforce 7d + physical+fulfilled block.

✅ Inventory DTO: refundable_until, fulfilled_at, can_cancel.

✅ FE: cancel enabled только если can_cancel; disabled + RU tooltip.

✅ Cancel API errors → HTTP 400 (InvalidArgument) + RU messages.

✅ Unit tests зелёные; tsc clean.

✅ Go tests перезапущены — 0.

Уточнения получили:

DB categories: merch (291), game_skin (14) — английские.

IsPhysicalItem: digital whitelist (game_skin, profile_theme, skin, theme, card, карточка), всё остальное → physical.

purchase.fulfilled — уже публикуется Fulfillment. Shop подписан.

Что ответить
Стоп — сначала rebuild + verify, потом commit, потом Track A #2.

Промпт:

text
Track A #1 reviewed. Plan matches spec.

Before commit + push:
  - I'll rebuild + redeploy shop + gateway.
  - Then run smoke (buy/cancel/expired/fulfilled/digital).

After smoke passes locally, I'll commit with your message and
push.

Now — start Track A #2 in parallel: author notifications.

Track A #2 plan first, no code.

Scope:
  1. Event `author.application.submitted` → notify admins.
     - Emitted by Authors service on POST /apply (after
       DB insert, before/with the response).
     - Consumer: Notification service.
     - Delivery: what channels exist today?
       (in-app notification table? email? Telegram?)
     - If only in-app: write a `notifications` row for each
       admin user.
     - If email / Telegram not configured → skip and log.

  2. Event `author.application.approved` → notify author.
     - Emitted by Authors service after approve success
       (before/after Auth.UpdateRole; recommend after, so
       the author is really an author when they receive it).
     - Consumer: Notification service.
     - Delivered to the applicant's user_id.

  3. Notification shape:
     - title, body (RU), link (optional), created_at, read_at.
     - FE: bell / dropdown in navbar? (Check what exists
       today.)

  4. Do NOT touch C2 approvals, refund window, games,
     hero, nav (except the bell if missing).

Questions to answer before code:
  Q1. What's the current Notification service API? Table?
      Existing events? (score.updated? purchase.paid?)
  Q2. Is there an in-app notification UI already (bell /
      dropdown / list)?
  Q3. Do we have email / Telegram sender configured, or
      should we keep in-app only for v1?

Show the plan with answers. No code until I OK.

Rules:
  - Notification-only PR. Do NOT touch refund window,
    approvals, Auth, games.
  - tsc + go test clean.
  - Smoke: submit application → admin gets a notification;
    approve → author gets a notification.
  - No push until I OK.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: C1–C4, hero, nav, disk, games, seed.
Что я держу в карте
Закрыто (main, c05f8b6):

v1.1.0.

Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Локально (не запушено):

Track A #1 — Refund window.

Track A — в работе:

✅ Item 1 — Refund window (verify → commit → push).

🟧 Item 2 — Author notifications (plan → approve).

🟧 Item 3 — JWT role refresh.

Track B (после A):

🟧 3 новые игры, лампочки, уровни.

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.