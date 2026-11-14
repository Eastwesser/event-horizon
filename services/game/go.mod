module github.com/Eastwesser/event-horizon/services/game

go 1.26.0

require (
	go.opentelemetry.io/otel v1.47.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0
	go.opentelemetry.io/otel/sdk v1.47.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12 // indirect
)

replace event_horizon/services/game/proto => ./proto

require (
	github.com/nats-io/nats.go v1.55.0
	github.com/prometheus/client_golang v1.25.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.17 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261110000943-09e21b237f2c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c // indirect
)

replace github.com/Eastwesser/event-horizon/services/game/proto => ./proto

replace github.com/Eastwesser/event-horizon/services/billing/proto => ../billing/proto

require (
	github.com/Eastwesser/event-horizon/services/billing/proto v0.0.0
	github.com/Eastwesser/event-horizon/services/game/proto v0.0.0-20261114100144-58d6b213e89b
	github.com/lib/pq v1.12.3
)

replace github.com/Eastwesser/event-horizon/pkg/migrator => ../../pkg/migrator

replace github.com/Eastwesser/event-horizon/platform => ../../platform

require (
	github.com/Eastwesser/event-horizon/pkg/migrator v0.0.0
	github.com/Eastwesser/event-horizon/platform v0.0.0
	github.com/google/uuid v1.6.0
)
