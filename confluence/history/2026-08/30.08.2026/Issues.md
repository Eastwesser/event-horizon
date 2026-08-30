# Issues & fixes — 26–30.08.2026 (**v1.0.8**)

> Release notes companion. README / CHANGELOG bumped to **v1.0.8**. Tag + GitHub push when you ask.

Related:  
[`OPTIMIZATION.md`](../26.08.2026/OPTIMIZATION.md) · [`DB_ISSUES.md`](../26.08.2026/DB_ISSUES.md) · [`FRONTEND_STATUS.md`](../26.08.2026/FRONTEND_STATUS.md) · [`GITGUARD_ASSESSMENT.md`](../26.08.2026/GITGUARDIAN/GITGUARD_ASSESSMENT.md) · [`INSTRUCTION_STATUS.md`](../23.08.2026/INSTRUCTION_STATUS.md)

---

## Proposed version: v1.0.8

| Item | Value |
|------|--------|
| Previous | v1.0.7 |
| Next | **v1.0.8** |
| Why patch | Secrets externalization, thin NATS deploy, login/Gateway fixes, Game outbox already in tree, observability in thin, NATS for fulfillment/notification |

**Release checklist (after your code review):**

1. Review diffs (Gateway, shop, fulfillment, notification, analytics, nats-hub, compose, Makefile, frontend Auth).
2. Smoke: register/login (≥8 chars), shop purchase → fulfillment log → notification log, Grafana up, game metrics up.
3. Commit + tag `v1.0.8` + push (only when you ask).
4. Optional: `make docker-push-all` or subset of rebuilt images.

---

## Issue index

| # | Symptom | Root cause | Fix | Status |
|---|---------|------------|-----|--------|
| 1 | GitGuardian on Patroni stubs | Dev passwords in compose/yaml | `${PATRONI_*}` / `.env.example`; assessment: not live secrets | Done |
| 2 | `POSTGRES_*` empty on deploy | Secrets moved to env; `.env` missing keys | `.env` + `${VAR:-eventhorizon}` defaults | Done |
| 3 | Register/login “broken” (SPA 500) | UI min 6 vs Auth min 8; Gateway 500 on wrapped gRPC | Frontend minLength 8; Gateway → 400 | Done |
| 4 | Prometheus **game** down | Game exited before PG ready | `depends_on: postgres-game (healthy)`; restart | Done |
| 5 | Prometheus **notification** 404 | No `/metrics` endpoint | `promhttp` on notification | Done |
| 6 | Fulfillment / notification / analytics “fail” | Profiles + `stop-heavy` + Kafka-only path | NATS path; all three on thin `make deploy` | Done |
| 7 | Analytics NATS subscribe fail | Durable names with `.` | Dashes (`analytics-payment-completed`) | Done |
| 8 | Analytics ClickHouse race | Started before CH / missing CH | Thin includes CH; `depends_on: healthy` | Done |
| 9 | Thin stack too heavy / Kafka always on | Kafka in default compose | Profiles: Kafka → `deploy-heavy`; obs stays in thin | Done |
| 10 | `rebuild-all-backend.sh` failed | Called missing `task` CLI | Bash `go build` loop | Done |
| 11 | Gateway whoami/profile load risk | No response cache / no `_partial` | Gateway 5s cache + `_partial` fallback | Done |
| 12 | Boosty webhook weak auth | Shared secret only | HMAC (`X-Boosty-Signature`) at Gateway | Done |

---

## 1. GitGuardian / secrets

**Found:** Alerts on Patroni stub passwords (`patroni`, `replicator`, `eventhorizon`). No `ghp_` / `dckr_pat` / committed `.env` in git history.

**Fixed:**

- Repo-root `.env.example` (JWT, POSTGRES, PATRONI, Grafana, payment placeholders).
- Compose / modular compose / Patroni overlay use env substitution.
- k3s `secret.yml` placeholders + `secretKeyRef` for DB user/password.
- Assessment doc: stubs ≠ production leak.

**Ops note:** Keep local `.env` out of git. Existing PG volumes still use original `eventhorizon` password unless you `make clean`.

---

## 2. Empty Postgres env → deploy warnings / fragile DB

**Found:** `make deploy` warned `POSTGRES_USER` / `POSTGRES_PASSWORD` unset → blank strings in containers. Volumes from older deploys still had `eventhorizon`, so Auth often still worked via Go defaults when env was empty.

**Fixed:**

- Appended `POSTGRES_*` / `GRAFANA_*` to local `.env`.
- Compose: `${POSTGRES_USER:-eventhorizon}`, `${POSTGRES_PASSWORD:-eventhorizon}`, Grafana password default.

---

## 3. Frontend cannot register / login

**Found:**

- SPA: password min **6**; Auth: **8–128**.
- Short password → gRPC `InvalidArgument` → Gateway returned **500** (otel-wrapped error; `status.FromError` failed) → UI “Ошибка соединения”.
- Register used `userId` vs API `user_id`.

