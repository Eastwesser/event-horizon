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

CREATE INDEX idx_inventory_author ON inventory_items(author_id);
CREATE INDEX idx_inventory_type ON inventory_items(type);
CREATE INDEX idx_inventory_attributes ON inventory_items USING GIN(attributes);
CREATE INDEX idx_inventory_name ON inventory_items USING GIN(to_tsvector('russian', name));

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_inventory_updated_at ON inventory_items;
CREATE TRIGGER update_inventory_updated_at BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_inventory_updated_at ON inventory_items;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_inventory_author;
DROP INDEX IF EXISTS idx_inventory_type;
DROP INDEX IF EXISTS idx_inventory_attributes;
DROP INDEX IF EXISTS idx_inventory_name;
DROP TABLE IF EXISTS inventory_items;
