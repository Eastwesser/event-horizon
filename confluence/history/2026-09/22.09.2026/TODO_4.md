Разбор
Две вещи. Одна — данные, вторая — баг рендера.

1. Описание отсутствует — это данные, не баг
Три Автоматический брелок, Супер-пупер брелок, Брелок с медведем, Медный брелок — созданы без описания. Это те же старые тестовые данные от 6 августа, про которые Cursor уже говорил.

Карточка просто не рендерит пустое описание — что правильно. Когда ты создашь новый предмет — описание будет, если ты его заполнишь.

Что можно сделать: почистить тестовые данные вручную (у тебя author-права). Или оставить, они не мешают.

2. 🔴 В наличии: без числа — это баг
Смотри внимательно:

Ключница Дракон → В наличии: 10 ✅

Супер-пупер брелок → В наличии: и пусто ❌

Автоматический брелок → В наличии: и пусто ❌

Медный брелок → В наличии: и пусто ❌

Проблема: либо stock = null, либо stock = 0 рендерится как пустая строка, либо значение обрезается.

Что должно быть:

Если stock есть → В наличии: N

Если stock null/undefined → всю строку скрыть (не показывать «В наличии:» без числа)

Если stock = 0 → Нет в наличии (красным/серым), не пустая строка

Это баг рендера. Логика неполная.

Промпт для Cursor
text
Two issues on the Inventory card. #2 is a bug.

================================================================
1. "В наличии:" renders without a number (BUG)
================================================================

On the Inventory grid, several cards show the label "В наличии:"
with NO number after it. Only "Ключница Дракон" shows
"В наличии: 10". Others are empty.

Expected behavior:
  - stock is a number  → "В наличии: N"
  - stock is 0         → "Нет в наличии" (muted, not empty)
  - stock is null/undefined → hide the whole line entirely

Never render the label alone.

Investigate first: what is the actual stock value for those
items in the DB? Is it null, 0, empty string, or a rendering
issue with number formatting? Report the value you find.

Fix in InventoryItemCard.tsx (or wherever the stock line lives).

================================================================
2. Empty description — data, not a bug (no code change)
================================================================

Several items have no description in the DB (old test data from
2026-08-06). The card correctly renders nothing — that's fine.

No fix needed. I will clean the test rows manually.

================================================================
PROCESS
================================================================
Step 1: Report the actual stock value for the empty cards.
Step 2: Fix the rendering (three cases above).
Step 3: Report the file changed.
Что я держу в карте
Закрыто:

✅ Task 1 (image URL) — работает

✅ Ред. — влезает

✅ Role boundaries подтверждены

Новое:

🔴 В наличии: без числа — баг рендера

Осталось:

🟧 Task 2 (admin panel) — план ждёт

🟧 Чистка тестовых данных вручную

🟧 Коммит dev/frontend → main