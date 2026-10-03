🔴 Bottleneck — честно назван
Largest cost is still the one-shot searchAllItems() (~280 cards) on shop/authors mount.

Это правда. Сейчас:

Каталог грузится целиком при заходе.

280 items → один большой ответ.

После — всё client-side (filter / sort / page — бесплатно).

Что предлагает Cursor:

Меньше DTO (без полного card_text и т.п.).

Серверная пагинация.

Не React-оптимизация.

Согласен. React тут ни при чём. Сетевой I/O — вот что медленно.

Вопрос в конце
Запускать make fe-build и деплой фронта, чтобы ты проверил волну на живом стенде?

Ответ:

Сначала — да. make fe-build → деплой → ты чекаешь A–H визуально.

Потом — approve на attribute filters (следующая волна).

Потом — bottleneck fix (server-side page или smaller DTO).

Сейчас не надо сразу чинить bottleneck — он не блокер, просто медленно на первом входе.

Готовый ответ
text
Yes, run `make fe-build` and deploy frontend. I'll verify A–H
on the live stand.

Bottleneck noted — one-shot searchAllItems() at ~280 items is
acceptable for now. Next win is server-side pagination or a
smaller list DTO, not FE optimization. Not blocking this wave.

After I verify A–H visually:
  - If clean: next wave is attribute filters (element / rarity /
    artist 1-to-1 / foil / noir / flying / companion / unique /
    cost_tier / set_number / year / stats with operators / icons
    multi-select), proberserk-style two-column search panel.
  - Bottleneck fix goes after attribute filters.
Коротко, по делу, с планом.

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

8 Pagination ✅

9 Sorting ✅

10 Artist pages ✅

A–H visual + UX polish ✅ (tsc зелёный, ждёт deploy + verify)

Ждёт:

🟧 make fe-build + деплой фронта

🟧 Визуальная проверка A–H (карты, sorting, авторы, prev/next, swipe)

Следующая волна:

🟧 Attribute filters (element / rarity / artist 1-в-1 / foil / stats / icons) — proberserk-style

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии /home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/NOIZ_COMMENTS.md

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Опционально / отложено:

🟧 «Страж сокровищ» — проверить арт

🟧 Visual audit 280 карт