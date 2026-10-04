OK. Реализуй C1
Что подтверждаю
Register — правильная находка
Hole: anyone can POST /api/auth/register with "role":"author" and skip approval. FE is clean; API is not.

Это критическая дыра. user может стать author через API без approval. Fix — обязателен.

Fix:

Auth Register — force role=user, ignore/reject non-user.

Drop author из validPublicRoles.

Update tests (TestRegisterAuthorRole должен ожидать reject).

Gateway — strip/ignore role в body.

C1 plan — чистый
Auth (extra #1) — public register → always user.

Authors migration — author_applications.

Proto — SubmitApplication, GetMyApplication.

Service — idempotent, reject if already author/admin.

Gateway — POST /api/authors/apply, GET /api/authors/me/application.

FE — /register-author, form, confirmation.

Extras locked: close Register → author, no notifications в MVP.

Out of C1: Admin «Заявки», UpdateRole, notifications.

Ответ Cursor'у
text
OK — implement C1.

Confirmations:
  - Auth: force role=user on public register, drop author
    from validPublicRoles, update tests.
  - Migration: author_applications (UUID, soft user_id,
    status, payload JSONB, created_at, reviewed_*, note).
    Unique partial index: one pending per user_id.
  - Proto: SubmitApplication, GetMyApplication. Payload:
    display_name, portfolio, motivation, contact_email.
  - Service: idempotent (existing pending → return it);
    reject if JWT role is author/admin.
  - Gateway: POST /api/authors/apply (RequireAuth),
    GET /api/authors/me/application (RequireAuth).
    OpenAPI sync.
  - FE: /register-author route, form → confirmation,
    authorsApi.apply / getMyApplication. Link from
    Profile or Authors.

Rules:
  - C1-only PR. Do NOT touch C2 (admin approvals), other
    waves, hero, nav, disk, games, seed, refund.
  - Migration must be idempotent — safe to re-run.
  - After implement:
      * tsc + build clean.
      * Smoke:
          - Register with role=author → rejected (400).
          - Login as user, POST /api/authors/apply → row
            created.
          - Submit again → returns existing pending row
            (no dupe).
          - Try apply as author/admin → 400.
          - FE /register-author: form, confirmation
            screen.
          - GET /api/authors/me/application returns the row.
      * One screenshot of the form and confirmation.
  - No push until I OK.

Report:
  - Files changed by layer (auth/authors/gateway/FE/proto/
    migration).
  - New HEAD.
Что я держу в карте
Закрыто (в origin/main, c0920f7):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 — в работе:

🟧 C1 — Author application (approve → implement).

🟧 C2 — Admin approval (после C1).

🟧 C3 — Author dashboard (L).

🟧 C4 — Payouts (L).

Критический fix в C1:

🔴 Public register → role=author — дыра.

Fix — force role=user.

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку — не забудь.

Действия
Отправь OK C1 — implement.

Cursor реализует — migration + proto + service + gateway + FE.

Verify — smoke 6 пунктов:

Register с role=author → 400.

Apply под user → row создан.

Apply снова → idempotent.

Apply под author → 400.

FE /register-author — form + confirmation.

GET /api/authors/me/application — row.

Push после OK.

Потом C2 (Admin approve flow).

Важно: дыра в register — реальный security fix, не только фича.