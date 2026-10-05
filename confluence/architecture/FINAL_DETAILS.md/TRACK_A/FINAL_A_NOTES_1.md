OK на оба. Разбор + старт Track A #2
Track A #1 — commit
Smoke 5/5 — pass. ✅

Digital cancel 200.

Double cancel → already_refunded 200.

Digital + fulfilled → 200 (allowed — правильно).

Physical + fulfilled → 400 «Товар уже отправлен».

Past window → 400 «Срок возврата истёк».

Migration columns present.

Коммит — OK. Message Cursor дал — хороший.

Пуш — сам (standing rule).

Track A #2 — plan approved с одним уточнением
Ответы — понятны:

Notification сейчас — только ops Telegram consumer. Нет proto / DB / API / repo.

FE bell — нет.

Channels — ops Telegram или log-skip. No per-user email.

Authors events — не публикуются. Approve outbox'ит author.upserted (источник application_approved).

Proposed scope — правильный:

Authors emit author.application.submitted + .approved.

Notification service — новый inbox (Postgres notifications + consumer).

Gateway GET /api/notifications, POST /:id/read.

FE — bell + dropdown + unread badge.

Один уточняющий вопрос
author.application.submitted — куда? По схеме:

Apply — auth user подаёт.

Notif — каждому admin.

Как найти admins? Через Auth ListUsers (filter role=admin)? Или через auth DB?

Notification service не должен лезть в auth DB. Варианты:

NATS request-reply к Auth (есть ListUsers RPC).

Или — Authors service эмитит event с admin user_ids в payload (Authors тоже не имеет доступа к auth).

Или — gateway обогащает event admin_ids (gateway уже умеет ListUsers).

Мой выбор — gateway обогащает. Причина: gateway уже имеет Auth client + ListUsers. Notification остаётся domain-only.

Промпт Cursor'у:

text
OK on both:

1. Commit Track A #1 — go ahead with your message.
   I'll push myself.

2. Track A #2 plan approved with one clarification:

   author.application.submitted → notify each admin.
   How does Notification know who the admins are?

   Options:
     a) Notification → Auth.ListUsers (gRPC). Adds Auth
        client to Notification.
     b) Authors emits event WITH admin user_ids in payload.
        But Authors doesn't have access to Auth either.
     c) Gateway enriches the event: on POST /apply, gateway
        calls Auth.ListUsers(role=admin), emits
        author.application.submitted with admin_ids.
        Notification stays domain-only.

   Recommended: (c) — gateway already has Auth client +
   ListUsers; Notification stays simple.

   Confirm (c) or explain why (a) is better.

Scope confirmed otherwise:
  - New Postgres notifications table.
  - Consumer: submitted → insert per admin; approved →
    insert for applicant.
  - GET /api/notifications, POST /:id/read.
  - FE bell + dropdown + unread badge in navbar (allowed
    exception).
  - Skip ops Telegram for these; ops bot stays purchase-only.
  - OpenAPI sync.

Rules:
  - Notification-only PR.
  - Do NOT touch refund window, C2 approval logic, Auth
    JWT refresh, games, hero, disk, seed.
  - tsc + go test clean.
  - Smoke: submit → admin inbox row; approve → author
    inbox row; mark-read works; unread badge matches DB.
  - No push until I OK.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (main, c05f8b6):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Локально (не запушено):

Track A #1 — Refund window (smoke 5/5 pass, commit approved).

Track A — в работе:

✅ Item 1 — Refund window (commit → push).

🟧 Item 2 — Author notifications (plan approved, уточнение).

🟧 Item 3 — JWT role refresh.

Track B (после A):

🟧 3 новые игры, лампочки, уровни.

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.