# Postmortem — API `/api/v1` cutover vs stale Gateway (09–10.10.2026)

**Severity:** user-facing outage of balances, leaderboard, shop inventory, submits, notifications  
**Status:** mitigated (gateway rebuilt + redeployed; FE already on `/api/v1`)  
**Related:** `CRUCIAL.md`, `BUGS_GAMS/*`, `LEFT_TO_FINISH.md`, `DAILY_OPS.md`

---

## 1. Summary

Frontend was switched to `baseURL: '/api/v1'`. Running **Docker images** for Gateway still served only legacy `/api/*`. Result: every protected call from Vite (`:5173`) returned **404**. That looked like “games broken / LB empty / lamps 0 / can’t fails” — mostly one routing mismatch, not eight separate game bugs.

Power cuts during remediations made cold `make deploy` fail on the **first** try (Postgres crash-recovery longer than healthcheck budget).

---

## 2. Timeline (compressed)

| When | What |
|------|------|
| v1.1 wave | Gateway source + OpenAPI + FE + k6 moved to `/api/v1`; legacy rewrite added in source |
| Deploy without rebuild | Compose kept old `eastwesser/gateway:latest` → host still answered `/api/*` with **200**, `/api/v1/*` with **404** |
| Denis QA | CRUCIAL + per-game console dumps: all `GET/POST …/api/v1/… 404` |
| Diagnosis | `curl :8079/api/v1/leaderboard` → 404; `curl :8079/api/leaderboard` → 200 |
| Fix | `bash scripts/rebuild-services.sh gateway` + recreate gateway×3 |
| Follow-on | Gin rewrite middleware only mutated `URL.Path` **after** route match → legacy stayed 404 until `HandleContext` re-dispatch |
| Cold boot | Unclean PG shutdown → fsync 30–90s; healthcheck `retries:5` marked PG unhealthy → billing/game cascade |
| Verify | `curl …/api/v1/leaderboard` → **200** (Denis, 10.10 ~05:10) |

---

## 3. Root causes

1. **Image drift:** code/docs said `/api/v1`; running containers did not.  
2. **Incomplete legacy rewrite:** Gin matches routes before middleware; path rewrite without `HandleContext` does not re-route.  
3. **Fragile first boot:** PG/billing healthchecks too tight after power loss (compose patched: longer `start_period` / `depends_on`).

**Not the root cause:** Vite proxy ( `/api` → `:8079` was fine), shop seed wipe (orthogonal), or per-game FE logic for the 404 storm.

---

## 4. Impact

- Lamps/tickets UI stuck at **0 0** (`/billing/balance/all` 404)  
- Leaderboard load / game submit / whoami / notifications / shop inventory 404  
- Mis-triage: hours spent in game bug files that were mostly API noise  
- `make seed-v110` **not required** once volumes survived — correct call by Denis  

---

## 5. Detection gap

No alert fired for “public `/api/v1` returns 404 while `/api` works.” Prometheus `up` stayed green (metrics ports healthy).

---

## 6. Fixes shipped

- Rebuild/redeploy Gateway with `/api/v1` routes  
- Legacy rewrite via `r.HandleContext(c)` after path remap  
- Compose: billing `depends_on` healthy PG/Redis/NATS; longer start periods; PG health `start_period: 120s`  
- Docs: `DAILY_OPS.md`, this postmortem  

---

## 7. Follow-ups (separate from this incident)

| Item | Note |
|------|------|
| Pancaker score → LB **0** | Different bug: hexagon validator ignored client score when `moves=[]` — fixed in game service (trust client like memory) |
| Boost at 0 lamps | FE gate (BoostCheckbox) |
| Observability “gaps” (Loki/Tempo/…) | **Not** this incident — see `DREAM_DEVOPS_IDEAS.md` |

---

## 8. Action items / prevention

1. After any public API prefix change: **rebuild + recreate gateway×3** before claiming done (`scripts/rebuild-services.sh gateway`).  
2. Smoke: `curl -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8079/api/v1/leaderboard?game_id=hexagon&limit=1` expect **200** (or 401), never 404.  
3. Optional later: blackbox probe on `/api/v1/...` (dream list).  
4. Do not advise “just run deploy twice” without fixing healthcheck budgets — first-try cold boot is the bar.

---

## 9. Lesson

**Version the contract and the image together.** FE on `/api/v1` + old gateway = silent product death with healthy metrics.
