# Helm chart — Event Horizon (Wave 4.5 + Track C data plane)

Wraps the legacy multi-container manifests under [`deployments/k3s/`](../../k3s/) as a Helm chart.

## Install

```bash
# preview (app only)
helm template eh ./deployments/helm/event-horizon

# preview with data plane
helm template eh ./deployments/helm/event-horizon --set dataPlane.enabled=true

# install / upgrade (release name `eh`)
helm upgrade --install eh ./deployments/helm/event-horizon \
  --set secret.jwtSecret="$(openssl rand -base64 32)" \
  --set secret.postgresPassword='…'

# app + in-cluster NATS ×3 + Postgres STS (7 DBs from values.db.hosts)
make deploy-k3s-dataplane

# or via Make (app only)
make deploy-k3s
make undeploy-k3s
```

## Scope

- **In chart (always):** Deployment (auth, billing, game, leaderboard, profile, shop, inventory, gateway), Service, Ingress, Secret.
- **In chart (`dataPlane.enabled=true`):** NATS StatefulSets + headless Services (`nats-1`…`nats-3`), Postgres StatefulSets + Services matching `db.hosts` (auth…inventory).
- **Still external / out of scope:** Redis, HPA, Patroni HA, nats-hub, payment/authors/history/notification DBs (not in this app Deployment).

Raw YAML in `deployments/k3s/` remains as a reference snapshot; `make deploy-k3s` prefers Helm when `helm` is on `PATH`.
