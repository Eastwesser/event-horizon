# Payment Smoke Test Results (19.08.2026)

## What was tested

- Register a fresh user
- Login and get JWT
- `POST /api/payment/checkout` with `plan=present`
- `POST /api/payment/yookassa/webhook` (stub payload `payment.succeeded`)
- `GET /api/payment/subscription`
- Repeat webhook once (idempotency check)
- `GET /api/payment/can-purchase-merch`

## Results

- Checkout returns `status: "pending"` and valid `payment_id`
- First webhook returns `success: true`, `message: "subscription activated"`
- Subscription check returns:
  - `active: true`
  - `plan: "present"`
  - `status: "active"`
- Second webhook with the same `payment_id` also succeeds and returns the same `subscription_id` (idempotent behavior)
- Merch gate check returns `{"allowed": true, "reason": ""}`

Conclusion: local YooKassa-stub flow works as expected end-to-end.

## Runtime fixes applied before successful run

- Rebuilt/restarted `auth` and `payment` images (`exec /...-service: no such file` was causing startup failures)
- Rebuilt/restarted `gateway` so the new route `/api/payment/yookassa/webhook` is present

## Reusable one-command smoke script

Use:

`confluence/history/2026-08/19.08.2026/yookassa/smoke_yookassa_stub.sh`