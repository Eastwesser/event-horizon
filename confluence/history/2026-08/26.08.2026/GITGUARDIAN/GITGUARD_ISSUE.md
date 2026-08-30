Commit e6bb5ab
Eastwesser
Eastwesser
committed
1 hour ago
chore(devops): fix compose env/JWT loading and add Auth Patroni stubs
Use --env-file .env for compose; allow ClickHouse default from Docker
network; add optional Patroni/etcd/HAProxy stubs for Auth on k3s.

1.0.7 wip:
main
1 parent 
7266ddb
 commit 
e6bb5ab
12 files changed

+444
-10
Lines changed: 444 additions & 10 deletions
Search within code
 
‎Makefile‎
+10
-7
Lines changed: 10 additions & 7 deletions
Original file line number	Diff line number	Diff line change
@@ -1,20 +1,23 @@
.PHONY: up down logs ps clean migrate-all migrate-profile restart status deploy test-all test-unit test-smoke test-k6

# Always pass repo-root .env so ${JWT_SECRET} etc. substitute correctly.
COMPOSE := docker compose --env-file .env -f deployments/docker-compose.cluster.yml
# ===== DOCKER =====
up:
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	$(COMPOSE) up -d

down:
	docker-compose -f deployments/docker-compose.cluster.yml down
	$(COMPOSE) down

logs:
	docker-compose -f deployments/docker-compose.cluster.yml logs -f
	$(COMPOSE) logs -f

ps:
	docker-compose -f deployments/docker-compose.cluster.yml ps
	$(COMPOSE) ps

clean:
	docker-compose -f deployments/docker-compose.cluster.yml down -v
	$(COMPOSE) down -v

# BuildKit (DOCKER_BUILDKIT=1) replaces the deprecated legacy builder.
# Full `docker buildx` needs the plugin: Arch `sudo pacman -S docker-buildx`.
@@ -123,7 +126,7 @@ deploy:
	@echo "🚀 Building nats-hub..."
	$(MAKE) build-nats-hub
	@echo "🚀 Starting infrastructure..."
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	$(COMPOSE) up -d
	@sleep 5
	@echo "📦 Running migrations..."
	$(MAKE) migrate-all
@@ -138,7 +141,7 @@ restart: down deploy
# ===== STATUS =====
status:
	@echo "🔍 Checking services..."
	docker-compose -f deployments/docker-compose.cluster.yml ps
	$(COMPOSE) ps

# ===== DELIVERY =====
delivery-dev:
‎deployments/clickhouse/users.d/default-network.xml‎
+11
Lines changed: 11 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,11 @@
<!-- Allow default user from Docker network (analytics → clickhouse).
     Official image users.d/default-user.xml restricts default to 127.0.0.1 only. -->
<clickhouse>
  <users>
    <default>
      <networks>
        <ip>::/0</ip>
      </networks>
    </default>
  </users>
</clickhouse>
‎deployments/docker-compose.cluster.yml‎
+9
-3
Lines changed: 9 additions & 3 deletions
Original file line number	Diff line number	Diff line change
@@ -927,8 +927,14 @@ services:
    ports:
      - "8123:8123"
      - "9000:9000"
    environment:
      # Official image otherwise writes users.d/default-user.xml that only allows
      # default@127.0.0.1 — analytics (other container) then gets Authentication failed.
      CLICKHOUSE_SKIP_USER_SETUP: "1"
      CLICKHOUSE_DB: eventhorizon
    volumes:
      - clickhouse_data:/var/lib/clickhouse
      - ./clickhouse/users.d/default-network.xml:/etc/clickhouse-server/users.d/default-network.xml:ro
    ulimits:
      nofile:
        soft: 262144
@@ -976,7 +982,7 @@ services:
    ports:
      - "8081:8080"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:8080/ready"]
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/ready"]
      interval: 15s
      timeout: 5s
      retries: 5
@@ -1013,7 +1019,7 @@ services:
    ports:
      - "8082:8080"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:8080/ready"]
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/ready"]
      interval: 15s
      timeout: 5s
      retries: 5
@@ -1050,7 +1056,7 @@ services:
    ports:
      - "8083:8080"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:8080/ready"]
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/ready"]
      interval: 15s
      timeout: 5s
      retries: 5
‎deployments/patroni/auth/README.md‎
+81
Lines changed: 81 additions & 0 deletions


