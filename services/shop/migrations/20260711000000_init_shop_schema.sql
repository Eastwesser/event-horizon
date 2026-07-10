-- +goose Up
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price INT NOT NULL,
    category TEXT NOT NULL,
    game_id TEXT,
    image_url TEXT,
    available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Начальные товары (кастомизация для Flappy)
INSERT INTO items (name, description, price, category, game_id, image_url) VALUES
('Радужные трубы', 'Сделайте трубы в Flappy радужными!', 100, 'game_skin', 'flappy', '/images/rainbow_pipes.png'),
('Золотая птичка', 'Птичка становится золотой!', 200, 'game_skin', 'flappy', '/images/golden_bird.png');

CREATE TABLE IF NOT EXISTS inventory (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    purchased_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE IF NOT EXISTS purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    price INT NOT NULL,
    currency_type TEXT DEFAULT 'tickets',
    purchased_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS purchases;