-- +goose Up
-- Gears flower palettes: digital merch, bought in the shop (Скины).

INSERT INTO items (id, name, description, price, category, game_id, image_url, available, stock, version)
VALUES
  (
    'c3333333-3333-4333-8333-333333333301',
    'Подсолнух',
    'Тёплые шестерни Gears в палитре подсолнуха. Косметика игры.',
    400, 'game_skin', 'gears', '/images/brand/planet-hero.png', true, 999, 1
  ),
  (
    'c3333333-3333-4333-8333-333333333302',
    'Роза',
    'Шестерни Gears в розовой палитре. Косметика игры.',
    400, 'game_skin', 'gears', '/images/brand/logo-minimal.png', true, 999, 1
  ),
  (
    'c3333333-3333-4333-8333-333333333303',
    'Анютины глазки',
    'Шестерни Gears в фиолетово-жёлтой палитре анютиных глазок. Косметика игры.',
    500, 'game_skin', 'gears', '/images/brand/planet-hero.png', true, 999, 1
  )
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price = EXCLUDED.price,
  category = EXCLUDED.category,
  game_id = EXCLUDED.game_id,
  image_url = EXCLUDED.image_url,
  available = true,
  stock = EXCLUDED.stock;

-- +goose Down
DELETE FROM items
WHERE id IN (
  'c3333333-3333-4333-8333-333333333301',
  'c3333333-3333-4333-8333-333333333302',
  'c3333333-3333-4333-8333-333333333303'
);
