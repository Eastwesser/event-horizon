Скриншот 1 — Home (главная страница)
Контент:

Hero: заголовок «Выбери игру и стань рекорды», подзаголовок «Играй в мини-игры, зарабатывай лампочки и билетики, становись лучшим в лидерборде.»

Кнопки Все игры (золото), Лидерборд (тёмная).

Справа — accretion disk с EVENT HORIZON.

Секция Игры — 8 карточек в сетке 4×2:

Pancaker 🥞 — «Гексагональный пазл с блинчиками»

Flappy Bird 🐦 — «Лети и не врезайся в трубы»

Builder 🗼 — «Строй башню из падающих блоков»

Hanoi 🪈 — «Классическая головоломка с кольцами»

Memonia 🎴 — «Найди пары фруктов»

Горизонт 2048 🔢 — «Сдвинь плитки — собери 2048»

Орбиты ⚙️ — «Сливай шестерёнки до восьмой»

Компаньон ⭐ — «Мягкий тамагочи без FOMO-смерти»

Проблема: все 8 карточек используют emoji-иконки (🥞 🐦 🗼 🪈 🎴 🔢 ⚙️ ⭐). Не заменены на SVG.

Откуда: Home.tsx (games array / game config).

Скриншот 2 — Burger menu (dropdown)
Контент:

Верхнее nav: 👤 Профиль, 🏆 Лидерборд, 🛒 Магазин, бургер ☰, Выйти.

Открытый burger dropdown:

📦 Инвентарь

📜 История

👥 Авторы

📋 Подписка

(разделитель)

⚙️ Админ-панель

📊 Аналитика

Статус: ✅ Emoji заменены на иконки (Cursor справился). Это твой референс "как должно быть".

Скриншот 3 — Profile (профиль)
Контент:

Плашка user: admin 🚀, admin@eventhorizon.local, ▶ Показать ID.

9 stat-карточек (3×3):

310 Всего очков 🏆

0 Pancaker 🥞

0 Memonia 🎴

0 Flappy Bird 🐦

190 Builder 🗼

0 Hanoi 🪈

0 2048 🔢

0 Орбиты ⚙️

0 Компаньон ⭐

Секция Достижения: badge Builder Master 🏆

Валюты: 💡 1000036 лампочек, 🎫 975243 билетиков

Секция Рекорды по играм — список 8 игр:

Pancaker 🥞 — 0

Memonia 🎴 — 0

Flappy Bird 🐦 — 0

Builder 🗼 — 190

Hanoi 🪈 — 0

2048 🔢 — 0

Орбиты ⚙️ — 0

Компаньон ⭐ — 0

Проблемы:

Emoji в stat-карточках (🥞 🎴 🐦 🗼 🪈 🔢 ⚙️ ⭐).

Emoji в валютах (💡 🎫).

Emoji в достижениях (🏆).

Emoji в списке рекордов.

Откуда: Profile.tsx.

Скриншот 4 — Leaderboard
Контент:

Заголовок 🏆 Лидерборд — emoji в заголовке.

Chips игр (2 ряда):

Ряд 1: 🥞 Pancaker (выбран), 🐦 Flappy Bird, 🎴 Memonia, 🗼 Builder, 🪈 Hanoi, 🔢 2048

Ряд 2: ⚙️ Орбиты, ⭐ Компаньон

Таблица:

# НИК ОЧКИ

Пустая строка → ? АНОНИМ — 0 🏆

Проблемы:

Emoji в заголовке (🏆).

Emoji в chips игр (🥞 🐦 🎴 🗼 🪈 🔢 ⚙️ ⭐).

Emoji в колонке очков (🏆).

Откуда: LeaderboardFull.tsx — конкретно массив GAME_TABS:

ts
const GAME_TABS: { id: GameId; label: string; icon: string }[] = [
  { id: 'hexagon', label: 'Pancaker', icon: '🥞' },
  { id: 'flappy', label: 'Flappy Bird', icon: '🐦' },
  { id: 'memory', label: 'Memonia', icon: '🎴' },
  { id: 'towers', label: 'Builder', icon: '🗼' },
  { id: 'hanoi', label: 'Hanoi', icon: '🪈' },
  { id: 'twenty48', label: '2048', icon: '🔢' },
  { id: 'gears', label: 'Орбиты', icon: '⚙️' },
  { id: 'companion', label: 'Компаньон', icon: '⭐' },
];
Скриншот 5 — Shop (магазин, вкладка «Товары»)
Контент:

Заголовок 🛍 Магазин — emoji в заголовке.

Справа: 🎫 975243 билетиков.

Вкладки: 🛒 Товары (выбран), 📦 Мой инвентарь (14).

Chips типов с emoji:

📦 Все

🎴 Карточки

🎨 Скины

🎭 Темы

👕 Мерч

🔑 Брелок

🖼 Картина

🪶 Фенечка

Сортировка: Новые ▾

Фильтры ▾

Показано 100 · отфильтровано 289 из 289

Сетка карточек (4 в ряд): карты ККИ (Дикий сугед, Живая баллиста, Пардус, Поварёнок) — с реальными картинками.

Каждая карточка: карточка badge, С Компаньон badge, В наличии: N, 🎫 10000, кнопка Купить.

Проблемы:

Emoji в заголовке (🛍).

Emoji в валюте (🎫).

Emoji в вкладках (🛒 📦).

Emoji в chips типов (📦 🎴 🎨 🎭 👕 🔑 🖼 🪶).

Emoji в цене (🎫).

Скриншот 6 — Shop → Мой инвентарь
Контент:

Заголовок 🛍 Магазин.

🎫 975243 билетиков.

Вкладки: 🛒 Товары, 📦 Мой инвентарь (14) (выбран).

Сетка карточек (4 в ряд, 14 штук) — 4 ряда.

Каждая карточка:

Иконка подарка (🎁 emoji) вместо превью товара.

Обрезанное название: Пр..., Ра..., Ко..., Бл..., Зо..., Бр..., Авт..., Су..., Кл... — все обрезаны.

Обрезанное описание: Куплено: 03.10.2026 (видно только дата).

Кнопка Отменить — большая, занимает правую половину карточки, наезжает на текст.

Есть ещё кнопка Отменить покупку где-то (в detail).

Проблемы:

Emoji 🎁 вместо иконки инвентаря/превью товара.

Обрезанные названия — карточки слишком узкие.

Кнопка Отменить слишком большая — занимает много места, выглядит громоздко.

Нет превью картинки товара — только иконка подарка. У купленных карт ККИ есть картинка, но в инвентаре показывается 🎁.

Сводка — что где искать
Файл	Что искать
Home.tsx	games array — emoji-иконки (🥞 🐦 🗼 🪈 🎴 🔢 ⚙️ ⭐)
Profile.tsx	stat-карточки, валюты, достижения — emoji
LeaderboardFull.tsx	GAME_TABS — emoji; заголовок — 🏆; колонка очков — 🏆
Shop.tsx	заголовок — 🛍; вкладки — 🛒 📦; chips типов — emoji
Shop.tsx / ShopInventoryTab	инвентарь — карточки с 🎁, обрезанные названия, кнопка Отменить слишком большая
Поиск по коду:

bash
rg "[\x{1F300}-\x{1FAFF}]" frontend/src --type tsx
Это даст полный список файлов и строк с emoji. Cursor должен пройтись по нему и заменить всё, кроме игр (игры — отдельная волна).

Обновлённый промпт Cursor'у — с описанием скриншотов
text
Follow-ups before Wave 2. Three items. I'll describe the
screenshots in text since you can't see images.

================================================================
1. EMOJI SWEEP — finish the job
================================================================

You replaced emoji in nav / shop chrome. But there are still
emoji left in these places (verified visually):

A) Home page game grid
   File: Home.tsx (games array / config)
   8 cards, all with emoji icons:
     Pancaker 🥞, Flappy Bird 🐦, Builder 🗼, Hanoi 🪈,
     Memonia 🎴, 2048 🔢, Орбиты ⚙️, Компаньон ⭐
   → replace each with the SVG icon used elsewhere.

