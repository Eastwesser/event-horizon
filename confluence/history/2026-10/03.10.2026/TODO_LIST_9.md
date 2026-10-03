Что в плане правильно
URL params — все ключи правильные:

element, rarity, cost_tier, set, year — multi через запятую.

Flags — foil=1 и т.д.

Class — multi, но из каталога (не хардкод).

Artist — single + autocomplete 1-в-1.

Stats — cost_op/cost, hp_op/hp, move_op/move, операторы eq|gt|lt|gte|lte.

Icons — multi, card has ANY selected (не все).

Page resets при смене фильтра.

UI — collapsible two-column panel. Совпадает с proberserk-референсом. Логика client-side на загруженный каталог — ок, потому что каталог уже грузится один раз.

filterCatalogItems в catalogQuery.ts — правильно, туда же.

Active filter count в toggle label — хороший UX, сразу видно, что фильтры активны.

Один момент — уточни Cursor'у
Из плана не ясно:

1. Reset — что сбрасывает?

Все filter params + sort + page?

Или только filters, sort остаётся?

Мой выбор: только filters. Sort — отдельный контрол.

2. attack_dice — как?

Plan: optional dice string.

Как именно? Поиск точного совпадения 2-2-3? Или содержит? Или список?

На практике — кто-то будет искать 2-2-3? Скорее редко. Можно отложить или сделать text field.

3. flavor в правой колонке (из старого промпта) — здесь нет. Видимо, ты убрал. Ок.

4. Class — откуда? Из уникальных attributes.class[] в каталоге. Правильно. Но заглавные/строчные — унифицировать? Гном vs гном?

5. Когда фильтров много (например, 5 активны) — что показывать? Счётчик Фильтры (5)? Или раскрывать автоматически, если какие-то активны?

Финальный ответ Cursor'у
text
OK — implement the filters wave.

Three clarifications:
  1. «Сбросить фильтры» clears only filter params (element,
     rarity, cost_tier, set, year, flags, class, artist,
     stats, icons). Sort stays as-is. Page resets to 1.
  2. attack_dice: make it a simple text field matching the
     dice string exactly (e.g. "2-2-3"). If nothing entered,
     ignore.
  3. class values: normalize case-insensitively in the
     filter (Гном === гном). Show title-case in the UI.
  4. Panel toggle: show «Фильтры (N)» where N = active filter
     count. If N > 0, auto-expand panel on mount? Your call —
     report the choice.

Everything else as written.

Order: implement filters panel + filterCatalogItems extension,
verify, then move on.

DO NOT TOUCH: seed, purchase, games, nav, admin.
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

1→4 final polish ✅ (Купить перенесён)

Attribute filters — plan approved ⏳

В работе (после approve):

🟧 Filters panel (two-column, proberserk-style)

🟧 filterCatalogItems extension

🟧 URL params для всех фильтров

🟧 Сбросить фильтры

🟧 Active filter count

Следующая волна:

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии в UI

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр