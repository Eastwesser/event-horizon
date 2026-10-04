-- +goose Up
CREATE TABLE IF NOT EXISTS author_applications (
    id             UUID PRIMARY KEY,
    user_id        TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'approved', 'rejected')),
    payload        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at    TIMESTAMPTZ,
    reviewed_by    TEXT NOT NULL DEFAULT '',
    reviewer_note  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_author_applications_user_id
    ON author_applications (user_id);

CREATE INDEX IF NOT EXISTS idx_author_applications_status
    ON author_applications (status);

-- At most one pending application per user.
CREATE UNIQUE INDEX IF NOT EXISTS uq_author_applications_pending_user
    ON author_applications (user_id)
    WHERE status = 'pending';

-- +goose Down
DROP INDEX IF EXISTS uq_author_applications_pending_user;
DROP INDEX IF EXISTS idx_author_applications_status;
DROP INDEX IF EXISTS idx_author_applications_user_id;
DROP TABLE IF EXISTS author_applications;
