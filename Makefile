.PHONY: up down logs ps clean

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
	# Тут позже добавим реальную генерацию

.PHONY: all
all: stop-services
	@echo "🚀 Starting all services..."
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service &
	cd services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service &
	cd services/game && go build -o game-service ./cmd/main.go && ./game-service &
	cd services/billing && go build -o billing-service ./cmd/main.go && ./billing-service &
	cd services/gateway && go build -o gateway ./cmd/main.go && ./gateway &
	@echo "✅ All services started"

.PHONY: start-services
start-services:
	@echo "🚀 Starting all services..."
	cd services/auth && ./auth-service > /tmp/auth.log 2>&1 &
	cd services/leaderboard && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
	cd services/game && ./game-service > /tmp/game.log 2>&1 &
	cd services/billing && ./billing-service > /tmp/billing.log 2>&1 &
	cd services/gateway && ./gateway > /tmp/gateway.log 2>&1 &
	@echo "✅ All services started"
	@echo "📝 Logs: /tmp/{auth,leaderboard,game,billing,gateway}.log"

.PHONY: stop-services
stop-services:
	@echo "🛑 Stopping all services..."
	@pkill -f "auth-service" || true
	@pkill -f "leaderboard-service" || true
	@pkill -f "game-service" || true
	@pkill -f "billing-service" || true
	@pkill -f "gateway" || true
	@echo "✅ All services stopped"	

.PHONY: restart-services
restart: stop all	

.PHONY: restart
restart: stop-services all