-- +goose Up
CREATE TABLE IF NOT EXISTS run_boosts (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    game_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    consumed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_run_boosts_active
    ON run_boosts (user_id, game_id)
    WHERE consumed_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_run_boosts_active;
DROP TABLE IF EXISTS run_boosts;
