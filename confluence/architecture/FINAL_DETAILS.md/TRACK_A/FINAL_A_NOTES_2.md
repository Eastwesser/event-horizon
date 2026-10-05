Track A #2 accepted. Smoke 3/3 pass. Committing now.

Next: Track A #3 — JWT role refresh after author approve.

Plan first, no code.

Problem:
  After admin approves an application and Auth.UpdateRole runs,
  the applicant's existing access JWT still has role=user.
  They see the change only after re-login.
  Worse: if the access token is still valid, calls that
  require role=author (e.g. /author/dashboard, inventory
  create) return 403.

Options to investigate:
  A) On approve, force-revoke the user's access/refresh tokens
     in Auth (Redis session invalidation), so the next request
     hits refresh → new token with role=author.
     - Then the FE needs to handle the 401 → refresh → retry
       cleanly. Does it today?
  B) Make Auth refresh re-read role from DB (currently it
     probably copies role from the old token). Then any
     token refresh picks up the new role.
     - Simpler if refresh already goes to DB.
  C) Frontend polls /auth/whoami after approve and caches
     the new role. No token change; only cosmetic.

Recommended: A + B together (revoke + DB role on refresh).
C is cosmetic and doesn't fix the 403.

Questions before code:
  Q1. Where does Auth.RefreshToken build the new JWT — from
      the old token's claims, or from a fresh DB read of the
      user?
  Q2. Is there a revocation list / Redis session store that
      the admin approve flow can write to?
  Q3. Does FE handle 401-on-refresh gracefully today, or
      does the user just get logged out?

Show the plan with answers. No code until I OK.

Rules:
  - Auth + gateway + FE only.
  - Do NOT touch refund, notification, approvals, games,
    hero, disk, seed.
  - tsc + go test clean.
  - Smoke:
      - Login as user, apply, approve, then WITHOUT re-login
        try to hit /author/dashboard → should work after
        refresh.
      - Double-check that user-only endpoints still 403 for
        role=user.
  - No push until I OK.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (main, 9ac8158):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A #1 — Refund window.

Локально (коммит сейчас):

Track A #2 — Author notifications inbox.

Track A — в работе:

✅ Item 1 — Refund window.

✅ Item 2 — Author notifications.

🟧 Item 3 — JWT role refresh (plan → approve).

Track B (после A):

🟧 3 новые игры, лампочки, уровни.

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.