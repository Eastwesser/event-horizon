# Gateway local redeploy (binary → image → compose)

You do **not** need `docker push` to run a new image on this machine.
Compose uses the local `eastwesser/gateway:latest` tag.

```bash
cd /home/denismatveev/event_horizon

# 1) binary + local image (no Hub)
bash scripts/rebuild-services.sh gateway

# 2) recreate the 3 gateway replicas (+ balancer if you changed routing)
docker compose --env-file .env \
  -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate gateway gateway-2 gateway-3 balancer

# 3) smoke
curl -sS http://127.0.0.1:8081/health
curl -sS http://127.0.0.1:8079/health
```

## When you *do* need Docker Hub (`docker push`)

Hub is only for other hosts / fresh machines pulling `eastwesser/*`.

If push fails with:

```text
lookup registry-1.docker.io on 8.8.8.8:53: write udp …->8.8.8.8:53: write: operation not permitted
```

Amnezia VPN (`amn0`) is blocking UDP DNS to Google, and `/etc/resolv.conf`
only lists `8.8.8.8`. dockerd’s Go resolver hits that and dies.

```bash
bash scripts/dns-check.sh
# then one of:
sudo ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
# or: echo 'nameserver 1.1.1.1' | sudo tee /etc/resolv.conf
# or: disconnect VPN → docker push → reconnect

docker push eastwesser/gateway:latest
```

Manual equivalent of rebuild-services (same as above, explicit):

```bash
cd services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd ../..
docker build -t eastwesser/gateway:latest -f Dockerfile.gateway.bin .
```
