# Security baseline note (v1.1.0 polish)

Quick posture + checklist pass 09.10.2026. Full narrative (RU, numbers): `SECURITY_AUDIT_RU.md`.

## Present

- JWT sessions in Redis; roles `user` | `author` | `admin`
- bcrypt cost 12; secrets via env (`internal/config`), `.env` gitignored
- gRPC `Validate()` interceptor on services; gateway HTTP→gRPC
- Rate limiter surfaces (login 5/s IP, submit 10/s, global ~100/s, WS 100/min IP)
- Notifications soft-fail when logged out (no 401 spam loop)
- FE axios `baseURL: '/api/v1'` (no double `/api/api`; legacy `/api/*` rewritten on gateway)

## Checklist pass (09.10)

| Area | Result |
|------|--------|
| CSRF | OK — Bearer JWT, not cookie session |
| XSS | OK — no `dangerouslySetInnerHTML` in FE |
| SQL | OK — parameterized `$n` in repos (spot-check) |
| AuthZ | OK baseline — admin/author writes behind `RequireRole` |
| DDoS | Documented — login blast fails by design |

## Observed under load (08.10.2026)

- CORE `browse.js` (~20 VU) is the sanity suite (`make test-k6`).
- Legacy ~500 VU login blast → ~99% auth failures: **bcrypt cost 12 + ~100 rps login rate limit**.
- Concurrently, shop / inventory / submit / leaderboard stayed green → read path OK; not an auth “outage” bug.
- Do not use `deployments/k6/legacy/loadtest.js` for pass/fail CORE reports.

**Verdict:** baseline + checklist OK for local/demo / v1.1.0; not a claim of production pen-test hardening.
