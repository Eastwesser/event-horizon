# Incident: Shop purchases without ticket deduction

**Date discovered:** 2026-09-20  
**Severity:** Critical (economy)  
**Environment:** cluster / docker-compose (dev). Prod impact TBD — confirm if same build shipped.

## Summary

Shop `PurchaseItem` granted inventory even when billing `SpendCurrency` failed. Buyers received items without ticket deduction.

## Root cause (two bugs)

1. **Overflowing `reference_id`**  
   Shop generated  
   `shop-spend-{userUUID}-{itemUUID}-{unixNano}` (~104 chars).  
   `billing.transactions.reference_id` was `varchar(100)`.  
   Postgres rejected the insert (`SQLSTATE 22001`).

2. **Silent failure**  
   Billing `SpendCurrency` returned `{success: false, message: "..."}` with a **nil gRPC error**.  
   Shop only checked `err != nil`, ignored `success`, and continued to `PurchaseItemWithStock`.

## Timeline (this cluster DB)

| Metric | Value |
|--------|-------|
| Completed purchases | 22 (3 users) |
| Successful `shop_purchase` spends | 3 |
| Ticket value of purchases | ~24 949 |
| Ticket value actually spent | ~300 (pre-fix) |

Notable gap: user `7fc8a659-…` — 9 purchases / 1150 tickets on 2026-07-13 with **0** matching spends.  
Admin seed user `1502a3fa-…` — 12 purchases on 2026-09-20 before the column widen; only later verification spends counted.

Long `shop-spend-…` format introduced with purchase-path hardening (`a34a38a` / present in `aef4a33` shop image from 2026-08-30).

## Fix

1. Live: `ALTER TABLE transactions ALTER COLUMN reference_id TYPE varchar(200);`  
2. Migration: `services/billing/migrations/20260920170000_widen_reference_id.sql`  
3. Shop: shorter refs (`shop-{item8}-{nano}`) + require `spendResp.GetSuccess()`  
4. Billing: return real gRPC errors on Spend/Add failure (not `success:false` + nil err)  
5. Frontend: invalidate balance cache after purchase  

## Decision (recovery)

**No automatic clawback** in this wave — leave historical free grants as-is unless product asks otherwise. Going forward, purchases must deduct.

## Lesson

Any money-moving RPC that uses a `success` bool **must** be checked by callers; prefer gRPC status errors for failures so they cannot be ignored by accident.

## Follow-up

- [x] Widen column + migration  
- [x] Shop Success check + short refs (source)  
- [x] Billing returns gRPC errors (source)  
- [x] Confirm shop/billing images redeployed with the source fix (2026-09-20 ~21:20 local; verified `shop-{item8}-{nano}` + `newBalance` deduct)  
- [x] Scan other callers of `AddCurrency` / `SpendCurrency` for ignored `Success` — only shop was a silent-fail economy path; game SubmitScore uses domain `Success` correctly; billing NATS consumer checks `err`  
- [ ] If prod ran the Aug–Sep shop image: audit purchases vs spends and decide clawback  
