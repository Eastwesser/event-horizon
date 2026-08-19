# HTTP API routes — Event Horizon

> Source of truth: `services/gateway/internal/app/gateway.go` + `docs/openapi.yaml`.  
> Companion: [`STATUS_CODES.md`](STATUS_CODES.md) (HTTP/gRPC error mapping).

---

## Prefix convention (v1.0.x)

| Pattern | Example | Notes |
|---------|---------|-------|
| **Public HTTP API** | `/api/<domain>/…` | No `/api/v1/` segment in v1.0.x |
| **WebSocket** | `/ws/…` | Not under `/api` |
| **Ops / docs** | `/health`, `/ready`, `/openapi.yaml`, `/docs` | Gateway liveness, Swagger |

**Frontend (React):** axios `baseURL: '/api'` in `frontend/src/services/api.ts`. Service modules pass paths **without** the `/api` prefix (`/shop/purchase` → `/api/shop/purchase`).

**When we need breaking HTTP changes:** introduce `/api/v2/…` (or mount `/api/v1/` as alias during migration). Do not sprinkle `v1` everywhere preemptively.

---

## Role matrix

| Role | Assigned how | Typical access |
|------|----------------|----------------|
| **user** | Default on register | Shop, games, profile, history (own) |
| **author** | Register with `role: "author"` or admin `update-role` | + inventory CRUD (own items), authors profile |
| **admin** | Only via `POST /api/auth/update-role` (admin caller) | + analytics, inventory stats, any-user role changes |

**Never:** self-assign `admin` at registration (blocked in Auth service).

Gateway enforcement: `RequireAuth` → JWT validated via Auth `ValidateToken` (Redis session). `RequireRole(...)` → 403 if role not allowed.

---

## Route table

Legend: **Auth** = Bearer JWT required. **Roles** = minimum role(s); `-` = any authenticated user.

### Ops & docs

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/health` | — | — | Liveness |
| GET | `/ready` | — | — | Readiness (Redis ping) |
| GET | `/openapi.yaml` | — | — | OpenAPI spec |
| GET | `/docs` | — | — | Swagger UI |
| WS | `/ws/leaderboard` | — | — | Real-time leaderboard push |

### Auth

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| POST | `/api/auth/register` | — | — | Register (`user` or `author` only) |
| POST | `/api/auth/login` | — | — | Login → JWT pair |
| POST | `/api/auth/refresh` | — | — | Refresh tokens |
| GET | `/api/auth/whoami` | ✓ | user+ | Current user from token |
| POST | `/api/auth/logout` | ✓ | user+ | Revoke session |
| POST | `/api/auth/update-role` | ✓ | **admin** | Change user role |
| GET | `/api/auth/user` | ✓ | user+ | User + scores |
| POST | `/api/auth/update-nickname` | ✓ | user+ | Update nickname |

### Profile & billing

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/api/profile` | ✓ | user+ | Aggregated profile |
| GET | `/api/billing/balance/all` | ✓ | user+ | Lamps + tickets balance |

### Shop

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/api/shop/items` | ✓ | user+ | Catalog |
| POST | `/api/shop/purchase` | ✓ | user+ | Buy item (403 `subscription_required` for merch without sub) |
| GET | `/api/shop/inventory` | ✓ | user+ | Owned shop items |

### Payment

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| POST | `/api/payment/checkout` | ✓ | user+ | Create checkout |
| GET | `/api/payment/subscription` | ✓ | user+ | Subscription status |
| GET | `/api/payment/can-purchase-merch` | ✓ | user+ | Merch gate probe (200 + allowed/reason) |
| POST | `/api/payment/webhook` | — | — | Boosty/provider webhook |
| POST | `/api/payment/yookassa/webhook` | — | — | YooKassa stub (local dev) |

### Authors

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| PUT | `/api/authors/me` | ✓ | author, admin | Upsert author profile |
| GET | `/api/authors` | — | — | List authors (public) |
| GET | `/api/authors/:user_id` | — | — | Get author (public) |

### History & analytics

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/api/history` | ✓ | user+ | User event history |
| GET | `/api/analytics/dau` | ✓ | **admin** | DAU |
| GET | `/api/analytics/mau` | ✓ | **admin** | MAU |
| GET | `/api/analytics/retention` | ✓ | **admin** | Retention |

### Inventory (author catalog / stock)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/api/inventory/items` | ✓ | user+ | Search/list |
| POST | `/api/inventory/items` | ✓ | author, admin | Create item |
| POST | `/api/inventory/items/bulk` | ✓ | author, admin | Bulk create |
| GET | `/api/inventory/items/:id` | ✓ | user+ | Get by ID |
| PUT | `/api/inventory/items/:id` | ✓ | author, admin | Update (409 on version conflict) |
| DELETE | `/api/inventory/items/:id` | ✓ | author, admin | Hard delete |
| POST | `/api/inventory/items/:id/reserve` | ✓ | author, admin | Reserve stock |
| DELETE | `/api/inventory/items/:id/soft` | ✓ | author, admin | Soft delete |
| POST | `/api/inventory/items/:id/restore` | ✓ | author, admin | Restore soft-deleted |
| GET | `/api/inventory/stats` | ✓ | **admin** | Inventory stats |

### Game & leaderboard

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| POST | `/api/game/submit` | ✓ | user+ | Submit score (uses authenticated user id) |
| GET | `/api/leaderboard` | — | — | Top scores (public) |

---

## Frontend module map

| Module | Paths (relative to `baseURL /api`) |
|--------|-------------------------------------|
| `services/api.ts` | auth, game, billing, shop, profile, leaderboard |
| `services/authorsApi.ts` | `/authors/…` |
| `services/paymentApi.ts` | `/payment/…` |
| `services/historyApi.ts` | `/history` |
| `services/analyticsApi.ts` | `/analytics/…` |
| `services/inventoryApi.ts` | `/inventory/items/…` |

WebSocket: `ws://${host}/ws/leaderboard` (Vite dev proxy forwards `/ws` → gateway).

---

## Common mistakes

1. **Double `/api` prefix** — `api.get('/api/shop/items')` with `baseURL: '/api'` → `/api/api/shop/items` (404).
2. **Missing JWT on game submit** — use shared `api` client, not raw `fetch`, so `Authorization` is attached.
3. **Confusing shop vs inventory** — `/api/shop/inventory` = purchased shop items; `/api/inventory/items` = author product catalog (different domain).

---

## Related

- OpenAPI: `docs/openapi.yaml` (also served at `GET /openapi.yaml`)
- Status codes: `confluence/architecture/STATUS_CODES.md`
- JWT roles: `user` \| `author` \| `admin` — see `.cursor/rules/event-horizon.mdc`
