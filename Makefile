.PHONY: up down logs ps clean migrate-all migrate-profile restart status deploy deploy-heavy deploy-full deploy-kafka stop-heavy test-all test-unit test-smoke test-k6

# Always pass repo-root .env so ${JWT_SECRET} etc. substitute correctly.
COMPOSE := docker compose --env-file .env -f deployments/docker-compose.cluster.yml
# Optional: Kafka broker only (apps already use NATS; set KAFKA_BROKERS when enabling)
COMPOSE_HEAVY := --profile kafka

# ===== DOCKER =====
up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

clean:
	$(COMPOSE) down -v
	-$(COMPOSE) $(COMPOSE_HEAVY) down

# BuildKit (DOCKER_BUILDKIT=1) replaces the deprecated legacy builder.
# Full `docker buildx` needs the plugin: Arch `sudo pacman -S docker-buildx`.
export DOCKER_BUILDKIT=1

# ===== DOCKER BUILD =====
docker-build-all:
	@echo "Building all services..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory payment authors history analytics fulfillment notification; do \
		DOCKER_BUILDKIT=1 docker build -t eastwesser/$$service:latest -f Dockerfile.$$service.bin .; \
	done

docker-push-all:
	@echo "Pushing all services..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory payment authors history analytics fulfillment notification; do \
		docker push eastwesser/$$service:latest; \
	done

# ===== BUILD =====
SERVICES ?= auth billing game leaderboard profile shop gateway balancer nats-hub inventory payment authors history analytics fulfillment notification

build-all:
	@echo "Building all services..."
	for service in $(SERVICES); do \
		(cd services/$$service && go build -o $$service-service ./cmd/main.go) || exit 1; \
	done

test-unit:
	@echo "Unit tests..."
	@failed=0; for service in $(SERVICES); do \
		echo "===== $$service ====="; \
		(cd services/$$service && GOWORK=off go test ./... -count=1) || failed=1; \
	done; \
	(cd pkg/sqb && GOWORK=off go test ./...) || failed=1; \
	(cd pkg/migrator && GOWORK=off go test ./...) || failed=1; \
	(cd platform && GOWORK=off go test ./...) || failed=1; \
	exit $$failed

test-smoke:
	@echo "Smoke: curl /health|/ready on local metrics ports (cluster must be up)"
	@set -e; \
	for url in \
	  http://127.0.0.1:9091/health \
	  http://127.0.0.1:9092/health \
	  http://127.0.0.1:9093/health \
	  http://127.0.0.1:9103/ready \
	  http://127.0.0.1:9104/ready \
	  http://127.0.0.1:9105/ready \
	  http://127.0.0.1:9106/ready \
	  http://127.0.0.1:8081/ready; do \
	  echo "→ $$url"; \
	  curl -fsS -o /dev/null --max-time 3 "$$url" || echo "SKIP/FAIL $$url (is compose up?)"; \
	done

test-k6:
	@echo "Optional k6 load (requires k6 + running stack)"
	@command -v k6 >/dev/null || { echo "k6 not installed — skip"; exit 0; }
	k6 run deployments/k6/loadtest.js

test-integration:
	@echo "Integration tests (testcontainers; needs Docker OR *_TEST_DATABASE_URL)"
	@set -e; \
	(cd services/billing && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/shop && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/inventory && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/payment && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/authors && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/history && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m); \
	(cd services/analytics && GOWORK=off go test -tags=integration ./internal/repository/clickhouse/ -count=1 -timeout 5m)

test-all: test-unit
	@echo "Unit OK. Optional: make test-smoke | make test-k6 | make test-integration"
# ===== MIGRATIONS =====
migrate-auth:
	cd services/auth && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up

migrate-billing:
	cd services/billing && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up

migrate-game:
	cd services/game && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up

migrate-leaderboard:
	cd services/leaderboard && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up

migrate-profile:
	cd services/profile && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5464/eventhorizon_profile?sslmode=disable" up

migrate-shop:
	cd services/shop && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5465/eventhorizon_shop?sslmode=disable" up

migrate-inventory:
	cd services/inventory && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5466/eventhorizon_inventory?sslmode=disable" up

migrate-all: migrate-auth migrate-billing migrate-game migrate-leaderboard migrate-profile migrate-shop migrate-inventory
	@echo "✅ All migrations applied"

# ===== NATS HUB =====
build-nats-hub:
	@echo "🔨 Building nats-hub..."
	cd services/nats-hub && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub ./cmd/main.go
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.nats-hub.bin -t eastwesser/nats-hub:latest .

# ===== PROTO / REBUILD HELPERS =====
gen-proto-local:
	bash scripts/rebuild-proto.sh

rebuild-services:
	@test -n "$(SVC)" || (echo "Usage: make rebuild-services SVC='game analytics'"; exit 1)
	bash scripts/rebuild-services.sh $(SVC)

docker-push:
	@test -n "$(SVC)" || (echo "Usage: make docker-push SVC='game analytics' or SVC=--all"; exit 1)
	bash scripts/docker-push-images.sh $(SVC)

patroni-auth-up:
	docker compose --env-file .env -f deployments/patroni/auth/docker-compose.patroni-auth.yml up -d

# ===== DEPLOY =====
# Thin (default): NATS + apps + fulfillment/notification/analytics + ClickHouse + obs.
# Heavy: same + Kafka broker (optional dual-write when KAFKA_BROKERS=kafka:9092).
# k3s: make deploy-k3s — separate, leave manifests untouched.
deploy:
	@echo "🚀 Building nats-hub..."
	$(MAKE) build-nats-hub
	@echo "🚀 Starting thin stack (NATS path; Kafka OFF)..."
	$(COMPOSE) up -d
	@sleep 5
	@echo "📦 Running migrations..."
	$(MAKE) migrate-all
	@echo "✅ Thin stack ready"
	@echo "   🎯 Gateway: http://localhost:8079"
	@echo "   📊 Grafana: http://localhost:3000"
	@echo "   🔍 Jaeger: http://localhost:16686"
	@echo "   📈 Prometheus: http://localhost:9090"
	@echo "   Stronger box later: make deploy-heavy  (adds Kafka broker)"
	@echo "   Stop Kafka only: make stop-heavy"

deploy-kafka:
	@echo "🚀 Starting Kafka broker (apps keep NATS; set KAFKA_BROKERS=kafka:9092 to dual-write)..."
	$(COMPOSE) --profile kafka up -d

# Prefer this name over deploy-full on a stronger machine.
deploy-heavy: deploy-kafka
	@echo "✅ Heavy extras up (Kafka). Thin stack already includes fulfillment/notification/analytics."
	@echo "   To dual-write shop→Kafka: export KAFKA_BROKERS=kafka:9092 and recreate shop/fulfillment/notification"

# Alias kept for muscle memory.
deploy-full: deploy-heavy

# Stop Kafka / qdrant only (do not stop thin apps or ClickHouse).
stop-heavy:
	@echo "⏹ Stopping Kafka / qdrant (thin stack left running)..."
	-docker stop event-horizon-kafka event-horizon-qdrant 2>/dev/null || true
	@echo "✅ Done"

# ===== RESTART =====
restart: down deploy

# ===== STATUS =====
status:
	@echo "🔍 Checking services..."
	$(COMPOSE) ps
	@echo "---"
	@$(COMPOSE) $(COMPOSE_HEAVY) ps 2>/dev/null || true

# ===== DELIVERY =====
delivery-dev:
	cd delivery && ansible-playbook -i inventory/dev.ini ansible/site.yml

delivery-staging:
	cd delivery && ansible-playbook -i inventory/staging.ini ansible/site.yml

delivery-prod:
	cd delivery && ansible-playbook -i inventory/prod.ini ansible/site.yml

# ===== K3S =====
deploy-k3s:
	@echo "🚀 Deploying to k3s..."
	kubectl apply -f deployments/k3s/secret.yml
	kubectl apply -f deployments/k3s/deployment.yml
	kubectl apply -f deployments/k3s/service.yml
	kubectl apply -f deployments/k3s/ingress.yml
	kubectl rollout status deployment/event-horizon

undeploy-k3s:
	@echo "🗑️ Removing from k3s..."
	kubectl delete -f deployments/k3s/deployment.yml
	kubectl delete -f deployments/k3s/service.yml
	kubectl delete -f deployments/k3s/ingress.yml
	kubectl delete -f deployments/k3s/secret.yml
	