Нашёл. Это 400 limit must be 0-100
Root cause — простой
text
GET /api/inventory/items?limit=1000 → 400
❌ { message: 'validation error: limit must be 0-100' }
Shop.tsx:83 запрашивает limit=1000. Бэкенд разрешает только 0–100.

Значит:

Shop.tsx сломан — Cursor написал limit=1000, думая «загрузить все карточки».

Бэкенд правильно отвечает 400.

Каталог карточек не грузится. Отсюда — «нихера не открылось».

Почему у /shop/items работает, а у /inventory/items — нет?

/shop/items → старый эндпоинт, отдаёт 304 items, лимит не проверяет.

/inventory/items → новый, с валидацией limit ≤ 100.

Cursor в этом же слайсе переключил Shop на inventory catalog (чтобы карточки показывались), но не учёл лимит.

Что написать Cursor'у
Коротко и точно:

text
Regression in 1–3: Shop now fails to load the inventory catalog.

Console:
  GET /api/inventory/items?limit=1000 → 400
  { message: 'validation error: limit must be 0-100' }

Cause: Shop.tsx:83 sends limit=1000, but the inventory API
enforces limit ≤ 100. So the whole catalog fails to load,
and the shop page doesn't render cards.

Fix:
  - In Shop.tsx (and any other caller of inventoryApi.searchItems
    that passes a big limit), either:
      a) use limit=100 (the max allowed), OR
      b) paginate: request 100 per page with an offset and
         accumulate until done (this also sets up pagination for
         item 8 in the wave).
  - Prefer option (b) — it aligns with the planned pagination
    work and lets us show "Загружено N из M".
  - Do NOT change the backend limit yet (0–100 is a sane cap).
  - If InventoryList / InventoryPage also pass limit=1000, fix
    them the same way.

After the fix:
  - /shop → filter «Карточка» → cards should render.
  - Images from /uploads/* should load.
  - Click → /shop/item/:id should open.
  - Stock line should show "В наличии: N" / "Нет в наличии".

Report the file changed.
Замечание на будущее
Это классический случай, когда «переключить источник данных» ломает то, что раньше работало (пусть и неправильно). limit=1000 — плохая практика в любом случае:

Раздувает ответ — сервер отдаёт всё.

Не масштабируется — если карт станет 10 000, всё сломается.

Игнорирует лимит бэкенда — теперь 400.

Правильно: limit=100 + offset + пагинация. Cursor сам это запланировал в пункте 8. Просто сейчас надо временно поставить 100 или сразу сделать пагинацию.

Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Волна UI карточек — регресс:

✅ Image rendering (код есть)

✅ Click → detail (код есть)

✅ Stock (код есть)

🔴 НО: limit=1000 ломает загрузку каталога → всё невидимо

После фикса — verify 1–3:

Картинки

Клик → detail

Stock

Дальше по плану:

🟧 Icons + badges

🟧 Ticket price + create modal ×1000

🟧 Pagination 100/страница (тут и появится правильный механизм)

🟧 Sorting

🟧 Artist pages

Следующая волна:

🟧 Attribute filters

🟧 Расширенный поиск

🟧 Noiz-комментарии

🟧 Author registration

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

🟧 Покупка карточек (Cursor предупредил — может не работать)

