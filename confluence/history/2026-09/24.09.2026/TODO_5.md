✅ Что подтвердил Cursor
Rebuild — ✅
Три gateway пересозданы, свежий образ sha256:eed45255….

Все три healthy.

Host make rebuild-services не сработал (sandbox GOMODCACHE) — Cursor обошёл через реальный module cache.

Curl — ✅
Запрос	Ответ	HTTP
?category=profile_theme	[]	200
?category=nonexistent_cat_xyz	[]	200
?category=merch (контроль)	10 items	200
null больше нет. Shop fix живой.

Shop response shape — ✅
Оставить голый массив — намеренный контракт.

OpenAPI уже декларирует type: array для /shop/items.

/shop/inventory — тоже голый массив.

Фронт уже ждёт массив.

Inventory использует gin.H{items, total} — потому что у него пагинация. У shop каталога — нет.

Логика правильная. Разные эндпоинты — разные контракты, но осмысленно. Не менять.

Что осталось
🟧 5 backend twin'ов — отложено
Все защищены на фронте. Но бэкенд всё ещё возвращает null для пустых:

authors

history/events

analytics/dau

analytics/retention

leaderboard/entries

Это отдельная волна. Не блокер для коммита.

🟧 Остальное отложено
2d — Analytics

2c v2, proto3 medium, gin.H refactor

File upload, navbar refactor

DNS fix для docker push

Теперь можно коммитить
Финальный чек-лист перед коммитом
#	Проверка	Статус
1	npm run build — clean	⏳ прогнать
2	npx tsc --noEmit — clean	⏳ прогнать
3	git status — видно все изменения	⏳ проверить
4	/inventory — фильтры не крашатся	✅
5	/shop — пустые категории не крашатся	✅
6	/admin — 3 таба работают	✅
7	Timeout — исчез	✅
Промпт для Cursor — pre-commit
text
Final pre-commit sweep on dev/frontend.

Run and report ONLY:
  1. npm run build — pass / fail
  2. npx tsc --noEmit — pass / fail
  3. ReadLints — warnings (if any)
  4. git status --short — list of changed files
  5. git diff --stat — summary (files, +/- lines)

Do NOT commit. Do NOT touch code. Just report.

Then suggest a commit message for this batch of work,
covering:
  - proto3 zero-omission fixes (stock, price, edit prefill)
  - inventory filter null crash + backend [] fix
  - shop/items null fix + frontend coerce
  - image URL field on inventory create/edit
  - admin panel: users/roles, inventory stats
  - nav swap, burger reorder
  - role gating (user/author/admin)
После отчёта — коммитишь.

Что я держу в карте
Закрыто (полностью):

✅ proto3 фиксы (stock, price, edit prefill)

✅ Image URL

✅ Inventory filter crash — фронт + бэкенд + живой образ

✅ Shop items null — фронт + бэкенд + живой образ

✅ Admin panel 2a + 2b + 2c

✅ Nav, burger, Block 3

✅ Timeout исчез

✅ Все 3 gateway — свежий образ, healthy

Отложено (следующие волны):

🟧 5 backend twin'ов — authors, history, analytics×2, leaderboard

🟧 2d — Analytics

🟧 2c v2 (top-5, total stock, author names)

🟧 proto3 medium, gin.H refactor

🟧 File upload, navbar refactor

🟧 DNS fix для docker push

Ждёт:

🟧 Коммит dev/frontend → потом main