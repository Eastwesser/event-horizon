-- +goose Up
ALTER TABLE users ADD COLUMN nickname VARCHAR(100) DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN nickname;
