Отлично, командир! Ты закоммитил рабочее состояние - это правильный ход! 👍

Теперь проблема: Gateway не слушает порт 8080. Судя по curl: (7) Failed to connect, gateway либо не запущен, либо запущен но не на 8080.

Проверяем:

bash
# 1. Смотрим логи gateway
docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=50

# 2. Проверяем запущен ли gateway
docker ps | grep gateway

# 3. Смотрим статус через make
make status

# 4. Проверяем порты
sudo ss -tlnp | grep 808
Скорее всего, gateway не запущен в Docker (мы его останавливали) или упал.

Решение - запустить gateway локально:

bash
# 1. Останавливаем gateway в Docker (если запущен)
docker-compose -f deployments/docker-compose.cluster.yml stop gateway gateway-2 gateway-3 2>/dev/null || true

# 2. Проверяем что бинарник существует
ls -la services/gateway/gateway-service

# 3. Запускаем локально
cd services/gateway
./gateway-service &

# 4. Проверяем
curl http://localhost:8080/health

# 5. Проверяем через балансер
curl http://localhost:8079/health
ИЛИ пересобираем Docker-образ с последним бинарником:

bash
cd /home/denismatveev/event_horizon

# 1. Пересобираем образ gateway из бинарника
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

# 2. Запускаем
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway

# 3. Смотрим логи
docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=30

# 4. Проверяем
curl http://localhost:8080/health
Если gateway стартует но не отвечает - проверь на каком порту он слушает:

