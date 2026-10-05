-- +goose Up
CREATE TABLE IF NOT EXISTS notifications (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL,
    title         TEXT NOT NULL,
    body          TEXT NOT NULL,
    link          TEXT NOT NULL DEFAULT '',
    source_event  TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_user_source
    ON notifications(user_id, source_event)
    WHERE source_event <> '';

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
    ON notifications(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id)
    WHERE read_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_notifications_user_created;
DROP INDEX IF EXISTS idx_notifications_user_source;
DROP TABLE IF EXISTS notifications;
