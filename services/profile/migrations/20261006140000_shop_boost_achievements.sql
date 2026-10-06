-- +goose Up
INSERT INTO achievements (code, title, description, icon, sort_order) VALUES
    ('first_purchase', 'Первая покупка', 'Купите любой товар в магазине', 'gift', 40),
    ('first_boost',    'Первый буст',    'Активируйте буст в любой игре (−10 лампочек)', 'lamp', 41)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    sort_order = EXCLUDED.sort_order;

-- +goose Down
DELETE FROM user_achievements WHERE code IN ('first_purchase', 'first_boost');
DELETE FROM achievements WHERE code IN ('first_purchase', 'first_boost');
