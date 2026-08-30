# Patroni — Event Horizon Postgres HA roadmap

> Default `make deploy` uses **single-instance Postgres per service**. Patroni overlays are **optional HA pilots**, not production defaults until drilled on k3s.

Architecture doc: [`confluence/architecture/PATRONI.md`](../../confluence/architecture/PATRONI.md).

---

## Service matrix (Postgres-backed)

| Service | Compose DB | Port (host) | Patroni status | Notes |
|---------|------------|---------------|----------------|-------|
| **Auth** | `event-horizon-postgres` | 5460 | **Pilot stubs** | [`auth/`](auth/) — compose + k8s |
| Game | `postgres-game` | 5461 | Planned | Copy Auth template after pilot drill |
| Billing | `postgres-billing` | 5462 | Planned | |
| Leaderboard | `postgres-leaderboard` | 5463 | Planned | |
| Profile | `postgres-profile` | 5464 | Planned | |
| Shop | `postgres-shop` | 5465 | Planned | |
| Payment | `postgres-payment` | 5467 | Planned | Money path — drill after Auth |
| Authors | `postgres-authors` | 5468 | Planned | |
| History | `postgres-history` | 5469 | Planned | |
| Inventory | PG + Mongo | — | Later | PG slice only if needed |

**Order of rollout:** Auth → Payment → Shop/Billing → Game/Leaderboard → rest.

---

## What exists today

```
deployments/patroni/
  README.md                 ← this file
  auth/                     ← only implemented stub set
    docker-compose.patroni-auth.yml
    patroni.yml
    haproxy.cfg
    k8s/
```

Other services: add `deployments/patroni/<service>/` **after** Auth failover drill succeeds — duplicate `auth/k8s/` pattern with new `PATRONI_SCOPE` and HAProxy Service name.

---

## Passwords in stubs (GitGuardian)

Compose/k8s credentials come from **`.env`** (compose) or **Kubernetes Secrets** (k3s). See repo-root [`.env.example`](../../../.env.example). Do not commit real passwords to git.

```bash
kubectl create secret generic patroni-auth \
  --namespace=eh-patroni-auth \
  --from-literal=superuser-password='…' \
  --from-literal=replication-password='…'
```

Do not commit real values. See [`confluence/history/2026-08/26.08.2026/GITGUARD_ISSUE.md`](../../confluence/history/2026-08/26.08.2026/GITGUARD_ISSUE.md).

---

## App wiring (any service)

Only **`DB_HOST`** (and secrets) change — no Patroni logic in Go services:

| Var | Example (Auth k8s) |
|-----|---------------------|
| `DB_HOST` | `auth-pg-haproxy.eh-patroni-auth.svc.cluster.local` |
| `DB_PORT` | `5432` |
| `DB_NAME` / user | unchanged per service |

---

## Related

- Auth pilot README: [`auth/README.md`](auth/README.md)
- k8s apply order: [`auth/k8s/README.md`](auth/k8s/README.md)
