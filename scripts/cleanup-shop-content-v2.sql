-- =============================================================================
-- Shop cleanup v2 — KEEP Berserk cards, 5 skins, 4 examples; hide/delete junk.
-- =============================================================================
--
-- ██████  HARD RULE — DO NOT TOUCH BERSERK CCG CARDS (~280 / 281)  ██████
-- They are REAL merch. Mutations below use EXPLICIT UUID lists ONLY.
-- Never: WHERE category='merch', WHERE type='карточка', or name ILIKE on cards.
-- Before/after: card count MUST stay identical (assert in apply wrapper).
--
-- Apply (after Emma OK + backup):
--   1) inventory DB  — this file section A
--   2) shop DB       — this file section B
-- Idempotent: safe to re-run.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- SECTION A — run against: eventhorizon_inventory
-- -----------------------------------------------------------------------------

BEGIN;

-- A1) Soft-delete unpurchased test junk (NOT cards).
--     HARD RULE: never soft-delete type='карточка' — C3 Smoke Card stays in inventory;
--     it is only hidden from the shop grid in section B2.
UPDATE inventory_items
SET deleted_at = COALESCE(deleted_at, NOW()),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND type <> 'карточка'
  AND id IN (
    '9a32f9e8-193e-4a6c-b88a-21661d0a3ac2', -- ProtoZeroTest
    '495a56a7-d8ec-4e2f-906c-673e3bcad2d5', -- Kitsune (картина) — replaced by example
    'fd96dd05-f38f-4157-ab63-fce7c3812681', -- Vika
    '35b06c27-96d1-46ae-92fe-8ad6e8c9b757'  -- Berserk generic test (not a card)
  );

-- A2) Soft-delete purchased placeholders + C3 Keychain from catalog
--     (owners keep rows in shop.inventory / purchases; catalog listing hidden via shop.available).
UPDATE inventory_items
SET deleted_at = COALESCE(deleted_at, NOW()),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND type <> 'карточка'
  AND id IN (
    'ff137485-f356-4ada-9b00-bfb84546204a', -- Ключница Дракон
    '5e3254d2-ed02-4584-af17-dcd3b4b0cfef', -- Супер-пупер брелок
    'c96419b1-f283-4fb5-b03f-d73baf6efd52', -- Автоматический брелок
    '4fb5c062-d212-4dff-bab1-5d346f76adff', -- Брелок с медведем
    'b5b332fd-146d-464c-b5f3-389a8fa48b83', -- Блинный мерч
    '676c9155-c0b1-4733-b710-d2ce326aac6b'  -- C3 Keychain (purchased in smoke)
  );

-- A3) Soft-delete EXTRA skin duplicates (canonical skins stay active).
--     Canonical KEEP (oldest with purchase): see comments in section B.
UPDATE inventory_items
SET deleted_at = COALESCE(deleted_at, NOW()),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND type <> 'карточка'
  AND id IN (
    -- Золотая птичка extras
    '5a3b1a32-dcae-4025-a8f5-da01748dab92',
    '80fcd762-aefc-4e1a-ac1e-b8884f2a00a1',
    '51b44ecf-0e01-4254-a0dc-8c45d5a81ebf',
    -- Карточки со зверями extra
    '521ada79-8821-4304-83d8-d903f3faf987',
    -- Космические блины extra
    '7bbca382-403c-400c-b1ef-ee03af49e4de',
    -- Радужные блоки extra
    'd4734224-ab7b-46a6-b503-6886bfbf4bd2',
    -- Радужные трубы extras
    'a61fb41c-2448-43fd-b088-518690a5ab1b',
    '8a43b702-bfe8-470a-8e60-65a3692cfc0f',
    '5598845d-8126-4041-b4eb-ea59f88bf59e'
  );

