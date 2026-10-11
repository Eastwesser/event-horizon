-- +goose Up
-- Orchestrator row for a purchase that spans billing and shop.
-- state: started -> spent -> stocked, or spent -> compensating -> compensated.
CREATE TABLE IF NOT EXISTS purchase_sagas (
    id           UUID PRIMARY KEY,
    user_id      TEXT NOT NULL,
    item_id      TEXT NOT NULL,
    purchase_id  UUID NOT NULL,
    spend_ref    TEXT NOT NULL UNIQUE,
    refund_ref   TEXT NOT NULL,
    price        INT NOT NULL,
    state        TEXT NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchase_sagas_state ON purchase_sagas (state, updated_at);

-- +goose Down
DROP TABLE IF EXISTS purchase_sagas;
