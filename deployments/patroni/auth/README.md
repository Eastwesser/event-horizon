# Patroni stub — Auth Postgres HA (pilot)

> Status: **stubs for k3s / optional compose overlay**. Not wired into default `make deploy`.
> Pilot DB: Auth (`eventhorizon` on shared auth Postgres).

---

## Goal

Give Auth a Patroni-managed primary + replica behind a single VIP so `DB_HOST` for the Auth service stays stable across failover.

Default local compose keeps `event-horizon-postgres` as a single instance. Use this overlay only for HA experiments / k3s.

---

## Layout

```
deployments/patroni/auth/
  docker-compose.patroni-auth.yml   # optional local overlay (etcd + 2× Patroni + haproxy stub)
  patroni.yml                       # Patroni config template
  haproxy.cfg                       # TCP proxy → Patroni leader
  k8s/
    namespace.yaml
    etcd.yaml                       # single-node etcd stub
    patroni-statefulset.yaml        # 2-pod Patroni stub
    haproxy-service.yaml            # Service VIP for Auth DSN
    README.md
```

---

## App wiring (when enabled)

| Var | Value |
|-----|--------|
| `DB_HOST` | `auth-pg-haproxy` (k8s) or `event-horizon-auth-pg-vip` (compose overlay) |
| `DB_PORT` | `5432` |
| `DB_NAME` | `eventhorizon` |
| User/password | unchanged (`eventhorizon` / from secret) |

Auth code does **not** need Patroni-awareness — only the DSN host changes.

---

## Local overlay (optional)

```bash
# From repo root — does NOT replace default postgres; use a separate project name:
docker compose -f deployments/patroni/auth/docker-compose.patroni-auth.yml up -d
```

Then point a **test** Auth container at `DB_HOST=auth-pg-haproxy`.

---

## k3s

```bash
kubectl apply -f deployments/patroni/auth/k8s/
# Wait for patroni-0 leader, then:
# Auth Deployment env: DB_HOST=auth-pg-haproxy
```

These manifests are **stubs**: images/tags and storage class must match your cluster. They will not magically work without etcd readiness and PVC provisioning.

---

## Failover drill (when live)

1. Identify leader: `patronictl list`
2. Kill leader pod / stop container
3. Confirm HAProxy routes to new leader within ~30s
4. Auth `/ready` stays OK; login still works

---

## Related

- [`../../architecture/PATRONI.md`](../../architecture/PATRONI.md)
- Default Auth Postgres: `event-horizon-postgres` in `docker-compose.cluster.yml`
