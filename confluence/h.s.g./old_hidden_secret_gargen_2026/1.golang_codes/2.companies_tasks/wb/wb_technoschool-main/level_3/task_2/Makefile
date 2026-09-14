.PHONY: help run build docker-up docker-down

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Запустить приложение
	go run main.go

build: ## Собрать приложение
	go build -o url-shortener.exe main.go

docker-up: ## Запустить Docker контейнеры
	@docker compose up -d 2>/dev/null || docker-compose up -d

docker-down: ## Остановить Docker контейнеры
	@docker compose down 2>/dev/null || docker-compose down

clean: ## Очистить сборку
	rm -f url-shortener.exe

install: ## Установить зависимости
	go mod download
	go mod tidy

.DEFAULT_GOAL := help

