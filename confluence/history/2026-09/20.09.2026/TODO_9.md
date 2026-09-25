Разбор ответа Cursor'а по Priority 1
Он заявил, что закрыл 4 пункта. Но в трёх из четырёх — компромиссы, а не то, что ты просил. Разберу.

1a. Towers LMB — правдоподобно
Click was bound while canvas wasn't mounted (skinsLoading). Now onClick on the canvas — same as SPACE.

Root cause: обработчик привязывался пока canvas не смонтирован (skinsLoading ещё true). Переехал на onClick canvas.

Правдоподобно. Но вопросы:

Работает ли LMB в любом месте canvas, или только когда canvas в фокусе?

Что происходит, пока skinsLoading = true? Пользователь видит что — пусто? спиннер?

Проверить: кликни в разных местах поля. Клик работает?

1b. Hanoi — ⚠️ компромисс, не откат
Ты просил:

«Откатить до предыдущего фикса, где мы ебались с размером стойки под максимальное кольцо на 8 уровне.»

Cursor сделал:

Pegs get min-width: 160px, chunkier rings (~42–92%)

Это НЕ откат. Это новая формула: стойка min 160px, кольца 42–92% от стойки.

Что было раньше: стойка = ширина максимального кольца (уровень 8). Кольца = разной ширины, но не «% от стойки».

Риск: «chunkier rings 42–92%» — это диапазон. Кольцо 1 = 42% от стойки, кольцо 8 = 92%. Визуально это может быть норм, а может — всё ещё не то. Он не видел результат.

Плюс: «Select shouldn't be clipped» — «shouldn't», не «is not». Он не уверен. Не проверял.

Проверить:

Кольца нормального размера, не «бусины»?

Селектор уровня 5–8 доступен?

1c. Hexagon — ⚠️ тоже компромисс
Ты просил:

«Он должен видеть весь экран. Без скроллов.»

Cursor сделал:

Board min-height: 380px, larger tray; slight scroll OK if the viewport is short.

Это отступление от требования. Он разрешил скролл на коротких вьюпортах. Компромисс между «мелкое поле» и «без скролла».

Вопрос: на твоём экране — скролл есть или нет? Если да — насколько «slight»?

Если скролл маленький (10–20px) — может, ок.

Если нужно скроллить чтобы увидеть поднос — не ок, ты просил без скролла.

Проверить: скроллится страница Hexagon? Счёт + поле + поднос — всё видно сразу?

1d. Leaderboard widget — похоже, правильно
Null-safe score / empty entries — no mid-game crash.

Правдоподобно. Тот же паттерн, что в LeaderboardFull.

Проверить: кликни «Лидерборд» в Hexagon. Не крашится? Не перезагружает?

Общее замечание
Cursor закрыл 4 пункта — но 2 из 4 (1b, 1c) через компромисс, а не через то, что ты просил.

Hanoi: ты просил «откат к предыдущему». Он сделал «новую формулу». Может сработать, может нет.

Hexagon: ты просил «без скроллов». Он разрешил «slight scroll». Может быть ок, может быть нет.

Он снова не видел результат. Только tsc + build. Визуальная проверка — на тебе.

Что делаем
Чек-лист визуальной проверки (5 минут):

#	Проверка	Ожидание
1	Towers — кликни мышью по полю	Блок падает, как по SPACE
2	Towers — кликни в разных местах	Работает везде
3	Hanoi — посмотри на кольца	Нормального размера, не бусины
4	Hanoi — открой селектор уровня	5–8 доступны
5	Hexagon — открой игру	Поле читаемое? Скролл есть?
6	Hexagon — клик на «Лидерборд»	Не крашится, не перезагружает
Развилка после проверки:

Всё ОК → запускаем Priority 2 (английские имена, грамматика, дубли).

Hanoi всё ещё не то → отдельный фикс: «не компромисс, откати к состоянию X».

Hexagon скроллится сильно → отдельный фикс: «min-height не нужен, сделай fit без скролла или уменьши поднос».

Towers LMB не везде → уточнить, где именно не работает.

===

Разбор проверки + промпт
Что подтвердилось ✅
#	Пункт	Статус
1a	Towers — LMB работает	✅
2	Towers — LMB внутри поля, не вне поля	✅ (правильно)
1b	Hanoi — кольца, селектор 5–8	✅ ИДЕАЛЬНО
1d	Hexagon — Leaderboard не крашится	✅
Priority 1 — фактически закрыт. Осталась мелочь на Hexagon.

Что осталось
🟧 Hexagon — «лишний миллиметр скролла»
Ты сказал:

«Сейчас фиксируем размер как 100%. Нужно 90%. Слегка скролл пустой миллиметр скроллит. НЕ критично, но хотел бы довести до ума.»

