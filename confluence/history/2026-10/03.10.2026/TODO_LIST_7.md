Что подтверждаю
1. Detail hero white corners.
Ты прав: контейнер уже rounded-md / bg-nebula / overflow-hidden, но contain всё равно может flash white. Причина:

Либо aspect-ratio контейнера конфликтует с картинкой.

Либо img рисуется поверх углов, не клипается.

Либо фон не полностью заполняет.

Правильно: убрать aspect-ratio на detail hero (там contain — карта не 5/7, потому что пропорции разные), форсить bg-nebula за картой, обязательно overflow-hidden на внешней обёртке. Тогда углы = фон, не белые.

2. Shop card — drop description. Точно. Убрать обрезанный текст — чище, короче, лучше выравнивание.

3. Purchase modal — attrs. Правильно. Редкость / стихия / фойл / HP / удар / стоимость / класс. Заполненные только.

4. Detail — inline Купить. Правильно. В инвентаре / Нет в наличии / клик → модалка → покупка → локальный state.

Один момент — про 4
Уточни Cursor'у:

text
On item 4 — after purchase, when the user stays on the detail
page:
  - decrement stock locally (item.stock -= 1)
  - if stock reaches 0 → button shows "Нет в наличии"
  - also update the /shop grid cache so returning to the grid
    shows the new stock without a refetch (or trigger a refetch)
  - if user opens the same item later, purchase doesn't work
    (owned → "В инвентаре")

Also: when the user already owns the item, show "В инвентаре"
AND (optionally) a "Перейти в инвентарь" link — so they don't
have to navigate manually. Up to you if it's worth adding.
Опционально — линк в инвентарь. Не обязательно.

Ответ Cursor'у — approve
text
OK. Ship 1 → 4.

One addition on 4 (detail inline buy):
  - After purchase, decrement stock locally.
  - If stock reaches 0 → "Нет в наличии".
  - If user already owns → "В инвентаре" (disabled) + optional
    link "Перейти в инвентарь".
  - Update the /shop grid cache (so returning to grid shows
    fresh stock) OR trigger a refetch on return.

Everything else as written.
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix + idempotency + purchase

4 Icons + grouped attrs

5 Badges, 6 Tickets, 7 Create modal ×1000

Пращник fixed

8 Pagination, 9 Sorting, 10 Artist pages

A–H visual + UX polish (verify ok, 4 фикса в работе)

Текущий полиш (plan approved):

🟧 1. Detail hero — white corners fix

🟧 2. Shop card — drop description

🟧 3. Purchase modal — attrs summary

🟧 4. Detail — inline Купить

Следующая волна:

🟧 Attribute filters (element / rarity / artist / foil / stats / icons) — proberserk-style

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр