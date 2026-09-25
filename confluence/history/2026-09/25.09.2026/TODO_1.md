TODO на завтра
🔴 Первое — открыть PR
GitHub даёт ссылку:

text
https://github.com/Eastwesser/event-horizon/pull/new/dev/frontend
Что сделать:

Открой ссылку.

PR: dev/frontend → main.

Title: feat(frontend): admin panel, null-safety, UI refresh.

В описании — список изменений (Cursor уже сгенерил commit message, можно взять оттуда).

Не мержить сразу. Сначала smoke-тест на dev/frontend через PR-превью или локально.

🟧 Второе — smoke-тест перед мержем в main
Прогони по чек-листу на dev/frontend:

#	Проверка
1	Home — играется, hero, 5 игр
2	Login / Register
3	Shop — покупка, списание билетиков
4	Inventory — CRUD (под author/admin)
5	Inventory — фильтры (Картина / Фенечка → empty)
6	Games — все 5
7	Subscription (3 состояния)
8	Profile
9	Admin /admin — users/roles, stats
10	Leaderboard — без крашей
Если всё ок → merge.

🟧 Третье — после мержа
Следующие волны. По приоритету:

А. Backend twin'ы — 5 штук:

authors

history/events

analytics/dau

analytics/retention

leaderboard/entries

Все защищены на фронте, но бэкенд отдаёт null. Фикс — по паттерну inventory/shop.

Б. 2d — Analytics в админке:

DAU / MAU / retention.

Big-number cards + bar/list.

В. 2c v2 — Inventory stats:

Top-5 expensive.

Total stock.

Author name enrichment (UUID → email).

Г. Proto3 medium / gin.H refactor:

Долгосрочный фикс: gateway → gin.H DTO или protojson.EmitUnpopulated.

Уберёт весь класс багов с null.

🟧 Мелочи (отложено)
File upload (S3/disk) — если понадобится.

Navbar refactor — вынести из Home.tsx в отдельный Layout.

Разделители в бургере — когда меню вырастет.

Empty state по фильтру — «В категории пусто» вместо «Товаров пока нет».

DNS fix для docker push.

make rebuild-services — сейчас ломается на sandbox GOMODCACHE. Может, поправить.

⚠️ Не забыть
.agents/ и skills-lock.json — остались вне коммита. Реши, нужны ли в репо. Если нет — добавь в .gitignore.

confluence/history/2026-09/ — ты закоммитил? Cursor пишет docs: seed-admin + September history. Значит закоммичено. Ок.

Пароли seed'ов — changeme-dev-author, changeme-dev-user. Перед продом сменить.

.env.seed.admin — gitignored, не ушёл.

Что закрыто (для истории)
За сессию закрыто:

Дизайн-система (индиго + золото, токены)

Все страницы на Tailwind + примитивы

Navbar, burger, footer

Все 5 игр — chrome, drag/flip, GAME OVER

Subscription — 3 состояния, none → человеческий текст

Shop + билетики + инцидент reference_id

Inventory CRUD + role gating

Block 3 — безопасность (401/403 на API)

Admin panel — 2a + 2b + 2c

Proto3 zero-omission фиксы (stock, price, edit prefill)

Image URL для товаров

Inventory filter crash

Shop items null twin

Timeout 15s — ушёл

Gateway cluster — все 3 свежие

6 коммитов на dev/frontend. Всё в git. Точка отката есть.