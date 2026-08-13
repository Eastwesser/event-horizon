-- +goose Up
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('user', 'author', 'admin'));
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- +goose Down
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users DROP COLUMN role;
