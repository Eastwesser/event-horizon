Вопрос 1 — Keep set
Смотрю твои скрины:

Товары → Мой инвентарь: Золотая птичка, Радужные трубы, Радужные блоки, Космические блины, Карточки со зверями — скины.

Золотая птичка — скин для Flappy (птичка золотая).

Радужные трубы — скин для Flappy (трубы радужные).

Радужные блоки — скин для Towers.

Космические блины — скин для Hexagon.

Карточки со зверями — скин для Memory.

То есть — это НЕ placeholder. Это реальные игровые скины, которые ты сам создал в v1.0.5. Их продают, покупают, применяют.

Что реально placeholder (мусор):

Ключница Дракон — брелок, тестовые данные.

Супер-пупер брелок — брелок.

Автоматический брелок ×3 — дубликаты теста.

Брелок с медведем, Медный брелок, Тестовый брелок — тесты.

Блинный мерч — мерч.

Возможно ещё пара.

Мой ответ — А.

Keep:

Все карточка (280).

Все game_skin (Золотая птичка, Радужные трубы, Радужные блоки, Космические блины, Карточки со зверями).

Все profile_theme (если есть).

Все merch — только если это реальный мерч (Блинный мерч — возможно да).

Drop:

Placeholder брелок — Ключница Дракон, Супер-пупер брелок, Автоматический брелок ×3, Брелок с медведем, Медный брелок, Тестовый брелок.

Placeholder мерч — если Блинный мерч — тест, дропнуть.

Почему A:

Скины — реальный контент. Игроки их покупают (14 в инвентаре — часть скинов).

Placeholder — тестовые брелоки. Они бесполезны в магазине, выглядят как мусор.

Профиль темы — если есть — тоже оставить.

Вопрос 2 — Push
Мой ответ — push сначала.

Причины:

Wave 1 + Wave 2 #1 — уже закоммичены (0ee567f + 6 коммитов). Они готовы.

DB cleanup — необратимая операция. Если push не сделан, а cleanup сломает → откатить сложнее.

Точка отката — если после cleanup что-то пойдёт не так — git даст откатить код.

Правило: сначала код в git, потом данные.

Порядок:

bash
# 1. Push Wave 1 + Wave 2 #1
git push origin main

# 2. Backup DB
pg_dump ... > backup_20261004_pre_cleanup.sql

# 3. Cleanup
# ...

# 4. Verify
# ...

# 5. Commit + push cleanup script (если есть)
Вопрос 3 — Промпт
Промпт Cursor'у — три ответа:

text
Answers:

1. Keep set = A.
   - Keep all карточка (280).
   - Keep all game_skin / profile_theme — they are REAL
     content (Золотая птичка, Радужные трубы, Радужные блоки,
     Космические блины, Карточки со зверями). Users bought
     them; they are not placeholders.
   - Drop only the placeholder inventory: брелок / мерч /
     картина / фенечка test items:
       - Ключница Дракон
       - Супер-пупер брелок
       - Автоматический брелок (×3)
       - Брелок с медведем
       - Медный брелок
       - Тестовый брелок
       - Блинный мерч (if it's test data — report, don't delete
         yet)
   - Do NOT delete anything that has been purchased (any row in
     shop.inventory / purchases).

2. Push FIRST.
   - I'll push Wave 1 + Wave 2 #1 (0ee567f) now.
   - After push: you draft the DB cleanup plan (counts →
     backup → delete list for my OK → delete unpurchased only).

3. Proceed with the plan draft after I confirm push is done.

Rules:
  - No DELETE without a backup file first.
  - No DELETE of purchased items.
  - Show the full delete list (id + name + type + purchased?)
    before running.
  - Drop items from both inventory and shop tables, idempotent.
Что делать сейчас
1. Push сначала:

bash
git log --oneline -5   # проверить, что 7 коммитов
git push origin main
2. Бэкап DB перед cleanup:

bash
# PG inventory
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  exec -T event-horizon-postgres-inventory pg_dump -U ... > \
  ~/backup_$(date +%Y%m%d)_inventory.sql

# PG shop
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  exec -T event-horizon-postgres-shop pg_dump -U ... > \
  ~/backup_$(date +%Y%m%d)_shop.sql
3. Отправь Cursor'у промпт выше.

4. Пока Cursor готовит план — бэкап на флешку (git уже актуален).

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f thin DTO (not pushed).

Сейчас:

🔴 Push — Wave 1 + Wave 2 #1.

🔴 Ответить Cursor'у:

Keep set = A (карточки + все скины/темы).

Push first.

🟧 DB cleanup plan — draft → approve → run.

🟧 Backup DB перед cleanup.

🟧 Бэкап на флешку.

Wave 2 (осталось):

🟧 #2 Реальные товары — в работе

🟧 #3 Лампочки как бусты

🟧 #4 Уровни сложности

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py