Original file line number	Diff line number	Diff line change
@@ -0,0 +1,81 @@
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
‎deployments/patroni/auth/docker-compose.patroni-auth.yml‎
+72
Lines changed: 72 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,72 @@
# Optional local Patroni pilot for Auth DB (NOT part of make deploy).
# Uses a dedicated compose project so it does not collide with event-horizon-postgres.
services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.15
    container_name: auth-patroni-etcd
    command:
      - etcd
      - --advertise-client-urls=http://etcd:2379
      - --listen-client-urls=http://0.0.0.0:2379
    networks: [patroni-auth-net]
  patroni-1:
    image: aginies/patroni:latest
    container_name: auth-patroni-1
    environment:
      PATRONI_NAME: patroni-1
      PATRONI_SCOPE: eh-auth
      PATRONI_ETCD3_HOSTS: etcd:2379
      PATRONI_RESTAPI_CONNECT_ADDRESS: patroni-1:8008
      PATRONI_RESTAPI_LISTEN: 0.0.0.0:8008
      PATRONI_POSTGRESQL_CONNECT_ADDRESS: patroni-1:5432
      PATRONI_POSTGRESQL_LISTEN: 0.0.0.0:5432
      PATRONI_POSTGRESQL_DATA_DIR: /home/postgres/pgdata
      PATRONI_SUPERUSER_USERNAME: postgres
      PATRONI_SUPERUSER_PASSWORD: patroni
      PATRONI_REPLICATION_USERNAME: replicator
      PATRONI_REPLICATION_PASSWORD: replicator
      # App role (stub — create DB/user on first bootstrap if image supports POST_INIT)
      PATRONI_admin_USERNAME: eventhorizon
      PATRONI_admin_PASSWORD: eventhorizon
    volumes:
      - ./patroni.yml:/etc/patroni/patroni.yml:ro
    depends_on: [etcd]
    networks: [patroni-auth-net]
    # STUB: many Patroni images expect a different entrypoint; adjust when bringing live.
  patroni-2:
    image: aginies/patroni:latest
    container_name: auth-patroni-2
    environment:
      PATRONI_NAME: patroni-2
      PATRONI_SCOPE: eh-auth
      PATRONI_ETCD3_HOSTS: etcd:2379
      PATRONI_RESTAPI_CONNECT_ADDRESS: patroni-2:8008
      PATRONI_RESTAPI_LISTEN: 0.0.0.0:8008
      PATRONI_POSTGRESQL_CONNECT_ADDRESS: patroni-2:5432
      PATRONI_POSTGRESQL_LISTEN: 0.0.0.0:5432
      PATRONI_POSTGRESQL_DATA_DIR: /home/postgres/pgdata
      PATRONI_SUPERUSER_USERNAME: postgres
      PATRONI_SUPERUSER_PASSWORD: patroni
      PATRONI_REPLICATION_USERNAME: replicator
      PATRONI_REPLICATION_PASSWORD: replicator
    volumes:
      - ./patroni.yml:/etc/patroni/patroni.yml:ro
    depends_on: [etcd]
    networks: [patroni-auth-net]
  auth-pg-haproxy:
    image: haproxy:2.9-alpine
    container_name: auth-pg-haproxy
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    ports:
      - "5433:5432"   # host:5433 → VIP (avoid clash with other Postgres host ports)
    depends_on: [patroni-1, patroni-2]
    networks: [patroni-auth-net]
networks:
  patroni-auth-net:
    driver: bridge
‎deployments/patroni/auth/haproxy.cfg‎
+20
Lines changed: 20 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,20 @@
global
  maxconn 100
  stats timeout 2m
defaults
  mode tcp
  timeout connect 5s
  timeout client 30m
  timeout server 30m
  option clitcpka
# STUB: health checks against Patroni REST (leader only).
# Adjust server names to match running Patroni containers.
listen postgres_write
  bind *:5432
  option httpchk GET /primary
  http-check expect status 200
  default-server inter 3s fall 3 rise 2 on-marked-down shutdown-sessions
  server patroni-1 patroni-1:5432 check port 8008
  server patroni-2 patroni-2:5432 check port 8008
‎deployments/patroni/auth/k8s/README.md‎
+28
Lines changed: 28 additions & 0 deletions


Original file line number	Diff line number	Diff line change
@@ -0,0 +1,28 @@
# STUB manifests for Auth Postgres HA on k3s.
Apply order:
```bash
kubectl apply -f namespace.yaml
kubectl apply -f etcd.yaml
# Create secret before Patroni (optional in stub):
# kubectl -n eh-patroni-auth create secret generic patroni-auth-secrets \
#   --from-literal=superuser-password=patroni \
#   --from-literal=replication-password=replicator
kubectl apply -f patroni-statefulset.yaml
kubectl apply -f haproxy-service.yaml
```
Point Auth chart/deployment:
```yaml
env:
  - name: DB_HOST
    value: auth-pg-haproxy.eh-patroni-auth.svc.cluster.local
  - name: DB_PORT
    value: "5432"
  - name: DB_NAME
    value: eventhorizon
```
**Not production-ready until:** PVC + storageClass, real Patroni image/config, etcd HA, network policies, and a failover drill.
‎deployments/patroni/auth/k8s/etcd.yaml‎
+39
Lines changed: 39 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,39 @@
# STUB: single-node etcd for Patroni DCS. Replace with HA etcd/consul for real prod.
apiVersion: apps/v1
kind: Deployment
metadata:
  name: etcd
  namespace: eh-patroni-auth
