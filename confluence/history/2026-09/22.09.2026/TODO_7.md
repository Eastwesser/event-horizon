Разбор плана Task 2 — admin panel
План сильный. Cursor нашёл блокер, который ты не видел: эндпоинта ListUsers нет. Без него 2b — «вставь user_id руками». Разберу и дам ответ.

✅ Фиксы (Fix 1–3) — закрыто
#	Что	Проверка
1	InventoryItemDetail stock	✅
2	formatPrice — «Бесплатно» при 0	✅
3	Edit prefill — ?? 0, submit числа	✅ проверено на данных
Fix 3 проверен руками: price=0 → переименовал через API с явным 0 → в БД осталось 0.00. Данные не портятся. Правильно.

🟧 Блокер — ListUsers отсутствует
Cursor:

No ListUsers / search-users in auth proto or gateway.
Without ListUsers, 2b cannot be a real user table — only «paste user_id + change role», which is a weak MVP.

Он прав. Без ListUsers:

Нет таблицы пользователей.

Нет поиска по email.

Нет пагинации.

Только «вставь UUID руками» → это не админка, это костыль.

Соглашайся на добавление ListUsers. Это небольшой эндпоинт — читает таблицу users, отдаёт список. Стандартный.

Плюс Cursor предлагает fan-out из billing + payment для lamps / tickets / sub — то есть gateway сам подтягивает данные из 3 сервисов и склеивает. Нормальный паттерн для админки (не критично к latency).

🟧 2c — v1 без top-5 / total stock
Cursor:

Plan: ship read-only view from current stats (counts by type + total items). Top-5 / total stock = follow-up.

Согласен.

v1 — что есть: total_items, by_type, by_author.

v2 (потом) — top-5 expensive, total stock. Требует расширения StatsResponse или нового эндпоинта.

Почему v1 — быстрее запустить. Расширение — не блокер.

🟧 2d — Analytics
Cursor уже разведал структуры:

DAU → { days: [{ date, count }] }

MAU → { mau, window_days }

Retention → { cohort_day, cohort_size, points: [{ day_n, rate }] }

Big-number cards для MAU + bar/list для DAU/retention. Разумно.

⚠️ Нюанс: DAU с пропущенными днями (proto3 zero-omission) — график может пропускать дни. Cursor пишет «acceptable for now» — ок для v1, но в v2 можно поправить.

Ответ Cursor'у — approve + два «да»
text
Plan approved. Two answers:

1. YES — add admin ListUsers + gateway enrich for 2b.
   - ListUsers(query, limit, offset) in auth proto.
   - GET /api/admin/users?q=&limit=50&offset= (admin-only).
   - Enrich lamps/tickets from billing, sub from payment via
     parallel fan-out in the gateway.
   - Keep it read-only. No user deletion, no user creation.

2. YES — 2c v1 uses current stats only.
   - total_items, by_type, by_author from the existing endpoint.
   - Mark in UI as "v1" (or just don't show top-5 / total stock
     yet).
   - Top-5 / total stock is a separate follow-up.

Implementation order — as proposed:
  2a + 2b (incl. ListUsers)
  2c
  2d

Before each step, verify with tsc + build. After each step,
report what changed + screenshot-free description for me to
check visually.

DO NOT TOUCH
- Game mechanics.
- Existing role gating (confirmed working).
- Nav, footer, burger.
- Anything from prior waves.

Process
Step 1: Implement 2a + 2b together. Report.
Step 2: I visually verify. Then 2c.
Step 3: I visually verify. Then 2d.
Do NOT bundle 2c and 2d with 2a+2b.
Что я держу в карте
Закрыто:

✅ Fix 1–3 (stock, price, edit prefill)

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

В работе (после approve):

🟧 Task 2a — Route /admin + guard

🟧 Task 2b — Users/Roles + новый ListUsers endpoint

🟧 Task 2c — Inventory stats v1

🟧 Task 2d — Analytics

Отложено:

🟧 2c v2 (top-5, total stock)

🟧 Medium proto3 risks

🟧 Long-term gin.H / protojson

🟧 File upload, navbar refactor

Ждёт финала:

🟧 Коммит dev/frontend → main

Один момент по процессу
Проси Cursor'а показывать план перед каждым шагом. Особенно 2b — там новый эндпоинт, миграция не нужна, но нужен proto-контракт. Пусть покажет сигнатуру ListUsers до реализации.

Если после 2a+2b что-то сломается в других местах — не давать сразу 2c. Сначала фикс.