**Fixed:**

- `Register.tsx` / `Login.tsx`: min 8, clearer API errors.
- Gateway `writeGRPCError`: unwrap / string-fallback → **400** for InvalidArgument.
- API smoke: register + login with `secret123` OK.

---

## 4. Game service / Prometheus

**Found:** `deployments-game-1` Exited (1); scrape `lookup game: no such host`. Logs: PG not ready / connection refused.

**Fixed:** Compose `depends_on: postgres-game` with `condition: service_healthy`. Restarted game → metrics **200**, outbox worker running.

---

## 5. Notification metrics 404

**Found:** Container up; Prometheus scraped `/metrics` → 404 (only `/health`, `/ready`).

**Fixed:** `promhttp` `/metrics` on notification HTTP server.

---

## 6–8. Fulfillment, notification, analytics on thin stack

**Found:**

- Put behind `profiles: [kafka]` / `[analytics]` then `make stop-heavy` → Exited — looked like “failures”.
- Fulfillment/notification were **Kafka-only**; thin had no Kafka → noop or stopped.
- Analytics: CH DNS race; NATS durables with dots → `invalid consumer name`.

**Fixed:**

| Change | Detail |
|--------|--------|
| NATS purchase bus | Shop publishes `purchase.paid`; fulfillment → `purchase.fulfilled`; notification subscribes both |
| Kafka left in repo | Optional via `KAFKA_BROKERS` + `make deploy-heavy` |
| nats-hub | EVENTS subjects + `UpdateStream` for `purchase.*`, `author.upserted`, … |
| Thin compose | fulfillment, notification, clickhouse, analytics **without** kafka profile |
| Analytics | Durable names with `-`; wait for healthy ClickHouse |

**Verified (30.08):** fulfill/health 200, notif/metrics 200, analytics/ready 200; NATS connected in logs.

---

## 9. Deploy profiles (memory)

**Found:** Full stack (Kafka + CH + everything) hard on ~10 GB + ZRAM.

**Fixed:**

| Target | Meaning |
|--------|---------|
| `make deploy` | **Thin:** NATS + apps + fulfillment/notification/analytics + CH + **obs** |
| `make deploy-heavy` | Thin + **Kafka broker** (preferred name over `strong-deploy`) |
| `make deploy-full` | Alias → `deploy-heavy` |
| `make stop-heavy` | Stop Kafka/qdrant only — **not** the three apps |
| `make deploy-k3s` | Untouched; separate |

Observability stays in thin (Prometheus/Grafana/Jaeger are light vs Kafka/k3s).

---

## 10. Rebuild tooling

**Found:** `scripts/rebuild-all-backend.sh` called `task` → `command not found`.

**Fixed:** Inline `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` loop. Also: `rebuild-proto.sh`, `rebuild-services.sh`, `docker-push-images.sh`, `coverage-gate.sh`.

---

## 11–12. Load resilience / payment (earlier in week)

| Item | Fix |
|------|-----|
| whoami / user / profile cache | Gateway `authReadCache` (~5s) |
| Profile degradation | `_partial: true` from cached whoami when profile CB open |
| Grafana alerts | `deployments/grafana/provisioning/alerting/event-horizon.yml` |
| Boosty webhook | HMAC-SHA256 header verify at Gateway (+ shared secret fallback) |

---

## Intentionally not “finished”

| Item | Notes |
|------|--------|
| Kafka / k3s removal | **Kept**; optional only |
| Qdrant embeddings | Scaffold / profile only |
| Coverage ≥70% everywhere | Gate script exists; many packages still below |
| Boosty official docs | HMAC scheme is industry-standard; no official Boosty public doc in-repo |
| Push to GitHub / Docker Hub | **Waiting for your review & explicit ask** |

---

## Suggested smoke after you approve

```bash
# Stack
make deploy   # or recreate only changed images

# Auth
curl -sS -X POST http://localhost:8079/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"v108@example.com","password":"secret123","nickname":"V108"}'

# Metrics
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost:9092/metrics   # game
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost:9101/health    # fulfillment
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost:9102/metrics   # notification
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost:9106/ready     # analytics

# Optional push later
# make docker-push SVC="gateway shop fulfillment notification analytics nats-hub game auth"
```

---

## Review focus (files)

- `deployments/docker-compose.cluster.yml` — profiles, env defaults, depends_on  
- `Makefile` — `deploy` / `deploy-heavy` / `stop-heavy`  
- `services/gateway/internal/app/gateway.go` — errors, cache, webhook HMAC  
- `services/shop/.../shop_service.go` — NATS `purchase.paid`  
- `services/fulfillment/**`, `services/notification/**` — NATS consumers  
- `services/nats-hub/**` — stream subjects  
- `services/analytics/internal/worker/ingest.go` — durable names  
- `frontend/src/components/Auth/{Register,Login}.tsx`  
- `.env.example` (local `.env` never commit)

When you’re happy: say the word to commit, tag **v1.0.8**, and push.
