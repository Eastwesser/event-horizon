✅ Рабочая инструкция (тот самый момент, когда всё завелось)

Перестал пересобирать gateway через Docker, а просто скопировал готовый локально собранный бинарник в образ.
Сделал NATS необязательным, чтобы gateway стартовал даже без подключения к шине.

Собрал финальный рабочий Docker-образ:
bash
cd /home/denismatveev/event_horizon

# 1. Правим код, чтобы NATS не фейлил gateway
cd services/gateway
sed -i 's/log.Fatalf("Failed to connect to NATS: %v", err)/log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)/g' cmd/main.go
sed -i 's/log.Fatalf("Failed to create JetStream context: %v", err)/log.Printf("⚠️ Failed to create JetStream context: %v", err)/g' cmd/main.go

# 2. Собираем бинарник ЛОКАЛЬНО (без генерации proto)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway-service ./cmd/main.go

# 3. Возвращаемся в корень
cd /home/denismatveev/event_horizon

# 4. Создаём максимально простой Dockerfile – просто копируем бинарник
cat > Dockerfile.gateway.bin << 'EOF'
FROM scratch
COPY services/gateway/gateway-service /gateway
EXPOSE 8080
CMD ["/gateway"]
EOF

# 5. Собираем образ gateway БЕЗ protobuf-генерации
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

# 6. Поднимаем только инфраструктуру и сам gateway в контейнере
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 7. Проверяем статус
docker-compose -f deployments/docker-compose.cluster.yml ps
curl http://localhost:8080/health

---
[denismatveev@c0der event_horizon]$ cd services/gateway
sed -i 's/log.Fatalf("Failed to connect to NATS: %v", err)/log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)/g' cmd/main.go
sed -i 's/log.Fatalf("Failed to create JetStream context: %v", err)/log.Printf("⚠️ Failed to create JetStream context: %v", err)/g' cmd/main.go
[denismatveev@c0der gateway]$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway-service ./cmd/main.go
[denismatveev@c0der gateway]$ cd /home/denismatveev/event_horizon
[denismatveev@c0der event_horizon]$ cat > Dockerfile.gateway.bin << 'EOF'
FROM scratch
COPY services/gateway/gateway-service /gateway
EXPOSE 8080
CMD ["/gateway"]
EOF
[denismatveev@c0der event_horizon]$ docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
            Install the buildx component to build images with BuildKit:
            https://docs.docker.com/go/buildx/

Sending build context to Docker daemon  866.9MB
Step 1/4 : FROM scratch
 ---> 
Step 2/4 : COPY services/gateway/gateway-service /gateway
 ---> d82d1090fdfe
Step 3/4 : EXPOSE 8080
 ---> Running in 7bf521944f33
 ---> Removed intermediate container 7bf521944f33
 ---> 5f04abd862d0
Step 4/4 : CMD ["/gateway"]
 ---> Running in 69d46ec04db9
 ---> Removed intermediate container 69d46ec04db9
 ---> c374817410f9
