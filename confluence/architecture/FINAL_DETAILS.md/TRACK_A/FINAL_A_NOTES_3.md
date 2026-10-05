OK. Реализуй A + B
Что подтверждаю
Q1: Refresh строит JWT из old claims. Значит B обязателен.

Q2: Redis есть:

auth:session:{jti} — access.

auth:refresh:{jti} + auth:user_refresh:{userID} — refresh.

auth:user:{id} — user cache.
Gap: нет user → access jti index. DeleteAllRefreshForUser есть, access revoke — нужен новый.

Q3: FE — graceful refresh (не logout). Cosmetic gap — role не переписывается после refresh.

Plan (A + B) — правильный:

B: RefreshToken → GetUser → issuePair(..., user.Role).

A: UpdateRole → DeleteAllRefreshForUser + new DeleteAllSessionsForUser.

CreateSession → auth:user_session:{userID} set.

Unit tests: refresh after role change, revoke old access, revoke old refresh.

FE: persist role from refresh response + invalidate whoami cache.

C отвергнут правильно — не фиксит 403.

Один комментарий — про race
Сценарий:

User логинится → access token A1, refresh R1.

Admin approve → UpdateRole revokes A1 + R1.

User делает запрос с A1 → ValidateToken fails (401) → FE делает refresh с R1.

Но R1 тоже отозван → refresh fails → clearAuth → logout.

Значит: user вылетит на login сразу после approve. Не получит новую роль без re-login.

Что нужно доработать:

Либо — не отзывать refresh сразу. Только access. Тогда user сможет refresh → получит author.

Либо — отзывать всё, но FE после logout показать сообщение «Роль изменена, войдите заново».

Либо — отзывать access, refresh оставить с пометкой «перечитать role из DB». Тогда refresh даст author.

Мой выбор — третий (не отзывать refresh, только access). Плюс B (refresh читает role из DB) — даст author сразу.

Промпт Cursor'у:

text
OK — implement A + B with one adjustment on the revoke side.

Race to consider:
  1. User logs in → access A1, refresh R1.
  2. Admin approves → UpdateRole revokes A1 AND R1.
  3. User hits API with A1 → 401 → FE tries refresh with R1
     → R1 is also revoked → refresh fails → forced logout.

So revoking BOTH sides kills the seamless transition.

Adjusted plan:
  - Revoke ONLY access sessions on UpdateRole:
      DeleteAllSessionsForUser(userID) → auth:user_session set
      removed → A1 fails on next ValidateToken.
  - Keep refresh tokens ALIVE.
  - RefreshToken (already reading role from DB per B) then
    returns a NEW access token with role=author.
  - FE 401 → refresh → retry → 200 as author. No logout.

If you disagree (e.g., security wants refresh revoked too),
say so and explain the UX tradeoff. Otherwise implement as
above.

Also:
  - Unit tests: refresh after role change → author; old
    access → 401; old refresh → still valid until used.
  - FE: persist role from refresh response +
    invalidateWhoamiCache().
  - Optional: gateway bust whoami cache for that user (not
    strictly needed; new access is new key).

Rules:
  - Auth + gateway + FE only.
  - Do NOT touch refund, notifications, C2, games, hero,
    disk, seed.
  - tsc + go test clean.
  - Smoke:
      - Login as user, apply, approve.
      - Same session, no re-login → hit /api/authors/me or
        author-gated API → expect 401 → auto-refresh → 200
        with role=author.
      - Never-approved user still 403 on author routes.
  - No push until I OK.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (main, c6888de):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A #1 — Refund window (9ac8158).

Track A #2 — Author notifications (c6888de).

Track A — в работе:

✅ Item 1 — Refund.

✅ Item 2 — Notifications.

🟧 Item 3 — JWT role refresh (approve → implement + race fix).

Track B (после A):

🟧 3 новые игры, лампочки, уровни.

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.