-- +goose Up
CREATE TABLE IF NOT EXISTS payments (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL,
    plan          TEXT NOT NULL,
    amount_rub    INT  NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    provider      TEXT NOT NULL DEFAULT 'boosty',
    provider_ref  TEXT,
    checkout_url  TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

CREATE TABLE IF NOT EXISTS subscriptions (
    id            UUID PRIMARY KEY,
    user_id       TEXT NOT NULL,
    plan          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    amount_rub    INT  NOT NULL,
    payment_id    UUID REFERENCES payments(id),
    starts_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_user_active
    ON subscriptions(user_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    processed     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payment_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_payment_outbox_processed;
DROP TABLE IF EXISTS outbox;
DROP INDEX IF EXISTS idx_subscriptions_user_active;
DROP TABLE IF EXISTS subscriptions;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_user;
DROP TABLE IF EXISTS payments;
