ВЫНОШУ ПРАВИЛО НАВЕРХ - НЕ ТРОГАЙ КАРТОЧКИ ККИ БЕРСЕРК, ИХ 280!!! У НАС ОНИ - ЭТО РЕАЛЬНЫЙ МЕРЧ, НЕ НАВРЕДИ!!!

Понял. Не «выкинуть всё», а «оставить по одному примеру каждого типа»
Что ты хочешь
Сейчас в shop:

281 карточка (ККИ) — ок.

5 скинов — ок.

291 merch — мусор (тестовые брелоки, junk).

1 картина (Kitsune — test).

Фенечек — нет.

Ты хочешь:

Убрать generic junk.

Оставить по 1 примеру на каждый тип: брелок, картина, фенечка, мерч + карточки, скины.

Каждый пример — с описанием, картинкой, ценой.

Можно — сгенерировать шаблон карточки товара-примера (описание, картинка, атрибуты).

Дальше — ты сам засеиваешь реальные товары вручную (не сейчас).

План — 3 шага
Шаг 1 — Keep / Delete list
KEEP:

Все карточки (281).

5 скинов (dedupe).

По 1 примеру каждого типа:

Брелок — 1 (создать example-брелок).

Картина — 1 (создать example-картина).

Фенечка — 1 (создать example-фенечка).

Мерч — 1 (создать example-мерч).

HIDE от shop (available=false, но visible в inventory владельца):

Ключница Дракон, Супер-пупер брелок, Автоматический брелок, Брелок с медведем, Блинный мерч, C3 Keychain — purchased (у кого-то есть).

DELETE (soft, если не purchased):

ProtoZeroTest, Kitsune (текущий), Vika, Berserk (generic), C3 Smoke Card.

Шаг 2 — Создать по 1 example каждого типа
4 новых item (seed SQL или seed script):

Example Брелок — описание, картинка, цена.

Example Картина — описание, картинка, цена.

Example Фенечка — описание, картинка, цена.

Example Мерч — описание, картинка, цена.

Стиль:

Generic, но осмысленный — «Космический брелок», «Космическая картина», и т.п.

В стиле Event Horizon (тёмный космос).

Картинка — placeholder или SVG из gameIcons? Или ты предоставишь реальные картинки позже?

Шаг 3 — SQL/script
Idempotent.

Backup перед apply.

Verify — /shop grid: 281 карточка + 5 скинов + 4 example = 290 items.

Вопросы к тебе
1. Example items — какие картинки?

Placeholder (📦 emoji)?

SVG из Icon набора?

Ты предоставишь картинки позже?

2. Example prices — какие?

100 билетиков? 1000?

Или — placeholder = 1?

3. profile_theme — есть в типах, но в БД нет. Нужен example? Или пропустить?

4. Дубликаты скинов (14 → 5):

Правило dedupe — твой выбор:

Keep самый старый (первый).

Keep самый новый.

Keep purchased (если owned, скрыть остальные).

5. Seed script или SQL?

SQL — быстро, одноразово.

Seed script — повторно idempotent.

Промпт Cursor'у
text
Shop cleanup v2 — keep examples, drop junk.

1. Keep / hide / delete list — as in your plan, adjusted:

   KEEP in shop:
     - All Berserk cards (~281)
     - 5 real skins (dedupe 14 → 5): Золотая птичка,
       Радужные трубы, Радужные блоки, Космические блины,
       Карточки со зверями
     - NEW: 1 example of each remaining type:
         · example-брелок
         · example-картина
         · example-фенечка
         · example-мерч

   HIDE from shop (available=false, keep in owner inventory):
     - Ключница Дракон, Супер-пупер брелок,
       Автоматический брелок, Брелок с медведем,
       Блинный мерч, C3 Keychain — purchased
     - extra skin duplicates that are owned

   SOFT-DELETE (unpurchased test junk):
     - ProtoZeroTest, Vika, Berserk generic, C3 Smoke Card
     - Kitsune (картина) — replaced by example-картина

2. NEW examples — content:
   - Style: Event Horizon / cosmic, in Russian.
   - Image: use an SVG placeholder from Icon set
     (e.g. generic icons per type) — I'll replace with
     real art later.
   - Price: 100 tickets each (placeholder).
   - Description: one short RU line each, in cosmic tone.

3. Dedupe rule for skins:
   - Keep one canonical id per (name, game_id).
   - Prefer the id that has purchases; else oldest.
   - Extras: if owned → hide (available=false); else soft
     delete.

4. SQL / seed:
   - Idempotent (available / deleted_at guards).
   - Backup shop + inventory before apply.
   - Report row counts before/after.

5. Verify:
   - /shop: cards + 5 skins + 4 examples only.
   - Owner inventories: still see what they bought.
   - No "already owned" regressions.

6. profile_theme: no rows today. Skip; add example later
   only if we introduce that category.

Show the SQL/seed script first. No DB write until I OK.
Что я держу в карте
Закрыто (main, f784419):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

8 игр playable.

Сейчас — Shop cleanup v2:

🟧 Keep — карточки, 5 скинов, 4 example (брелок, картина, фенечка, мерч).

🟧 Hide — purchased placeholders.

🟧 Soft-delete — unpurchased junk.

🟧 Dedupe — 14 → 5 скинов.

🟧 New examples — 4 item с описанием, картинкой, ценой.

Дальше — options:

🟧 Track B parity (boost, levels).

🟧 Пауза / собесы.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.