# TODO FINAL LIST — 14.08.2026

Prioritized from the audit. **Excluded** (not needed now): squirrel, swaggo, zap migration, Redis on game/history/analytics, whole-service coverage ≥70%, Envoy, CQRS rewrites, Qdrant MCP upgrade.

---

## P0 — Do next (small, high value)

| # | Task | Why |
|---|------|-----|
| 1 | OpenAPI: add `POST /api/auth/refresh`, `GET /api/auth/whoami`, `POST /api/auth/logout`, `POST /api/auth/update-role` | Routes sync gap; 15–30 min |
| 2 | Bump version label README / badge **1.0.6 → 1.0.7** (or keep 1.0.6 explicitly with changelog note) | Docs match stack |
| 3 | Rebuild + push **gateway** + **inventory** (CB + proto version) when awake | Runtime matches hardening |

---

## P1 — Product gaps

| # | Task | Why |
|---|------|-----|
| 4 | FE **History** page (`/history`) wired to `GET /api/history` | Last FE↔BE lag |
| 5 | Integration tests: **≥1** each for payment, authors, history, analytics | Audit Q2 bar |

---

## P2 — Nice hardening (when free)

| # | Task | Why |
|---|------|-----|
| 6 | Unit smoke for balancer / fulfillment / notification (even config/health) | Zero tests today |
| 7 | Shop: optional outbox for `shop.purchased` (today direct NATS publish) | Align with inventory/billing pattern |
| 8 | Replace `log.Printf` in outbox workers / gateway WS with `slog` | Logging consistency |

---

## Explicitly deferred / skip

| Item | Reason |
|------|--------|
| Squirrel everywhere | Raw SQL/pgx is fine |
| swaggo annotations | Hand OpenAPI + `/docs` by design |
| zap instead of slog | slog is the standard |
| Redis on game/history/analytics | Intentional absence |
| Full `internal/` ≥70% coverage | Critical packages OK; needs interface refactors |
| MCP fine-tune / Qdrant | Binary works; optional later |
| Live Boosty signature verification | Stub webhook OK for now |

---

## Suggested order when back

1 → 2 → 3 (ship) → 4 → 5 → then P2 as appetite allows.

Sleep well. 💤
