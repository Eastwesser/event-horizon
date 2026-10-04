Wave 2 закрыт. Решаем Q2
Поздравляю с финалом Wave 2
Wave 2 — 6/6 + 5/5 игр:

✅ #1 thin DTO

✅ #2 DB cleanup

✅ #3 lamps as boosts

✅ #4 Flappy levels

✅ #5 achievements

✅ #6 5/5 игр (Flappy, Towers, Hanoi, Memory, Hexagon)

Всё в origin/main, HEAD c0920f7.

Q2 — Wave 3 или Wave 4 первым
Мой выбор — Wave 3 C1+C2.

Причины:

Portfolio value. Ты готовишься к собесам (помнишь memo про деньги?). Author registration — это системная фича с ролями, approval flow, notification — отличный разговор на интервью. «Как вы строили модерацию авторов?» — сильный ответ.

Small win. C1+C2 — S+M. Быстрее закрыть, чем Wave 4 инфра (Helm, Consul, k6 — каждая M–L).

User-facing. Демонстрируется в 5 минутах на демо: «подай заявку → админ approve → ты author → создаёшь карты».

Wave 4 — параллельно можно запустить юнит-тесты — не блокирует Wave 3. Но старт — Wave 3.

Wave 4 — после C1+C2 или параллельно (юнит-тесты — независимы).

Промпт Cursor'у
text
Q2 answer: Wave 3 first. Start with C1 + C2 (author
application + admin approval). Defer C3 + C4 until C1+C2
are proven in production and monetization is decided.

Push confirmed: HEAD c0920f7.

================================================================
WAVE 3 C1 — Author application (S)
================================================================

Plan first, no code.

Scope:
  - Route /register-author (authenticated user, role=user).
  - Form fields: display name, portfolio URL or text,
    motivation (free text), contact email (default = account
    email).
  - POST /api/authors/apply — creates a row in a new
    author_applications table:
      id, user_id, status(pending|approved|rejected),
      payload(jsonb), created_at, reviewed_at, reviewed_by,
      reviewer_note.
  - Idempotent: if a user has a pending application, second
    submit returns the existing one (no duplicate rows).
  - If the user is already author/admin, return 400 with a
    clear message.
  - After submit: show a confirmation screen («Заявка
    отправлена, ожидайте решения»).

Questions to answer before code:
  1. Which service owns this? Authors service or new?
     (Recommended: Authors service — you already have it.)
  2. Does Authors service have a `users` FK or only its own
     schema? Check before designing the FK.
  3. Notification on submit: notify admins via existing
     Notification service (NATS event author.application.submitted)?

================================================================
WAVE 3 C2 — Admin approval (M)
================================================================

Plan first, no code.

Scope:
  - /admin → tab «Заявки» (new tab; next to Пользователи,
    Инвентарь, Аналитика).
  - Table: applicant (email + display name), submitted at,
    status.
  - Actions: Approve / Reject with optional note.
  - Approve → POST /api/authors/applications/:id/approve →
    sets status=approved, changes the user's role to author
    (or emits an event that the auth service consumes).
  - Reject → status=rejected + reviewer_note.
  - Filter: pending | approved | rejected.
  - Role change mechanics: who updates auth.users.role?
    Auth service or via NATS event?

Questions to answer before code:
  1. Where does `role` live? Auth service (single source of
     truth)? If so, admin approval needs to call Auth
     service.
  2. Should approved authors get a welcome notification
     (Notification service)?
  3. RBAC: approve/reject endpoints admin-only. Confirm the
     middleware is enough or need explicit check.

================================================================
PROCESS
================================================================
  1. Investigate the data model (users, roles, authors) in
     the codebase.
  2. Show C1 plan (migrations, endpoints, proto, FE route).
  3. Show C2 plan.
  4. Wait for my OK.
  5. Implement C1 first, verify, then C2.

DO NOT TOUCH: hero, nav, disk, seed, refund, games.
Что я держу в карте
Закрыто (в origin/main, c0920f7):

v1.0.9.

Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3.

Smooth scroll.

Следующее — Wave 3:

🟧 C1 — Author application (S).

🟧 C2 — Admin approval (M).

🟧 C3 — Author dashboard (L) — после C1+C2.

🟧 C4 — Payouts (L) — после monetization decision.

Параллельно (Wave 4):

🟧 Helm, Consul, k6, юнит-тесты.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку — сейчас хороший момент (Wave 2 закрыт).

Действия
Отправь Cursor'у — Wave 3 C1+C2 plan.

Сам — сделай бэкап на флешку.

Изучи план Cursor'а — 3 вопроса по C1, 3 по C2.

Approve реализации.

Реализация → verify → push.

Wave 2 — закрыт. Wave 3 — старт новой большой фичи.