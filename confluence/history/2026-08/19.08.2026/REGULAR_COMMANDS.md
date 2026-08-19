cd ~/event_horizon/services/game

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o game-service ./cmd/main.go

cd ~/event_horizon

docker build -f Dockerfile.game.bin -t eastwesser/game:latest .


--
Your side (no net from me)
Docker rebuild for changed images when ready:

# binaries already built under services/*/*-service
# or rebuild + image:
--
for svc in auth gateway billing inventory profile shop leaderboard game; do
  (cd services/$svc && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${svc}-service ./cmd/main.go)
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
--

РЕГЕН ПРОТОК ФАЙЛОВ

cd services/auth
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/auth.proto


cd services/billing
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/billing.proto


cd services/inventory
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/inventory.proto


cd services/profile
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/profile.proto


cd services/shop
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/shop.proto


cd services/leaderboard
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/leaderboard.proto


cd services/game
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/game.proto   



       PUSH


       Docker push only takes one image per call — that multi-arg form won’t work.

Use a loop:

for img in \
  eastwesser/auth:latest \
  eastwesser/gateway:latest \
  eastwesser/inventory:latest \
  eastwesser/shop:latest \
  eastwesser/billing:latest \
  eastwesser/game:latest \
  eastwesser/leaderboard:latest \
  eastwesser/profile:latest \
  eastwesser/balancer:latest \
  eastwesser/nats-hub:latest
do
  echo "===== PUSH $img ====="
  docker push "$img" || exit 1
done
If you also built W5 images:

for img in eastwesser/fulfillment:latest eastwesser/notification:latest; do
  docker push "$img" || exit 1
done
For only what you just rebuilt for W6:

for img in eastwesser/auth:latest eastwesser/gateway:latest; do
  docker push "$img" || exit 1
done



Stale “yet to do” below is **done as of 19.08 P0–P2**. Payment / authors / history / analytics shipped. Deferred items (squirrel, swaggo, zap, Envoy, live Boosty, ≥70% coverage) are **skip**, not a next phase.





docker build -f Dockerfile.authors.bin -t eastwesser/authors:latest .
docker build -f Dockerfile.history.bin -t eastwesser/history:latest .
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker push eastwesser/authors:latest
docker push eastwesser/history:latest
docker push eastwesser/analytics:latest
docker push eastwesser/gateway:latest
docker compose -f deployments/docker-compose.cluster.yml up -d \
  postgres-authors redis-authors authors \
  postgres-history history \
  clickhouse analytics \
  gateway gateway-2 gateway-3


docker build -f Dockerfile.shop.bin -t eastwesser/shop:latest .
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker build -f Dockerfile.inventory.bin -t eastwesser/inventory:latest .
docker build -f Dockerfile.billing.bin -t eastwesser/billing:latest .

docker push eastwesser/shop:latest
docker push eastwesser/gateway:latest
docker push eastwesser/analytics:latest
docker push eastwesser/inventory:latest
docker push eastwesser/billing:latest


(cd services/mcp && CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-server ./cmd/main.go)


Priority 5 (MCP + RAG) is in. Billing/dot issue acknowledged — good catch.

What you got
services/mcp — stdio MCP server (mark3labs/mcp-go)
Tools: nats_list_streams, nats_list_consumers, postgres_query (SELECT-only), redis_get / redis_keys, search_prydwen
RAG: offline TF-IDF over Prydwen (no embedding API)
Config example: .cursor/mcp.json
Status: MCP_RAG_STATUS.md
Binary is already built: services/mcp/mcp-server

Wire Cursor
Enable the event-horizon MCP server from .cursor/mcp.json (paths already point at your machine). With compose up, set NATS_URL / REDIS_ADDR / MCP_POSTGRES_DSN as needed.

Then try: search_prydwen → "how does Outbox work in Shop", or nats_list_streams.

Rebuild if you change code:

(cd services/mcp && CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-server ./cmd/main.go)
No Docker push needed for MCP — Cursor runs it as a local stdio process.


===

Thanks for prydwen knowledge.

I also need to check one more thing- status codes.
/home/denismatveev/event_horizon/confluence/history/2026-08/13.08.2026/STATUS_CODES.md

One more thing: /home/denismatveev/event_horizon/confluence/architecture
This is the main folder of architecture. Check all the files there and write to this one
(/home/denismatveev/event_horizon/confluence/architecture/SYSTEM_DESIGN/HOW_DOES_EH_WORK.md),
how does EH work. ANALYZE this picture (/home/denismatveev/event_horizon/confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.6.png) and also suggest schema worth showing on system design interview. This one is made in miro, you may do it differently


YET TODO LIST:
Answers (short)
#      Verdict
1 Unit tests
Core yes (auth/billing/shop/inventory/gateway/payment). Not every service; balancer/fulfillment/notification = 0.
2 Integration
No — only auth + billing + shop + inventory. Need ≥1 for payment/authors/history/analytics/…
3 Routes sync
Mostly. OpenAPI missing auth refresh/whoami/logout/update-role. FE has no History page.
4 Squirrel
Nowhere — raw SQL/pgx. Not required.
5 Redis
Where needed (auth/billing/shop/inventory/leaderboard/gateway/payment/authors/profile). Not on game/history/analytics — intentional.
6 FE↔BE
Connected now (payment/authors/analytics/merch). Badge still says 1.0.6; stack is past that. Only History UI lags.
7 MCP
Binary + .cursor/mcp.json ready. Rebuild only if you change MCP. Optional fine-tune later.
8 OpenAPI
~full minus 4 auth routes.
9 Patterns
CB all gRPC, rate limit, outbox (billing/inventory/authors/payment), optimistic lock, interceptors, health/ready.
10 swaggo
No — hand YAML + /docs Swagger UI. By design.
11 Logging
slog standard; log.Printf leftovers in workers/gateway. No zap.
12 TX/outbox
TX: shop/billing/inventory/payment/authors. Outbox: billing/inventory/authors/payment. Shop publishes NATS directly (no outbox table).