Successfully built c374817410f9
Successfully tagged eastwesser/gateway:latest
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml up -d
[+] Running 14/20
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 14/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 15/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 15/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 19/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
[+] Running 20/20t-horizon-redis-leaderboard     Running                                                                                                               0.0s 
 ✔ Container event-horizon-redis-leaderboard     Running                                                                                                               0.0s 
 ✔ Container event-horizon-prometheus            Running                                                                                                               0.0s 
 ✔ Container event-horizon-grafana               Running                                                                                                               0.0s 
 ✔ Container deployments-auth-1                  Running                                                                                                               0.0s 
 ✔ Container event-horizon-postgres              Running                                                                                                               0.0s 
 ✔ Container event-horizon-redis-game            Running                                                                                                               0.0s 
 ✔ Container event-horizon-nats                  Running                                                                                                               0.0s 
 ✔ Container deployments-gateway-1               Started                                                                                                               3.1s 
 ✔ Container event-horizon-postgres-billing      Running                                                                                                               0.0s 
 ✔ Container event-horizon-jaeger                Running                                                                                                               0.0s 
 ✔ Container event-horizon-postgres-leaderboard  Running                                                                                                               0.0s 
 ✔ Container event-horizon-redis                 Running                                                                                                               0.0s 
 ✔ Container deployments-gateway-3-1             Started                                                                                                               3.0s 
 ✔ Container deployments-gateway-2-1             Started                                                                                                               3.7s 
 ✔ Container event-horizon-redis-billing         Running                                                                                                               0.0s 
 ✔ Container event-horizon-postgres-game         Running                                                                                                               0.0s 
 ✔ Container deployments-balancer-1              Running                                                                                                               0.0s 
 ✔ Container deployments-leaderboard-1           Started                                                                                                               2.8s 
 ✔ Container deployments-billing-1               Started                                                                                                               2.8s 
 ✔ Container deployments-game-1                  Started                                                                                                               2.8s 
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml ps
curl http://localhost:8080/health
NAME                                 IMAGE                             COMMAND                  SERVICE                CREATED          STATUS                    PORTS
deployments-auth-1                   eastwesser/auth:latest            "/auth-service"          auth                   13 minutes ago   Up 13 minutes             0.0.0.0:50051->50051/tcp, [::]:50051->50051/tcp
deployments-balancer-1               eastwesser/balancer:latest        "/balancer"              balancer               13 minutes ago   Up 12 minutes             0.0.0.0:8079->8079/tcp, [::]:8079->8079/tcp
event-horizon-grafana                grafana/grafana:latest            "/run.sh"                grafana                13 minutes ago   Up 12 minutes             0.0.0.0:3000->3000/tcp, [::]:3000->3000/tcp
event-horizon-jaeger                 jaegertracing/all-in-one:latest   "/go/bin/all-in-one-…"   jaeger                 13 minutes ago   Up 13 minutes             9411/tcp, 14250/tcp, 0.0.0.0:4317-4318->4317-4318/tcp, [::]:4317-4318->4317-4318/tcp, 0.0.0.0:16686->16686/tcp, [::]:16686->16686/tcp, 14268/tcp
event-horizon-nats                   nats:2.10-alpine                  "docker-entrypoint.s…"   nats                   13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:4222->4222/tcp, [::]:4222->4222/tcp, 0.0.0.0:8222->8222/tcp, [::]:8222->8222/tcp, 6222/tcp
event-horizon-postgres               postgres:16-alpine                "docker-entrypoint.s…"   postgres               13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:5460->5432/tcp, [::]:5460->5432/tcp
event-horizon-postgres-billing       postgres:16-alpine                "docker-entrypoint.s…"   postgres-billing       13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:5462->5432/tcp, [::]:5462->5432/tcp
event-horizon-postgres-game          postgres:16-alpine                "docker-entrypoint.s…"   postgres-game          13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:5461->5432/tcp, [::]:5461->5432/tcp
event-horizon-postgres-leaderboard   postgres:16-alpine                "docker-entrypoint.s…"   postgres-leaderboard   13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:5463->5432/tcp, [::]:5463->5432/tcp
event-horizon-prometheus             prom/prometheus:latest            "/bin/prometheus --c…"   prometheus             13 minutes ago   Up 13 minutes             0.0.0.0:9090->9090/tcp, [::]:9090->9090/tcp
event-horizon-redis                  redis:7-alpine                    "docker-entrypoint.s…"   redis                  13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp
event-horizon-redis-billing          redis:7-alpine                    "docker-entrypoint.s…"   redis-billing          13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:6381->6379/tcp, [::]:6381->6379/tcp
event-horizon-redis-game             redis:7-alpine                    "docker-entrypoint.s…"   redis-game             13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:6380->6379/tcp, [::]:6380->6379/tcp
event-horizon-redis-leaderboard      redis:7-alpine                    "docker-entrypoint.s…"   redis-leaderboard      13 minutes ago   Up 13 minutes (healthy)   0.0.0.0:6382->6379/tcp, [::]:6382->6379/tcp
curl: (7) Failed to connect to localhost port 8080 after 0 ms: Could not connect to server
[denismatveev@c0der event_horizon]$ 
---