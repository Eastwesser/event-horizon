ОК. Шип 1 → 2 → 3
План точный, ничего лишнего.

Что подтверждаю
1. Dead icons — drop counter / uchr / tap / instant. ✅
Оставить: armor, zoal, zov, zoz, zot, zor, zoo, zom, regen, stamina, direct, ova, ovz, ovs. 15 иконок.

2. Collapse long groups. ✅

Class — first 8 + Показать все (N).

Icons — то же, если > 8 (после удаления 4 dead — станет ~15, всё ещё много → collapse).

Остальные (element / rarity / type / flags / set / year / cost_tier) — always open. Правильно, они маленькие.

3. Curl sanity. ✅
Проверить:

attributes.element / rarity / class / hp / icons

icons — array, не string

price / stock — числа

card_no / set_number / artist_id — есть

хотя бы одна foil + одна noir

artifact с hp, без attack_dice (напр. Склеп Керсам)

companion (напр. Беллигемин)

Один дополнительный пункт к 3 — заодно проверить
Раз уж curl — проверь заодно counter / uchr / tap / instant:

text
Also, for step 3, explicitly confirm that counter / uchr /
tap / instant do NOT exist in attributes.icons for any card
(that's why we're removing them from the filter). Query:
  curl .../items?... | jq '[.items[].attributes.icons[]?.type] | unique'

The output should NOT contain counter / uchr / tap / instant.
If any of those DO appear, keep that icon in the filter list
and report it.
Зачем: если Cursor ошибается и эти иконки реально где-то есть — фильтр должен их сохранить. Curl покажет правду.

Финальный ответ Cursor'у
text
OK — ship 1 → 2 → 3.

Addition on step 3:
  Query the distinct values of attributes.icons[].type across
  the whole catalog (or a big page):
    ... | jq '[.items[].attributes.icons[]?.type] | unique'
  Confirm counter / uchr / tap / instant do NOT appear.
  If any of them DO appear, keep that icon in the filter and
  report it back. Otherwise drop as planned.
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

1→4 final polish (Купить в detail)

Attribute filters ✅

Следующий полиш (в работе):

🟧 Drop dead icons (counter / uchr / tap / instant)

🟧 Collapse Class + Icons (8 + «Показать все»)

🟧 Curl sanity-check бэкенда + distinct icons

Следующая волна (после этого):

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии в UI

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Оценка
Ты быстро движешься. Attribute filters — большая волна закрыта. Осталось 3 полишных фикса — и следующие волны.

Мой приоритет на потом:

Bottleneck fix — архитектурно важно, каталог не масштабируется на 1000+.

Noiz-комментарии — у тебя данные уже готовы (сопоставлены с картами). Быстрый win.

Refund / cancel purchase — обкатать на Сугед. Быстро.

Author registration — крупная фича (авторы сами добавляют). Отдельная волна.