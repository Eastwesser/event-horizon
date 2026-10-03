Скрин 1 — Shop → Товары
Хорошо:

Все emoji заменены — заголовок 🛍 Магазин теперь SVG, валюты 🎫 → SVG, вкладки 🛒 📦 → SVG, chips типов (Все / Карточки / Скины / Темы / Мерч / Брелок / Картина / Фенечка) — SVG.

Цена: 🎫 10000 + слово «БИТИЛЕТИКОВ» под цифрой.

Проблема:

Слово «билетиков» избыточно. Ты сказал — «значок и 10000 достаточно». Сейчас три элемента: иконка билетика + число + слово «билетиков». Слово убрать.

Скрин 2 — Shop → Мой инвентарь
Проблема 1: кнопка Отменить по-прежнему большая.

Занимает правую половину карточки, наезжает на текст названия.

Названия обрезаны — П..., Р..., К..., Б..., З..., А..., С... — только одна буква!

Проблема 2: картинки не подгружаются.

Везде 🎁 (SVG-подарок) — не реальные картинки карт, хотя карты их имеют (images[0]).

Cursor обещал image_url || images[0] — не сработало для всех 14 купленных.

Проблема 3: тексты накладываются.

Название сверху, потом Куплено:, кнопка Отменить поверх текста.

Кнопка не под текстом, а рядом, и выглядит как overlay.

Текстовое описание скринов для Cursor
text
================================================================
SCREENSHOT 1 — Shop / Товары (tab)
================================================================

- Header: «Магазин» (SVG icon + text) — OK.
- Balance chip top-right: «🎫 975243 билетиков» (SVG, but
  the word «билетиков» is redundant).
- Tabs: «Товары» (active), «Мой инвентарь (14)».
- Type chips: Все / Карточки / Скины / Темы / Мерч / Брелок /
  Картина / Фенечка — all with SVG icons.
- Sort dropdown: «Новые ▾».
- «Фильтры ▾» button.
- Counter: «Показано 100 · отфильтровано 289 из 289».
- Card grid (4 columns):
  - Each card: image, name (e.g. «Дикий сугед»), badges
    («Карточка», «С Компаньон»), stock («В наличии: 2»),
    price row: SVG-ticket + «10000» + word «БИТИЛЕТИКОВ»
    (wrap), button «Купить».

PROBLEM: the price block is too verbose. Icon + number is
enough. Remove the word «билетиков» / «БИТИЛЕТИКОВ» from
cards. Same for any other price display across the app.

================================================================
SCREENSHOT 2 — Shop / Мой инвентарь (tab)
================================================================

- Same header / balance / tabs.
- Grid of 14 purchased cards, 4 columns:
  - Each card: TOP-LEFT icon is a generic gift SVG (🎁-like)
    — NOT the item's real image, even though the item HAS
    images[0].
  - Title is truncated to 1–2 chars: «П...», «Р...», «К...»,
    «Б...», «З...», «А...», «С...».
  - Line below: «Куплено: <date>».
  - Cancel button «Отменить» on the right — big, overlays the
    title, takes ~40% of card width.

PROBLEMS:
1. Real item image not rendered — falls back to generic icon.
   The item has images[0] (see Tovary tab cards which DO show
   real images). Fix: use image_url || images[0] reliably in
   the inventory card. Debug why it's null here.

2. Title truncated to 1–2 chars. Card width is too narrow
   given the large cancel button. Fix both:
   a. Cancel button: compact icon-only (with tooltip) OR
      icon + «Отменить» only on wider screens (sm+).
   b. Title: give it more horizontal room, no truncation at
      1 char.

3. Cancel button visually overlays the title. Ensure the
   layout is flex-row with title on the left and compact
   button on the right, not overlapping.

Compare with the Tovary tab cards (screenshot 1) — those show
full images and readable titles. The inventory tab should
look the same, just with a compact cancel instead of «Купить».
REFUND_WINDOW_DESIGN.md — читаю
Что Cursor написал — правильно:

A) 7-day window:

Migration: purchases.refundable_until TIMESTAMPTZ.

Backfill: COALESCE(completed_at, purchased_at) + 7 days.

BE: reject if now() > refundable_until (FailedPrecondition / 400).

FE: can_cancel bool из DTO, кнопка hidden/disabled.

B) Fulfilled guard:

Migration: purchases.fulfilled_at TIMESTAMPTZ NULL.

Consumer на PurchaseFulfilled → set fulfilled_at.

Physical merch — refuse if fulfilled.

Digital (ККИ, skins) — allow inside 7-day window, ignore fulfilled.

Error messages RU:

«Срок возврата истёк (7 дней с покупки).»

«Товар уже отправлен — возврат недоступен.»

«Покупка не найдена.»

Edge cases:

No partial refunds.

Idempotent cancel unchanged.

