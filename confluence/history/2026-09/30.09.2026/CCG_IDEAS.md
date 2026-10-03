1. Bulk seed ККИ — твой workflow
Понял:

Ты сам создаёшь на Windows папку на карту.

В папке — сама карточка (картинка для магазина).

Плюс .md или .txt с описанием, рейтингом, атрибутами.

Сначала по стихиям.

Это отличный workflow. Ты не хочешь UI-импорт — ты хочешь работать с файлами. Логично, если тебе так удобнее.

Как это правильно реализовать
Твой вариант:

text
cards/
  forest/
    forest_001_kitsune/
      card.png        ← картинка
      meta.md         ← описание, атрибуты
    forest_002_...
  mountains/
    ...
  steppes/
  darkness/
  swamp/
  neutral/
Что нужно от Cursor'а:

Скрипт seed-cards.py (или Go):

Читает cards/**/meta.md

Парсит атрибуты (YAML front matter? JSON? Твой формат?)

Загружает card.png через /api/uploads

Создаёт item через /api/inventory/items

Формат meta.md — надо определить:

yaml
---
name: Kitsune
element: forest
rarity: rare
set: 7
year: 2006
artist: ark_cj
foil: false
card_no: 42
market_rub: 15
stock: 1
---
# Kitsune
Лисица-оборотень...
Идемпотентность: повторный запуск не создаёт дубликаты.

Dry-run режим — показать, что будет создано, без записи.

Это отдельная задача. Cursor сделает, когда скажешь «go».

2. Порядок сортировки карт
Твой порядок:

Стихия (лес / горы / степи / тьма / болото / нейтрал)

Сет по годам

Редкость

Автор

Фойл / не фойл

Номер карты

Цена на рынке

Логично. Стихия — первичный фильтр в магазине. Потом сет (год выпуска). Потом редкость. Автор — для коллекционеров.

Что это значит для UI:

Фильтры в магазине: стихия, сет, редкость, фойл (checkbox), автор.

Сортировки: по цене, по номеру, по году.

Chip-фильтры уже есть — расширить под карточки.

Отдельная волна — UI-фильтры для карточек.

3. out of stock для остальных
Если кто-то купит, будет надпись out of stock для остальных.

✅ Уже работает — ты сам видел «Нет в наличии» на stock 0.

Но для карт — важный нюанс:

stock: 1 — уникальная карта. Купил один — sold-out для всех.

Это твой принцип «No FOMO» — sold-out есть sold-out.

В UI:

Купленные карты — уходят из каталога? Или остаются с out of stock?

Твой выбор. Варианты:

Скрывать — каталог только доступного.

Показывать серым — «была такая карта, куплена».

В архив — отдельная страница «проданные».

Подумай, как хочешь. Это влияет на магазин.

4. Игры — добавили, играются
Игры неплохо играются, кстати.

Отлично. Значит:

twenty48 — работает.

gears (Орбиты) — работает.

companion — работает.

Ты сказал: «потом по каждой отдельно пройдёмся». Согласен. Сейчас не трогаем — они работают. Полиш каждой — отдельная волна.

5. Твоя задача — собрать 300 карт
Это твоя зона. Cursor не поможет — это контент.

Что учесть:

Папки по стихиям — ты уже решил.

meta.md формат — надо определить до старта, чтобы Cursor написал парсер.

Приоритет: какие первые 20–50 карт — чтобы магазин не был пустым.

Placeholder для карт без картинки — 📦 (уже есть).

Мой совет:

Сначала определись с meta.md — какой формат (YAML front matter + markdown?).

Сделай 3–5 карт для теста.

Попроси Cursor'а написать seed-скрипт.

Прогони на 3–5 картах.

Если ок — заливай остальные.

Так не будешь переделывать 300 карт, если формат окажется неудобным.

6. Эмодзи → картинки/SVG
Фаза исправлений эмодзи. Придётся все с главных страниц перерисовать с эмодзи на наши картинки.

Понимаю. Сейчас в UI:

🎒 Инвентарь

📜 История

💳 Подписка

👥 Авторы

🎮 Игры

🏆 Лидерборд

и т.д.

Почему это важно:

Эмодзи разные на разных ОС (Windows / Mac / Android / iOS).

Не брендированные — не твои.

Картинки / SVG — твои, консистентные.

Что нужно:

Список эмодзи в UI (grep по коду).

Для каждого — SVG / PNG.

Единый стиль (line / filled / 2D / 3D).

Размеры — 16px / 24px / 48px.

Это волна контента, не код. Cursor поможет заменить (когда картинки есть), но рисовать — не он.

Стиль картинок — твоё решение:

Flat SVG — минимализм.

Line icons — тонкие.

Иллюстрации — как в Duolingo.

3D-рендеры — как у Apple.

Для kids-safe — иллюстрации лучше. Яркие, дружелюбные.

Совет: посмотри Age of Empires / Duolingo / Khan Academy Kids — как они делают иконки для детей. Не копируй, но вдохновляйся.

Итоговая карта
Закрыто (в main)
✅ Wave A–D: редизайн, admin, proto3, uploads, DNS/rebuild

