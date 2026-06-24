.PHONY: up down logs ps clean migrate-all restart status deploy

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

# ===== MIGRATIONS =====
migrate-auth:
	cd services/auth && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up

migrate-billing:
	cd services/billing && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up

migrate-game:
	cd services/game && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up

migrate-leaderboard:
	cd services/leaderboard && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up

migrate-all: migrate-auth migrate-billing migrate-game migrate-leaderboard
	@echo "✅ All migrations applied"

# ===== DEPLOY =====
deploy:
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