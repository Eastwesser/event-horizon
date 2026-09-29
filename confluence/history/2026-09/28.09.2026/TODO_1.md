Разбор отчёта Cursor'а
Смотри, тут не просто fix — тут сразу 5 крупных кусков в одной волне. Разберу.

Что он сделал
1. gin.H DTO — долгосрочный фикс proto3
Это то, о чём я говорил несколько раз. Он вынес маппинг в internal/dto, теперь price: 0, stock: 0, images: [] не теряются при сериализации. Применено к list / create / get / update inventory.

Это правильное архитектурное решение. Убирает весь класс багов с proto3 omitempty, не только те 5 twin'ов, которые мы нашли.

2. Navbar — вынесен из Home.tsx
То, что я флагил как «архитектурный запах». Теперь AppNavbar / AppFooter отдельно. Правильно.

3. Uploads — file upload!
Не URL, а настоящая загрузка файлов.

POST /api/uploads (author/admin).

Статика /uploads/*.

FE — file picker вместо URL.

Compose volume — общий для контейнеров.

Это был deferred item. Сделано.

4. DNS / rebuild — инфраструктурные фиксы
GOMODCACHE в rebuild скриптах.

scripts/dns-check.sh.

Это значит — make rebuild-services больше не падает на sandbox, и есть диагностика DNS. То, что мы вручную обходили.

5. Всё билдится
Gateway builds.

Frontend tsc — clean.

Что от тебя требуется
Две команды + smoke:

bash
bash scripts/rebuild-services.sh gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3 balancer
Почему balancer тоже — Cursor явно включил его в команду. Возможно, нужно перечитать конфиг или зарегистрировать обновлённые бэкенды. Внимательно следуй его команде, не убирай balancer.

Потом smoke:

Inventory create с загрузкой файла — работает?

Item со stock: 0 — показывает «Нет в наличии»?

⚠️ На что обратить внимание
1. Загрузка файлов — где хранятся?
shared compose volume

Вопрос: это volume на диске или S3? Если volume — при пересоздании контейнера данные сохранятся? Проверь конфиг compose.

Проверь:

Файл загружен → лежит в ./uploads/ или внутри Docker volume?

После docker compose down/up — файл на месте?

В проде — что с бэкапами?

2. Что с URL-полем?
Раньше был URL input, теперь file picker. Но что стало с существующими предметами, у которых images: [url]? Работают? Или старый URL-формат сломался?

Проверь:

Открой существующий предмет с картинкой по URL.

Превью отображается?

3. internal/dto — не сломал ли существующие эндпоинты?
Cursor переписал маппинг inventory. Проверь:

Список items — все поля на месте?

Create — работает?

Update — работает?

Get one — все поля?

4. Navbar extraction — не сломал ли burger / admin link?
Бургер работает?

Админ-панель показывается под admin?

Nav swap (Профиль · Лидерборд · Магазин) — сохранён?

Что ответить Cursor'у
Если всё ок после smoke:

text
Ships look clean. Rebuild + recreate done, smoke passes:
  - Inventory create with file upload: works
  - Item with stock 0: shows «Нет в наличии»
  - Navbar/burger: works, admin link visible under admin
  - Old items with URL images: still render
  - /admin tabs: fine

Commit on this branch:
  feat(deferred): gin.H DTOs, navbar extraction, uploads, dns/rebuild

Before commit — confirm:
  1. .gitignore covers the uploads volume dir (do not commit files).
  2. The uploads volume path in compose is documented.
  3. POST /api/uploads validates file type/size (security).
Если что-то сломалось — конкретно, по пунктам.

Что я держу в карте
Закрыто (в main):

✅ Вся первая волна редизайна + admin panel + null-safety

В работе (feat/deferred-wave-finish):

🟧 gin.H DTO

🟧 Navbar extraction

🟧 Uploads (file upload)

🟧 DNS/rebuild infra

После smoke:

🟧 Commit + push feat/deferred-wave-finish

🟧 PR → merge в main

Отложено (следующее):

🟧 Backend twin'ы (authors, history, analytics×2, leaderboard) — частично закрыты gin.H DTO? Уточнить у Cursor'а

🟧 2d — Analytics

🟧 2c v2 — top-5, total stock, author names

🟧 Разделители в бургере

🟧 Empty state по фильтру (UX-мелочь)

Вопрос к Cursor'у перед commit'ом
Стоит спросить: закрыл ли gin.H DTO только inventory, или все 5 twin'ов тоже?

Если только inventory — твины остались. Отдельная волна.

text
Quick clarification before commit:

The gin.H DTO layer — does it cover ONLY inventory items, or did
you also route the following endpoints through it?

  GET /api/authors
  GET /api/history/events
  GET /api/analytics/dau
  GET /api/analytics/retention
  GET /api/leaderboard/...

Report status for each. If they still return raw proto (nil →
null for empty), they need the same treatment in a follow-up.
Твои шаги
Запусти rebuild + recreate (две команды выше).

Smoke 5 пунктов (список выше).

Отправь Cursor'у:

approve на commit, если всё ок,

или конкретику, если что-то сломалось,

или уточняющий вопрос про twin'ы (промпт выше).

После commit'а — push feat/deferred-wave-finish, PR → merge в main.

Если хочешь — можно уточнен