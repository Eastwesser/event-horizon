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

all:
	@echo "🚀 Starting all services..."
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service &
	cd services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service &
	cd services/game && go build -o game-service ./cmd/main.go && ./game-service &
	cd services/gateway && go build -o gateway ./cmd/main.go && ./gateway &
	@echo "✅ All services started"		