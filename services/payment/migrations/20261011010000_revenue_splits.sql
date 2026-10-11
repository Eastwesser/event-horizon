-- +goose Up
-- Commercial rule: author pool 85%, platform 15% of a completed subscription payment.
-- This is a ledger row, not a bank transfer.
CREATE TABLE IF NOT EXISTS revenue_splits (
    payment_id          UUID PRIMARY KEY REFERENCES payments(id),
    amount_rub          INT NOT NULL,
    author_share_rub    INT NOT NULL,
    platform_share_rub  INT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS revenue_splits;
