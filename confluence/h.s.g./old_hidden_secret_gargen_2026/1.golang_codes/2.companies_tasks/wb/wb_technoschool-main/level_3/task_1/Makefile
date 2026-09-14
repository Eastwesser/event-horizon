.PHONY: help run build test clean docker-up docker-down

help: ## Показать это сообщение
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Запустить приложение
	go run main.go

build: ## Собрать приложение
	go build -o delayed-notifier main.go

test: ## Запустить тесты
	go test -v ./...

clean: ## Очистить сборку
	rm -f delayed-notifier
	rm -f delayed-notifier.exe

docker-up: ## Запустить Docker контейнеры (RabbitMQ + Redis)
	@docker compose up -d 2>/dev/null || docker-compose up -d

docker-down: ## Остановить Docker контейнеры
	@docker compose down 2>/dev/null || docker-compose down

docker-logs: ## Показать логи Docker контейнеров
	@docker compose logs -f 2>/dev/null || docker-compose logs -f

install: ## Установить зависимости
	go mod download
	go mod tidy

.DEFAULT_GOAL := help

