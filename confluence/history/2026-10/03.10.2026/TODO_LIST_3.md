Что вижу я (плюс к твоему)
🔴 Скрин 2 — «Живая баллиста» — реальные проблемы
Картинка не вписана в контейнер. Карта в полный рост, но края контейнера видны прямоугольником — карта «плавает» внутри, letterbox сверху/снизу, но сам контейнер не закруглён. Углы карты не совпадают с углами контейнера.

Верхний отступ слишком маленький — карта почти липнет к верху.

Слишком много пустого места справа — карта узкая, object-contain сохраняет пропорции, но контейнер шире, чем нужно.

Художник: Григорий Давыдов — сломанная ссылка? Смотри — текст не кликабельный? А должен вести в /authors/:id. Проверь — работает ли.

🟧 Скрин 1 — сетка /shop — общая проблема
Карты в сетке разной высоты — Живая баллиста и Пардус имеют разные высоты описаний. Нижний край не выровнен. В наличии: N у карт на разной высоте.

Отступы между картинкой и названием — тесновато.

Кнопки Купить — не на одной линии по вертикали.

🟧 Скрин 3 — dropdown сортировки
Sort dropdown — выглядит ок, но не хватает опций:

Нет стихий (ты сам попросил).

Нет «только в наличии» — фильтр.

Нет поиска по имени — отдельное поле.

Все / Карточки / Скины / ... — много фильтров в одну строку, тесновато.

🟧 Скрин 4 — модалка покупки
Картинка карты слишком маленькая. Пастух Ульфгарда — превью мелкое, не вписывается в модалку.

Углы картинки — не закруглены (в модалке видишь квадрат).

Много пустого места вокруг.

🟧 Скрин 5 — /authors
Хорошо сделано. Но:

Один художник — один ряд? Или collapse? Сейчас все 108 раскрыты — длинный скролл.

Нет поиска по художнику — если хочешь найти конкретного, листаешь.

А. Кретов. А. Масалимова — два автора в одной строке. Это правильно (оба автора карты), но при клике → только один artist_id? Или оба?

Счёт 3 карт — не "3 карты". Грамматика.

108 — если суммарно 280, но 108 уникальных, то среднее ~2.6 карты. Ок, но стоит проверить — нет ли «художника» с 30+ картами (что странно).

🟧 Скрин 5 — «Авторы» (коммьюнити)
Вижу: вкладка Художники карт (108) — это отдельная секция над Авторы сообщества Event Horizon. Правильно. Но:

Заголовок Авторы — большой, а секция Художники карт — внутри. Иерархия может быть лучше.

Общее — визуальные паттерны
Проблема одна: картинки карт не вписываются идеально в контейнеры.

В сетке — разной высоты, тесновато.

В detail — letterbox, отступы разные.

В модалке — мелко, без скругления.

Решение: единый <CardImage> компонент с фиксированными пропорциями:

Aspect ratio — карта ККИ обычно 63×88 мм (≈ 0.716). Можно зафиксировать aspect-ratio: 5 / 7 в контейнере.

object-fit: cover — заполняет весь контейнер, обрезает по краям.

border-radius + overflow: hidden — углы закруглены.

background: var(--color-nebula) — тёмный фон за краями.

Что меняется: карта заполняет контейнер полностью, углы закруглены, никаких белых/чёрных просветов, одинаковый размер во всех местах.

Минус: cover обрезает край карты. Но для игровой карты это ок — самое важное в центре.

Альтернатива: contain с фиксированными пропорциями контейнера = карта вписывается целиком, но letterbox по краям. Ты уже видел — тебе не нравится.

Мой выбор: cover + фиксированный aspect-ratio. Карты будут заполнять превью. Обрезка — минимальная (обычно 2-5% по краям).

Промпт для Cursor — обновлён
text
Wave: fixes + sorting expansion + UX polish.

================================================================
A. CARD IMAGE — unified, fills container, rounded
================================================================

Problem: card images do not fill their containers.
  - Detail: letterbox, uneven top spacing, too much side space.
  - Grid: cards vary in height, tight spacing.
  - Purchase modal: small preview, square corners.

Solution: one shared <CardImage> component used everywhere.
  - Container: fixed aspect-ratio: 5 / 7 (KKI card ≈ 63×88mm).
  - Image inside: object-fit: cover.
  - Container: border-radius (md), overflow: hidden,
    background: var(--color-nebula).
  - Optional prop: `fit="contain"` for cases where the full
    card must be visible (e.g. detail hero) — but default is
    `cover`.
  - Result: cards fill the container; corners are rounded;
    no white or awkward letterbox anywhere.