Что делать: уменьшить общий размер GameShell примерно на 10%, чтобы всё влезло в вьюпорт без скролла.

Конкретное решение: не «min-height: 380px», а max-height: 90dvh на stage. Тогда:

Stage не превысит 90% высоты экрана.

Хедер + поднос + футер займут оставшиеся 10%.

Скролл не появится.

Поле будет читаемым.

🟧 Leaderboard в Hexagon — «показывает остальные игры»
Ты сказал:

«Показывает остальные игры, кста…»

Это подозрительно. Лидерборд должен показывать либо:

лидеров текущей игры (Hexagon / Блинопёк),

либо лидеров всех игр (если это общий виджет).

Если виджет показывает все игры — это может быть by design (общий топ игроков по всем играм). Или баг — если ты ожидал топ только по Hexagon.

Надо уточнить у Cursor'а: что именно должен показывать виджет Leaderboard в игре? Только текущую игру или общий топ?

🔴 Скины — не грузятся
Ты:

«Надо понять, куда делись скины. Не прогружаются потому что не куплены, или сломаны?»

Мы это уже видели в логе:

text
useSkins - ответ от API: null
⚠️ Неизвестный формат инвентаря: null
items: []
Три гипотезы (не разделить без БД):

Скины не куплены → null ожидаемо, но UI должен показывать человеческий empty state, а не молчаливый warning.

Скины куплены, но эндпоинт возвращает null → это баг эндпоинта.

Скины потерялись при миграции (Wave 3 трогала Inventory).

Что нужно: Cursor должен посмотреть эндпоинт /shop/inventory — что он возвращает, когда у пользователя ничего нет. Правильный ответ: [], не null. Если null — эндпоинт сломан.

🟧 Admin-аккаунт + god-доступ
Ты:

«Давай Cursor'у дадим таску создать админа с 1 млн билетиками и лампочками, и полным god доступом. Ну я же его разрабатываю, так?»

Абсолютно согласен. Ты разработчик, тебе нужен god-аккаунт для тестирования всех путей.

Что должен сделать Cursor:

Найти, где в БД хранятся роли (user / author / admin).

Найти, где хранятся балансы (лампочки / билетики).

Создать seed-скрипт или SQL, который:

Создаёт пользователя с ролью admin,

Ставит ему 1 000 000 лампочек и 1 000 000 билетиков,

Даёт полный доступ ко всем эндпоинтам.

Не хардкодить в код — сделать через seed / migration / CLI-команду, чтобы можно было повторить.

Дать тебе логин/пароль (или инструкцию, как залогиниться).

Важно: не должен пушить это в прод. Только dev-сид.

Промпт для Cursor
text
CONTEXT
Priority 1 is confirmed working (Towers LMB, Hanoi, Leaderboard
widget). Three follow-ups + Priority 2.

================================================================
FOLLOW-UPS (do first)
================================================================

--- F1. Hexagon — remove the last millimeter of scroll ---

Current state: works, one-screen, but a tiny empty scroll appears.
Target: reduce the GameShell stage so it fits 100% without scroll.

Fix approach (pick the cleanest):
  - Stage max-height: 90dvh (or calc(100dvh - header - tray)),
    so header + stage + tray = viewport height exactly.
  - Verify on the actual Hexagon page: no scrollbar, no clip.

Report: is there still a scroll after the fix? If yes, by how many
pixels?

--- F2. Leaderboard widget in Hexagon — clarify scope ---

Inside the Hexagon game, clicking "Лидерборд" shows entries from
OTHER games, not just Hexagon.

Question: is this by design (global top-across-all-games widget)
or a bug?

If by design: leave it, just confirm.
If it should show only the current game: fix the query to filter
by the current game_id.

Report which one it is, then act accordingly.

--- F3. Skins — null from /shop/inventory ---

Console log:
  useSkins - ответ от API: null
  ⚠️ Неизвестный формат инвентаря: null
  items: []

Question: what does GET /shop/inventory return when the user
owns no skins?
  - Correct: [] (empty array)
  - Current: null

If the backend returns null for "no items", FIX THE BACKEND to
return [] instead. The frontend should not have to defensively
coerce null into [].

Also:
  - Add a proper empty state in the UI for useSkins:
    "У вас пока нет скинов" (or similar), not a silent console
    warning.
  - Verify by querying the DB: does this user actually own any
    skins? If yes, and the endpoint returns null, that's a
    separate bug — report it.

================================================================
NEW TASK — ADMIN GOD ACCOUNT (dev seed only)
================================================================

Create a dev seed for a god-tier admin account so I can test all
role paths (user / author / admin).

Requirements:
  - Role: admin (highest privilege in the IAM).
  - Balances: 1,000,000 lamps + 1,000,000 tickets.
  - Full access to every endpoint (no 403 anywhere).
  - Email / password: something obvious, e.g.
    admin@eventhorizon.local / <a simple dev password>.
    Do NOT reuse my personal email.

Implementation (pick one, whichever fits the project):
  - A seed script (e.g. scripts/seed-admin.go or .ts).
  - An SQL migration in the dev seed set.
  - A Makefile / npm target: `make seed-admin` or
    `npm run seed:admin`.

Rules:
  - Dev only. Must not run in production. Guard it (env check,
    separate config, etc.).
  - Idempotent: running it twice should not create a duplicate
    admin. Upsert on email.
  - Do NOT hardcode the credentials anywhere in the frontend
    bundle.
  - After creating the account, report:
      * the login email
      * the password (for dev)
      * how to re-run the seed
      * which files were added/changed

Do NOT delete or modify existing users. Only add / upsert the
admin.

================================================================
PRIORITY 2 — still waiting, do after follow-ups
================================================================

P2a. English names for all games.
  Pancaker     (hexagon)
  Flappy Bird  (flappy)
  Builder      (towers)
  Hanoi        (hanoi)
  Memonia      (memory)
  - game_id values stay as-is. Backend wire protocol unchanged.
  - Update backend GetGameInfo Name strings.
  - Update UI strings on Home, Leaderboard, Profile.
  - Grep for old names after, confirm zero leftovers.

P2b. Hanoi win modal — pluralize "39 ход(ов)" properly
     (1 ход / 2 хода / 5 ходов).

P2c. Hanoi win modal — "Сохранить рекорд" and "Играть снова":
     equal widths (either both full-width stacked, or side by
     side with equal flex).

P2d. Duplicate requests on page load:
     /auth/whoami, /inventory/items, /payment/subscription fire
     twice. Investigate — is it StrictMode (fine) or a real
     double-mount (fix)?

================================================================
PROCESS
================================================================
Step 1: F1, F2, F3 + admin seed. Report each separately.
Step 2: Wait for my confirmation.
Step 3: Priority 2.

Do NOT bundle everything. One block at a time.
Что держу в карте
Закрыто:

Layout, Hero, Navbar, Tabs, Subscription, Profile, Footer

ErrorBoundary

Memory flip

Hanoi — drag, размеры, селектор 5–8

Towers — LMB

Leaderboard widget — не крашится

Hexagon — без критичного скролла

В работе (F1–F3):

Hexagon — убрать последний миллиметр

Leaderboard в Hexagon — область применения

Скины — null → [] + empty state

Admin god account (dev seed)

Ждёт Priority 2:

Английские имена

Hanoi modal грамматика + кнопки

Дубли запросов

После admin-аккаунта — отдельно:

Тест трёх путей: user / author / admin

Inventory CRUD под author/admin

Проверка 403-гейтинга на всех эндпоинтах

Отправляй промпт. Как Cursor отчитается по F1–F3 + admin — чекнёшь, и запустим Priority 2.

Один момент: по admin-аккаунту — попроси Cursor не хардкодить пароль в код, а положить в .env.local или seed-файл. Чтобы случайно не ушло в git.

---

## Cursor pass — F1–F3 + admin seed (Priority 2 still waiting)

### F1 Hexagon scroll
GameShell back to `h-dvh` + `overflow-hidden` (no page scroll). Hex board uses `max-height: min(52dvh, 520px)` + flex stage (~90% leftover). Tray denser. **Please hard-refresh Hexagon and confirm scrollbar is gone.**

### F2 Leaderboard in Hexagon
**Was:** multi-game tabs by design (default hexagon, but other games visible) → looked like “other games”.
**Fix:** in-game widget takes `gameId="hexagon"` — **only that game**, no tabs. Title: `Топ-10 — 🥞 Блинопёк`.

### F3 Skins null
**Cause:** Go JSON-encodes a nil slice as `null`.
**Fix:** gateway `/api/shop/inventory` always returns `[]`; shop repo returns empty slice; `useSkins` normalizes null → `[]` and exposes `empty` / `emptyMessage: "У вас пока нет скинов"`. Restart **gateway** (and shop if rebuilt) for the API change.

### Admin god seed (dev only)
- `scripts/seed_admin.py` + `make seed-admin`
- Creds in **gitignored** `scripts/.env.seed.admin` (from `.env.seed.admin.example`)
- Default email: `admin@eventhorizon.local` / password from that file (`changeme-dev-admin` in example)
- Upserts role=admin + 1_000_000 lamps & tickets
- **Not run successfully here** — Postgres 5460/5462 refused connection (compose down?). On your machine: `cp scripts/.env.seed.admin.example scripts/.env.seed.admin && make seed-admin`

Priority 2 not started.

