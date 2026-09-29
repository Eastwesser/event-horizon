Что делать сейчас
Twin'ы закрыты → можно смело коммитить deferred wave.

Шаг 1 — Rebuild + recreate
bash
bash scripts/rebuild-services.sh gateway
docker compose --env-file .env \
  -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate gateway gateway-2 gateway-3 balancer
И для inventory (если менялся):

bash
bash scripts/rebuild-services.sh inventory
docker compose --env-file .env \
  -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate inventory
Шаг 2 — Smoke (5 пунктов)
#	Проверка	Ожидание
1	Inventory create с файлом (не URL)	Файл загружается, картинка в карточке
2	Item с stock: 0	«Нет в наличии»
3	Item с price: 0	«Бесплатно»
4	Существующие items с URL-картинками	Превью работает (не сломалось)
5	Navbar + burger	Работают, admin link виден под admin
Бонус (если быстро):

Открой /admin → таблица users — грузится.

Игра в Hexagon — drag работает.

Шаг 3 — Commit
Если всё ок:

text
Commit on this branch:
  feat(deferred): gin.H DTOs, navbar extraction, uploads, dns/rebuild

Push to feat/deferred-wave-finish, then PR → main.
Перед коммитом уточни у Cursor'а:

text
Before commit, confirm:
  1. .gitignore covers the uploads volume dir (do not commit
     uploaded files).
  2. Uploads volume path documented in compose.
  3. POST /api/uploads validates file type and size (security).
Это важно — иначе загруженные файлы улетят в git, или через upload можно будет залить что угодно.

Что я держу в карте
Закрыто (в main):

✅ Вся первая волна: редизайн, admin panel, proto3 фиксы, null-safety, image URL

В работе (feat/deferred-wave-finish):

🟧 gin.H DTO (inventory)

🟧 Navbar extraction

🟧 Uploads (file upload)

🟧 DNS/rebuild infra

🟧 twin'ы — уже закрыты (не в этой волне)

После smoke:

🟧 Commit + push feat/deferred-wave-finish

🟧 PR → merge в main

Отложено (следующее):

🟧 2d — Analytics в админке

🟧 2c v2 — top-5, total stock, author names

🟧 Разделители в бургере

🟧 Empty state по фильтру (UX)

🟧 Full DTO mapper для 5 twin'ов (сейчас — только nil→[])

Мой совет
Rebuild → smoke → commit → push → PR.

Перед commit'ом — обязательный вопрос про uploads (gitignore + валидация). Это единственное белое пятно в этой волне.

После мержа — выбираешь:

2d (Analytics) — относительно быстро.

2c v2 (top-5, total stock).

Full DTO для 5 twin'ов — архитектурный долг.

Мелочи (разделители, empty state).

Держу карту. Что делаем — rebuild + smoke, или сразу коммит?