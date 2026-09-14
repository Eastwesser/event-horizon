.PHONY: help run build clean install

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Запустить приложение
	go run main.go

build: ## Собрать приложение
	go build -o comment-tree.exe main.go

clean: ## Очистить сборку
	rm -f comment-tree.exe

install: ## Установить зависимости
	go mod download
	go mod tidy

.DEFAULT_GOAL := help

