Что принято — 021f7a3
✅ Authors: List / Approve / Reject / Revert + tx upsert (display_name / portfolio / bio / verified_at).

✅ Gateway: admin routes orchestrate Auth.UpdateRole(author) + RevertApplication if Auth fails.

✅ Admin tab «Заявки» с Pending / Approved / Rejected.

✅ JWT role refresh — заметка в map (re-login нужен).

✅ Smoke 8/8.

Хорошая работа. RevertApplication — правильный паттерн (compensation transaction). Fail loudly вместо silent half-state — хорошо.

Ответ Cursor'у
text
C2 accepted. Push 021f7a3 → origin/main. Report new HEAD.

================================================================
WAVE 3 C3 — Author dashboard (L)
================================================================

Plan first, no code.

Goal: authors get a personal workspace to manage their own
cards, see sales, and track status. Admin panel already
exists for admins; C3 is author-facing.

Scope for v1 (read + manage own cards):

1. Route /author/dashboard (author or admin only).
   - If user is not author/admin → redirect to / or show
     a «Подать заявку» CTA if there's no pending application.

2. Section «Мои карты»:
   - List of inventory items where author_user_id === me.
   - Reuse InventoryItemCard / table.
   - Filters: type (карточка / мерч / ...), status
     (published / draft), search by name.
   - Actions:
       - Edit (link to existing inventory edit modal).
       - Create new (reuse InventoryCreateModal, but force
         author_user_id = me).
       - Soft-delete / unpublish — confirm first.

3. Section «Продажи» (basic):
   - Aggregated: number of sales, total tickets earned.
   - List of purchases of my items (buyer email masked,
     price, purchased_at, refunded at).
   - Read-only.

4. Section «Профиль автора»:
   - display_name, portfolio, bio from author profile.
   - Edit → POST /api/authors/me (new endpoint).

5. Top nav: link «Автор» in burger (authors + admins only).

Questions to answer before code:
  Q1. Is there an author_user_id field on inventory items
      already? If not — we need a migration to link items
      to authors. C1 used soft user_id; same pattern here?
  Q2. Inventory service has GET /items?author=... or
      similar filter? If not — add.
  Q3. Sales list: shop exposes purchases per item? If not,
      need a new endpoint GET /api/authors/me/sales.
  Q4. Author profile edit endpoint — Auth update or
      Authors service? Recommend Authors service
      (display_name/portfolio/bio live there).

Constraints:
  - Read-only sales, no payouts (that's C4).
  - No notifications yet.
  - C3-only PR. Do NOT touch C1/C2, C4, other waves, hero,
    nav (except the burger link), disk, games, seed, refund.
  - Authors only see their own items. Admin can see all.

Show the C3 plan with answers to Q1–Q4. No code until I OK.
Что я держу в карте
Закрыто (в origin/main, 7110d1b):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 C1 — Author application.

Локально (пуш сейчас):

021f7a3 — C2 Admin approval.

Wave 3 — в работе:

✅ C1 — 7110d1b.

✅ C2 — 021f7a3 (push сейчас).

🟧 C3 — Author dashboard (следующее, L).

🟧 C4 — Payouts (L, после monetization).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

🟧 JWT role refresh — UX-gap (re-login needed).

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку — после C2 push хороший момент.

Действия
Отправь Cursor'у — push + C3 plan.

Cursor ответит на Q1–Q4 — изучи.

Approve реализации.

Verify → push.

Потом C4 (Payouts — после monetization decision).

C3 — крупная (L). Author dashboard + sales + profile edit. Может потребовать migration для author_user_id в inventory.