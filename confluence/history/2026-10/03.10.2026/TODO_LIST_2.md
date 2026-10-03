Visual fix: card image corners show white, not rounded.

Symptom: on card detail (and grid), the card image is a
rectangular JPG inside a rounded container. The four corners
of the image show white — because the JPG has no transparency
and does not fill the rounded corners.

Fix:
  - Container: rounded corners + overflow: hidden +
    background = --color-nebula (or another dark token).
  - Image inside: object-fit: contain (so the full card fits)
    with the dark background showing in the letterbox areas.
  - Result: no white corners anywhere; corners match the app
    background; the full card stays visible.

Apply to:
  - ShopItemCard image
  - ShopItemDetail hero image
  - InventoryItemCard image
  - InventoryItemDetail hero image
  - Any other place we render card JPGs.

Do NOT change the image asset. Do NOT change aspect ratio of
the image itself. Only the container styling + object-fit.

Report files changed.
Про audit — опционально, но разумно
Cursor сказал:

Slug↔name mismatches — almost all transliteration. Пращник-style content bug: none left after the fix.

Формально — audit закрыт.

Но — Пращник показал, что визуальный mismatch возможен и не ловится скриптом. Если хочешь — можешь сам пролистать 280 карт в /shop (быстро, ~3–5 минут), искать глазами:

Имя карты ≠ арт на картинке.

Но — не блокер. Пращник — единичный случай. Остальные — вероятно ок.

Мой совет: отложи. Если в будущем заметишь ещё — фикс по одной. Не гоняйся за всеми сразу.

Да, двигаемся дальше — 8 → 9 → 10
Промпт для Cursor:

text
Now: 8 pagination UI → 9 sort → 10 artist pages.

8. PAGINATION UI
   - Currently /shop loads all 289 items with hidden pagination
     (searchAllItems under the hood).
   - Add a visible pagination control:
       * 100 items per page (configurable constant)
       * URL: /shop?page=1 (default), preserve other filters
       * Buttons: «Назад» / «Далее» + «Стр. N из M»
       * Reuse the /admin Users pagination pattern
   - /inventory: same treatment (289 items there too).

9. SORTING
   - Add a sort control above the grid (small dropdown next to
     filter chips):
       * Новые (created_at desc) — default
       * Имя A–Z / Z–A
       * Редкость (common → ultra / ultra → common)
       * Цена (asc / desc)
       * Художник A–Z
       * Выпуск (#, asc / desc)
   - URL param: ?sort=...&order=... (shareable).
   - Persist across pages.

10. ARTIST PAGES
   - /authors — list of all artists:
       * name, count of cards
       * sorted alphabetically
   - /authors/:artist_id — grid of that artist's cards:
       * reuse ShopItemCard
       * header: artist display name + "N карт"
   - Link from card detail: artist name → /authors/:artist_id

Process: implement 8, verify, then 9, verify, then 10.

DO NOT TOUCH: seed, purchase flow, game mechanics, nav,
admin panel, attribute filters (next wave).
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix + idempotency + purchase работают

4 Icons + grouped attrs ✅

5 Badges ✅

6 Tickets ✅

7 Create modal ×1000 ✅

Пращник fixed ✅

Текущее:

🔴 Visual fix: белые углы картинок (border-radius + object-fit)

🟧 8 Pagination UI (100/страница)

🟧 9 Sorting

🟧 10 Artist pages

Следующая волна:

🟧 Attribute filters

🟧 Расширенный поиск (референс proberserk)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Опционально / отложено:

🟧 «Страж сокровищ» — проверить арт

🟧 Visual audit 280 карт (глазами)