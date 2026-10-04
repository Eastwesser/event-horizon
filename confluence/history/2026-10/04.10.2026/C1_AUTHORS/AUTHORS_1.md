OK C1. Два решения по «extras»
Extras
1. Close public Register → role=author?
Вопрос: сейчас публичный /register может создать author'а? Или только user?

Мой ответ: да, закрыть — рекомендую.

Причины:

Ролевая модель должна быть строгой. author — только через approval flow (C1+C2). Иначе — дыра в безопасности.

admin — тоже только через seed или DB (уже так).

user — публичная регистрация, всё остальное — через approval.

Что сделать:

Проверить текущий /register — принимает ли role параметр?

Если да — убрать.

Если нет — ок, дополнительно проверить, что нигде в коде нет обходных путей (например, seed-скрипты — это ок, UI — нет).

2. Notifications в C1/C2 MVP?
Вопрос: добавить notification сейчас или позже?

Мой ответ: MVP — без notification.

Причины:

Notification service — есть, но путь для author.application.submitted и author.application.approved — новый.

C1+C2 — цель — workflow applications + approval. Notification — украшение.

Добавишь notification позже — отдельный маленький PR (S). Не блокер.

MVP — DB + confirmation UI + admin approve + role update.

Промпт:

text
OK — proceed with C1 implementation.

Two extras answered:

1. YES — close public Register so it can only create
   role=user.
   - Check the current /register flow: does it accept a role
     parameter? If yes — remove.
   - Verify no other code path (outside seed scripts) creates
     author/admin via public API.
   - Report what you find before changing anything.

2. NO notifications in C1/C2 MVP.
   - DB + confirmation UI + admin approve + Auth.UpdateRole
     is enough.
   - Add notifications as a later small PR (author.application.
     submitted → admins; author.application.approved → author).
   - Note it as a follow-up in the map.

C1 scope confirmed:
  - Authors service owns author_applications.
  - No users FK — soft user_id string.
  - POST /api/authors/apply — idempotent on (user_id, pending).
  - Reject if user is already author/admin.
  - FE /register-author route + form + confirmation screen.

Show the C1 plan: migrations, proto changes if any, endpoints,
FE route, RBAC. No code until I OK.
Что я держу в карте
Закрыто (в origin/main, c0920f7):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Следующее — Wave 3:

🟧 C1 — Author application (план → approve).

🟧 C2 — Admin approval (после C1).

🟧 C3 — Author dashboard (L).

🟧 C4 — Payouts (L, после monetization).

Решения по extras:

✅ Close public Register → role=user.

✅ Notifications — позже (отдельный PR).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь OK C1 + 2 extras.

Cursor покажет C1 plan — изучи:

Migration author_applications.

Endpoints — POST /api/authors/apply.

FE — /register-author.

RBAC — middleware.

Approve реализации.

Verify → push.

Потом C2.