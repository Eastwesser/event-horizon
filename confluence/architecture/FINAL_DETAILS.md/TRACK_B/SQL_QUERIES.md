[denismatveev@c0der event_hobash scripts/apply-shop-cleanup-v2.sh
=== Shop cleanup v2 ===
SQL: /home/denismatveev/event_horizon/scripts/cleanup-shop-content-v2.sql
Berserk cards (inventory type=карточка, active) BEFORE: 281
Backup → /home/denismatveev/backups/shop-cleanup-v2-20261006_041548
Applying SECTION A → inventory…
BEGIN
UPDATE 1
UPDATE 5
UPDATE 0
INSERT 0 4
          metric           | value 
---------------------------+-------
 inventory_cards_active    | 281
 inventory_examples_active | 4
(2 rows)

COMMIT
Applying SECTION B → shop…
BEGIN
UPDATE 6
UPDATE 5
UPDATE 9
UPDATE 5
INSERT 0 4
          metric          | value 
--------------------------+-------
 shop_available_total     | 289
 shop_available_game_skin | 5
 shop_examples_available  | 4
(3 rows)

COMMIT
Berserk cards AFTER: 281
=== Verify counts ===
 category  | available | total 
-----------+-----------+-------
 game_skin |         5 |    14
 merch     |       284 |   295
(2 rows)

             name             | available 
------------------------------+-----------
 Значок Event Horizon         | t
 Золотая птичка               | t
 Картина «Туманность Horizon» | t
 Космический брелок           | t
 Радужные трубы               | t
 Фенечка «Орбита»             | t
(6 rows)

✅ Done. Cards untouched (281). Backup: /home/denismatveev/backups/shop-cleanup-v2-20261006_041548
[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c \
  "SELECT COUNT(*) FROM inventory_items WHERE type='карточка' AND deleted_at IS NULL;"
 count 
-------
   281
(1 row)

[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c \
  "SELECT category, COUNT(*) FILTER (WHERE available) AS avail, COUNT(*) AS total
   FROM items GROUP BY category ORDER BY 1;"
 category  | avail | total 
-----------+-------+-------
 game_skin |     5 |    14
 merch     |   284 |   295
(2 rows)

[denismatveev@c0der event_horizon]$ 



[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c \
  "SELECT id, name, category FROM items WHERE id IN (
    SELECT id FROM items WHERE available=true
    EXCEPT SELECT id FROM items WHERE category='merch' AND name LIKE '%Сердце%'
  ) ORDER BY name LIMIT 20;"
                  id                  |        name         | category 
--------------------------------------+---------------------+----------
 50bfd4a4-7c8d-47e8-922f-f209d091abc2 | Ёж-воитель          | merch
 0b145c81-2e01-447a-a0ab-26ade03e69ef | Авгур               | merch
 a60795ca-8585-48a8-a416-a6a5c995e1ee | Автоматон           | merch
 9d35398d-f8b5-4923-b383-6e300189ea0b | Адепт пламени       | merch
 3e23a88c-5a6d-4f8f-a723-d78957dbf053 | Адская гончая       | merch
 c7dd2618-76c7-4d19-b393-f01c9cd9d94a | Акванит-лучник      | merch
 be586392-3d7a-4d75-a39d-f99ed86d78ea | Алая тетива         | merch
 f436f03d-2f3c-4239-af97-01e9136bdc8d | Альхарис            | merch
 b3054faa-734d-4f71-b12a-509c6c5d6810 | Ангел возмездия     | merch
 2f970a2a-7cf9-43e5-9fa1-9c4f6d782255 | Ангел познания      | merch
 0adf1e70-2b0c-49a9-8015-456fafd7a0dd | Арбалетчик Братства | merch
 3d89f65a-287c-4bb7-a7e1-0f290574e52d | Аримас              | merch
 042c8177-c5b4-4043-a655-8c015dcd5db0 | Асгара              | merch
 b2401812-7da5-4058-9d9c-05d471f19042 | Ассасин Братства    | merch
 3f00c102-4524-4f36-9764-bb54f371d9cd | Ашаби               | merch
 599a8e79-d3f5-4a08-8373-5328d82104fe | Баалит              | merch
 8191543e-0979-466d-b25c-eba8d48a4fc8 | Безбородый          | merch
 4ba057ea-524a-4abd-9dee-ad809b2d1c4e | Беллигемин          | merch
 15b5eeb2-3a07-4c7b-a944-452184bef758 | Белый шквал         | merch
 c6cee94a-2f4c-4bf4-bdd4-ef5c02d73a34 | Бер                 | merch
(20 rows)

[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c \
  "SELECT id, name, category, available, price, created_at
   FROM items
   WHERE created_at > NOW() - INTERVAL '1 hour'
   ORDER BY created_at DESC;"

                  id                  |             name             | category | available | price |  created_at
--------------------------------------+------------------------------+----------+-----------+-------+----------------------------
 a1111111-1111-4111-8111-111111111101 | Космический брелок           | merch    | t         |   100 | 2026-10-06 01:15:55.947839
 a1111111-1111-4111-8111-111111111102 | Картина «Туманность Horizon» | merch    | t         |   100 | 2026-10-06 01:15:55.947839
 a1111111-1111-4111-8111-111111111103 | Фенечка «Орбита»             | merch    | t         |   100 | 2026-10-06 01:15:55.947839
 a1111111-1111-4111-8111-111111111104 | Значок Event Horizon         | merch    | t         |   100 | 2026-10-06 01:15:55.947839
(4 rows)

[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c \

  "SELECT id, name, type, price, created_at
   FROM inventory_items
   WHERE created_at > NOW() - INTERVAL '1 hour'
   ORDER BY created_at DESC;"

                  id                  |             name             |  type   | price  |         created_at         
--------------------------------------+------------------------------+---------+--------+----------------------------
 a1111111-1111-4111-8111-111111111101 | Космический брелок           | брелок  | 100.00 | 2026-10-06 01:15:55.074874
 a1111111-1111-4111-8111-111111111102 | Картина «Туманность Horizon» | картина | 100.00 | 2026-10-06 01:15:55.074874
 a1111111-1111-4111-8111-111111111103 | Фенечка «Орбита»             | фенечка | 100.00 | 2026-10-06 01:15:55.074874
 a1111111-1111-4111-8111-111111111104 | Значок Event Horizon         | мерч    | 100.00 | 2026-10-06 01:15:55.074874
(4 rows)



Что вижу
Name	Type	Price
Космический брелок	брелок	100
Картина «Туманность Horizon»	картина	100
Фенечка «Орбита»	фенечка	100
Значок Event Horizon	мерч	100
Всё правильно. Значок = мерч — логично (значок — это мерч).

Итог Shop cleanup v2
Что получилось:

inventory: 281 карточка + 5 скинов + 4 examples = 290 items.

type=карточка — 281 (Berserk ККИ).

type=брелок — 1 (example) + 1 (Ключница, hidden)** = 2?

type=картина — 1 (example) + 1 (Kitsune, soft-deleted)** = 1?

type=фенечка — 1 (example).

type=мерч — 1 (Значок).

shop: 289 available.

merch — 284 available (280 Berserk + 4 examples).

game_skin — 5 available (5 unique).

11 placeholder hidden (available=false).

Cards 281 — не тронуты. ✅

Один момент — счётчик
В inventory_items.type='брелок' active:

1 example.

1 Ключница Дракон (из старого seed, purchased).



[denismatveev@c0der event_horizon]$ docker exec event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c \
  "SELECT type, COUNT(*) FILTER (WHERE deleted_at IS NULL) AS active,
          COUNT(*) AS total
   FROM inventory_items GROUP BY type ORDER BY 1;"
   type   | active | total 
----------+--------+-------
 брелок   |      1 |    10
 картина  |      1 |     2
 карточка |    281 |   281
 мерч     |      1 |     1
 фенечка  |      1 |     1
(5 rows)

[denismatveev@c0der event_horizon]$ 
