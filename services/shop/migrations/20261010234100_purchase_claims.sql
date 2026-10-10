-- +goose Up
-- One open spend cycle per player+item. Retries share the cycle; a refund bumps it
-- so the next purchase gets a new billing reference_id.
CREATE TABLE IF NOT EXISTS purchase_claims (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    cycle   INT  NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, item_id)
);

-- +goose Down
DROP TABLE IF EXISTS purchase_claims;
