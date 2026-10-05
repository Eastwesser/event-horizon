# Unit test coverage (Wave 4)

**Gate:** `MIN_COVERAGE=70 bash scripts/coverage-gate.sh`  
**Target:** ≥70% on `internal/service` (business logic).  
**CI:** `coverage` job in `.github/workflows/main.yml` (blocks `build` / image push).  
**Gateway:** excluded (thin HTTP↔gRPC adapters).

## Status (2026-10-05, `go test ./internal/service/`)

| Service | Service % | Status |
|---------|----------:|--------|
| authors | **93.2%** | ✅ |
| shop | **83.9%** | ✅ |
| inventory | **98.5%** | ✅ |
| auth | **86.7%** | ✅ |
| billing | **85.7%** | ✅ |
| payment | **91.1%** | ✅ |
| profile | **78.8%** | ✅ |
| game | **77.6%** | ✅ |
| leaderboard | **100%** | ✅ |
| history | **100%** | ✅ |
| analytics | **100%** | ✅ |
| gateway | — | excluded |

## Progress log

| Date | Service | Before → After (service) | Notes |
|------|---------|--------------------------|-------|
| 2026-10-05 | — | baseline | plan C4_TODO_3 |
| 2026-10-05 | authors | 0% → **93.2%** | AuthorStore/AuthorCache ports |
| 2026-10-05 | shop | ~2% → **83.9%** | dial/NATS → app; merch_gate assert fix |
| 2026-10-05 | inventory | 7% → **98.5%** | include_deleted + soft-delete/restore |
| 2026-10-05 | auth | 65% → **86.7%** | miniredis sessions |
| 2026-10-05 | leaderboard | 0% → **100%** | hot-read clamp |
| 2026-10-05 | history | 0% → **100%** | HistoryStore port |
| 2026-10-05 | analytics | 0% → **100%** | AnalyticsStore port |
| 2026-10-05 | billing | ~partial → **85.7%** | BillingStore/Cache; lamps+tickets |
| 2026-10-05 | payment | 0% → **91.1%** | CanPurchaseMerch + confirm idempotent |
| 2026-10-05 | profile | evaluate-only → **78.8%** | Get/Update/Unlock |
| 2026-10-05 | game | 0% → **77.6%** | SubmitScore + StartBoost |
| 2026-10-05 | k6 | browse+purchase ready | smoke needs live stack |
| 2026-10-05 | gate | package → **service layer** | wired in CI |

## Notes

- Package-wide totals stay low (handlers/repos/workers) — intentional; gate is service-only.  
- Indexes / k6 load reports: separate PR after `make deploy` + EXPLAIN.
