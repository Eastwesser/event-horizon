Reorder the burger menu items only.

Current:
  Инвентарь · Подписка · Авторы · История

New:
  Инвентарь · История · Авторы · Подписка

Location: the burger dropdown (the ☰ menu, between "Профиль"
and "Выйти").

Do exactly this and nothing else:
  - Reorder the 4 items in the JSX.
  - Do NOT touch the top-bar nav (Профиль · Лидерборд · Магазин).
  - Do NOT touch Выйти, logo, or any styling.
  - Do NOT change icons, labels, or hrefs.
  - Do NOT add group headings or separators — leave it as a flat
    list of 4 items.

Report the file changed.

Then:

Two tasks. Do them in this order. Do NOT bundle.

================================================================
TASK 1 — IMAGE URL FIELD (UI only)
================================================================

Confirmed: backend already has images TEXT[], create/update
accept images: string[]. Card already renders images[0] with
📦 fallback. No migration, no new endpoint.

Authors are adults (this is a public platform; players are
children 4–18). URL input is acceptable for now. File upload
(S3/disk) is a separate wave — skip.

Implement:
  - Create modal: one field "Изображение (URL)".
  - Edit modal: same field, prefilled with images[0] if present.
  - Preview thumbnail in the modal after the URL is entered
    (with graceful fallback if the URL is broken — show the 📦).
  - Clear button (×) to remove the image → sends images: [].
  - On submit: send images: [url] if non-empty, images: [] if empty.
  - Do NOT touch other fields.
  - Do NOT redesign the modal.

Before writing code: show me a brief layout plan of where the
new field sits in Create/Edit modals (text description, no code).
Then wait for my OK. Then implement.

================================================================
TASK 2 — ADMIN PANEL (build it for real)
================================================================

Backend already exposes admin-gated APIs:
  POST /api/auth/update-role          (admin)
  GET  /api/analytics/dau|mau|retention (admin)
  GET  /api/inventory/stats           (admin)
  inventory create/update/delete      (author or admin)

There is currently NO /admin route and NO admin UI.

Build a minimal, kids-safe-agnostic (this is an internal tool,
not a public page) admin panel, gated to role=admin only.

--- 2a. Route + guard ---
  - Add a route /admin (route exists, but is visible only to
    role=admin).
  - Non-admin users hitting /admin → redirect to Home (or 404).
    Backend must also reject — do NOT rely on UI alone.
  - Add a link to /admin in the burger menu ONLY when role=admin
    (label: "Админ-панель"). For non-admins — hidden.
  - Reuse the existing layout (nav, footer). This is not a
    separate app.

--- 2b. Section 1 — USERS / ROLES (most important) ---
  - Table of users: email, role, lamps, tickets, subscription
    status, created_at.
  - Search by email.
  - Pagination (50 per page).
  - Inline action: change role via POST /api/auth/update-role.
    Roles: user, author, admin.
  - Guardrails:
      * Cannot demote yourself (avoid locking out).
      * Confirmation dialog before changing any role.
      * Show the target user's current role in the confirm dialog.
  - Do NOT add user deletion in this pass.

--- 2c. Section 2 — INVENTORY STATS ---
  - Simple view: totals per type (брелок / картина / фенечка),
    total items, total stock, top-5 most expensive.
  - Read-only. Uses GET /api/inventory/stats.
  - No edit here — editing stays on /inventory.

--- 2d. Section 3 — ANALYTICS ---
  - Three numbers: DAU, MAU, retention (whatever the endpoint
    returns).
  - Simple chart or big-number cards — whatever's simplest.
  - Read-only.
  - If the endpoint shape is unclear, report it first and ask
    before assuming.

--- 2e. Layout ---
  - Tabs or sections inside /admin: "Пользователи" · "Инвентарь" ·
    "Аналитика".
  - Use the existing primitives: PageHeader, Card, Badge, Button,
    Table (or list), Spinner, FilterChip.
  - Do NOT introduce a new design language.

--- 2f. Kids-safe note ---
  - This is an INTERNAL tool for admins (adults). It does not need
    to be playful. It DOES need to be readable, calm, and
    consistent with the rest of the app's palette.
  - No red used decoratively. Red only for destructive actions
    (role demotion confirmation).

--- 2g. Process ---
  Step 1: Show the plan for /admin (routes, tabs, which endpoints,
          which primitives). No code.
  Step 2: Wait for my OK.
  Step 3: Implement 2a + 2b first (roles). Verify.
  Step 4: Then 2c (inventory stats). Verify.
  Step 5: Then 2d (analytics). Verify.

================================================================
DO NOT TOUCH
================================================================
- Game mechanics.
- Recent fixes (Hexagon drag, Memory, Subscription, tickets,
  game names, Hanoi modal, nav swap, burger order).
- Role gating (user/author/admin boundaries — confirmed working).
- Nav, footer.

================================================================
PROCESS (overall)
================================================================
1. Task 1 (image URL) — plan → OK → implement → verify.
2. Task 2 (admin panel) — plan → OK → implement in 4 sub-steps.
Do NOT start Task 2 before Task 1 is done and verified.