# Event Horizon — Final ToDo (19.08.2026)

This file aggregates the **current “still to do”** ideas from:
- [`confluence/history/2026-08/14.08.2026/TODO_FINAL_LIST.md`](../14.08.2026/TODO_FINAL_LIST.md)
- [`confluence/history/2026-08/19.08.2026/CONTINUE.md`](CONTINUE.md)
- [`confluence/history/2026-08/19.08.2026/P2_RESULTS.md`](P2_RESULTS.md)
- [`confluence/history/2026-08/19.08.2026/REGULAR_COMMANDS.md`](REGULAR_COMMANDS.md)
- [`confluence/tech_debt/CURRENT_DEBT/*`](../../tech_debt/CURRENT_DEBT/)

Note: as of 19.08, P0–P2 were completed; what remains is “optional later” hardening.

---

## Backend (hardening / optional later)

### 1) Live Boosty signature verification (real verification, not only shared-secret equality)
**Status:** still technically outstanding (official signed verification not implemented yet), but webhook contract is now integration-ready for “connect Boosty today”.

**Research note (19.08):**
- Current implementation is a stub-style webhook confirmation flow, not a documented provider-signed webhook flow.
- Public quick research did not reveal a clear Boosty webhook-signature specification.
- This item may require either:
  - Boosty’s actual webhook docs / sample signed request, or
  - an explicit project decision to keep the current shared-secret flow and mark “official signature verification” as not applicable for the present integration shape.

**What exists now (19.08, for integration testing):**
- `services/gateway` webhook handler now extracts `webhook_secret` from:
  - JSON `webhook_secret` / camelCase `webhookSecret`,
  - headers: `X-Webhook-Secret`, `X-Boosty-Webhook-Secret`,
  - query: `webhook_secret`.
- `services/payment` `ConfirmPayment` is now idempotent: if the same `payment_id` is already completed, it returns the existing active subscription (so merch unlock stays consistent).

**What to implement**
- Add Boosty’s official signature verification:
  - validate signature headers (per Boosty spec),
  - validate timestamp/nonce if required,
  - protect against replay/incorrect payloads.
- Change the current webhook contract so verification can use the real request shape:
  - pass provider signature headers and raw request body from `services/gateway`,
  - extend `services/payment/proto/payment.proto` beyond just `webhook_secret`,
  - update `docs/openapi.yaml` / gateway OpenAPI description to match.

**Where it’s described**
- `confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`
  - (Boosty signature verification is listed as “partial” and “missing HMAC/signature flow”)
- `confluence/history/2026-08/14.08.2026/TODO_FINAL_LIST.md`
  - “Live Boosty signature verification” is deferred (“stub webhook OK”) but mentioned as optional later.

Suggested next step:
- Today (connect Boosty + verify webhook end-to-end):
  - set `PAYMENT_WEBHOOK_SECRET` in your env/compose,
  - configure Boosty webhook URL to `POST /api/payment/webhook`,
  - set the same webhook secret in Boosty,
  - trigger a webhook test delivery and confirm we return `HTTP 200` and subscription becomes `active`.
- Then (only if Boosty provides signed callbacks / signature headers): Research Boosty webhook signature spec and then implement in this order:
  - update Gateway webhook endpoint to forward raw payload + signature metadata,
  - add verifier logic in `services/payment`,
  - add replay protection if Boosty provides timestamp/nonce semantics,
  - add unit tests for valid signature / invalid signature / tampered payload / replay case,
  - update service README and OpenAPI docs.
- If no official Boosty signature spec exists for this integration path, decide whether to:
  - keep the shared-secret webhook as the final design, or
  - redesign the payment integration around a provider flow that does expose verifiable signed callbacks.

---

### 2) ≥70% coverage enforcement (measure + add tests where it drops)
**Status:** not fully achieved / not enforced by CI.

**What to do**
- Run per-service coverage measurement:
  - `task test-coverage` (uses `go test -coverprofile` + prints `go tool cover -func`)
- Identify which packages fall below the target and add unit tests.
- (Optional) add a CI gate so it doesn’t regress.

**Where it’s described**
- `confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`
  - coverage is explicitly called out as “missing coverage gate”
- `confluence/history/2026-08/14.08.2026/TODO_FINAL_LIST.md`
  - “Full internal/whole-service ≥70% coverage” is deferred.

---

### 3) MCP/Qdrant polish (embeddings + vector DB upgrade path)
**Status:** MCP works with offline TF-IDF; Qdrant embeddings upgrade is still missing.

**What to implement**
- Upgrade Stage 2 RAG to:
  - generate embeddings (per your chosen plan),
  - store them in Qdrant,
  - keep MCP tool contract stable (e.g. `search_prydwen` tool name/signature).

**Where it’s described**
- `confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`
  - MCP is “done” with offline TF‑IDF; Qdrant embeddings are “missing”
- `confluence/history/2026-08/14.08.2026/TODO_FINAL_LIST.md`
  - “MCP fine-tune / Qdrant” is deferred but listed as optional later.

---

## Docs / Verification (optional sanity checks)

### 4) Verify HTTP/gRPC status mapping rules
**Status:** optional sanity-check.

**Where it’s described**
- `confluence/history/2026-08/19.08.2026/REGULAR_COMMANDS.md`
  - “I also need to check one more thing- status codes: `confluence/history/2026-08/13.08.2026/STATUS_CODES.md`”

---

## Frontend (sanity checks)

### Verify “History UI lags” from the latest notes
**Status:** optional sanity-check (not a strict requirement list item).

**Where it’s described**
- `confluence/history/2026-08/19.08.2026/REGULAR_COMMANDS.md`
  - “Only History UI lags” (context for earlier FE mismatch notes)

