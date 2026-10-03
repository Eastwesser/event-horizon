Approve. План — прямо то, что нужно
Что подтверждаю
Storage — attributes.noiz_review (object). ✅

{ text, rating, verdict, author }.

В detail — уже грузится полный attributes → новый эндпоинт не нужен.

В list — omit через catalogAttrOmit → grid остаётся тонким.

UI data-driven — показывать iff noiz_review exists. Не hardcode set_number === 8. ✅

Правильно: если завтра добавишь review для 7 сета — сразу появится.

Data load — parse NOIZ_COMMENTS.md → match к 8 сет по нормализованному имени. ✅

Strip (фойл), (нуар, …) и т.п.

Idempotent backfill — seed или one-shot PUT.

Unmatched → log + skip, не invent. ✅

FE layout:

ShopItemDetail + InventoryItemDetail.

Below attributes, after flavor / card_text.

Quote-style, отличный от flavor (не курсив, не gold-border).

Header «Мнение Noiz» + rating badge + verdict chip.

Footer — Noiz.

Empty / missing → ничего.

Out of scope — правильно:

UGC, likes, admin edit.

Reviews на list cards.

Сеты кроме 8.

C / thin shop DTO / emoji→SVG.

Один момент — уточни
1. verdict — где взять?

Из текста Noiz? Или отдельно указать?

У тебя NOIZ_COMMENTS.md — там только текст?

Если verdict нет — опционально, не показывать.

2. rating — есть ли у тебя?

Ты говорил «Noiz — заядлый игрок, комментирует».

Rating (0-10) — твой или Noiz'а?

Если в тексте нет цифры — не выдумывать. rating: null → не показывать.

3. Формат NOIZ_COMMENTS.md.

Как связаны комментарии с картами?

По имени («Лак-нак»)? По card_no? По позиции в блоке?

Покажи пример одной записи — Cursor должен понять парсинг.

4. verdict словарь.

Ты упомянул: имба / норм / филлер / скип / …

Фиксированный словарь? Или свободный текст?

Если свободный — chip будет длинным. Fixed — лучше.

Ответ Cursor'у
text
OK — implement A (Noiz reviews).

Three clarifications before code:

1. NOIZ_COMMENTS.md format
   Show me one example record (name + body + any structured
   fields) so we agree on the parser. If rating / verdict are
   not in the source, leave them null and don't render them.

2. Verdict
   If the source has no explicit verdict, do not invent one.
   Render the verdict chip only when attributes.noiz_review.
   verdict is a non-empty string.

3. Matching
   Normalize both sides: strip "(фойл)", "(нуар, …)", "(х2,
   …)", "(из журнала …)", trim, lowercase, drop punctuation.
   Match within set_number = 8 only.
   Report unmatched names (log list) so I can fix the source
   later.

Everything else as written.

Order: backfill storage first (idempotent script), then
detail UI. Verify on Лак-нак, Суgед, Карбагаш + one non-8th
card (should show nothing). Report file changes + match
count.
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

Attribute filters ✅

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅ (verified end-to-end)

RBAC fix на purchase-path ✅

Текущая волна (A — approve):

🟧 Storage: attributes.noiz_review = {text, rating, verdict, author}.

🟧 Backfill — parse NOIZ_COMMENTS.md, match к 8 сета, idempotent.

🟧 FE: блок «Мнение Noiz» в detail (только если есть).

🟧 В list — omit через catalogAttrOmit.

Следующая волна:

🟧 C. Author registration — крупная.

Отложено:

🟧 /shop/items thin DTO

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр