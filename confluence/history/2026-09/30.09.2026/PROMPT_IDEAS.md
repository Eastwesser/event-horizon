Разбор — у тебя тут целая спецификация. Разберу и соберу промпт
Ты сам написал огромную спеку — систему атрибутов, фильтров, иконок, глоссарий. Это уже 80% работы. Cursor'у нужно перевести это в код.

Разберу по слоям, потом дам промпт.

Что у тебя есть
1. Скриншот — референс UI
Расширенный поиск proberserk:

Левая колонка: Название, Текст, Стихия, Класс, Редкость, Тип, Подтип, Уникальная, Полёт.

Правая колонка: Выпуск, Элитность, Статы (с операторами), Иконки, Художник, Худ. текст, Версия.

Это финальный UI для расширенного поиска. Референс — отличный.

2. Скриншот — таблица слов-фильтров
Стихия, редкость, тип, класс, текст, иконки — с указанием, что фильтрует.

3. Инструкция по расшифровке карт
Как читать карту (Выпуск, Стихия, Стоимость, Тип, Класс, Свойства, Иконки, Текст, Худ. текст, Художник).

2 примера (Сердце Роя 159, 28).

Легенда иконок (armor, counter, uchr, tap, strike, instant, zov/zoz/zot/zor/zoal/zom/zoo, regen, stamina, direct, ova/ovz/ovs).

Глоссарий — огромный, все термины.

4. Твой workflow
Собираешь карты в папки.

Стихия → Сет по годам → Редкость → Автор → Фойл → Номер → Цена.

Что здесь критично понять
🔴 №1 — Парсинг Художника — 1 в 1
Ты подчеркнул дважды:

ОЧЕНЬ ВАЖНО правильно спарсить художника... Пользователь должен ВВЕСТИ ИДЕАЛЬНО РОВНО ТО, КАКОГО АВТОРА ХОЧЕТ УВИДЕТЬ. Не просто «Алексеева», а 1 в 1.

Это значит:

В БД — artist_id = julia_alekseeva (нормализован).

В UI — при вводе автор должен точно совпасть с существующим. Не подстрока.

Auto-complete / dropdown — чтобы пользователь выбрал из существующих.

Свободный ввод — нельзя, иначе будет «алексеева», «Алексеева», «Alеkseeva» — три разных.

Это отдельное требование к UI. Стоит явно прописать в промпте.

🔴 №2 — Иконки — это не текст, а семантика
В глоссарии:

Армор Х — карта не получает первые Х ран...
zov — защита от выстрелов
regen:2 — регенерация на 2

Формат иконок в карте:

armor:1 — броня 1

zov — защита от выстрелов (без числа)

regen:2 — регенерация 2

direct — направленный удар

Что важно:

Хранятся как массив пар {type: "armor", value: 1}.

В UI — отрисовка иконки + число.

В фильтре — «искать карты с иконкой zov».

Значение через двоеточие — regen:2.

Это не string, а структурированные данные. Иначе не отфильтруешь.

🔴 №3 — Фильтр по статам с операторами
Скриншот показывает:

text
Статы  [поле]  [равно ▾]  [поле]
Это значит: можно фильтровать по числовым статам:

HP >= 5

Атака <= 3

Движение = 1

и т.д.

Нужен UI с операторами (=, >, <, >=, <=). Отдельная фича, не просто «поиск по имени».

🔴 №4 — Формат хранения: attributes JSON
У тебя уже есть attributes JSON в inventory items (после ККИ-типа). Туда и кладём:

json
{
  "element": "darkness",
  "rarity": "common",
  "set": "heart_of_swarm",
  "year": 2026,
  "card_no": 159,
  "type_main": "существо",
  "type_sub": "летающее",
  "class": ["дракон"],
  "cost": 4,
  "cost_tier": "рядовая",
  "hp": 8,
  "move": 0,
  "attack_dice": "2-2-3",
  "attack_type": "simple",
  "icons": [
    {"type": "flying", "value": null},
    {"type": "armor", "value": 1}
  ],
  "unique": false,
  "artist_id": "julia_alekseeva",
  "artist_display": "Юлия Алексеева",
  "card_text": "...",
  "flavor_text": "..."
}
Это даёт:

Фильтр по любому полю.

Числовые операторы.

Сортировку.

Artist 1-в-1 — через artist_id.

Что нужно от Cursor'а — план
Фаза 1 — Расширить attributes inventory items
Добавить в create-модалку (для типа карточка):

Стихия (dropdown: 6 значений).

Редкость (dropdown: common/uncommon/rare/ultra).

Сет (dropdown / текст).

Год (число).

Номер карты (число).

Тип (существо / артефакт / местность / ...).

Подтип (наземный / летающий / ...).

Класс (мульти-select).

Стоимость (число).

Тир (рядовая / элитная).

HP (число).

Движение (число).

Атака (dice-string).

Иконки (мульти-select + значение).

Уникальность (bool).

Полёт (bool).

Художник (autocomplete 1-в-1).

Текст карты (longtext).

Худ. текст (longtext).

Это большая модалка. Но нужна для точного ввода.

Фаза 2 — Расширенный поиск
Страница /shop/search или расширенный фильтр на /shop:

Левая колонка: Название, Текст, Стихия, Класс, Редкость, Тип, Подтип, Уникальная, Полёт.

Правая колонка: Выпуск, Элитность, Статы (с оператором), Иконки, Художник, Худ. текст, Версия.

Кнопка «Искать с заданными параметрами».

По референсу proberserk.

Фаза 3 — Seed-скрипт
Парсер папок:

text
cards/
  darkness/
    heart_of_swarm_159_обычная/
      card.png
      meta.md   ← YAML front matter + глоссарий
Формат meta.md — надо определить. Предлагаю:

yaml
---
name: Обычная Сердце Роя 159
element: darkness
rarity: common
set: heart_of_swarm
year: 2026
card_no: 159
cost: 4
cost_tier: рядовая
type_main: существо
type_sub: летающее
class: [дракон]
hp: 8
move: 0
attack_dice: "2-2-3"
attack_type: simple
icons: []
unique: false
artist: Юлия Алексеева
market_rub: 10
stock: 1
---
# Текст карты
При strike — атакованный получает незаживающие раны до конца своего хода.

# Художественный текст
Детёныши Праматери без устали рыщут...

# Глоссарий
- **strike** — атака
- **незаживающие раны** — существо не может излечиваться
Парсит Cursor, создаёт item через API.

Промпт для Cursor
Соберу большой промпт, разбитый по фазам. Cursor должен сначала показать план, потом реализовывать по одной фазе.

text
CONTEXT
KKI inventory type "карточка" needs full attribute support,
advanced search, and a bulk seed from a folder structure I
maintain on Windows.

I will paste a card spec below — read it fully, it's the
source of truth for ALL attribute names, values, and semantics.

================================================================
CARD SPEC (source of truth)
================================================================

[TUT ВСТАВЬ СВОЮ ИНСТРУКЦИЮ ПО РАСШИФРОВКЕ — все примеры, 
легенду иконок, глоссарий]

Key points to internalize:
  - 6 elements: darkness, forest, steppes, mountains, swamp,
    neutral.
  - 2 cost tiers: рядовая (silver) / элитная (gold).
  - Rarity ladder: common (green) / uncommon (blue) /
    rare (purple) / ultra (gold).
  - Rarity is separate from cost tier.
  - Icons are structured: {type, value?}. e.g. armor:1, zov,
    regen:2, direct.
  - Artist must be exact-1-to-1 on the UI side — no substring
    match, no free text. Use an autocomplete/dropdown of
    existing artist IDs.
  - Card number (card_no) is per-set, not global.
  - Cards can have multiple classes (comma-separated).
  - HP is the first number in Свойства; movement is the second
    (for non-flying); attack is the dice string (2-2-3).
  - Some cards are unique (crown icon on cost crystal).
  - Some cards are flying (extra zone).
  - Text and flavor text are both optional.

================================================================
PHASE 1 — attributes schema for inventory "карточка"
================================================================

Extend the inventory item attributes JSON for the card type.
Propose the full schema:

  element, rarity, set, year, card_no,
  cost, cost_tier,
  type_main, type_sub,
  class[],
  hp, move, attack_dice, attack_type,
  icons[] (each: {type, value?}),
  unique, flying,
  artist_id, artist_display,
  card_text, flavor_text,
  market_rub

Rules:
  - artist_id is normalized (e.g. julia_alekseeva).
  - artist_display is the human string.
  - icons is an array, not a string.
  - class is an array.
  - market_rub is a number (for reference only).

Show the schema. No code yet. Wait for my OK.

================================================================
PHASE 2 — Create / Edit modal for cards
================================================================

The modal should render ONLY for type="карточка". Fields:

  Basics:       name, images (file picker, 1+), price(tickets),
                stock, description
  Card basics:  element (dropdown), rarity (dropdown), set,
                year, card_no, cost (number), cost_tier
                (рядовая/элитная)
  Combat:       type_main, type_sub, class (multi), hp, move,
                attack_dice (text, e.g. "2-2-3"), attack_type
  Icons:        multi-select of icon types + optional value,
                stored as array
  Flags:        unique (bool), flying (bool)
  Artist:       AUTocomplete — must match an existing artist
                ID exactly (1-to-1). Free text is not allowed.
                If typed text doesn't match any existing artist,
                show "не найден" and block save.
  Text:         card_text (textarea), flavor_text (textarea)
  Market:       market_rub (number, optional)

