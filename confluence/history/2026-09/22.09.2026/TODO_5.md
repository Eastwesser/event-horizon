Принято — фикс применён
Ты уже прислал этот отчёт. Дублирую вывод коротко.

Что закрыто
Root cause: proto3 опускает stock: 0 из JSON → фронт видит undefined → рендерит В наличии: без числа.

Fix: 0/omitted → «Нет в наличии», >0 → «В наличии: N», null → скрыть.

Логика правильная.

Что дальше — 3 шага
1. Визуальная проверка (2 минуты)
Открой /inventory.

Карточки с stock = 0 → «Нет в наличии».

Ключница Дракон → «В наличии: 10».

Никаких пустых В наличии: без числа.

2. Аудит proto3 zero-omission
Тот же баг может кусать в других полях — price, lamps, tickets, scores, is_active. Отправь Cursor'у:

text
Root cause was proto3 JSON omitting zero values. Stock is fixed,
but the same pattern may bite other numeric/boolean fields.

Audit the proto schema + frontend:
  - Which numeric/boolean fields exist?
  - Which are optional (preserve zero) vs implicit zero-omitted?
  - Which frontend components read them and might hit the same
    undefined-when-zero bug?

Report only. No fix yet.
3. Task 2 — admin panel plan
После аудита — запрашиваем план /admin.

Что я держу
Закрыто:

✅ Stock rendering

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

В работе:

🟧 Аудит proto3 (отправить)

🟧 Task 2 plan (после аудита)

Осталось:

🟧 Чистка тестовых данных (вручную)

🟧 Коммит dev/frontend → main