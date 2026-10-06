OK на apply. Разбор + проверка
Что подтверждаю
Жёсткое правило учтено:

✅ C3 Smoke Card (type=карточка) убран из soft-delete inventory — иначе count 281 → 280 → abort.

✅ На все UPDATE inventory_items добавлен AND type <> 'карточка'.

✅ C3 Smoke — только hide в shop (available=false).

✅ Apply wrapper — abort, если count карточек изменится.

✅ Backup перед apply.

Файлы:

scripts/cleanup-shop-content-v2.sql.

scripts/apply-shop-cleanup-v2.sh (--dry-run ок, cards BEFORE: 281).

Ожидание после apply:

inventory карточки: 281 (без изменений).

shop: ~280 Berserk + 5 skins + 4 examples ≈ 289.

Твой вопрос — про терминологию
Надеюсь, ты не перепутал понятия карточки берсерка и карточки товара!

Разберём:

«карточка Берсерка» = товар типа карточка в inventory (281 штук). Это ККИ карты — реальный мерч.

«карточка товара» = UI-элемент (превью товара в магазине). Не БД сущность.

Cursor использует терминологию корректно:

AND type <> 'карточка' — защита именно от удаления Берсерк карт.

C3 Smoke Card — тоже type=карточка, но это test item. Cursor его не трогает в inventory, только hide в shop.

Значит — всё правильно. Термины не перепутаны.

Ответ Cursor'у
text
OK на apply.

Terminology check confirmed:
  - "карточка" in inventory = Berserk CCG cards (280 + C3 Smoke).
  - "карточка товара" in UI = shop item card. Different things.
  - Your SQL guards on `type <> 'карточка'` are correct.

Before apply — final sanity:

1. Run the dry-run once more and paste:
     - cards BEFORE: expected 281
     - cards AFTER dry-run: must equal BEFORE
     - any UPDATE/DELETE that would touch type='карточка'
       (should be zero — C3 Smoke only via shop available=false)

2. Show the exact commands:
     - backup step (shop + inventory DB dump paths)
     - apply step
     - verify queries (cards count, shop count, owner inventories)

3. Confirm idempotency: running apply-shop-cleanup-v2.sh twice
   must not double-hide, double-delete, or change counts.

After those — I'll run apply + verify + commit.

Rules:
  - Do NOT touch type='карточка' in inventory at all.
  - Shop C3 Smoke: available=false only. Never delete.
  - Purchased placeholders: hide only, keep in owner inventory.
  - No DB write until step 1–3 are confirmed.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (main, f784419):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

8 игр playable.

Сейчас — Shop cleanup v2:

🟧 scripts/cleanup-shop-content-v2.sql — готов.

🟧 scripts/apply-shop-cleanup-v2.sh — готов, dry-run ок.

🟧 Ждёт final sanity (dry-run, backup path, idempotency) перед apply.

🟧 После apply — verify + commit.

Дальше — options:

🟧 Track B parity (boost, levels).

🟧 Пауза / собесы.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.

Что добавить к промпту
Мой промпт уже был финальным — с final sanity (dry-run, backup, idempotency). Добавить можно одно — проверку терминологии в SQL:

text
One more thing before dry-run:

Confirm the guard is on the DB type, not the UI concept:

  - UPDATE inventory_items ... AND type <> 'карточка'
    → protects Berserk cards (inventory type).

  - C3 Smoke Card (type='карточка'):
      inventory: NOT touched (no UPDATE, no soft-delete)
      shop: UPDATE items SET available=false WHERE id=...
      → hide only, never delete.

  - All other UPDATE/DELETE on shop items:
      no guard on category = 'merch' (that's Berserk
      cards' category in shop DB — verify this does not
      collide with placeholder merch).

Wait — important check:
  shop.items.category = 'merch' holds both:
    · Berserk cards (same IDs as inventory карточка)
    · test merch placeholders (brelki etc.)
  So a blind UPDATE shop SET available=false WHERE
  category='merch' would hide Berserk cards too.

  Confirm your SQL does NOT filter by category alone —
  it must use explicit item IDs for hides/deletes.
Это важно. В shop.items — категория merch — 291 item, из которых ~280 — Berserk карты, а 11 — тестовые брелоки. Если Cursor сделает UPDATE ... WHERE category='merch' — скроет всё, включая карты.

Cursor в плане писал WHERE id IN (...), но стоит проверить явно.

Что я держу в карте
Закрыто (main, f784419):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

8 игр playable.

Сейчас — Shop cleanup v2:

🟧 scripts/cleanup-shop-content-v2.sql — готов.

🟧 scripts/apply-shop-cleanup-v2.sh — готов, dry-run ок.

🟧 Ждёт final sanity + проверка guard'ов на category vs id.

🟧 После apply — verify + commit.

Дальше — options:

🟧 Track B parity (boost, levels).

🟧 Пауза / собесы.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.