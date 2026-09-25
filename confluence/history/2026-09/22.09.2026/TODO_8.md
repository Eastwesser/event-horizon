2a + 2b — реализовано. Чек-лист
Что подтвердил Cursor
Контракт ListUsers: чистый. query (email substring), limit (1–100), offset. Возвращает users[] + total. Правильно.

Gateway: /api/admin/users?q=&limit=50&offset= — admin-only, fan-out в billing + payment.

Реализовано:

Backend: repo → service → gRPC → gateway → redeploy.

Frontend: /admin route + guard, бургер, Users tab.

Verified: admin 200 / player 403.

Чек-лист визуальной проверки (5 минут)
Как admin@eventhorizon.local
#	Проверка	Ожидание
1	Залогинься → открой бургер ☰	Админ-панель виден
2	Кликни Админ-панель	Открывается /admin
3	Вкладки	Пользователи · Инвентарь · Аналитика (последние две — placeholder)
4	Таблица Users	email / role / lamps / tickets / sub / created_at
5	Поиск по email	Введи часть admin → фильтрует
6	Пагинация	Если >50 — есть «next»
7	Смена роли — на другом	Выбери player@... → смени user → author → confirm dialog показывает текущую роль
8	Смена роли на себе	admin → заблокировано (guardrail)
Как player@eventhorizon.local
#	Проверка	Ожидание
9	Бургер ☰	Админ-панель НЕ виден
10	Набери localhost:5173/admin руками	Редирект на Home
Через DevTools (любая роль, кроме admin)
#	Проверка	Ожидание
11	fetch('/api/admin/users') под player	403 insufficient permissions
Пункт 11 — критичен. Подтверждает, что бэкенд защищён, а не только UI.

После проверки — 2c
Если всё ок — отправь:

text
2a + 2b verified. Go ahead with 2c (Inventory stats v1).
Show the layout plan first — which fields, how displayed, which
primitives. No code yet.
Если что-то не так:

text
2a + 2b — issue at step N: [что видишь] instead of [что
ожидалось]. Fix and re-verify.
Что я держу
Закрыто:

✅ Fix 1–3 (stock, price, edit prefill)

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

✅ 2a + 2b (admin route, guard, users table, ListUsers)

В работе:

🟧 2c — Inventory stats v1 (план ждёт)

🟧 2d — Analytics

Отложено:

🟧 2c v2 (top-5, total stock)

🟧 Medium proto3 risks, gin.H refactor

🟧 File upload, navbar refactor

Ждёт финала:

🟧 Коммит dev/frontend → main


UPDATED:

Всё работает. Один нюанс + approve на 2c
✅ Что подтвердилось
Как admin — 8/8: бургер, /admin, вкладки, таблица, поиск, пагинация (2852 юзера, 58 страниц!), смена роли, confirm dialog с текущей ролью, self-demotion заблокирован.

Как player — 2/2: бургер без Админ-панель, /admin редиректит домой.

Фикс роли сработал: testuser23 — user → author, badge обновился, всё чётко.

⚠️ Нюанс — 401 vs 403
Твой DevTools-тест:

text
fetch('/api/admin/users') → 401 (Unauthorized)
Это не 403 — это 401. Разница:

401 = «ты вообще не авторизован» (нет токена).

403 = «ты авторизован, но тебе нельзя» (есть токен, роль не та).

Почему так вышло: ты сделал сырой fetch без заголовка Authorization. Без токена → бэкенд говорит «кто ты?» → 401.

Чтобы проверить настоящий role gate (403), нужно вставить токен:

js
fetch('/api/admin/users', {
  headers: { 'Authorization': 'Bearer ' + localStorage.getItem('accessToken') }
}).then(r => console.log(r.status, r.statusText))
Ожидание: 403 insufficient permissions.

Но — Cursor уже подтверждал player → 403 в предыдущем отчёте (через api.ts с интерцептором). Так что бэкенд-защита работает. Твой тест просто не дошёл до проверки роли.

Не блокер. Если хочешь — проверь с токеном для чистоты. Если нет — идём дальше.

Approve на 2c
Отправляй Cursor'у:

text
2a + 2b verified. All checks pass:
  - admin: burger link, /admin, tabs, table, email search,
    pagination (2852 users / 58 pages), role change with
    confirm dialog showing current role, self-demotion blocked.
  - player: no burger link, /admin redirects to Home.
  - role change user → author worked, badge updated.

Go ahead with 2c (Inventory stats v1).
Show the layout plan first — which fields, how displayed, which
primitives. No code yet.
Что я держу
Закрыто:

✅ Fix 1–3 (stock, price, edit prefill)

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

✅ 2a + 2b (полностью)

В работе:

🟧 2c — Inventory stats v1 (план ждёт)

Отложено:

🟧 2d — Analytics

🟧 2c v2 (top-5, total stock)

🟧 Medium proto3 risks, gin.H refactor

🟧 File upload, navbar refactor

Ждёт финала:

🟧 Коммит dev/frontend → main

Отправляй approve. Как Cursor покажет план 2c — чекнём, дадим OK.