-- A4) Upsert 4 Event Horizon example catalog items (idempotent).
-- author_id = seed admin (not the Berserk card author).
INSERT INTO inventory_items (id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at, deleted_at, version)
VALUES
  (
    'a1111111-1111-4111-8111-111111111101',
    '1502a3fa-0e64-4873-a329-3d8fa1d5204d',
    'брелок',
    'Космический брелок',
    'Брелок с силуэтом горизонта событий. Пример физического товара — арт заменим позже.',
    100, 50,
    '{"example": true, "kind": "brelok"}'::jsonb,
    ARRAY['/images/brand/logo-icon.png'],
    NOW(), NOW(), NULL, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111102',
    '1502a3fa-0e64-4873-a329-3d8fa1d5204d',
    'картина',
    'Картина «Туманность Horizon»',
    'Печать космической туманности. Пример картины для витрины.',
    100, 20,
    '{"example": true, "kind": "kartina"}'::jsonb,
    ARRAY['/images/brand/planet-hero.png'],
    NOW(), NOW(), NULL, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111103',
    '1502a3fa-0e64-4873-a329-3d8fa1d5204d',
    'фенечка',
    'Фенечка «Орбита»',
    'Плетёная фенечка в тёмных тонах Event Horizon. Пример фенечки.',
    100, 100,
    '{"example": true, "kind": "fenechka"}'::jsonb,
    ARRAY['/images/brand/logo-minimal.png'],
    NOW(), NOW(), NULL, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111104',
    '1502a3fa-0e64-4873-a329-3d8fa1d5204d',
    'мерч',
    'Значок Event Horizon',
    'Металлический значок с логотипом. Пример мерча для витрины.',
    100, 50,
    '{"example": true, "kind": "merch"}'::jsonb,
    ARRAY['/images/brand/logo-icon.png'],
    NOW(), NOW(), NULL, 1
  )
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price = EXCLUDED.price,
  stock = EXCLUDED.stock,
  type = EXCLUDED.type,
  attributes = EXCLUDED.attributes,
  images = EXCLUDED.images,
  deleted_at = NULL,
  updated_at = NOW();

-- Safety report (inventory)
SELECT 'inventory_cards_active' AS metric, COUNT(*)::text AS value
FROM inventory_items WHERE type = 'карточка' AND deleted_at IS NULL
UNION ALL
SELECT 'inventory_examples_active', COUNT(*)::text
FROM inventory_items
WHERE id IN (
  'a1111111-1111-4111-8111-111111111101',
  'a1111111-1111-4111-8111-111111111102',
  'a1111111-1111-4111-8111-111111111103',
  'a1111111-1111-4111-8111-111111111104'
) AND deleted_at IS NULL;

COMMIT;

-- -----------------------------------------------------------------------------
-- SECTION B — run against: eventhorizon_shop
-- -----------------------------------------------------------------------------

BEGIN;

-- B1) HIDE purchased placeholders + C3 Keychain from shop grid.
UPDATE items
SET available = false
WHERE available IS DISTINCT FROM false
  AND id IN (
    'ff137485-f356-4ada-9b00-bfb84546204a', -- Ключница Дракон
    '5e3254d2-ed02-4584-af17-dcd3b4b0cfef', -- Супер-пупер брелок
    'c96419b1-f283-4fb5-b03f-d73baf6efd52', -- Автоматический брелок
    '4fb5c062-d212-4dff-bab1-5d346f76adff', -- Брелок с медведем
    'b5b332fd-146d-464c-b5f3-389a8fa48b83', -- Блинный мерч
    '676c9155-c0b1-4733-b710-d2ce326aac6b'  -- C3 Keychain
  );

-- B2) HIDE unpurchased junk (prefer hide over hard-delete — safer with purchase history).
UPDATE items
SET available = false
WHERE available IS DISTINCT FROM false
  AND id IN (
    '9a32f9e8-193e-4a6c-b88a-21661d0a3ac2', -- ProtoZeroTest
    '495a56a7-d8ec-4e2f-906c-673e3bcad2d5', -- Kitsune
    'fd96dd05-f38f-4157-ab63-fce7c3812681', -- Vika
    '35b06c27-96d1-46ae-92fe-8ad6e8c9b757', -- Berserk generic
    'fbf3a60c-60ce-4e52-858d-72092f057ad4'  -- C3 Smoke Card (hide only; inventory row untouched)
  );

