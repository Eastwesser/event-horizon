What we would need for YooKassa
Minimal implementation plan:

Provider abstraction in services/payment Instead of hardcoding Boosty redirect logic in CreateCheckout, add a payment-provider interface.

YooKassa config Likely envs such as:

YOOKASSA_SHOP_ID
YOOKASSA_SECRET_KEY
YOOKASSA_RETURN_URL
YOOKASSA_WEBHOOK_SECRET or equivalent, depending on their model
Checkout creation POST /api/payment/checkout should:

create internal pending payment
call YooKassa create-payment API
store provider ref / confirmation URL
return redirect URL to frontend
Webhook handling Add/provider-specific confirm flow:

verify YooKassa notification
map provider payment id back to internal payment_id
activate subscription once
keep idempotency on duplicate deliveries
Admin/manual fallback remains Keep Boosty/manual grant path untouched.

Best first pass
I’d do it in this order:

Add provider-agnostic skeleton
Add stub YooKassa provider
Add unit tests for provider flow
Add manual smoke test for checkout/confirm
Only then connect real YooKassa credentials
Smoke tests we should have
At minimum:

create checkout returns URL + internal payment_id
confirm payment activates subscription
duplicate confirm does not double-activate or fail incorrectly
GET /api/payment/subscription shows active
shop merch gate opens after activation
## Local smoke test for the YooKassa stub (no real credentials)

1. Login to EH (get JWT).
2. Create payment:
   - `POST /api/payment/checkout`
   - header: `Authorization: Bearer <JWT>`
   - body: `{ "plan": "present" }`
   - read `payment_id` from the JSON response.
3. Simulate YooKassa webhook (stub endpoint):
   - `POST /api/payment/yookassa/webhook`
   - headers:
     - `Content-Type: application/json`
     - `X-Webhook-Secret: <PAYMENT_WEBHOOK_SECRET>`
   - body:
     - `{ "event": "payment.succeeded", "object": { "id": "<payment_id>", "status": "succeeded" } }`
4. Check unlock:
   - `GET /api/payment/subscription` (JWT)
   - then try a shop purchase (it should pass `checkMerchAllowed`).
5. Idempotency:
   - repeat the webhook request once; it should still succeed.

Important note
I would not replace or break the current Boosty flow.
Best path is:

keep Boosty as fallback/manual helper
add YooKassa as the real primary provider later
If you want, the next concrete step can be: I patch final_todo.md with this decision and then prepare a small YooKassa integration plan/stub checklist as the next backend item.