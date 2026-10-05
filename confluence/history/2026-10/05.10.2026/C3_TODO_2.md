OK. Реализуй C3
Что подтверждаю
Q1–Q4 — ответы принимаю:

Q1: author_id — уже есть (soft string). No migration.

Q2: GET /api/inventory/items?author_id=... — есть. Gateway create forces AuthorId = JWT user. Хорошо.

Q3: sales per author — нет. Нужен новый GET /api/authors/me/sales.

Q4: Authors service owns profile. PUT /api/authors/me extend с portfolio. Keep PUT.

Proposed v1 shape — правильный:

/author/dashboard — RequireAuth + role author|admin.

«Мои карты» — reuse InventoryItemCard.

«Продажи» — read-only, aggregate + list.

«Профиль автора» — load + edit.

Burger — «Автор» → /author/dashboard.

Out of scope — правильно (C4, notifications, draft status).

Risk: L — в основном sales endpoint + FE dashboard.

Один уточняющий вопрос
Про soft-delete:

Soft-delete via existing inventory soft-delete (confirm).

Уточни:

Soft-deleted items — видны автору в его dashboard или нет?

Мой выбор — видны с бейджем «Удалено» + кнопка «Восстановить».

Иначе автор не увидит, что удалил, и захочет вернуть — а не сможет.

Промпт Cursor'у — этот вопрос + approve.

Ответ Cursor'у
text
OK — implement C3.

Confirmations:
  - Q1–Q4 answers accepted.
  - /author/dashboard — RequireAuth + role author|admin.
  - «Мои карты» — reuse InventoryItemCard; create keeps
    author_id = me.
  - «Продажи» — new GET /api/authors/me/sales (admin:
    ?author_id=), read-only, aggregates + list. No payouts.
  - «Профиль автора» — load + edit via extended PUT
    /api/authors/me (+ portfolio).
  - Burger — «Автор» → /author/dashboard for author|admin
    only.

One clarification before you code:

  Soft-delete display in author dashboard:
    - Soft-deleted items should be VISIBLE in «Мои карты»
      with a «Удалено» badge + «Восстановить» action.
    - Otherwise authors lose track of deleted items.
    - If inventory API can't currently return soft-deleted
      items to their owner, add an ?include_deleted=1
      param (author|admin only).
    - If it's simpler to skip v1 — say so and we defer.

Rules:
  - C3-only PR. Do NOT touch C1/C2, C4, other waves, hero,
    nav (except the burger link), disk, games, seed, refund.
  - Do NOT invent draft status.
  - tsc + build clean.
  - Smoke:
      - Login as author → /author/dashboard loads.
      - «Мои карты» lists own items only.
      - Create new card → author_id = me.
      - Edit existing → saved.
      - «Продажи» aggregates match (buy + cancel in the
        same session).
      - «Профиль автора» loads / saves portfolio.
      - Burger shows «Автор» for author, hidden for user.
      - user hitting /author/dashboard → redirect / CTA.
  - Screenshots: dashboard «Мои карты», «Продажи»,
    «Профиль».
  - No push until I OK.
Что я держу в карте
Закрыто (в origin/main, 021f7a3):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 C1 — 7110d1b.

Wave 3 C2 — 021f7a3.

Wave 3 — в работе:

✅ C1 — Author application.

✅ C2 — Admin approval.

🟧 C3 — Author dashboard (approve → implement, L).

🟧 C4 — Payouts (L, после monetization).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

🟧 JWT role refresh — UX-gap (re-login needed).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.