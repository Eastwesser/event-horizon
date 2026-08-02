.PHONY: proto build run test docker

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/inventory.proto

build:
	go build -o inventory-service ./cmd/main.go

run:
	go run cmd/main.go

test:
	go test -v ./...

docker:
	docker build -t inventory-service .