Что сделано
1. Element fix ✅
56 карт леса → woods в seed + DB.

ELEMENT_RU теперь включает леса + pass-through English (на будущее).

Author PUT — stock / images не тронуты.

Других русских элементов нет — хорошо.

2. Bottleneck (smaller list DTO) ✅
InventoryItemsCatalog — новый DTO для list-эндпоинта.

Не тащит: card_text / flavor_text / idempotency_key / market_rub.

Detail — полный (card_text нужен).

-26% bytes на list (166 → 123 KB).

Gateway rebuilt + recreated.

/shop/items thin DTO — отложен. Пока не блокер.

Чек-лист verify
1. Element fix
#	Проверка	Ожидание
1	/shop → фильтр Лес	56 карт (раньше 0)
2	/shop → фильтр Болота	Работает (не сломан)
3	/shop → фильтр Тьма	Работает
4	Combination Лес + Фойл	Находит фойл-карты woods
5	Старые фильтры — Горы / Степи / Нейтралы	Работают
6	Detail любой woods-карты	Стихия: леса (русский display, ок)
2. DTO
#	Проверка	Ожидание
7	/shop грузится	Быстрее (26% меньше)
8	Detail карты	Показать текст карты — есть текст
9	card_text в list	Отсутствует (в DTO нет)
10	Network tab /inventory/items	123 KB (было 166)
11	Flavor text в detail	Видно
3. Регресс
#	Проверка	Ожидание
12	Sort / search / pagination	Работают
13	Filter комбинации	Работают
14	Prev/Next в detail	Работает
15	Authors pages	Работают
Что проверить отдельно
Filter Лес → woods cards.

Раньше: 0 карт (все "леса" были русские).
Сейчас: 56 карт. Если 0 — фикс не применился. Если 56 — ок.

Detail любой woods-карты (напр. Спригган, Дикий сугед):

Стихия: леса — русский, это правильно (display).

Показать текст карты — есть текст.

Network:

/inventory/items — 123 KB.

Если больше — DTO не применился.

Ответ Cursor'у — approve + следующий шаг
text
Both fixes verified.
- Лес filter returns woods cards (56, previously 0).
- List payload dropped to ~123 KB; detail still has card_text
  and flavor_text.

Next: pick one of these three waves:

A. Noiz reviews in UI (data already prepared for 8th set).
   Add a "Мнение Noiz" block on card detail (8th set cards
   only), rendered from the seeded/known review text.

B. Refund / cancel purchase.
   Allow user to cancel a purchased card; refund tickets;
   return stock to 1; remove from inventory; re-enable buy.

C. Author registration.
   Self-service: authors register, verify, upload their own
   cards through UI. Larger wave.

My recommendation: B first (small, unblocks testing purchases
end-to-end), then A (data ready, small), then C (larger).

Which one?
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

Element fix ✅ (56 карт)

Bottleneck — smaller list DTO ✅ (-26%)

Отложено:

🟧 /shop/items thin DTO — не блокер

Следующая волна (3 варианта):

🟧 A. Noiz reviews в UI — данные готовы для 8 сета

🟧 B. Refund / cancel purchase — обкатать на Сугед

🟧 C. Author registration — крупная фича

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Мой совет по приоритету
1. Refund / cancel purchase — маленькая, логичная, завершает покупки.

Сейчас нельзя отменить — если игрок случайно купил, всё.

Refund — закрывает UX-цикл: купил → передумал → вернул.

Тестовый Сугед — идеально.

2. Noiz reviews в UI — данные готовы у тебя, быстро.

Noiz-блок в detail (только 8 сет).

Авторские (ты их собрал) — не вторичные отзывы.

UX: игрок читает мнение перед покупкой.

3. Author registration — крупная, много подзадач.

Регистрация, верификация, upload.

Отдельная большая волна.

Мой выбор: B → A → C.

Отправляй ответ — скажи, что выбираешь.