# Security baseline note (v1.1.0 polish)

Quick posture check — not a full audit. Ticklist §3 stays open for deeper CSRF/XSS/SQL pass.

## Present

- JWT sessions in Redis; roles `user` | `author` | `admin`
- bcrypt cost 12; secrets via env (`internal/config`), `.env` gitignored
- gRPC `Validate()` interceptor on services; gateway HTTP→gRPC
- Rate limiter surfaces exist (gateway/balancer) — verify coverage under load
- Notifications soft-fail when logged out (no 401 spam loop)
- FE axios `baseURL: '/api'` (no double `/api/api` after profile fix)

## Watch / next

- CSRF on cookie-less JWT SPA is lower risk; still audit admin mutation forms
- XSS: prefer text nodes / React escaping; avoid `dangerouslySetInnerHTML` (spot-check remaining)
- SQL: repositories should use parameterized queries only (goose + pgx)
- AuthZ: re-check admin / author / inventory write routes after author seed
- DDoS: document limiter thresholds after `make test-k6` on live stack

**Verdict for now:** baseline OK for local/demo; not a claim of production hardening.
