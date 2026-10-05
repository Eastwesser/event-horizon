# Unit test coverage (Wave 4)

**Gate:** `MIN_COVERAGE=70 bash scripts/coverage-gate.sh`  
**Target:** ≥70% on `internal/service` first; package total via gate after priority services recover.  
**CI:** do **not** wire the gate into GitHub Actions until auth / inventory / shop / authors are ≥70% (avoid breaking `main`).

## Gateway exclusion

`gateway` is **not** in `coverage-gate.sh`. Handlers are thin HTTP↔gRPC adapters; middleware / circuit / DTO packages already have partial tests. Track gateway coverage opportunistically — not a Wave 4 gate blocker.

## Baseline (2026-10-05, package `go test ./internal/...`)

| Service | Total | Service layer notes | Status |
|---------|------:|---------------------|--------|
| authors | 13.5% pkg / **93.2% service** | Submit/Approve/Reject/Revert/Upsert covered | ✅ service target met |
| shop | 16.5% pkg / **83.9% service** | Purchase/Cancel/ListPurchasesByItemIDs covered | ✅ service target met |
| inventory | 8.8% pkg / **98.5% service** | GetByAuthor/SoftDelete/Restore/Reserve/Release | ✅ service target met |
| auth | — / **86.7% service** | Logout+session, Refresh revoke, ListUsers, role gate | ✅ service target met |
| gateway | 8.3% | excluded from gate | — |
| billing | — | after priority 4 | ⬜ |
| payment | — | after priority 4 | ⬜ |
| profile | — | after priority 4 | ⬜ |
| game | — | after priority 4 | ⬜ |
| leaderboard | — | after priority 4 | ⬜ |
| history | — | after priority 4 | ⬜ |
| analytics | — | after priority 4 | ⬜ |

## Progress log

| Date | Service | Before → After (service / total) | Notes |
|------|---------|----------------------------------|-------|
| 2026-10-05 | — | baseline recorded | plan approved C4_TODO_3 |
| 2026-10-05 | authors | service 0% → **93.2%** / pkg 0% → 13.5% | mock store+cache; thin `AuthorStore`/`AuthorCache` ports |
| 2026-10-05 | shop | service ~2% → **83.9%** / pkg 2% → 16.5% | mock store+billing; dial/NATS sync moved to `app`; fixed stale merch_gate assertions |
| 2026-10-05 | inventory | service 7% → **98.5%** | Search/GetByAuthor/include_deleted + soft-delete/restore/reserve/release |
| 2026-10-05 | auth | service 65% → **86.7%** | miniredis session tests; ListUsers; public role=user only |
| 2026-10-05 | k6 | browse.js + purchase.js ready | smoke deferred — needs live stack (`make deploy`) |

## Order

1. authors → shop → inventory → auth (high-risk)  
2. remaining gate services  
3. extend `coverage-gate.sh` + CI when priority 4 ≥70%
