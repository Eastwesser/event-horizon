# Zombie docker-proxy on :4317 (jaeger)

**Date:** 10.11.2026
**Symptom:** `make deploy` fails with `Bind for 0.0.0.0:4317 failed: port is already allocated`, but no container holds the port.

## Root cause

`event-horizon-jaeger` was in `Exited (255)` — killed by a signal (reboot / `systemctl restart docker` / `docker kill`). When a container is killed with SIGKILL, Docker **does not always** clean up its `docker-proxy` process. The proxy stays alive, holds `:4317`, and Docker thinks the port is taken.

Sequence:
1. `Exited (255)` — container killed.
2. `docker-proxy` (pids 1810, 1819) orphaned, still binds `:4317`.
3. `docker rm -f event-horizon-jaeger` — container removed, **proxy remains**.
4. `kill <pid>` — SIGTERM ignored; Docker daemon may recreate the proxy.
5. `systemctl restart docker` — **only reliable fix**: clears all docker-proxy + iptables DNAT.

## Diagnosis

```bash
# Port held, but no container claims it
sudo ss -tlnp | grep 4317                    # docker-proxy, pid ...
docker ps -a --format '{{.Names}}\t{{.Ports}}' | grep 4317   # empty
docker ps -a --filter "name=jaeger"          # Exited (255)

# iptables orphan (if systemctl restart docker is not an option)
sudo iptables -t nat -L DOCKER -n --line-numbers | grep 4317

Fix (in order)
docker rm -f event-horizon-jaeger

sudo kill -9 <docker-proxy pids> — may not help

sudo systemctl restart docker — guaranteed; containers with restart: unless-stopped come back, the rest via docker compose up -d

Prevention
restart: unless-stopped on every service (so a reboot brings the stack back).

Health checks on every service (so docker ps shows healthy / unhealthy).

If a container is stuck in Exited (255) — systemctl restart docker is the safe reset, do not try to kill individual pids.

Related
PORT_9190_BALANCER-2.md — another port collision (Profile held :9099).

BALANCER_FOUNDINGS.md — balancer health checks + nginx retry.