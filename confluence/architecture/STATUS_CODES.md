# HTTP / gRPC status codes — Event Horizon

> Companion to Prydwen `9.common_backend/02_STATUS_CODES.md`.  
> Focus: **what EH actually returns today** (Gateway Gin + gRPC services).

---

## 1. Quick map (HTTP via Gateway)

| Code | Meaning | EH where |
|------|---------|----------|
| **200** | OK | Almost all successful GETs/POSTs that return a body |
| **201** | Created | Prefer for create if we tighten OpenAPI later; many creates still use 200 |
| **400** | Bad request | Failed bind/validate, empty required fields, gRPC `InvalidArgument` |
| **401** | Unauthenticated | Missing/invalid JWT (`RequireAuth`), bad login/refresh |
| **403** | Forbidden | Authenticated but wrong role (`RequireRole`), inventory gRPC role interceptor |
| **404** | Not found | Author/item/user missing (`NotFound`) |
| **409** | Conflict | Optimistic lock / duplicate — **target**; inventory has `ErrVersionConflict` (map to 409 in gateway when wiring) |
| **429** | Rate limited | Gateway rate limiter (per user / IP) |
| **500** | Internal | Unexpected gRPC/`Internal`, unmapped errors |
| **503** | Unavailable | Circuit breaker **open**, `/ready` degraded (dependency down) |

---

## 2. Gateway middleware contract (source of truth)

| Middleware | HTTP | Body hint |
|------------|------|-----------|
| `RequireAuth` — no/invalid token | **401** | `missing authorization` / `invalid or expired token` / `auth service unavailable` |
| `RequireRole` — role not allowed | **403** | `insufficient permissions` |
| Circuit breaker open | **503** | `service temporarily unavailable` + `circuit` name |
| Rate limit | **429** | limiter response |
| Liveness `/health` | **200** | `{status: ok}` |
| Readiness `/ready` (Redis down) | **503** | `{status: degraded}` |

Rules of thumb for interview:

- **401** = we do not know who you are (or session revoked in Redis).
- **403** = we know who you are; policy says no (role).
- Never return 401 for “wrong role” — client will uselessly re-login.

---

## 3. gRPC → HTTP mapping (intended)

Gateway today routes RPC errors through `writeGRPCError()` / `handleRPCError()` on all gRPC-backed routes. Payment webhooks use `writePaymentConfirmError()` (duplicate delivery → **200**).

| gRPC `codes.*` | HTTP |
|----------------|------|
| `OK` | 200 / 201 |
| `InvalidArgument` | 400 |
| `Unauthenticated` | 401 |
| `PermissionDenied` | 403 |
| `NotFound` | 404 |
| `AlreadyExists` / `Aborted` (version conflict) | 409 |
| `ResourceExhausted` | 429 |
| `Unavailable` / circuit | 503 |
| `Internal` / unknown | 500 |

Auth examples already in code: login fail → `Unauthenticated`; bad email → `InvalidArgument`; missing user → `NotFound`.

Inventory mutating RPCs without `x-user-role` metadata → `PermissionDenied` (403 at edge if gateway maps it).

---

## 4. Domain-specific cases

### Auth
- Register validation fail → 400 / `InvalidArgument`
- Wrong password → 401 / `Unauthenticated` (do not leak “email exists” vs “bad password” in public messages if tightening security)
- Refresh revoked/expired → 401
- `UpdateRole` without admin → 403 at Gateway

### Shop / Payment
- Merch without subscription → **403** + `{ error: "subscription_required", code: "subscription_required" }` on `POST /api/shop/purchase`; probe via `GET /api/payment/can-purchase-merch` (200 + `allowed`/`reason`)
- Out of stock / insufficient tickets → **409** (`FailedPrecondition` / `AlreadyExists`)
- Circuit on Billing/Shop/Payment → **503**
- Duplicate `POST /api/payment/webhook` or `/api/payment/yookassa/webhook` (gRPC `AlreadyExists`) → **HTTP 200** (idempotency)

### Inventory
- Soft-deleted / missing item → 404
- `ErrVersionConflict` → **409** (ensure gateway maps this)
- Author-only stock mutations without role → 403

### Analytics
- Admin-only DAU/MAU/retention → 403 for non-admin (Gateway `RequireRole(admin)`)

---

## 5. What not to do

- Mask every failure as 500 “just in case”
- Return 404 for “exists but not yours” when that enables IDOR probing — often **403** or opaque 404 by policy
- Treat open circuit as 500 (hides that fail-fast is intentional)
- Empty list → still **200** with `[]`, not 404

---

## 6. Interview answers (short)

1. **401 vs 403** — identity vs authorization (EH: `RequireAuth` / `RequireRole`).
2. **409** — optimistic `version` or duplicate unique key.
3. **429 vs 503** — client too chatty vs we/dependency cut off (circuit).
4. **Empty collection** — 200 + `[]`.
5. **gRPC mapping** — table above; Gateway is the single HTTP façade.

---

## 7. Follow-ups (tech debt)

- [x] Normalize Gateway error mapping from gRPC status (`handleRPCError` / `writeGRPCError` on RPC routes)
- [x] Map inventory version conflict → 409 consistently (`Aborted` via `mapInventoryErr` + gateway mapping)
- [x] Merch-gate: **403** + domain code `subscription_required` on shop purchase; can-purchase-merch stays 200 probe
- [x] Keep `docs/openapi.yaml` in sync (YooKassa webhook stub, shop/payment status codes)