Apply to: ShopItemCard, ShopItemDetail hero, InventoryItemCard,
InventoryItemDetail hero, PurchaseModal preview.

Do NOT change the JPG asset.

================================================================
B. GRID ALIGNMENT
================================================================

In /shop and /inventory grids:
  - Cards should be equal height within a row (already via
    items-stretch, but verify with longer descriptions).
  - "В наличии: N" line should align across cards in the same
    row (add mt-auto before it).
  - "Купить" button should align at the bottom across cards.
  - More vertical spacing between the image and the name.

================================================================
C. DETAIL PAGE — reduce raw card text dump
================================================================

Question from me: does the user need the full card text on the
detail page?

Current: card text (2 lines) + flavor text + attributes table.

Proposal:
  - Keep: hero image, name, badges, price, stock.
  - Keep: flavor text (small, italic).
  - Collapse the full card text under a "Показать текст карты"
    toggle (default: hidden).
  - Attributes table: keep sections (Бой / Карта / Автор), but
    make them denser visually. Show only what matters to a
    buyer; hide internal fields (artist_id, market_rub,
    idempotency_key).
  - Header line under name: "Сет · Год · Художник" — one line.

Also fix: the artist name in detail should link to
/authors/:artist_id (currently may not be clickable).

================================================================
D. SORTING — expand
================================================================

Add to the sort dropdown:
  - Element A–Z (стихия: болота, горы, леса, ...)
  - Set number asc / desc (уже есть)
  - "Только в наличии" toggle (filter, not sort)
  - "Только фойл" / "Только нуар" / "Только летающие"
    (checkbox filters, not sort)

Add a SEARCH FIELD above the grid (separate from existing
"Поиск по названию..."):
  - Immediate name search (as user types, debounced).
  - URL param: ?q=...

================================================================
E. MODAL PREVIEW SIZE
================================================================

PurchaseModal: preview image is too small. Make it:
  - Same CardImage component.
  - Wider preview (e.g. max-w-[240px], aspect-ratio 5/7).
  - Rounded corners.

================================================================
F. AUTHORS PAGE — polish
================================================================

/authors:
  - Add a search field (name filter).
  - Pluralize count: "1 карта / 2 карты / 5 карт".
  - Multi-author rows ("А. Кретов. А. Масалимова") — click
    should link to a combined view, or at least pick the
    first author. Decide and report.
  - Optionally: collapse to top-N + "показать всех".
  - Consider sort: by count desc (popular artists first) as
    an option.

/authors/:id:
  - "N карт" header.
  - Sorted by default newest-first.

================================================================
G. SWIPE / ARROW NAVIGATION ON DETAIL PAGE
================================================================

On /shop/item/:id and /inventory/:id:
  - Add left/right arrows to move between adjacent items in
    the current list (uses the same filter+sort+page context
    the user came from).
  - Optional: swipe gesture on touch devices (basic enough —
    just touch start/end delta).
  - Keep browser back working.

================================================================
H. PAGINATION — performance
================================================================

"100 items per page" feels slow. Check:
  - Are images lazy-loaded? Add loading="lazy" + decoding="async".
  - Are we re-fetching on every page change, or caching pages?
  - Is the backend slow, or is it frontend rendering?
Report the bottleneck.

================================================================
PROCESS
================================================================

Order: A + B first (visual foundation). Then C. Then D + E.
Then F. Then G. Then H.

Show the plan first, no code. Then implement step by step.

DO NOT TOUCH: seed, purchase flow, game mechanics, nav,
admin panel.
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix + idempotency + purchase

4 Icons + grouped attrs ✅

5 Badges ✅

6 Tickets ✅

7 Create modal ×1000 ✅

Пращник fixed ✅

8 Pagination ✅ (работает, но медленно)

9 Sorting ✅ (базовая, надо расширить)

10 Artist pages ✅ (работают, нужен polish)

Текущая волна (visual + UX polish):

🔴 Card image — единый <CardImage> (cover + aspect-ratio 5/7 + rounded)

🟧 Grid alignment — equal heights, bottom-aligned

🟧 Detail page — collapse card text, denser attributes

🟧 Sorting — element / только в наличии / фойл / поиск

🟧 Modal preview — крупнее

🟧 Authors page — поиск, плюрализация, multi-author

🟧 Swipe / arrows — навигация между картами

🟧 Pagination perf — lazy load / кэш

Следующая волна:

🟧 Attribute filters (element / rarity / artist 1-в-1 / foil / stats / icons)

🟧 Расширенный поиск (proberserk-style)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр