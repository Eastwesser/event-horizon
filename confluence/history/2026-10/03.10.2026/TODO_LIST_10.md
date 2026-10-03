Что вижу
Хорошо:

Фильтры (N) — collapsible, две колонки.

Стихия / Класс / Редкость / Тип / Флаги — слева.

Выпуск / Год / Стоимость / Характеристики / Удар / Иконки / Художник — справа.

Мульти-select работает, Сбросить (наверное) внизу.

Проблемы:

Фишка, УЧР, Поворот, Мгновенно — не фильтруют. Ты сам сказал — удалить.

Класс — разросся в 20+ чипов. Выглядит перегружено.

Иконки — 14 чипов. Много, но по делу — это фильтр по иконкам карт.

Стихия, Класс, Редкость, Тип, Флаги — каждая группа с заголовком. Ок.

Выпуск, Год, Стоимость — тоже с заголовками.

Что тебя беспокоит:

Слишком много всего, давай добавим свертывание где класс хотя бы.

Правильно. Класс — самая большая группа (20+). Если её свернуть в dropdown или collapse — панель станет компактнее.

Промпт для Cursor — 3 фикса
text
Three polish items on the attribute filters panel, then the
bottleneck wave.

================================================================
1. REMOVE non-filtering icons
================================================================

«Фишка» (counter), «УЧР» (uchr), «Поворот» (tap), «Мгновенно»
(instant) do not filter anything (no cards carry them as
attributes.icons). Remove them from the icons chip list.

Keep only icons that actually exist in the catalog:
  armor, zoal, zov, zoz, zot, zor, zoo, zom, regen, stamina,
  direct, ova, ovz, ovs

================================================================
2. COLLAPSE large groups (Class especially)
================================================================

The «Класс» group has 20+ chips. Panel is too tall.

Solution:
  - Default state: show top 6–8 chips + «Показать все (N)».
  - Click → expand full list.
  - Same treatment for «Иконки» (15 chips) if it feels long
    after the removal of the dead ones.
  - Keep other groups (element, rarity, type, flags, set, year,
    cost_tier) as-is — they're small.

Apply only to groups with more than ~8 options.

================================================================
3. Sanity-check backend data (curl smoke)
================================================================

Before the bottleneck wave, run a quick sanity pass on the
inventory/shop APIs to confirm the data the frontend relies on:

  curl -s http://localhost:8079/api/inventory/items?type=карточка&limit=100 | jq
  curl -s http://localhost:8079/api/shop/items | jq '.[0]'

Check:
  - items have attributes.element / rarity / class / hp / icons
  - attributes.icons is an array (not a string)
  - price / stock are numbers
  - card_no, set_number, artist_id are present
  - at least one foil and one noir card in the set
  - an artifact with hp but no attack_dice
  - a companion
Report a short list: what's present, what's missing, what looks
wrong. Do NOT change the backend yet.

================================================================
PROCESS
================================================================

Order: 1, 2, 3. Show a short plan first (no code). After 3,
I'll decide whether we go to bottleneck next or fix a data
issue.
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

1→4 final polish

Attribute filters ✅ (verify: работают, комбинация, prev/next учитывает)

Следующий полиш:

🟧 Убрать Фишка / УЧР / Поворот / Мгновенно

🟧 Collapse больших групп (Класс, Иконки)

🟧 Curl sanity-check бэкенда

Следующая волна:

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии в UI (/home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/NOIZ_COMMENTS.md)

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр