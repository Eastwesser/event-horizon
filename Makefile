.PHONY: up down logs ps clean proto all start-services stop-services restart status

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

proto:
	@echo "Generating protobuf..."

# Остановить все сервисы
stop-services:
	@echo "🛑 Stopping all services..."
	-pkill -f "auth-service"
	-pkill -f "leaderboard-service"
	-pkill -f "game-service"
	-pkill -f "billing-service"
	-pkill -f "gateway"
	@echo "✅ All services stopped"

# Запустить все сервисы (каждая команда в своей подоболочке)
start-services:
	@echo "🚀 Starting all services..."
	cd /home/denismatveev/event_horizon/services/auth && go build -o auth-service ./cmd/main.go && ./auth-service > /tmp/auth.log 2>&1 &
	cd /home/denismatveev/event_horizon/services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
	cd /home/denismatveev/event_horizon/services/game && go build -o game-service ./cmd/main.go && ./game-service > /tmp/game.log 2>&1 &
	cd /home/denismatveev/event_horizon/services/billing && go build -o billing-service ./cmd/main.go && ./billing-service > /tmp/billing.log 2>&1 &
	cd /home/denismatveev/event_horizon/services/gateway && go build -o gateway ./cmd/main.go && ./gateway > /tmp/gateway.log 2>&1 &
	sleep 2
	@echo "✅ All services started"

# Перезапустить всё
restart: stop-services start-services
	@echo "🔄 Restart completed"

# Запустить всё
all: start-services
	@echo "🎮 EventHorizon is running!"

# Быстрая проверка статуса
status:
	@echo "🔍 Checking services..."
	@pgrep -f "auth-service" && echo "✅ Auth running" || echo "❌ Auth not running"
	@pgrep -f "leaderboard-service" && echo "✅ Leaderboard running" || echo "❌ Leaderboard not running"
	@pgrep -f "game-service" && echo "✅ Game running" || echo "❌ Game not running"
	@pgrep -f "billing-service" && echo "✅ Billing running" || echo "❌ Billing not running"
	@pgrep -f "gateway" && echo "✅ Gateway running" || echo "❌ Gateway not running"
	@echo ""
	@echo "🐳 Docker containers:"
	@docker-compose -f deployments/docker-compose.cluster.yml ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || echo "Docker not running"