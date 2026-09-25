-- +goose Up
-- Shop purchase reference_ids are ~104 chars (shop-spend-{user}-{item}-{nano}).
-- Old varchar(100) caused SpendCurrency to fail silently while inventory still granted.
ALTER TABLE transactions ALTER COLUMN reference_id TYPE varchar(200);

-- +goose Down
ALTER TABLE transactions ALTER COLUMN reference_id TYPE varchar(100);
