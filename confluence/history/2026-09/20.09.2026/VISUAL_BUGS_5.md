Что делать дальше — приоритет
🔴 Сначала Block 3 — пограничные проверки
Почему первым: это безопасность. Если user может через API создать предмет — вся UI-защита бесполезна. Лучше узнать сейчас, чем после мёржа.

Как проверить Block 3 (DevTools)
Пункт 1 — user → POST /inventory/items напрямую

Залогинься как player@eventhorizon.local.

Открой DevTools → Console.

Вставь:

js
fetch('/api/inventory/items', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + localStorage.getItem('access_token')
  },
  body: JSON.stringify({
    type: 'брелок',
    name: 'HackTest',
    description: 'test',
    price: 1,
    stock: 1
  })
}).then(r => r.json()).then(console.log)
Если ключ токена не access_token — посмотри в DevTools → Application → Local Storage, найди, как он называется.

Ожидание: 403 insufficient permissions или 403 forbidden.
Плохо: 200 OK + предмет создан.

Пункт 3 — user → открыть /admin напрямую

Просто в адресной строке набери localhost:5173/admin (или какой у тебя роут). Ожидание: 404 / редирект / 403. Плохо: страница открылась.

RESULTS:

Текстовый отчёт для Cursor — без скриншотов
Раз Cursor не видит картинки — описываю всё словами. Формат: что подтверждено, что нашлось, что проверить.

Готовый файл для Cursor
text
QA REPORT — three-role manual pass
(no screenshots — described in words)

================================================================
CONFIRMED WORKING
================================================================

Visual pass (Block 1):
  1. Hexagon drag & drop — pancake lands in the hex under the
     cursor. Center and edges both register correctly. No scroll.
  2. Memory — cards visible, flip animation works.
  3. Subscription (admin, active) — shows "Управление",
     "Текущий план" label, no raw "present" enum.
  4. Shop — tickets are deducted on purchase.
  5. Towers GAME OVER — warm tone (#E8D5A3-ish), Space Grotesk.
  6. Home — all game names updated to:
     Pancaker / Flappy Bird / Builder / Hanoi / Memonia.
  7. Hanoi win modal — moves pluralize correctly
     (1 ход / 2 хода / 5 ходов). Buttons equal width.
  8. Skins — auto-apply, can toggle off, work in all games.

user role (player@eventhorizon.local):
  1. Inventory — no "Создать / Редактировать / Удалить" buttons
     visible. Correct.
  2. Shop — purchase works, tickets deducted.
  3. Games — playable, skins apply.
  4. Subscription — "Не активна", plan cards visible.
  5. Leaderboard — loads without crashes.
  6. Profile — "Показать ID" disclosure present, no CRUD
     buttons.

author role (author@eventhorizon.local):
  1. Inventory — "Создать / Редактировать / Удалить" buttons
     visible.
  2. Create — works, item appears.
  3. Update — works, changes saved.
  4. Delete — works on own items.
  5. Shop — purchase works, tickets deducted.
  6. Subscription — active (from seed).
  7. Boundary test — attempted to delete ANOTHER user's item:
     backend returned 403 "you can only delete your own items".
     Correct — role boundary enforced at API level, not just UI.

admin role (admin@eventhorizon.local):
  Subscription management, shop, inventory all working.

================================================================
BUG FOUND — Inventory card button overflow
================================================================

Location: /inventory (Каталог товаров), grid of item cards.
Role: visible as author and admin (may also affect user? user has
no buttons, so not visible there).

Symptom (described):
  - Each card shows two buttons side by side:
    "Редактировать" (indigo) and "Удалить" (red).
  - "Редактировать" sits INSIDE the card border.
  - "Удалить" OVERFLOWS the right edge of the card — its red
    border pokes outside the card's rounded border.
  - This happens on EVERY card in the grid, not just one.

Likely cause: two buttons in a grid-cols-2 row; the "Удалить"
button's min-content width exceeds its allocated column width,
so it spills outside the parent.

Fix options (pick one):
  a) Equal width: each button w-full inside its grid column.
  b) Reduce horizontal padding on both buttons so they fit.
  c) Stack buttons vertically on narrow cards.
Pick one and apply consistently. Buttons must not overflow the
card at any viewport width.

================================================================
ENHANCEMENT REQUEST — item image upload
================================================================

For item create / edit, add support for a single image per item:
  - Upload a file OR paste a URL (whichever fits the stack).
  - Preview in the modal after selection.
  - In the card grid: image fills the placeholder area
    (object-fit: cover, rounded corners), auto-scaled.
  - If no image: keep the current emoji placeholder (📦).

Backend:
  - New column: image_url (or image_path) on inventory items.
  - Migration.
  - Accept in create / update payloads.
  - Serve / store files (local dir or S3 — whatever fits).

