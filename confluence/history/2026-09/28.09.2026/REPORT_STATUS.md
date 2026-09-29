# September 2026 — what actually shipped (status report)

> Branch under review: `feat/deferred-wave-finish`  
> Source of truth: git merges + code, not every TODO checkbox (TODOs duplicate a lot).

---

## Pre-commit upload checklist (2026-09-28)

| # | Check | Status | Notes |
|---|--------|--------|--------|
| 1 | `.gitignore` covers upload dirs | **OK** | `data/uploads/`, `**/data/uploads/`, `services/gateway/data/`. Compose uses named volume `gateway_uploads` (never a bind mount into the repo). |
| 2 | Uploads volume documented in compose | **OK** | Comments on gateway env + `volumes.gateway_uploads`. All three gateway replicas share `gateway_uploads:/data/uploads` + `UPLOAD_DIR=/data/uploads`. |
| 3 | `POST /api/uploads` validates type/size | **OK** | Auth + author/admin role; `MaxBytesReader` + size ≤ `UPLOAD_MAX_BYTES` (default 2 MiB); ext allowlist `.jpg/.jpeg/.png/.webp`; `http.DetectContentType` must be `image/*`; stored as UUID+ext (`O_EXCL`). |

Residual (acceptable for kids catalog, not blockers): sniff is not a full magic-byte audit; GET `/uploads/*` is public (needed for `<img src>`).

---

## Answer to TODO_2 (28.09) — map cleanup

### Do now (deferred wave)
| Item | Reality |
|------|---------|
| gin.H DTO (inventory) | **Done on branch** — `services/gateway/internal/dto` |
| Navbar extraction | **Done on branch** — `AppNavbar` / `AppFooter` |
| File upload | **Done on branch** — gateway + FE picker |
| DNS/rebuild infra | **Done on branch** — `dns-check.sh`, GOMODCACHE in rebuild scripts, redeploy.md |
| Twin nil→[] | **Already in main** (merge `fix/gateway-nil-slices`) — not part of this wave’s diff |
| Commit + PR | **Pending** — awaiting your go-ahead after smoke |

### Marked “deferred” in TODO_2 but already in main
| Item | Reality |
|------|---------|
| 2d Analytics in admin | **Done** — `AdminAnalytics`, merge `feat/admin-analytics-stats-v2` |
| 2c v2 (top-5, total stock, author emails) | **Done** — `AdminInventoryStats` + inventory stats API |
| Empty state by filter | **Done** — «В категории пусто» in `InventoryList` |

### Still truly deferred
| Item | Notes |
|------|--------|
| Burger separators | UX polish when menu grows |
| Full gin.H DTO for twin list endpoints | Only nil→[] today; same class as inventory DTO, optional hardening |
| Docker Hub push on this host | VPN blocks UDP to `8.8.8.8`; **local compose does not need push** |

### Rebuild / recreate (local, no Hub)
```bash
bash scripts/rebuild-services.sh gateway
docker compose --env-file .env \
  -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate gateway gateway-2 gateway-3 balancer
```

Smoke (from TODO_2): file upload create · stock 0 · price 0 · old URL images · navbar/burger+admin.

Suggested commit message when you approve:
`feat(deferred): gin.H DTOs, navbar extraction, uploads, dns/rebuild`

---

## What was actually done — September waves (confluence/history/2026-09)

### Wave A — Design system + page migration (~19–22.09)
- Tailwind tokens, UI primitives (`Button`, `Card`, `PageShell`, …), brand assets
- Pages migrated off ad-hoc CSS (Home, Shop, Inventory, Profile, Games, Admin shell, …)
- Nav / burger / role gating iterated through VISUAL_BUGS_* and early TODOs

**Git (main):** `b633147` design system · `e6d7900` page migration · `67267ed` admin users/roles + inventory stats v1

### Wave B — Proto3 / null-safety + shop incident (~20–24.09)
- FE coerce for stock/price/empty lists; edit modal prefill
- Shop `reference_id` widen + purchase harden (`de40842`)
- Gateway empty-list twins for inventory/shop, then remaining five list endpoints
- Rebuild path: host binary → `Dockerfile.*.bin` → local tag (Hub push often blocked by DNS)

**Git:** `99d5f71` coerce slices · merge `4329068` `fix/gateway-nil-slices`

### Wave C — Admin analytics + inventory stats v2 (~24–25.09)
- Admin tab: DAU / MAU / retention
- Stats v2: `total_stock`, `top_expensive`, `author_emails`
- Empty-list harden on FE

**Git:** `39765f9` · merge `dbe477d` `feat/admin-analytics-stats-v2`

### Wave D — Deferred finish (this branch, ~25–28.09)
Uncommitted / unpushed on `feat/deferred-wave-finish` until you commit:
1. Inventory **gin.H DTO** (zeros + `images:[]` survive JSON)
2. **AppNavbar / AppFooter** extract from Home
3. **Scoped uploads** (`POST /api/uploads`, static `/uploads`, shared volume, FE file picker + URL still works)
4. **DNS/rebuild**: `scripts/dns-check.sh`, GOMODCACHE in rebuild scripts, clearer local-redeploy docs

### Ops note (28–29.09)
`docker push` fails with `write udp …→8.8.8.8:53: operation not permitted` under Amnezia (`amn0`) while `/etc/resolv.conf` is only `8.8.8.8`. Fix resolv/stub or skip Hub — **recreate from local tags**.

---

## Honest “duplicate TODOs” map

Many 22–25.09 TODO files re-list the same backlog. Collapse to:

| Closed | Open |
|--------|------|
| Design system + page migrate | Burger separators |
| Proto3 FE coerce + gateway nil→[] twins | Optional full DTO on twin endpoints |
| Admin users/roles | (none blocking) |
| Admin analytics (2d) | |
| Inventory stats v2 (2c) | |
| Image URL field + filter empty state | |
| Deferred wave code (DTO/nav/upload/dns) — **on branch** | Commit / PR / merge |

---

## Recommended next step

1. You: smoke the 5 rows in TODO_2 (if not already).  
2. Say **commit** → agent commits on `feat/deferred-wave-finish`.  
3. You: push + PR → main (agent has no GitHub push rights unless you grant).  
4. After merge: pick polish (burger separators) or leave twin full-DTO as debt.