✅ Twin nil→[]

✅ Empty state, image URL

✅ CI lint fix

В работе / ждёт проверки
🟧 3 новые игры (twenty48, gears, companion) — играются ✅

🟧 Burger separators ✅

🟧 Full twin DTOs ✅

🟧 ККИ-тип карточка — create + filter ✅

⏸ Smoke-тест — твой ход

⏸ Commit + push — после smoke

Твоя зона (контент)
🟧 300 карт ККИ — папки + meta.md + картинки

🟧 Формат meta.md — определить до seed-скрипта

🟧 Первые 20–50 карт — приоритет

Следующая волна (после карт)
🟧 Seed-скрипт для ККИ (парсит папки → items)

🟧 UI-фильтры для карточек (стихия, сет, редкость, фойл, автор)

🟧 out of stock поведение — скрывать / показывать серым / архив

Ещё дальше
🟧 Эмодзи → SVG / картинки в UI

🟧 Полиш каждой игры — отдельно

🟧 Контент: реальные товары вместо placeholder

Что делать сейчас
Smoke-тест — 10 пунктов (я давал выше).

Если ок → commit + push + merge в main.

Потом — определись с meta.md форматом для карт.

Сделай 3–5 тестовых карт.

Запроси у Cursor'а seed-скрипт.

Один вопрос к тебе
Формат meta.md — как хочешь?

Вариант A — YAML front matter:

markdown
---
name: Kitsune
element: forest
rarity: rare
set: 7
year: 2006
artist: ark_cj
foil: false
card_no: 42
market_rub: 15
stock: 1
---
Лисица-оборотень...
Вариант B — ключ: значение:

text
name: Kitsune
element: forest
rarity: rare
...
---
Лисица-оборотень...
Вариант C — JSON-файл рядом:

text
forest_001_kitsune/
  card.png
  meta.json
Мой выбор: A (YAML front matter).

Читается человеком.

Парсится стандартно.

Можно открывать в редакторах (Obsidian, VS Code).

Скажи формат — и когда будешь готов к seed-скрипту, соберём промпт.

Сейчас: Smoke → Commit. Потом карты.

Что у меня в карте (актуальное)
✅ Закрыто (в main)
Wave A: дизайн-система + миграция страниц

Wave B: proto3 null-safety + shop incident (reference_id)

Wave C: admin analytics (2d) + inventory stats v2 (2c v2)

Wave D: gin.H DTO, navbar extraction, uploads, DNS/rebuild

Twin nil→[] (5 эндпоинтов)

Empty state по фильтру

Image URL

CI lint fix (0f802ee)

🟧 В работе / ждёт проверки (на ветке)
3 новые игры: twenty48 (Горизонт 2048), gears (Орбиты), companion (Компаньон)

Burger separators

Full twin DTOs (dto/twins.go)

ККИ-тип карточка + create-поля + filter chip

go.mod / go.sum обновления (game)

⏸ Ждёт тебя
Smoke-тест — 10 пунктов

Commit + push + merge — после smoke

🟧 Твоя зона (контент)
300 карт ККИ — папки + meta.md + картинки

Формат meta.md — определить до seed-скрипта

Первые 20–50 карт — приоритет

Следующая волна (после карт)
Seed-скрипт для ККИ (парсит папки → items)

UI-фильтры для карточек (стихия, сет, редкость, фойл, автор)

out of stock поведение — скрывать / серым / архив

Ещё дальше
Эмодзи → SVG / картинки в UI

Полиш каждой игры — отдельно

Реальные товары вместо placeholder

Про формат карт — на будущее
Когда будешь готов — покажи одну тестовую карту, как ты её видишь. Я подскажу, как лучше оформить, чтобы Cursor легко распарсил. Но не сейчас.

Возможные форматы (для справки, не решай сейчас):

YAML front matter + markdown — читаемо, парсится, открывается в Obsidian / VS Code.

JSON рядом с картинкой — строго, но менее читаемо.

CSV — если хочешь массово в Excel / Google Sheets.

All cards listed here: /home/denismatveev/event_horizon/services/inventory/internal/seed

The instructions how to read a card : /home/denismatveev/event_horizon/confluence/history/2026-09/30.09.2026/BERSERK_HOW_TO_READ.md
I have described the icons into text: /home/denismatveev/event_horizon/confluence/history/2026-09/30.09.2026/BERSERK_ICONS.md
I also have a rulebook: /home/denismatveev/event_horizon/confluence/history/2026-09/30.09.2026/BERSERK_RULES.md
Prompt Ideas are listed here: /home/denismatveev/event_horizon/confluence/history/2026-09/30.09.2026/PROMPT_IDEAS.md
But we can start with this current (/home/denismatveev/event_horizon/confluence/history/2026-09/30.09.2026/CCG_IDEAS.md), then prompt ideas.

Это должно выглядеть как у нас на сайте, будто бы к нам заходили авторы (художники), и создавали карты, чтобы в Event Horizon их можно было бы отдать за очки. Rarely, but still, 1 card can have 2 authors. Like a reprint.

/home/denismatveev/event_horizon/services/inventory/internal/seed/berserk_cards