-- +goose Up
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    attributes JSONB NOT NULL DEFAULT '{}',
    images TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_author ON inventory_items(author_id);
CREATE INDEX IF NOT EXISTS idx_inventory_type ON inventory_items(type);
CREATE INDEX IF NOT EXISTS idx_inventory_attributes ON inventory_items USING GIN(attributes);

-- +goose Down
DROP TABLE IF EXISTS inventory_items;