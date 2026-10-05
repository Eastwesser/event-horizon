# Helm chart — Event Horizon (Wave 4.5)

Wraps the legacy multi-container manifests under [`deployments/k3s/`](../../k3s/) as a Helm chart.

## Install

```bash
# preview
helm template eh ./deployments/helm/event-horizon

# install / upgrade (release name `eh`)
helm upgrade --install eh ./deployments/helm/event-horizon \
  --set secret.jwtSecret="$(openssl rand -base64 32)" \
  --set secret.postgresPassword='…'

# or via Make
make deploy-k3s
make undeploy-k3s
```

## Scope

- **In chart:** Deployment (auth, billing, game, leaderboard, profile, shop, inventory, gateway), Service, Ingress, Secret.
- **Not in chart:** NATS, Postgres, Redis StatefulSets, HPA, Patroni (same gaps as raw k3s).

Raw YAML in `deployments/k3s/` remains as a reference snapshot; `make deploy-k3s` prefers Helm when `helm` is on `PATH`.
