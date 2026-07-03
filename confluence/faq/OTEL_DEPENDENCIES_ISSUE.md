cd /home/denismatveev/event_horizon/services/billing

# Обновляем всё
go get go.opentelemetry.io/otel@v1.34.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
go get go.opentelemetry.io/otel/sdk@v1.34.0
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
go mod tidy

# Пересобираем
go build -o billing-service ./cmd/main.go

Повтори для game, auth и leaderboard:

cd ../game
go get go.opentelemetry.io/otel@v1.34.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
go get go.opentelemetry.io/otel/sdk@v1.34.0
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
go mod tidy
go build -o game-service ./cmd/main.go

cd ../leaderboard
go get go.opentelemetry.io/otel@v1.34.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
go get go.opentelemetry.io/otel/sdk@v1.34.0
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
go mod tidy
go build -o leaderboard-service ./cmd/main.go

cd ../auth
go get go.opentelemetry.io/otel@v1.34.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
go get go.opentelemetry.io/otel/sdk@v1.34.0
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
go mod tidy
go build -o auth-service ./cmd/main.go

JAEGER TRACES ISSUE 

[denismatveev@c0der event_horizon]$ cd /home/denismatveev/event_horizon
[denismatveev@c0der event_horizon]$ find services -name "main.go" -exec sed -i 's/localhost:4317/172.17.0.1:4317/g' {} \;
[denismatveev@c0der event_horizon]$ grep -r "172.17.0.1:4317" services/*/cmd/main.go
services/auth/cmd/main.go:        otlptracegrpc.WithEndpoint("172.17.0.1:4317"),
services/billing/cmd/main.go:        otlptracegrpc.WithEndpoint("172.17.0.1:4317"),
services/game/cmd/main.go:        otlptracegrpc.WithEndpoint("172.17.0.1:4317"),
services/leaderboard/cmd/main.go:        otlptracegrpc.WithEndpoint("172.17.0.1:4317"),
[denismatveev@c0der event_horizon]$ 

Проблема в том, что Go-сервисы на хосте не могут подключиться к localhost:4317 внутри Docker-контейнера. 
Нужно использовать IP шлюза Docker (172.17.0.1).