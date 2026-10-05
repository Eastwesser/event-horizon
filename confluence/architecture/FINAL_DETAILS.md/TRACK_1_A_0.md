Track A item 1 (refund window) with a plan only — reading the design and current cancel/purchase flow.

Track A, item 1 — Refund window (plan only, no code). Waiting for your OK.

Goal
Enforce 7-day cancel window + fulfilled guard for physical merch, per REFUND_WINDOW_DESIGN.md. Today cancel is always allowed for any COMPLETED purchase.

1. Migration (shop DB)
New goose file under services/shop/migrations/:

ALTER TABLE purchases
  ADD COLUMN IF NOT EXISTS refundable_until TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS fulfilled_at TIMESTAMPTZ;
UPDATE purchases
SET refundable_until = COALESCE(completed_at, purchased_at) + INTERVAL '7 days'
WHERE refundable_until IS NULL;
ALTER TABLE purchases
  ALTER COLUMN refundable_until SET NOT NULL;
-- (or leave nullable + always set on INSERT; same effect)
On PurchaseItemWithStock / purchase insert: set refundable_until = now() + 7 days (or COALESCE(completed_at, purchased_at) + 7d).

2. BE — RefundPurchase / CancelPurchase
In postgres_repo.go RefundPurchase (before marking REFUNDED), load refundable_until, fulfilled_at, and item category (join items):

Check	Reject
now() > refundable_until
ErrRefundWindowExpired → gRPC FailedPrecondition
Physical and fulfilled_at IS NOT NULL
ErrAlreadyFulfilled → FailedPrecondition
Already refunded
unchanged (idempotent OK)
IsPhysicalItem(category) as in design (merch / мерч / … vs game_skin / карточка / …). Unknown → physical (safer).

Wire errors in mapShopErr. Gateway already maps FailedPrecondition → HTTP 400.

PurchaseFulfilled → shop: subscribe shop (or small worker) to purchase.fulfilled / PurchaseFulfilled (purchase_uuid) and:

UPDATE purchases SET fulfilled_at = NOW() WHERE id = $1 AND fulfilled_at IS NULL.

3. DTO / proto / gateway
Expose on inventory/owned item responses (extend GetInventory / shop inventory DTO):

refundable_until (ISO)
fulfilled_at (ISO, optional)
can_cancel (computed: not refunded ∧ within window ∧ !(physical ∧ fulfilled))
OpenAPI + FE types accordingly.

4. FE gating
Shop.tsx / ShopItemDetail.tsx:

Show Отменить only when can_cancel === true
Else: hide or disable + tooltip (window expired / already shipped)
On API 400: surface RU messages from body
5. Error messages (RU)
Case	Message
Past window
Срок возврата истёк (7 дней с покупки).
Fulfilled merch
Товар уже отправлен — возврат недоступен.
Already refunded
keep current
Not found
Покупка не найдена.
6. Edge cases (v1)
No partial refunds · no admin override · ticket-shop only
Digital: allow inside 7d even if fulfilled_at set
Physical: block after fulfill
Window from COALESCE(completed_at, purchased_at)
7. Tests
Unit: window expired / fulfilled physical / digital after fulfill / idempotent refund
Optional: repo integration with migration columns
Out of scope (later Track A items)
Author notifications · JWT role refresh · C4 · Track B/C

