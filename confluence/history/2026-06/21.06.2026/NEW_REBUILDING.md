cd /home/denismatveev/event_horizon/services/billing
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags=-static" -o billing-service ./cmd/main.go

