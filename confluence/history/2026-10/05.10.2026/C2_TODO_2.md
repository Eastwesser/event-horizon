OK C2. И да — upsert profile при approve
Что подтверждаю
C2 plan — чистый:

Слой	Работа
Authors proto	ListApplications, ApproveApplication, RejectApplication
Authors service	List by status (default pending); approve/reject only if pending
Gateway	GET /applications, POST /:id/approve, POST /:id/reject, admin-only; оркестрирует Auth.UpdateRole
FE	Admin tab «Заявки» + filter + table + actions
Migration	Нет
Notify	Deferred
Ответы:

Gateway orchestration — правильно. Authors domain-only.

Double approve → 400 — правильно (не idempotent).

JWT / role pickup — UX-gap отмечен. Re-login нужен.

Admin tab — существует, добавить «Заявки».

Upsert profile при approve
Мой ответ — ДА, включить в C2.

Причины:

Approval — единственное событие, когда мы получаем portfolio, display_name, motivation. Логично сразу перенести в authors profile.

Author dashboard (C3) будет читать authors profile — если он пустой после approve, C3 придётся дозаполнять.

Fail loudly — если profile upsert сломался, approval тоже revert (transaction) — чисто.

Минимум работы — 2-3 строки в approve handler'е.

Что upsert:

authors.display_name — из application payload.

authors.portfolio — из payload.

authors.user_id — soft link.

authors.created_at, verified_at — timestamps.

Если authors profile уже существует — update.
Если нет — create.

Ответ Cursor'у
text
OK — implement C2.

Confirmations:
  - Gateway orchestrates: Authors approve → Auth
    UpdateRole(author). If Auth fails — fail loudly
    (rollback approval transaction).
  - Double approve / reject → 400 when status != pending.
  - Admin tab «Заявки» with filter + table + actions.
  - Note JWT role refresh as UX gap (re-login needed) —
    add a short follow-up note in the map.

Upsert on approve — YES, include in C2:
  - On successful approve (inside the same transaction):
      * upsert authors profile row for that user_id:
          display_name = application.payload.display_name
          portfolio    = application.payload.portfolio
          bio          = application.payload.motivation
          verified_at  = now()
      * if profile exists → update
      * if not → create
  - If profile upsert fails → whole approve rolls back.

Rules:
  - C2-only PR. Do NOT touch C1, C3/C4, other waves, hero,
    nav, disk, games, seed, refund.
  - Do NOT touch notifications (deferred).
  - tsc + build clean.
  - Smoke:
      - user submits application (C1 path).
      - admin opens /admin → tab «Заявки» → sees pending.
      - approve → status=approved, role=author.
      - login as that user → role=author; /register-author
        redirects.
      - authors profile upserted with display_name /
        portfolio / motivation.
      - reject on another application with note → saved.
      - double-approve → 400.
      - second admin refresh shows updated rows.
  - Screenshots: admin tab, approve success, profile row.
  - No push until I OK.
Что я держу в карте
Закрыто (в origin/main, 7110d1b):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 C1 — Author application.

Wave 3 — в работе:

✅ C1 — 7110d1b.

🟧 C2 — Admin approval (approve → implement).

🟧 C3 — Author dashboard (L).

🟧 C4 — Payouts (L).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

🟧 JWT role refresh — UX-gap (re-login needed).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь OK C2 + upsert profile — implement.

Cursor реализует — proto + service + gateway + FE.

Verify — smoke 8 пунктов.

Скинь — скрин admin tab, approve, profile row.

Push после OK.

Потом C3 (Author dashboard).OK C2. И да — upsert profile при approve
Что подтверждаю
C2 plan — чистый:

Слой	Работа
Authors proto	ListApplications, ApproveApplication, RejectApplication
Authors service	List by status (default pending); approve/reject only if pending
Gateway	GET /applications, POST /:id/approve, POST /:id/reject, admin-only; оркестрирует Auth.UpdateRole
FE	Admin tab «Заявки» + filter + table + actions
Migration	Нет
Notify	Deferred
Ответы:

Gateway orchestration — правильно. Authors domain-only.

Double approve → 400 — правильно (не idempotent).

JWT / role pickup — UX-gap отмечен. Re-login нужен.

Admin tab — существует, добавить «Заявки».

Upsert profile при approve
Мой ответ — ДА, включить в C2.

Причины:

Approval — единственное событие, когда мы получаем portfolio, display_name, motivation. Логично сразу перенести в authors profile.

Author dashboard (C3) будет читать authors profile — если он пустой после approve, C3 придётся дозаполнять.

Fail loudly — если profile upsert сломался, approval тоже revert (transaction) — чисто.

Минимум работы — 2-3 строки в approve handler'е.

Что upsert:

authors.display_name — из application payload.

authors.portfolio — из payload.

authors.user_id — soft link.

authors.created_at, verified_at — timestamps.

Если authors profile уже существует — update.
Если нет — create.

Ответ Cursor'у
text
OK — implement C2.

Confirmations:
  - Gateway orchestrates: Authors approve → Auth
    UpdateRole(author). If Auth fails — fail loudly
    (rollback approval transaction).
  - Double approve / reject → 400 when status != pending.
  - Admin tab «Заявки» with filter + table + actions.
  - Note JWT role refresh as UX gap (re-login needed) —
    add a short follow-up note in the map.

Upsert on approve — YES, include in C2:
  - On successful approve (inside the same transaction):
      * upsert authors profile row for that user_id:
          display_name = application.payload.display_name
          portfolio    = application.payload.portfolio
          bio          = application.payload.motivation
          verified_at  = now()
      * if profile exists → update
      * if not → create
  - If profile upsert fails → whole approve rolls back.

Rules:
  - C2-only PR. Do NOT touch C1, C3/C4, other waves, hero,
    nav, disk, games, seed, refund.
  - Do NOT touch notifications (deferred).
  - tsc + build clean.
  - Smoke:
      - user submits application (C1 path).
      - admin opens /admin → tab «Заявки» → sees pending.
      - approve → status=approved, role=author.
      - login as that user → role=author; /register-author
        redirects.
      - authors profile upserted with display_name /
        portfolio / motivation.
      - reject on another application with note → saved.
      - double-approve → 400.
      - second admin refresh shows updated rows.
  - Screenshots: admin tab, approve success, profile row.
  - No push until I OK.
Что я держу в карте
Закрыто (в origin/main, 7110d1b):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 C1 — Author application.

Wave 3 — в работе:

✅ C1 — 7110d1b.

🟧 C2 — Admin approval (approve → implement).

🟧 C3 — Author dashboard (L).

🟧 C4 — Payouts (L).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

🟧 JWT role refresh — UX-gap (re-login needed).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь OK C2 + upsert profile — implement.

Cursor реализует — proto + service + gateway + FE.

Verify — smoke 8 пунктов.

Скинь — скрин admin tab, approve, profile row.

Push после OK.

Потом C3 (Author dashboard).