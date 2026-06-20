module github.com/Eastwesser/event-horizon/services/gateway

go 1.25.7

require (
	github.com/gin-gonic/gin v1.10.1
	github.com/gorilla/websocket v1.5.3
	github.com/nats-io/nats.go v1.52.0
	github.com/prometheus/client_golang v1.20.5
	github.com/redis/go-redis/v9 v9.19.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.55.0
	google.golang.org/grpc v1.69.0-dev
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/Eastwesser/event-horizon/services/auth/proto => ../auth/proto
	github.com/Eastwesser/event-horizon/services/billing/proto => ../billing/proto
	github.com/Eastwesser/event-horizon/services/game/proto => ../game/proto
	github.com/Eastwesser/event-horizon/services/leaderboard/proto => ../leaderboard/proto
)

require (
	github.com/Eastwesser/event-horizon/services/auth/proto v0.0.0
	github.com/Eastwesser/event-horizon/services/billing/proto v0.0.0
	github.com/Eastwesser/event-horizon/services/game/proto v0.0.0
	github.com/Eastwesser/event-horizon/services/leaderboard/proto v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.1 // indirect
	github.com/bytedance/sonic/loader v0.5.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/gin-contrib/sse v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.2 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/arch v0.27.0 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
