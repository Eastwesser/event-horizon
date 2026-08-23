# Patroni — Postgres HA for Event Horizon

## Short answer

**Yes for k3s/prod (Auth pilot first). No as default `make deploy`.**

Local compose keeps single-instance Postgres per service. HA stubs for **Auth** live under:

`deployments/patroni/auth/` (compose overlay + k8s stubs + README).

---

## What Patroni gives

| Capability | Why EH wants it |
|------------|-----------------|
| Automatic leader election | Survive primary crash without manual restart |
| Streaming replicas | Optional read scaling later |
| Controlled switchover | Safer upgrades |
| Health endpoints | HAProxy / k8s probes |

Stack: **Patroni + etcd + HAProxy (VIP)** → Auth keeps one `DB_HOST`.

---

## What we have today

| Layer | State |
|-------|--------|
| Compose default | `event-horizon-postgres` single node for Auth |
| Patroni Auth pilot stubs | `deployments/patroni/auth/` ✅ |
| Payment / other DBs | Still single-instance (expand after Auth drill) |
| App code | No Patroni-specific logic — only DSN host |

Docker exit `255` after host reboot is **not** a Patroni problem (lifecycle). Patroni helps when Postgres/primary dies while the cluster stays up.

---

## Auth pilot (recommended first)

1. Review stubs: `deployments/patroni/auth/README.md`
2. On k3s: `kubectl apply -f deployments/patroni/auth/k8s/` (after fixing storage + secrets)
3. Set Auth `DB_HOST=auth-pg-haproxy.eh-patroni-auth.svc.cluster.local`
4. Run failover drill (`patronictl list` → kill leader → login still works)

Local optional overlay (separate network, host port **5433**):

```bash
docker compose -f deployments/patroni/auth/docker-compose.patroni-auth.yml up -d
```

Images/entrypoints are **stubs** — expect to adjust Patroni image when bringing live.

---

## Do not

- Wrap all 8+ EH Postgres instances in Patroni at once
- Replace Outbox / Redis with “just add Patroni”
- Point production Auth at stubs without a drill

---

## Related

- [`LOAD_RESILIENCE.md`](LOAD_RESILIENCE.md)
- [`EH_SCHEMAS.md`](EH_SCHEMAS.md)
- [`GAME_OUTBOX.md`](GAME_OUTBOX.md)
- Early scale notes: `confluence/history/2026-05/11.05.2026/CORE_IDEAS.md`
