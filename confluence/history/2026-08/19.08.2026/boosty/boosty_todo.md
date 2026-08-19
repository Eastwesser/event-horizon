# Boosty + Event Horizon: actual state and next steps

## 1. What is already implemented in Event Horizon

- `POST /api/payment/checkout` creates an internal payment record in `services/payment`.
- The returned checkout URL is currently just:
  - `BOOSTY_CHECKOUT_URL?payment_id=...&plan=...&amount=...`
- `POST /api/payment/webhook` can confirm a payment and activate a subscription.
- Webhook secret can now be accepted from:
  - JSON `webhook_secret`
  - JSON `webhookSecret`
  - header `X-Webhook-Secret`
  - header `X-Boosty-Webhook-Secret`
  - query `webhook_secret`
- Duplicate webhook deliveries now return `HTTP 200` at gateway level.

## 2. What is NOT implemented yet

- Event Horizon does **not** create a real payment/subscription inside Boosty via API.
- Event Horizon does **not** yet know Boosty’s official signed-webhook format.
- Current checkout is therefore a **redirect stub**, not a full provider integration.

## 3. What this means in practice

Right now there are two different layers:

1. **Boosty page setup**
   - your public page,
   - subscription levels,
   - posts / bundles / showcase,
   - profile text / presentation.

2. **Event Horizon backend integration**
   - internal payment record,
   - redirect to Boosty page,
   - webhook/callback back into EH,
   - activation of subscription in EH database.

At the moment, layer 1 is almost empty, and layer 2 is only partially real.

## 4. Today’s realistic goal

Today we should aim for this:

- make the Boosty page look real and usable,
- define matching payment plans on Boosty,
- verify what callback/webhook options Boosty actually exposes,
- connect the simplest working callback into Event Horizon,
- only then decide whether full signed verification is possible.

## 5. Concrete step-by-step plan

### Step A. Fill Boosty page

- Add avatar / branding for Event Horizon
- Fill “About” block
- Create at least 2 subscription levels:
  - `Present` — 200 RUB
  - `Future` — 300 RUB
- Add at least 1 visible post or showcase item

### Step B. Align Boosty plans with EH

EH currently supports only:

- `present` = 200 RUB
- `future` = 300 RUB

So Boosty should use the same naming / pricing as closely as possible.

### Step C. Check what Boosty can send back

Inside Boosty dashboard/settings, verify whether it supports:

- webhook URL,
- callback URL,
- success / return URL,
- custom metadata / external id / comment / tag,
- secret token,
- signature headers.

This is the key missing piece.

### Step D. Map Boosty → EH

Best case:
- Boosty sends a webhook/callback containing enough data to identify the EH payment:
  - `payment_id` directly, or
  - some reference we can map to EH record.

If Boosty does not support that, we may need:
- success redirect back to frontend,
- then manual confirmation flow,
- or a redesigned integration approach.

### Step E. Test with EH webhook

Once we know Boosty’s actual callback shape:

- set `PAYMENT_WEBHOOK_SECRET`,
- configure Boosty callback/webhook to `POST /api/payment/webhook`,
- send test request,
- confirm subscription becomes active,
- repeat same request to confirm idempotent `HTTP 200`.

## 6. Exact blocker right now

We do not yet know:

- whether Boosty supports webhooks for this page/setup,
- what payload it sends,
- whether it can pass custom metadata like EH `payment_id`,
- whether it signs callbacks.

Today’s finding: for the creator UI/settings you can access normally, there is no obvious/public option for:
- `webhook URL` / `callback URL`
- secret/token for payment callbacks
- signed payment notifications

So “native autosubscription” (automatic activation right after payment) is likely **not achievable** via Boosty in this setup.

Until we have either (a) official Boosty callback/webhook docs for your page type, or (b) an agreed unofficial polling/API approach, we should keep Boosty as a **redirect/helper** and rely on our current EH activation path.

## 7. What Emma should do in Boosty UI now

Open Boosty and collect screenshots or text for:

- subscription level creation page,
- payment/settings/integration page,
- webhook/callback settings page,
- success/fail redirect settings,
- whether custom fields / external id / secret are available.

Then we can map those fields directly to EH.

## 8. After Boosty

Planned next focus after server deep-dive:

- understand and use all backend services normally,
- later do per-service coverage work toward `>=70%`,
- then frontend redesign / beautification as the final pass.