bash
# В коде gateway
grep "ListenAndServe" services/gateway/cmd/main.go
# Должно быть :8080
Покажи что выдают эти команды - тогда будем чинить дальше.
---
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=50
gateway-1  | panic: runtime error: slice bounds out of range [-4:]
gateway-1  | 
gateway-1  | goroutine 1 [running]:
gateway-1  | google.golang.org/protobuf/internal/filedesc.(*File).unmarshalSeed(0xc000102a80, {0x134ca74, 0x8bf, 0x8bf})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:174 +0x1719
gateway-1  | google.golang.org/protobuf/internal/filedesc.newRawFile({{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1515020, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:30 +0x1f3
gateway-1  | google.golang.org/protobuf/internal/filedesc.Builder.Build({{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1515020, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/build.go:105 +0xf8
gateway-1  | google.golang.org/protobuf/internal/filetype.Builder.Build({{{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x0, ...}, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filetype/build.go:138 +0x185
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.file_billing_proto_init()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:998 +0x198
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.init.0()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:979 +0xf
gateway-1  | panic: runtime error: slice bounds out of range [-4:]
gateway-1  | 
gateway-1  | goroutine 1 [running]:
gateway-1  | google.golang.org/protobuf/internal/filedesc.(*File).unmarshalSeed(0xc000103c00, {0x134ca74, 0x8bf, 0x8bf})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:174 +0x1719
gateway-1  | google.golang.org/protobuf/internal/filedesc.newRawFile({{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1515020, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:30 +0x1f3
gateway-1  | google.golang.org/protobuf/internal/filedesc.Builder.Build({{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1515020, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/build.go:105 +0xf8
gateway-1  | google.golang.org/protobuf/internal/filetype.Builder.Build({{{0x1146b66, 0x3a}, {0x134ca74, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x0, ...}, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filetype/build.go:138 +0x185
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.file_billing_proto_init()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:998 +0x198
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.init.0()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:979 +0xf
[denismatveev@c0der event_horizon]$ docker ps | grep gateway
[denismatveev@c0der event_horizon]$ make status
🔍 Checking services...
36257
59536
✅ Auth running
59538
✅ Leaderboard running
59540
✅ Game running
59542
✅ Billing running
59544
✅ Gateway running

🐳 Docker containers:
NAME                                 STATUS
deployments-auth-1                   Up 7 minutes
deployments-balancer-1               Up 7 minutes
event-horizon-grafana                Up 7 minutes
event-horizon-jaeger                 Up 7 minutes
event-horizon-nats                   Up 7 minutes (healthy)
event-horizon-postgres               Up 7 minutes (healthy)
event-horizon-postgres-billing       Up 7 minutes (healthy)
event-horizon-postgres-game          Up 7 minutes (healthy)
event-horizon-postgres-leaderboard   Up 7 minutes (healthy)
event-horizon-prometheus             Up 7 minutes
event-horizon-redis                  Up 7 minutes (healthy)
event-horizon-redis-billing          Up 7 minutes (healthy)
event-horizon-redis-game             Up 7 minutes (healthy)
event-horizon-redis-leaderboard      Up 7 minutes (healthy)
[denismatveev@c0der event_horizon]$ sudo ss -tlnp | grep 808
[sudo] пароль для denismatveev: 
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml stop gateway gateway-2 gateway-3 2>/dev/null || true
[denismatveev@c0der event_horizon]$ ls -la services/gateway/gateway-service
-rwxr-xr-x 1 denismatveev denismatveev 38808493 июн 20 22:26 services/gateway/gateway-service
[denismatveev@c0der event_horizon]$ cd services/gateway
./gateway-service &
[1] 61278
[denismatveev@c0der gateway]$ panic: runtime error: slice bounds out of range [-4:]

goroutine 1 [running]:
google.golang.org/protobuf/internal/filedesc.(*File).unmarshalSeed(0xc000292540, {0x134b77e, 0x8bf, 0x8bf})
        /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:174 +0x1719
google.golang.org/protobuf/internal/filedesc.newRawFile({{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1513ac0, ...}, ...})
        /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:30 +0x1f3
google.golang.org/protobuf/internal/filedesc.Builder.Build({{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1513ac0, ...}, ...})
        /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/build.go:105 +0xf8
google.golang.org/protobuf/internal/filetype.Builder.Build({{{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x0, ...}, ...}, ...})
        /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filetype/build.go:138 +0x185
github.com/Eastwesser/event-horizon/services/billing/proto.file_billing_proto_init()
        /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:998 +0x198
github.com/Eastwesser/event-horizon/services/billing/proto.init.0()
        /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:979 +0xf
[denismatveev@c0der gateway]$ 


[denismatveev@c0der event_horizon]$ curl http://localhost:8080/health
curl: (7) Failed to connect to localhost port 8080 after 0 ms: Could not connect to server
[denismatveev@c0der event_horizon]$ curl http://localhost:8079/health
Backend error
[denismatveev@c0der event_horizon]$ docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
            Install the buildx component to build images with BuildKit:
            https://docs.docker.com/go/buildx/

Sending build context to Docker daemon  866.9MB
Step 1/4 : FROM scratch
 ---> 
Step 2/4 : COPY services/gateway/gateway-service /gateway
 ---> Using cache
 ---> 657e2ba69b82
Step 3/4 : EXPOSE 8080
 ---> Using cache
 ---> 18b58f14cc1c
Step 4/4 : CMD ["/gateway"]
 ---> Using cache
 ---> 4fb18b4852b4
Successfully built 4fb18b4852b4
Successfully tagged eastwesser/gateway:latest
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml up -d gateway
[+] Running 2/2
 ✔ Container event-horizon-nats     Running                                                                                                                            0.0s 
 ✔ Container deployments-gateway-1  Started                                                                                                                            3.4s 
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=30
gateway-1  | panic: runtime error: slice bounds out of range [-4:]
gateway-1  | 
gateway-1  | goroutine 1 [running]:
gateway-1  | google.golang.org/protobuf/internal/filedesc.(*File).unmarshalSeed(0xc000133880, {0x134b77e, 0x8bf, 0x8bf})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:174 +0x1719
gateway-1  | google.golang.org/protobuf/internal/filedesc.newRawFile({{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1513ac0, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/desc_init.go:30 +0x1f3
gateway-1  | google.golang.org/protobuf/internal/filedesc.Builder.Build({{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x1513ac0, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filedesc/build.go:105 +0xf8
gateway-1  | google.golang.org/protobuf/internal/filetype.Builder.Build({{{0x1145aa6, 0x3a}, {0x134b77e, 0x8bf, 0x8bf}, 0x1, 0xc, 0x0, 0x1, {0x0, ...}, ...}, ...})
gateway-1  |    /home/denismatveev/go/pkg/mod/google.golang.org/protobuf@v1.34.2/internal/filetype/build.go:138 +0x185
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.file_billing_proto_init()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:998 +0x198
gateway-1  | github.com/Eastwesser/event-horizon/services/billing/proto.init.0()
gateway-1  |    /home/denismatveev/event_horizon/services/billing/proto/billing.pb.go:979 +0xf
[denismatveev@c0der event_horizon]$ curl http://localhost:8080/health
curl: (7) Failed to connect to localhost port 8080 after 2 ms: Could not connect to server
[denismatveev@c0der event_horizon]$ grep "ListenAndServe" services/gateway/cmd/main.go
        if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
[denismatveev@c0der event_horizon]$ 
---