Scope: do NOT redesign the modal, do NOT touch other fields,
keep the emoji fallback intact.

================================================================
QUESTIONS — FINDINGS (Cursor, no fix applied)
================================================================

Q1. Three "Автоматический брелок" @ 200₽ — (a) old test data.
  Three distinct rows in inventory_items (same author_id):
    6e4d8ba9…  created 2026-08-06 00:12:19
    1590b92b…  created 2026-08-06 00:20:23
    c96419b1…  created 2026-08-06 00:32:22
  ~8–12 min apart → repeated manual creates, not a double-submit
  race. No UNIQUE on name; CreateItem always inserts a new UUID.
  UI create has a loading guard; not a create-side bug that
  triples on one save. Safe to delete extras manually.

Q2. "Vika" 110₽ red→indigo price chip — no intentional rule.
  InventoryItemCard price is always a gold span
  (text-horizon-gold), never a Badge / never stock-based.
  The chip next to price is type Badge, hard-coded tone="indigo".
  No Vika / price=110 row in DB now (deleted or renamed).
  Likely: gold misread as red, or confusion with danger
  "Удалить" / type Badge. No status→color mapping exists.
  If you want one tone: keep gold for price; leave type indigo.

================================================================
QUESTIONS TO INVESTIGATE (no fix yet, just report) — answered above
================================================================

Q1. Catalog shows three identical "Автоматический брелок" entries
    (same price 200). Is this:
      a) Old test data in the DB (I'll clean manually).
      b) A create-side bug that duplicates on save.
    Investigate and report.

Q2. Item "Vika" (110₽) rendered with a RED price chip at first,
    then switched to an INDIGO price chip after an update.
    Is chip color tied to something intentional (availability,
    stock, status), or is it random? Report the rule. If no
    rule exists, standardize to one tone.

================================================================
BLOCK 3.1 / 3.3 — RESULTS (Cursor, 2026-09-22)
================================================================

Block 3.1 — PASS
  player@eventhorizon.local → POST /api/inventory/items
  Response: HTTP 403 {"error":"insufficient permissions"}
  No HackTest row created in inventory_items.
  Note: SPA stores JWT as localStorage key `accessToken`
  (not `access_token`). API login body field is `access_token`.

Block 3.3 — PASS (no admin surface)
  App.tsx has no `/admin` route (and no catch-all).
  Direct /admin does not render an admin page — unmatched SPA path.
  Vite was down during check; conclusion from route table + code.

================================================================
DO NOT TOUCH
================================================================
- Game mechanics.
- Recent fixes (Hexagon drag, Memory, Subscription UI, tickets
  deduction, game names, Hanoi modal).
- Nav, footer, role gating.

================================================================
PROCESS
================================================================
Step 1: Fix the button overflow (Bug above). Report the actual
        root cause before fixing.
Step 2: Investigate Q1 (duplicates) and Q2 (price chip color).
        Report findings, no fix yet.
Step 3: Wait for my confirmation on Step 1 + Step 2.
Step 4: Only then — plan and implement image upload (Enhancement).
Что тебе осталось
Перед отправкой промпта — прогони Block 3.1 и 3.3. 30 секунд:

Залогинься как player@eventhorizon.local.

DevTools → Console → вставь fetch (см. промпт выше).

Смотри ответ: 403 → отлично, 200 → критично.

Открой localhost:5173/admin в адресной строке — что показал?

Результаты добавь в отчёт — в блоке STILL TO CHECK замени на фактические.

Формат твоего файла — предложение
Сейчас у тебя файл начинается с «Что делать дальше — приоритет / 🔴 Сначала Block 3». Это методичка, а не отчёт.

Предлагаю такую структуру:

CONFIRMED WORKING — что прошло QA (для Cursor'а это важно — не трогать).

BUG FOUND — что нашлось (карточки overflow).

ENHANCEMENT — что хочешь добавить (картинки).

QUESTIONS — что Cursor должен расследовать (дубли, цена).

STILL TO CHECK — что ты ещё не проверил (Block 3.1, 3.3).

DO NOT TOUCH — защита от регрессов.

PROCESS — порядок работ.

Причина: Cursor сейчас находится в режиме «чинить и не ломать». Ему важно сначала увидеть что уже работает — иначе он может случайно тронуть то, что ты подтвердил как ок.

Всё, что выше — уже в этом формате. Копируй.

Что я держу
Block 1: ✅

user: ✅

author: ✅ 6/7 (overflow)

admin: ✅

Block 3: ✅ boundary (403 на чужое)

Block 3.1 / 3.3: ⏳ прогнать

Bug: card overflow 🔴

Enhancement: image upload 🟧

Questions: duplicates, chip color 🟧