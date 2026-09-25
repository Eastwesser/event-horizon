# Success: false audit (2026-09-20)

## Pattern

Handlers that return `{success: false, …}, nil` (nil gRPC error) force every caller to check `GetSuccess()`. Missing that check = silent failure (shop tickets incident).

## Money path (critical)

| Location | Status |
|----------|--------|
| `shop` → `billing.SpendCurrency` | **Fixed** — checks `GetSuccess()` + short refs |
| `billing.SpendCurrency` handler | **Fixed** — returns gRPC `FailedPrecondition` / `Internal` |
| `billing.AddCurrency` handler | **Fixed** — returns gRPC `Internal` on error |
| `billing` NATS reward consumer | OK — checks `err` from service layer |

## Same anti-pattern still present (non-money / UX)

These return `Success: false` with **nil** gRPC error. Gateway usually forwards the bool to HTTP JSON; frontend must check. No inventory/currency grant on these paths.

| Service | RPC | Risk |
|---------|-----|------|
| `auth` | `Register`, `Logout`, `UpdateNickname`, `UpdateRole` | Client may show success if it ignores `success` |
| `profile` | `UpdateProfile` | Same |
| `leaderboard` | `UpdateScore` | Score may appear saved if client ignores `success` |
| `game` | domain `SubmitScoreResponse.Success` | Handler **does** branch on `resp.Success` — OK |

## Recommendation

Prefer gRPC status errors for failures (InvalidArgument / FailedPrecondition / Internal). Keep `success: true` only on happy-path responses, or require both: status OK **and** `success`.

### Hardened in this pass (same shape as billing)

- `leaderboard.UpdateScore` — gRPC Internal on error  
- `profile.UpdateProfile` — gRPC Internal on error  
- `auth.Register` / `Logout` / `UpdateNickname` / `UpdateRole` — gRPC errors on failure  

Redeploy those services when network/build allows. Frontend UI Priority 2 does not depend on that redeploy.
