# Consul evaluate (v1.1.0 / Wave 4.5)

**Decision: do not introduce Consul in Event Horizon for v1.1.0.**

## When Consul helps

- Dynamic service registration / deregistration (many short-lived instances).
- Multi-datacenter or multi-cluster discovery outside Kubernetes.
- Centralized health-driven routing beyond what the orchestrator already provides.
- Shared KV / feature flags (possible, but EH already uses env + Redis where needed).

## What EH uses today

| Environment | Discovery |
|-------------|-----------|
| Docker Compose | Static DNS names (`auth`, `shop`, `nats-1`, …) in compose network |
| k3s | Kubernetes Service + **CoreDNS** (`*.svc.cluster.local`); gateway env points at service names |
| Local / Ansible | Compose-first; k3s manifests assume external NATS/Postgres until StatefulSets exist |

Gateway and services dial fixed addresses from config (`AUTH_ADDR`, `SHOP_ADDR`, `NATS_URL`, …). There is no client-side service catalog.

## Why skip now

1. **k3s DNS is enough** for in-cluster gRPC/HTTP once Services exist — same pattern as every standard Kubernetes app.
2. **Compose does not need Consul** — static compose DNS already matches the interview “static addresses” story.
3. Adding Consul would mean another HA dependency (agents, ACL, upgrades) without fixing the real k3s gaps (NATS + Postgres in-cluster), which are documented in [`deployments/k3s/README.md`](../../deployments/k3s/README.md).
4. Portfolio / interview value of “we evaluated Consul and chose platform DNS” is clearer than a half-wired Consul sidecar.

## Revisit if

- Services run **outside** k3s/compose with changing IPs, or
- We need cross-cluster discovery, or
- We deliberately want Consul KV / Connect as a demo (not required for product).

Until then: keep **Compose DNS + k3s CoreDNS**; no Consul chart or compose service.
