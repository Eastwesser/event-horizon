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
| shop | 2.0% | Purchase/Cancel/ListPurchasesByItemIDs 0% | 🟧 queued |
| inventory | 7.2% | Search/GetByAuthor/Release 0% | 🟧 queued |
| auth | 22.7% | Logout/Refresh/ListUsers weak | 🟧 queued |
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

## Order

1. authors → shop → inventory → auth (high-risk)  
2. remaining gate services  
3. extend `coverage-gate.sh` + CI when priority 4 ≥70%