spec:
  replicas: 1
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      labels:
        app: etcd
    spec:
      containers:
        - name: etcd
          image: quay.io/coreos/etcd:v3.5.15
          args:
            - etcd
            - --advertise-client-urls=http://0.0.0.0:2379
            - --listen-client-urls=http://0.0.0.0:2379
          ports:
            - containerPort: 2379
              name: client
---
apiVersion: v1
kind: Service
metadata:
  name: etcd
  namespace: eh-patroni-auth
spec:
  selector:
    app: etcd
  ports:
    - port: 2379
      targetPort: 2379
      name: client
‎deployments/patroni/auth/k8s/haproxy-service.yaml‎
+63
Lines changed: 63 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,63 @@
# VIP for Auth: set Auth Deployment DB_HOST=auth-pg-haproxy
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-pg-haproxy
  namespace: eh-patroni-auth
data:
  haproxy.cfg: |
    global
      maxconn 100
    defaults
      mode tcp
      timeout connect 5s
      timeout client 30m
      timeout server 30m
    listen postgres_write
      bind *:5432
      option httpchk GET /primary
      http-check expect status 200
      default-server inter 3s fall 3 rise 2
      server patroni-0 patroni-0.patroni:5432 check port 8008
      server patroni-1 patroni-1.patroni:5432 check port 8008
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-pg-haproxy
  namespace: eh-patroni-auth
spec:
  replicas: 1
  selector:
    matchLabels:
      app: auth-pg-haproxy
  template:
    metadata:
      labels:
        app: auth-pg-haproxy
    spec:
      containers:
        - name: haproxy
          image: haproxy:2.9-alpine
          ports:
            - containerPort: 5432
          volumeMounts:
            - name: cfg
              mountPath: /usr/local/etc/haproxy
      volumes:
        - name: cfg
          configMap:
            name: auth-pg-haproxy
---
apiVersion: v1
kind: Service
metadata:
  name: auth-pg-haproxy
  namespace: eh-patroni-auth
spec:
  selector:
    app: auth-pg-haproxy
  ports:
    - port: 5432
      targetPort: 5432
      name: postgres
‎deployments/patroni/auth/k8s/namespace.yaml‎
+7
Lines changed: 7 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,7 @@
apiVersion: v1
kind: Namespace
metadata:
  name: eh-patroni-auth
  labels:
    app.kubernetes.io/part-of: event-horizon
    app.kubernetes.io/component: patroni-auth-stub
‎deployments/patroni/auth/k8s/patroni-statefulset.yaml‎
+58
Lines changed: 58 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,58 @@
# STUB StatefulSet — image/env must be validated on your k3s before go-live.
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: patroni
  namespace: eh-patroni-auth
spec:
  serviceName: patroni
  replicas: 2
  selector:
    matchLabels:
      app: patroni
  template:
    metadata:
      labels:
        app: patroni
    spec:
      containers:
        - name: patroni
          image: aginies/patroni:latest
          env:
            - name: PATRONI_SCOPE
              value: eh-auth
            - name: PATRONI_ETCD3_HOSTS
              value: etcd:2379
            - name: PATRONI_SUPERUSER_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: patroni-auth-secrets
                  key: superuser-password
                  optional: true
            - name: PATRONI_REPLICATION_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: patroni-auth-secrets
                  key: replication-password
                  optional: true
          ports:
            - containerPort: 5432
              name: postgres
            - containerPort: 8008
              name: rest
          # volumeClaimTemplates omitted in stub — add PVC + storageClass for real deploy
---
apiVersion: v1
kind: Service
metadata:
  name: patroni
  namespace: eh-patroni-auth
spec:
  clusterIP: None
  selector:
    app: patroni
  ports:
    - port: 5432
      name: postgres
    - port: 8008
      name: rest
‎deployments/patroni/auth/patroni.yml‎
+46
Lines changed: 46 additions & 0 deletions
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,46 @@
# STUB Patroni config — values overridden by env in compose where possible.
scope: eh-auth
namespace: /eventhorizon/
name: patroni-node
restapi:
  listen: 0.0.0.0:8008
  connect_address: 127.0.0.1:8008
etcd3:
  hosts: etcd:2379
bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
  initdb:
    - encoding: UTF8
    - data-checksums
  users:
    eventhorizon:
      password: eventhorizon
      options:
        - createrole
        - createdb
postgresql:
  listen: 0.0.0.0:5432
  connect_address: 127.0.0.1:5432
  data_dir: /home/postgres/pgdata
  pgpass: /tmp/pgpass
  authentication:
    replication:
      username: replicator
      password: replicator
    superuser:
      username: postgres
      password: patroni
  parameters:
    max_connections: 100
    shared_buffers: 128MB
    wal_level: replica
    hot_standby: "on"
    max_wal_senders: 5