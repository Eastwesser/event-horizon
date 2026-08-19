# TODO FINAL LIST — 14.08.2026 (updated 19.08.2026)

Prioritized from the audit. **Excluded**: squirrel, swaggo, zap, Redis on game/history/analytics, whole-service ≥70%, Envoy, CQRS rewrites, Qdrant MCP.

---

## P0 — Do next (small, high value)

| # | Task | Status |
|---|------|--------|
| 1 | OpenAPI: `POST /api/auth/refresh`, `GET /api/auth/whoami`, `POST /api/auth/logout`, `POST /api/auth/update-role` | ✅ 19.08 |
| 2 | Bump version label README / badge **1.0.6 → 1.0.7** | ✅ 19.08 |
| 3 | Rebuild + push **gateway** + **inventory** (CB + proto version) | ✅ 19.08 pushed + nats-hub |
| 3b | `make deploy` / nats-hub: `./cmd/main.go`, go.work + mcp `go 1.25.7` | ✅ 19.08 |
| 3c | Kafka host port clash with Game metrics `:9092` → Kafka **19092:9092** | ✅ 19.08 |

---

## P1 — Product gaps

| # | Task | Status |
|---|------|--------|
| 4 | FE **History** page (`/history`) → `GET /api/history` | ✅ 19.08 |
| 5 | Integration tests: **≥1** each for payment, authors, history, analytics | ✅ 19.08 `make test-integration` |

---

## P2 — Nice hardening (when free)

| # | Task | Status |
|---|------|--------|
| 6 | Unit smoke for balancer / fulfillment / notification | ✅ 19.08 |
| 7 | Shop: optional outbox for `shop.purchased` | ✅ 19.08 |
| 8 | Replace `log.Printf` in outbox workers / gateway WS with `slog` | ✅ 19.08 |

---

**Counted work is complete.** P0 / P1 / P2 are the audit list. There is no P3.

## Explicitly deferred / skip (not a backlog to grind — we chose not to)

| Item | Reason |
|------|--------|
| Squirrel everywhere | Raw SQL/pgx is fine |
| swaggo annotations | Hand OpenAPI + `/docs` by design |
| zap instead of slog | slog is the standard |
| Redis on game/history/analytics | Intentional absence |
| Full `internal/` ≥70% coverage | Critical packages OK |
| MCP fine-tune / Qdrant | Binary works |
| Live Boosty signature verification | Stub webhook OK |

---

## Next when back

Nothing required from this list. Optional later (only if a real need appears): live Boosty signatures, extra coverage, MCP/Qdrant polish.
