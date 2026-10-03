Один визуальный фикс — перед attribute filters
Твой пункт:

Кнопку Купить я бы добавил вправо и вниз на уровне Художник: Степан Гилев. Симметрично: от левого края до Художник — такое же расстояние от правого края до Купить.

Точно. Сейчас Купить — слева, выше блока. Ты хочешь — справа, на одной линии с Художник: ... внизу.

Что делать Cursor'у: вынести кнопку Купить в правый нижний угол страницы детали. По вертикали — на уровне последней строки атрибутов (Художник). По горизонтали — симметрично левому краю.

Промпт для Cursor — 2 задачи
text
Two things:

================================================================
FINAL POLISH — inline Купить position in detail
================================================================

Screen: /shop/item/:id (Карбагаш).

Current: Купить button sits left, above the attributes block.
New: move Купить to the bottom-right of the detail card, on the
same vertical line as the last attribute row («Художник: ...»).
Left-right symmetry: the same gap on the right side as the
left side has before the attribute labels.

Behavior unchanged: click → modal, owned → «В инвентаре», etc.

================================================================
NEXT WAVE — attribute filters
================================================================

Plan only, no code yet.

Filters (URL params, shareable):
  - element (multi: darkness / woods / mountains / steppes /
    swamps / neutral)
  - rarity (multi: common / uncommon / rare / ultra)
  - cost_tier (row / elite)
  - set_number (multi: 4 / 5 / 6 / 7 / 8)
  - year (multi)
  - foil / noir / flying / companion / symbiont / parasite /
    unique (checkbox toggles)
  - class (multi: Гном / Эльф / Демон / ...)
  - artist (single, 1-to-1 autocomplete over existing
    artists, exact match)
  - stats: cost / hp / move with operator (=, >, <, >=, <=)
  - attack_dice (optional: filter cards with specific dice)
  - icons (multi-select: armor / zoal / zov / zoz / zot /
    zor / zoo / zom / regen / stamina / direct / uchr /
    ova / ovz / ovs)

Layout: proberserk-style two-column panel
  left:  text / element / class / rarity / type / unique /
         flight
  right: set / cost_tier / stats / icons / artist / flavor /
         version

Reuse: FilterChip, Card, Button, existing sort/search.

URL: all filter params appended, page resets on filter change.
Reset: «Сбросить фильтры» button.

Show the plan first, no code.
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

A–H visual + UX polish

1→4 final polish ✅ (verified, кроме позиции кнопки)

В работе:

🟧 Detail — перенести Купить в правый низ (симметрично)

Следующая волна (после фикса):

🟧 Attribute filters (element / rarity / artist / foil / stats / icons) — proberserk-style

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Оценка
Модалка — attrs summary отлично: Редкость / Стихия / Фойл / HP / Удар / Стоимость / Класс. Ничего лишнего.

Detail — Показать текст карты collapsed, Сет · Год · Художник в строке, attributes компактно. Хорошо.

Финальный штрих — перенос Купить в правый низ — и detail закончен.