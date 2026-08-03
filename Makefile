.PHONY: up down logs ps clean migrate-all migrate-profile restart status deploy

# ===== DOCKER =====
up:
	docker-compose -f deployments/docker-compose.cluster.yml up -d

down:
	docker-compose -f deployments/docker-compose.cluster.yml down

logs:
	docker-compose -f deployments/docker-compose.cluster.yml logs -f

ps:
	docker-compose -f deployments/docker-compose.cluster.yml ps

clean:
	docker-compose -f deployments/docker-compose.cluster.yml down -v

# ===== DOCKER BUILD =====
docker-build-all:
	@echo "Building all services..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory; do \
		docker build -t eastwesser/$$service:latest -f Dockerfile.$$service.bin .; \
	done

docker-push-all:
	@echo "Pushing all services..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory; do \
		docker push eastwesser/$$service:latest; \
	done		

# ===== BUILD =====

build-all:
	@echo "Building all services..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory; do \
		cd services/$$service && go build -o $$service-service ./cmd/main.go; \
	done

test-all:
	@echo "Running all tests..."
	for service in auth billing game leaderboard profile shop gateway balancer nats-hub inventory; do \
		cd services/$$service && go test -v ./...; \
	done

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
	cd services/nats-hub && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub ./main.go
	docker build -f Dockerfile.nats-hub.bin -t eastwesser/nats-hub:latest .

# ===== DEPLOY =====
deploy:
	@echo "🚀 Building nats-hub..."
	$(MAKE) build-nats-hub
	@echo "🚀 Starting infrastructure..."
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	@sleep 5
	@echo "📦 Running migrations..."
	$(MAKE) migrate-all
	@echo "✅ Everything is ready!"
	@echo "   🎯 Gateway: http://localhost:8079"
	@echo "   📊 Grafana: http://localhost:3000 (admin/admin)"
	@echo "   🔍 Jaeger: http://localhost:16686"

# ===== RESTART =====
restart: down deploy

# ===== STATUS =====
status:
	@echo "🔍 Checking services..."
	docker-compose -f deployments/docker-compose.cluster.yml ps

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
	kubectl apply -f deployments/k3s/deployment.yml
	kubectl apply -f deployments/k3s/service.yml
	kubectl apply -f deployments/k3s/ingress.yml
	kubectl rollout status deployment/event-horizon

undeploy-k3s:
	@echo "🗑️ Removing from k3s..."
	kubectl delete -f deployments/k3s/deployment.yml
	kubectl delete -f deployments/k3s/service.yml
	kubectl delete -f deployments/k3s/ingress.yml
	