-- Platform P1: shop example art + hide картина example + optional C3 smoke hide.
-- HARD RULE: never touch Berserk type='карточка' except C3 Smoke (test) hide in shop only.
-- Idempotent.

-- ---------------------------------------------------------------------------
-- SECTION A — eventhorizon_inventory
-- ---------------------------------------------------------------------------
BEGIN;

-- Soft-delete example painting (Emma: remove from project)
UPDATE inventory_items
SET deleted_at = COALESCE(deleted_at, NOW()),
    updated_at = NOW()
WHERE id = 'a1111111-1111-4111-8111-111111111102'
  AND type <> 'карточка';

-- Update images for remaining examples
UPDATE inventory_items
SET images = ARRAY['/images/shop/brelok-cosmic.jpg'],
    updated_at = NOW()
WHERE id = 'a1111111-1111-4111-8111-111111111101'
  AND type <> 'карточка';

UPDATE inventory_items
SET images = ARRAY['/images/shop/fenechka-orbit.jpg'],
    updated_at = NOW()
WHERE id = 'a1111111-1111-4111-8111-111111111103'
  AND type <> 'карточка';

UPDATE inventory_items
SET images = ARRAY['/images/shop/badge-horizon.jpg'],
    updated_at = NOW()
WHERE id = 'a1111111-1111-4111-8111-111111111104'
  AND type <> 'карточка';

SELECT 'inv_examples_active' AS metric, COUNT(*)::text AS value
FROM inventory_items
WHERE id IN (
  'a1111111-1111-4111-8111-111111111101',
  'a1111111-1111-4111-8111-111111111103',
  'a1111111-1111-4111-8111-111111111104'
) AND deleted_at IS NULL
UNION ALL
SELECT 'inv_painting_soft_deleted', COUNT(*)::text
FROM inventory_items
WHERE id = 'a1111111-1111-4111-8111-111111111102' AND deleted_at IS NOT NULL
UNION ALL
SELECT 'inv_cards_active', COUNT(*)::text
FROM inventory_items WHERE type = 'карточка' AND deleted_at IS NULL;

COMMIT;

-- ---------------------------------------------------------------------------
-- SECTION B — eventhorizon_shop
-- ---------------------------------------------------------------------------
BEGIN;

-- Hide painting example from shop grid
UPDATE items
SET available = false
WHERE id = 'a1111111-1111-4111-8111-111111111102';

-- Hide C3 Smoke Card from shop (test item; inventory карточка row untouched)
UPDATE items
SET available = false
WHERE id = 'fbf3a60c-60ce-4e52-858d-72092f057ad4';

UPDATE items
SET image_url = '/images/shop/brelok-cosmic.jpg',
    available = true
WHERE id = 'a1111111-1111-4111-8111-111111111101';

UPDATE items
SET image_url = '/images/shop/fenechka-orbit.jpg',
    available = true
WHERE id = 'a1111111-1111-4111-8111-111111111103';

UPDATE items
SET image_url = '/images/shop/badge-horizon.jpg',
    available = true
WHERE id = 'a1111111-1111-4111-8111-111111111104';

SELECT 'shop_examples_avail' AS metric, COUNT(*)::text AS value
FROM items
WHERE available AND id IN (
  'a1111111-1111-4111-8111-111111111101',
  'a1111111-1111-4111-8111-111111111103',
  'a1111111-1111-4111-8111-111111111104'
)
UNION ALL
SELECT 'shop_painting_hidden', COUNT(*)::text
FROM items
WHERE id = 'a1111111-1111-4111-8111-111111111102' AND available = false
UNION ALL
SELECT 'shop_c3_smoke_hidden', COUNT(*)::text
FROM items
WHERE id = 'fbf3a60c-60ce-4e52-858d-72092f057ad4' AND available = false;

COMMIT;
