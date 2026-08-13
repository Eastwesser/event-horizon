-- +goose Up
CREATE TABLE IF NOT EXISTS authors (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    bio           TEXT NOT NULL DEFAULT '',
    avatar_url    TEXT NOT NULL DEFAULT '',
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_authors_active ON authors(active);

CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    processed     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_authors_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_authors_outbox_processed;
DROP TABLE IF EXISTS outbox;
DROP INDEX IF EXISTS idx_authors_active;
DROP TABLE IF EXISTS authors;
