Промпт Cursor'у:

text
Wave 2 #1 verified. Push + next: Реальные товары.

Task: drop placeholder items, keep Berserk CCG cards.

Screenshots of current /shop:
  - Cards (real): Дикий сугед, Живая баллиста, Пардус, Поварёнок
  - Placeholders (old test data):
      Ключница Дракон, Супер-пупер брелок,
      Автоматический брелок (×3), Брелок с медведем,
      Медный брелок, Тестовый брелок,
      Золотая птичка, Радужные трубы,
      Радужные блоки, Карточки со зверями,
      Космические блины, Блинный мерч

Plan (no code first):
  1. List all items in DB by type:
       - how many карточка (280 expected)
       - how many мерч / брелок / картина / фенечка / скин /
         тема
  2. Decide with me which types to KEEP:
       a) Keep карточка (all 280)
       b) Keep скины / темы (they have game_id fallback icons)?
       c) Drop мерч / брелок / картина / фенечка (placeholder)?
  3. SQL / script to delete by id (backup first — dump to a
     file before delete).
  4. Do NOT delete items that have been PURCHASED (any row in
     shop.inventory or purchases). Those must stay for users.
  5. Verify /shop after: only kept items remain.

Report counts before/after. Show the delete list for approval
before running.

DO NOT TOUCH: seed, code, games.
Что я держу в карте
Закрыто (v1.0.9 + Wave 1 + Wave 2 #1):

Все прошлые волны.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f thin DTO.

Сейчас:

⏸ Rebuild + verify payload меньше

⏸ Push Wave 1 + Wave 2 #1

⏸ Бэкап на флешку

⏸ Wave 2 #2 — Реальные товары

Wave 2 (осталось):

🟧 #2 Реальные товары

🟧 #3 Лампочки как бусты

🟧 #4 Уровни сложности

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py