B) Burger dropdown — DONE, no action. Reference only.

C) Profile page
   File: Profile.tsx
   - 9 stat cards: «310 Всего очков» 🏆, per-game cards
     with 🥞 🎴 🐦 🗼 🪈 🔢 ⚙️ ⭐
   - Валюта: 💡 лампочки, 🎫 билетики
   - Достижения: badge «Builder Master» 🏆
   - Рекорды по играм list: per-game emoji
   → replace all with SVG icons.

D) Leaderboard
   File: LeaderboardFull.tsx
   - Header «🏆 Лидерборд»
   - Game tabs (GAME_TABS array): 🥞 🐦 🎴 🗼 🪈 🔢 ⚙️ ⭐
   - Score column: 🏆
   → replace.

E) Shop — Товары
   File: Shop.tsx
   - Header «🛍 Магазин»
   - Balance chip: «🎫 N билетиков»
   - Tabs: «🛒 Товары», «📦 Мой инвентарь (N)»
   - Type chips: 📦 Все, 🎴 Карточки, 🎨 Скины, 🎭 Темы,
     👕 Мерч, 🔑 Брелок, 🖼 Картина, 🪶 Фенечка
   - Price chip in cards: 🎫 N
   → replace.

F) Shop — Мой инвентарь
   File: Shop.tsx (inventory tab)
   - Purchased cards show a generic 🎁 emoji instead of the
     item's real image (the item HAS images[]).
   - Titles are truncated: «Пр...», «Ра...», «Ко...» etc.
   → render the item's real image (like the Товары tab).
   → ensure the title/description have room.

Sweep command (report + fix):
  rg "[\x{1F300}-\x{1FAFF}]" frontend/src --type tsx

Do NOT touch in-game UI (Flappy / Towers / Hanoi / Memory /
Hexagon / 2048 / Orbits / Companion gameplay) — separate
polish wave.

================================================================
2. INVENTORY CANCEL BUTTON — too large
================================================================

Screenshot: /shop → Мой инвентарь.
- Each card has a big «Отменить» button taking most of the
  card width.
- Title and description truncate («Пр…», «Ку…»).

Fix:
  - Compact cancel button: icon + short label, or icon-only
    with tooltip.
  - Give title/description more room.
  - Verify on /shop → Мой инвентарь and on shop detail.

================================================================
3. REFUND WINDOW — design first, no code
================================================================

Right now cancel is always allowed. Need guardrails:

  a) 7-day refund window:
     - purchases.refundable_until = purchased_at + 7 days.
     - FE: cancel button active only before that time.
     - BE: reject with clear error after that.

  b) Fulfilled guard:
     - If a purchase is marked fulfilled (merch shipped), refund
       is not allowed.

Show a plan first (no code):
  - Which table/column tracks fulfilled state?
  - Migration needed?
  - BE + FE changes
  - Error messages (RU)
  - Edge cases: partial refunds? digital vs physical goods?

================================================================
DO NOT TOUCH: games, seed, Author registration.
================================================================
Что я держу в карте
Закрыто:

v1.0.9 на GitHub (45d1b4e)

Wave 1 — 6/6

Открытые follow-ups перед Wave 2:

🟧 Emoji sweep — Home, Profile, LeaderboardFull, Shop, Мой инвентарь (6 мест)

🟧 Cancel button compact (inventory)

🟧 Refund window 7 days + fulfilled guard (design first)

🟧 Игры emoji — отдельная волна

Следующее — Wave 2:

🟧 /shop/items thin DTO

🟧 Реальные товары

🟧 Полиш игр

🟧 Лампочки / уровни

Действия:

Rebuild + recreate gateway + balancer + alertmanager.

Отправь промпт Cursor'у — 3 follow-ups с описанием скринов.

Commit Wave 1 — по фазам (6 коммитов).

Потом — Wave 2.