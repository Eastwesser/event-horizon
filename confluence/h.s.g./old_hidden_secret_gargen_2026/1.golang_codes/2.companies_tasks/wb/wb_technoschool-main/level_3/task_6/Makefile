.PHONY: help docker-up docker-down run build clean test

help:
	@echo "SalesTracker - Available commands:"
	@echo "  make docker-up    - Start PostgreSQL container"
	@echo "  make docker-down  - Stop PostgreSQL container"
	@echo "  make run          - Run the application"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f postgres

run: docker-up
	@echo "Waiting for database..."
	@sleep 3
	go run main.go

build:
	go build -o task_6.exe main.go

test:
	go test ./...

clean:
	rm -f task_6.exe
	go clean

