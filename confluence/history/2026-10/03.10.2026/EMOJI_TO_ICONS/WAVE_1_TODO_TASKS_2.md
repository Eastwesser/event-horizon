Скрин 1 — Shop → Мой инвентарь
Хорошо:

✅ Компактная кнопка ↺ Отменить — уже не сосиска.

✅ Вертикальные карточки — как в Товарах.

✅ Названия не обрезаны — Пращник, Радужные блоки, Космические блины, Радужные трубы — читаются.

✅ Цена убрана (в инвентаре она не нужна).

✅ Дата Куплено: 03.10.2026 — видна.

Проблема — картинки:

Пращник → реальная картинка карты ✅

Радужные блоки, Космические блины, Радужные трубы → 🎁 placeholder ❌

Почему так:

Карта ККИ — есть images[0] в item → рендерится.

Скины (Радужные блоки, Космические блины, Радужные трубы) — НЕ карты, у них нет images[0].

Это скины для игр, у них свой формат — иконка по умолчанию 🎁.

Что делать:

Вариант A: скинам тоже дать картинку — например, превью того, как выглядит скин в игре.

Вариант B: оставить 🎁 для скинов — это осознанный fallback.

Вариант C: отдельная иконка для скинов — например, палитра 🎨.

Мой выбор — A или C. 🎁 для скинов — визуально странно. Игрок не понимает, что купил.

Скрин 2 — Shop → Товары
Хорошо:

✅ Все emoji заменены на SVG.

✅ Цена 🎫 10000 — без слова «билетиков» ✅

✅ Карты ККИ — с реальными картинками.

✅ Badges Карточка, С комманда — ок.

Чисто. Ничего не править.

Скрин 3 — Alertmanager UI
Хорошо:

✅ UI открывается на http://localhost:9193.

✅ Cluster Status: ready.

✅ Config: telegram_api_url: https://api.telegram.org — есть.

✅ receiver: "null" — по умолчанию, нормально (нет настроенного TELEGRAM_*).

Проверить:

TELEGRAM_BOT_TOKEN + TELEGRAM_CHAT_ID в .env.

Если есть — форсить alert → сообщение в Telegram.

Если нет — noop, но UI работает.

Что осталось — 1 фикс + push
1. Fix — иконки скинов в инвентаре
Промпт Cursor'у:

text
Screenshot: /shop → Мой инвентарь.

Cards with real images (Пращник, ККИ) — OK, image renders.
Cards with skins (Радужные блоки, Космические блины, Радужные
трубы) — show a generic 🎁 placeholder.

Cause: skins/items without images[0] fall back to a gift icon,
but this looks confusing. Player can't tell what they bought.

Fix (pick one, report which):
  a) Give skins a preview image (how the skin looks in-game).
     Requires content work — maybe later.
  b) Use a distinct icon for skins (palette / paint brush).
  c) Use a distinct icon per item type:
       карточка → card icon
       скин → palette
       тема → theme icon
       мерч → merch icon
       брелок → keychain
       картина → picture
       фенечка → bracelet

Recommend (c) — icon reflects item type. Falls back nicely.

Apply in InventoryItemCard (and ShopItemCard for shop items
without images).

================================================================
DO NOT TOUCH: games, seed, Author registration.
================================================================
2. Commit + push
Ты уже закоммитил 6 коммитов. Сейчас только один follow-up (скины иконки) — отдельный 7-й коммит:

text
feat(frontend): per-type fallback icons for inventory items

Items without images[0] (skins, themes, merch) no longer
show a generic gift — use a type-specific icon.
Потом:

bash
git push origin main
Проверка:

bash
git log --oneline -3
git status
Что я держу в карте
Закрыто (v1.0.9 + Wave 1, 6 коммитов локально):

✅ Emoji → SVG chrome

✅ Retry + jitter

✅ Alertmanager + Telegram (порт 9193)

✅ CB + Bulkhead

✅ Rate limiter

✅ Compact cancel + real image (для ККИ)

✅ Prices — icon + number

✅ Refund design

Проверено на живом:

✅ Emoji везде — SVG

✅ Alertmanager — работает

✅ Кнопка Отменить — компактная

✅ ККИ карточки — реальные картинки

✅ Цены — без «билетиков»

🟧 Скины — 🎁 placeholder (нужен фикс)

Открыто:

🟧 Fix — иконки скинов (по типу)

🟧 Commit #7 + push (ты сам)

Wave 2 (после push):

🟧 /shop/items thin DTO

🟧 Реальные товары (placeholder → Берсерк)

🟧 Полиш игр

🟧 Лампочки / уровни

Отложено:

🟧 Refund window implementation

🟧 Игры emoji