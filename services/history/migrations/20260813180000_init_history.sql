-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL DEFAULT '',
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_user_created ON events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_type_created ON events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_events_created;
DROP INDEX IF EXISTS idx_events_type_created;
DROP INDEX IF EXISTS idx_events_user_created;
DROP TABLE IF EXISTS events;
