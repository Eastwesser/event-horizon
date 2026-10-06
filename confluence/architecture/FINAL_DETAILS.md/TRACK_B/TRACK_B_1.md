Разбор Track B readiness. Мой выбор — shop content
Что вижу
Три «новых» игры — уже playable:

twenty48 ✅

gears (Орбиты) ✅

companion (Компаньон) ✅

Home — 8 игр (5 classic + 3 новые). Routes, BE allowlist, leaderboard — работает.

Что не сделано (parity):

Boost — только Flappy. Остальные 7 — нет boost.

Levels — только Flappy 1–10. Остальные — in-run progression или hardcoded level: 1.

1–20 — нигде.

Track B (parity) = boost + levels на 7 играх. Weeks, не days.

Quick win — real shop items (drop placeholders). 2–3 дня.

Мой выбор — Shop content (quick win)
Причины:

Quick. 2–3 дня vs недели на parity.

Видимо. Убираем placeholder'ы (Ключница Дракон и т.п.) — магазин становится чище.

Финализирует v1.0.9/v1.1.0 хвост.

Track B parity — большая, лучше отдельным спринтом с свежей головой.

Перед собесами — чистый магазин лучше, чем недоделанный parity.

Track B parity — потом. Или — после собесов.

Промпт Cursor'у — Shop content cleanup
text
Decision: Shop content cleanup (2–3 days) before Track B
parity.

Plan first, no code.

Goal:
  1. Inventory of current shop items by category.
  2. Decide which placeholders to drop (unpurchased only).
  3. Seed real product list (Berserk CCG cards + selected
     skins/themes) or clean up.

Context:
  - Berserk CCG cards: ~280 seeded, playable, purchasable.
  - Placeholders still in shop: Ключница Дракон, Супер-пупер
    брелок, Автоматический брелок ×1, Брелок с медведем —
    kept because they were purchased in early testing.
  - Real merch / real skins? — clarify what counts as
    "real" here.

Steps:

1. DB inventory:
   - List items by category (merch / game_skin / карточка /
     profile_theme / прочее) with count + sample names.
   - Mark: purchased (has row in inventory) vs unpurchased.
   - Mark: real content vs placeholder (test data).

2. Keep/delete list (draft for my review):
   - Cards: keep all 280.
   - Skins: keep real ones (Золотая птичка, Радужные трубы,
     Радужные блоки, Космические блины, Карточки со
     зверями) — verify all still in shop.
   - Placeholders: unpurchased → delete; purchased → hide
     from shop grid but keep for the buyer's inventory.

3. Real content gap:
   - Do we need to fill the shop with more real items
     besides CCG cards? (Real merch designs? Real skins?)
   - If yes → what exists, what's the source?

4. Migrations / scripts:
   - No new migration. Use a one-off SQL or existing
     inventory soft-delete.
   - Idempotent — safe to re-run.

5. Verify after:
   - /shop grid: only real content.
   - Purchased placeholders still in the buyer's inventory.
   - No "already owned" bugs.

Out of scope:
  - Track B parity (boosts/levels on other games).
  - C4, refund, notifications, auth.
  - Real card art polish.

Show the plan. I'll approve before any DB write.
Что я держу в карте
Закрыто (main, f784419):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3: Refund, Notifications, JWT refresh.

8 игр (5 classic + 3 новые) — playable.

Дальше — options:

🟧 Shop content cleanup (мой выбор, 2–3 days).

🟧 Track B parity (boost + levels на 7 играх) — weeks.

🟧 Пауза — подготовка к собесам.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.

🟧 Companion polish.