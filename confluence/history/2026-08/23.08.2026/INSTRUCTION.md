/home/denismatveev/event_horizon/confluence/architecture/GAME_OUTBOX.md
Have we finished all the works with this outbox pattern for game service? 

have we already done all the todos from this (/home/denismatveev/event_horizon/confluence/architecture/LOAD_RESILIENCE.md) file?

Also, can we use this file (/home/denismatveev/event_horizon/confluence/architecture/PATRONI.md) and implement patroni to all our services?

Check if we still have tech debt here: /home/denismatveev/event_horizon/confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md

Then, this: /home/denismatveev/event_horizon/deployments/patroni/auth/k8s
We have to organize work and complete it for all services we have/

Update this if needeed: /home/denismatveev/event_horizon/deployments/patroni/auth/k8s/README.md

ALso update all patroni related files and  services: /home/denismatveev/event_horizon/deployments/patroni

**Status (26.08.2026):** see [`INSTRUCTION_STATUS.md`](INSTRUCTION_STATUS.md)

___
cd services/game
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/game.proto   

cd services/analytics
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/analytics.proto

# Gateway has NO proto/gateway.proto — HTTP only. Use scripts/rebuild-proto.sh instead.


Nghh, help me to rebuild protofiles and binaries, my commands may be ruined (((

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
  eastwesser/analytics:latest \
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



yet to do

Payment — Boosty/subscription (~200₽), unlock shop/merch path
Authors — register authors, APIs to stock Inventory
History — event/audit trail (retention 30d), feeds Analytics
Analytics — ClickHouse (DAU/MAU/retention)
Harden notification + Auth RBAC end-to-end (gateway + role enforcement)
Then Stage 2 fine-tuning / MCP / RAG as in the voice doc





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