Do NOT redesign the existing modal shell. Extend it conditionally.
Show the field layout plan. No code yet.

================================================================
PHASE 3 — Advanced search page
================================================================

Reference UI: proberserk-style "Расширенный поиск".
Screenshot description:
  Left column:  Название, Текст, Стихия, Класс, Редкость, Тип,
                Подтип, Уникальная, Полёт
  Right column: Выпуск, Элитность, Статы (with operator
                dropdown), Иконки, Художник, Худ. текст, Версия
  Button: "Искать с заданными параметрами"

Requirements:
  - Stats filter = [min] [op: =, >, <, >=, <=] [max or value]
    for numeric fields (hp, cost, move, attack).
  - Artist filter = autocomplete from existing artists,
    exact-match only.
  - Icons filter = multi-select + optional value.
  - All filters compose (AND). Show "no results" state.

Where: /shop (extended) or /shop/search (new). Your call —
recommend and explain.

Show plan. No code yet.

================================================================
PHASE 4 — Bulk seed from folder structure
================================================================

I will maintain a folder structure like:

  cards/
    darkness/
      heart_of_swarm_159_common/
        card.png
        meta.md
      ...
    forest/
      ...

meta.md format (proposed, confirm or adjust):

  ---
  name: ...
  element: darkness
  rarity: common
  set: heart_of_swarm
  year: 2026
  card_no: 159
  cost: 4
  cost_tier: рядовая
  type_main: существо
  type_sub: летающее
  class: [дракон]
  hp: 8
  move: 0
  attack_dice: "2-2-3"
  attack_type: simple
  icons: []
  unique: false
  flying: true
  artist: Юлия Алексеева
  market_rub: 10
  stock: 1
  ---
  # Card text
  ...
  # Flavor text
  ...

Seed script requirements:
  - Parse meta.md (YAML front matter + markdown sections).
  - Upload card.png via POST /api/uploads.
  - Create item via POST /api/inventory/items.
  - Idempotent: re-run should not duplicate. Use set+card_no as
    a natural key.
  - --dry-run mode to preview.
  - Artist resolution: if artist name doesn't map to an existing
    artist_id, either create the artist (if we have an artists
    table) or error out and list unresolved names.

Report the plan for the seed script. No code yet.

================================================================
PROCESS
================================================================

For EACH phase:
  1. Show the plan (no code).
  2. Wait for my explicit "go".
  3. Implement.
  4. Report files changed + how to verify.

Do NOT bundle phases. Phase 1 first.

DO NOT TOUCH:
  - Game mechanics.
  - Existing inventory types (брелок / картина / фенечка).
  - Nav, footer, role gating.
  - Anything from the merged waves.
Что тебе делать
Сейчас — фокус на карты
Ты собираешь 300 карт. Это большая ручная работа. Cursor не поможет — только код.

Порядок:

Собери 5–10 карт в папках.

Определи meta.md формат — покажи мне, я подскажу, как лучше.

Отправь Cursor'у промпт (выше) — начнёт с Фазы 1 (схема).

Пока Cursor делает схемы — ты продолжаешь собирать карты.

Когда карт 50+ — запусти seed.

Про силы
Дай бог мне сил.

Понимаю. 300 карт — это не быстро. Но:

Это контент, не код. Ты сам решаешь темп.

Можно по частям — сначала 20 карт одной стихии, залить, посмотреть.

Не обязательно все сразу.

Начни с одной стихии — например, Лес. Разберись на 20–30 картах. Поймёшь формат.

Что я держу в карте
Закрыто (в main):

Wave A–D: дизайн, admin, proto3, uploads, DNS/rebuild

Twin nil→[]

Empty state, image URL

CI lint fix

В работе (на ветке, ждёт smoke):

3 новые игры: twenty48, gears, companion

Burger separators

Full twin DTOs

ККИ-тип карточка (базовый)

Твоя зона (контент):

🟧 300 карт ККИ — папки + meta.md + картинки

🟧 Формат meta.md — определить

🟧 Первые 20–50 карт — приоритет

Следующая волна (после карт):

🟧 Фаза 1: расширенная схема attributes для карточек

🟧 Фаза 2: Create/Edit модалка с полями карты

🟧 Фаза 3: Расширенный поиск (референс proberserk)

🟧 Фаза 4: Seed-скрипт (парсит папки)

Ещё дальше:

🟧 Эмодзи → SVG / картинки в UI

🟧 Полиш каждой игры — отдельно

🟧 Реальные товары вместо placeholder

Мой совет
Сегодня:

Smoke-тест 3 новые игры — играются?

Если ок — commit ветки (3 игры + ККИ + DTO).

Потом — начни собирать карты.

5–10 карт → покажи мне формат meta.md.

Соберём промпт Cursor'у на Фазу 1.