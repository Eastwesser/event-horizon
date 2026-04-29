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