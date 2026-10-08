-- Shop cosmetics: ensure skins tab + themes tab are non-empty.
-- Run against: eventhorizon_shop
-- Safe / idempotent. Does not touch Berserk карточка rows.
--
-- Usage:
--   psql "$SHOP_DSN" -f scripts/seed-shop-themes-skins.sql

BEGIN;

-- Rename rainbow → cosmic (kids-safe copy). Keep purchase history on same ids.
UPDATE items
SET name = 'Космические трубы',
    description = 'Трубы в Flappy в космической палитре горизонта событий.',
    available = true
WHERE id = '6a1de8dd-9457-4aa4-99a7-78267aee731d'
   OR (name ILIKE '%радужн%труб%' AND category = 'game_skin');

UPDATE items
SET name = 'Космические блоки',
    description = 'Блоки Builder в космической палитре.',
    available = true
WHERE id = '68da9951-0925-48d2-a755-511ecf05e5d7'
   OR (name ILIKE '%радужн%блок%' AND category = 'game_skin');

-- Canonical skins stay on the grid (same ids as cleanup-shop-content-v2).
UPDATE items
SET available = true
WHERE id IN (
  '82be50db-670b-48c6-beb9-7e00d584f6de', -- Золотая птичка
  'c586a47f-b6d3-49d3-a627-26903aa5e26e', -- Карточки со зверями
  'be87a17f-8208-4442-8395-19a3540edcf3', -- Космические блины
  '68da9951-0925-48d2-a755-511ecf05e5d7', -- Космические блоки
  '6a1de8dd-9457-4aa4-99a7-78267aee731d'  -- Космические трубы
);

-- Profile themes (≥1) — digital cosmetics, cheap tickets.
INSERT INTO items (id, name, description, price, category, game_id, image_url, available, stock, version)
VALUES
  (
    'b2222222-2222-4222-8222-222222222201',
    'Тема «Туманность»',
    'Тёмный профиль с мягким свечением горизонта. Косметика профиля.',
    500, 'profile_theme', '', '/images/brand/planet-hero.png', true, 999, 1
  ),
  (
    'b2222222-2222-4222-8222-222222222202',
    'Тема «Золотой рассвет»',
    'Тёплые золотые акценты для профиля. Косметика профиля.',
    750, 'profile_theme', '', '/images/brand/logo-minimal.png', true, 999, 1
  )
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price = EXCLUDED.price,
  category = EXCLUDED.category,
  image_url = EXCLUDED.image_url,
  available = true,
  stock = EXCLUDED.stock;

SELECT 'shop_skins_available' AS metric, COUNT(*)::text AS value
FROM items WHERE available AND category = 'game_skin'
UNION ALL
SELECT 'shop_themes_available', COUNT(*)::text
FROM items WHERE available AND category = 'profile_theme';

COMMIT;