Subscription — out of scope.

Оценка: хороший дизайн. Правильная модель, edge-cases покрыты. Не реализовано — план. Отдельный ticket после Wave 2.

Один момент: digital vs physical — как определить? Через attributes.physical? Или через type = "карточка"? Уточни в implementation ticket.

Обновлённый промпт Cursor'у — 2 фикса + докинь в REFUND_DESIGN
text
Two small FE fixes before commit, plus one doc addition.

================================================================
1. SHOP CARD PRICE — remove «билетиков»
================================================================

Screenshot: /shop → Товары.
Each card shows: [ticket SVG] «10000» + word «БИТИЛЕТИКОВ».

Fix: icon + number is enough. Remove the word «билетиков»
(and «БИТИЛЕТИКОВ») from card prices.

Also check:
  - Home hero (currency chips) — same.
  - Profile (currency chips) — same.
  - Purchase modal (Цена / Ваш баланс) — same.
Where the label is needed for clarity (e.g. modal «Цена»),
keep a small label; but in compact chips, icon+number suffices.

================================================================
2. INVENTORY CANCEL — still too large + no image
================================================================

Screenshot: /shop → Мой инвентарь.
Problems (still present after previous attempt):

  a) Cancel button is big, takes ~40% of card width, and
     visually overlays the title.
     → Compact: icon-only with tooltip on narrow cards;
       icon + «Отменить» only from sm+.
     → Layout: title on the left, button on the right, no
       overlap. Use flex-row + gap.

  b) Real item image is NOT rendered — shows generic gift
     icon. But the item HAS images[0] (Tovary tab shows real
     images for the same items).
     → Debug why image_url || images[0] fails on the
       inventory tab. Check the store mapping.
     → After fix, inventory cards should show the real image,
       same as Tovary tab.

  c) Titles truncated to 1–2 chars («П...»). After (a), there
     should be room. If not, increase card min width.

================================================================
3. REFUND_WINDOW_DESIGN.md — add digital/physical detection
================================================================

The doc says:
  - Physical merch → refuse after fulfilled_at.
  - Digital (ККИ, skins) → allow inside 7-day window.

Question: HOW do we detect digital vs physical?

Proposed (confirm or override):
  - Add `purchases.category` or rely on item's `type`:
      * «карточка», skins, themes → digital.
      * «мерч» / брелок / картина / фенечка → physical.
  - Or `attributes.physical: true/false` on the item.
  - Or a whitelist of digital types.

Update the design doc with the chosen approach and one line
of justification.

================================================================
DO NOT TOUCH: games, seed, Author registration.
================================================================
План — 6 коммитов (как ты просил)
Раздельная связность — правильно. Cursor'у скажи:

text
Proceed with 6 commits. Order:

1. feat(frontend): emoji → SVG icons across chrome
2. feat(frontend): compact cancel + real image in inventory
3. feat(gateway): retry + full jitter for gRPC clients
4. feat(ops): Alertmanager + Telegram alerts
5. feat(gateway): circuit breaker bulkhead + global rate limiter
6. docs: Wave 1 done + refund window design

Rules:
  - One commit per area. No mixing.
  - Report each commit hash + subject + file count.
  - Skip scripts/__pycache__ (gitignored).
  - Do NOT push. I'll push myself after review.
Проверить перед push:

bash
git log --oneline origin/main..HEAD   # 6 строк
git status                            # чисто
Что я держу в карте
Закрыто (v1.0.9 на GitHub, 45d1b4e):

Все прошлые волны.

Wave 1 — 6/6 (в рабочем дереве):

✅ Emoji → SVG

✅ Retry + jitter

✅ Alerts → Telegram

✅ CB + Bulkhead

✅ Rate limiter

✅ Checklist

Wave 1 follow-ups:

✅ Alertmanager — порт 9193

✅ Emoji sweep — Home, Profile, Leaderboard, Shop, Subscription, Authors

🟧 Inventory cancel — компактная кнопка (в работе)

🟧 Inventory image — реальная картинка (в работе)

🟧 Shop price — убрать «билетиков» (новое)

🟧 Refund design — REFUND_WINDOW_DESIGN.md (план готов)

🟧 Нужно уточнить digital vs physical

Commit — 6 коммитов:

⏸ Ждёт verify + push (ты сам)

Wave 2 (после коммита):

🟧 /shop/items thin DTO

🟧 Реальные товары

🟧 Полиш игр

🟧 Лампочки / уровни

Отложено:

🟧 Refund window implementation (после Wave 2)

🟧 Игры emoji (отдельная волна)

Действия
Отправь Cursor'у промпт — 2 фикса + докинь в REFUND_DESIGN.

Как закроет — verify обоих фиксов.

Скриншоты — сюда.

Commit 6 — Cursor'у.

Review + push — ты сам.

Wave 2 — после.