-- B3) Skin dedupe — KEEP canonical (oldest with purchase), HIDE extras.
-- KEEP:
--   82be50db… Золотая птичка
--   c586a47f… Карточки со зверями
--   be87a17f… Космические блины
--   68da9951… Радужные блоки
--   6a1de8dd… Радужные трубы
UPDATE items
SET available = false
WHERE available IS DISTINCT FROM false
  AND id IN (
    '5a3b1a32-dcae-4025-a8f5-da01748dab92',
    '80fcd762-aefc-4e1a-ac1e-b8884f2a00a1',
    '51b44ecf-0e01-4254-a0dc-8c45d5a81ebf',
    '521ada79-8821-4304-83d8-d903f3faf987',
    '7bbca382-403c-400c-b1ef-ee03af49e4de',
    'd4734224-ab7b-46a6-b503-6886bfbf4bd2',
    'a61fb41c-2448-43fd-b088-518690a5ab1b',
    '8a43b702-bfe8-470a-8e60-65a3692cfc0f',
    '5598845d-8126-4041-b4eb-ea59f88bf59e'
  );

-- Ensure canonical skins stay available.
UPDATE items
SET available = true
WHERE id IN (
  '82be50db-670b-48c6-beb9-7e00d584f6de', -- Золотая птичка
  'c586a47f-b6d3-49d3-a627-26903aa5e26e', -- Карточки со зверями
  'be87a17f-8208-4442-8395-19a3540edcf3', -- Космические блины
  '68da9951-0925-48d2-a755-511ecf05e5d7', -- Радужные блоки
  '6a1de8dd-9457-4aa4-99a7-78267aee731d'  -- Радужные трубы
);

-- B4) Upsert 4 example merch into shop (same ids as inventory).
INSERT INTO items (id, name, description, price, category, game_id, image_url, available, stock, version)
VALUES
  (
    'a1111111-1111-4111-8111-111111111101',
    'Космический брелок',
    'Брелок с силуэтом горизонта событий. Пример физического товара — арт заменим позже.',
    100, 'merch', '', '/images/brand/logo-icon.png', true, 50, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111102',
    'Картина «Туманность Horizon»',
    'Печать космической туманности. Пример картины для витрины.',
    100, 'merch', '', '/images/brand/planet-hero.png', true, 20, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111103',
    'Фенечка «Орбита»',
    'Плетёная фенечка в тёмных тонах Event Horizon. Пример фенечки.',
    100, 'merch', '', '/images/brand/logo-minimal.png', true, 100, 1
  ),
  (
    'a1111111-1111-4111-8111-111111111104',
    'Значок Event Horizon',
    'Металлический значок с логотипом. Пример мерча для витрины.',
    100, 'merch', '', '/images/brand/logo-icon.png', true, 50, 1
  )
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price = EXCLUDED.price,
  category = EXCLUDED.category,
  image_url = EXCLUDED.image_url,
  available = true,
  stock = EXCLUDED.stock;

-- Safety report (shop) — cards ≈ merch rows that are NOT in junk/hide/example lists
-- Prefer inventory card count as source of truth; shop lists cards as category=merch.
SELECT 'shop_available_total' AS metric, COUNT(*)::text AS value
FROM items WHERE available = true
UNION ALL
SELECT 'shop_available_game_skin', COUNT(*)::text
FROM items WHERE available = true AND category = 'game_skin'
UNION ALL
SELECT 'shop_examples_available', COUNT(*)::text
FROM items
WHERE available = true
  AND id IN (
    'a1111111-1111-4111-8111-111111111101',
    'a1111111-1111-4111-8111-111111111102',
    'a1111111-1111-4111-8111-111111111103',
    'a1111111-1111-4111-8111-111111111104'
  );

COMMIT;

-- Expected /shop after apply:
--   inventory active карточка: unchanged (~281, includes C3 Smoke Card row — never soft-deleted)
--   shop available: ~280 Berserk cards (C3 Smoke hidden) + 5 skins + 4 examples ≈ 289
-- Owners of hidden placeholders still see them in personal inventory via purchase history.