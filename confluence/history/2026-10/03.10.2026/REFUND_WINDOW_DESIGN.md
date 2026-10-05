# Refund window + fulfilled guard (design)

**Status:** implemented (Track A item 1) — migration + CancelPurchase guards + inventory `can_cancel` + FE gating.  
**Context:** Wave 1 follow-up; abuse case was physical merch received → cancel.

## Current state

| Piece | Today |
|-------|--------|
| Table | `purchases` (`user_id`, `item_id`, `price`, `status`, `purchased_at` / `completed_at`, `refunded_at`) |
| Cancel path | Shop `RefundPurchase` → status `REFUNDED`, restore stock, billing `AddCurrency` |
| Fulfilled | Fulfillment service publishes `PurchaseFulfilled` (NATS/Kafka) — **not persisted** on `purchases` |
| FE | Cancel button always shown for owned inventory / detail |

Abuse case: physical merch received → cancel → tickets back, goods kept.

## Proposed model

### A) 7-day window

1. Migration on `purchases`:
   - `refundable_until TIMESTAMPTZ NOT NULL` (default `purchased_at + interval '7 days'`, backfill from `COALESCE(completed_at, purchased_at) + 7 days`).
2. BE `CancelPurchase` / `RefundPurchase`:
   - If `now() > refundable_until` → reject with gRPC `FailedPrecondition` / HTTP 400.
3. FE:
   - Expose `refundable_until` (ISO) on inventory / purchase DTO.
   - Show cancel only when `Date.now() < refundable_until` (and not already refunded).
   - Disabled + tooltip when expired (optional).

### B) Fulfilled guard

1. Migration: `purchases.fulfilled_at TIMESTAMPTZ NULL`.
2. Consumer (shop or thin fulfillment→shop handler) on `PurchaseFulfilled` sets `fulfilled_at = NOW()` for that purchase UUID.
3. Refund rules:
   - **Physical**: refuse if `fulfilled_at IS NOT NULL`.
   - **Digital**: allow inside 7-day window; ignore `fulfilled_at` (instant delivery).

Prefer `fulfilled_at` column over a new status `FULFILLED` so cancel still uses `COMPLETED` → `REFUNDED` without status soup.

### C) Digital vs physical detection (chosen)

**Primary:** item `type` / `category` string (shop + inventory already use these).

| Kind | Match (case-insensitive) |
|------|--------------------------|
| Digital | `карточка`, `game_skin`, `profile_theme`, `skin`, `theme` |
| Physical | `merch`, `мерч`, `брелок`, `картина`, `фенечка` |

**Override:** if `attributes.physical === true` → treat as physical; if `attributes.physical === false` → digital.

**Justification:** types are already on every catalog/purchase path; no new column. `attributes.physical` covers edge cases without a migration. Unknown types default to **physical** (safer: block after fulfill) unless clearly digital by name above.

Helper (implementation ticket): `func IsPhysicalItem(type, category string, attrs map) bool`.

## Error messages (RU)

| Case | Message |
|------|---------|
| Past window | `Срок возврата истёк (7 дней с покупки).` |
| Fulfilled merch | `Товар уже отправлен — возврат недоступен.` |
| Already refunded | Keep current idempotent OK / «Уже возвращено». |
| Not found | `Покупка не найдена.` |

## Edge cases (v1)

- No partial refunds.
- Idempotent cancel (already `REFUNDED`) unchanged.
- Window starts at `completed_at` if set, else `purchased_at`.
- Admin override: out of scope for v1.
- Subscription / payment-plan purchases: out of scope (shop ticket cancel only).

## Implementation sketch (later ticket)

1. Shop migration + repo checks in `RefundPurchase`.
2. Proto/DTO: `refundable_until`, `fulfilled_at`, `can_cancel` bool (computed).
3. Gateway pass-through.
4. FE: hide/disable cancel from `can_cancel`.
5. Wire `PurchaseFulfilled` → `UPDATE purchases SET fulfilled_at`.

**Do not block Wave 2** content work on this; schedule as a short shop/security ticket after Wave 1 chrome follow-ups.
