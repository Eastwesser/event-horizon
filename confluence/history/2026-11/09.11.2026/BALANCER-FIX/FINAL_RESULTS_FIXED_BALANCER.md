✅ Что подтвердилось
nginx -t зелёный (и в отдельном контейнере с правильной сетью deployments_event-horizon-net, и через docker exec deployments-nginx nginx -t).

Балансеры healthy: balancer-1 → health: starting (через пару секунд станет healthy), balancer-2 → healthy. Разные состояния — норм, это race на старте health-check'а (5s ticker). Через 5-10s оба будут healthy.

nginx running на :80.

curl http://localhost/health → 200 через всю цепочку Nginx → Balancer → Gateway, с заголовком X-Balancer-Backend: gateway:8080 — значит балансер реально проксирует и отдаёт response header (как задумано).

curl http://localhost:9098/ready → 200 {"status":"ready","service":"balancer",...} — осмысленный /ready работает.

make deploy → ✅ Thin stack ready без Ошибка 1.

Цепочка живая: Клиент :80 → Nginx → Balancer ×2 → Gateway ×3 → gRPC. 🎉

🔍 Одна вещь для проверки — balancer-1 "health: starting"
На момент ps он был health: starting, а balancer-2 — healthy. Это не баг, а разница тайминга: health-check горутина стартует в RunUntilSignal, первый check() сразу, но gateway'и на тот момент могли ещё не отвечать → первый check() → unhealthy → потом через 5s второй check() → healthy. У balancer-2 первый check() попал в момент, когда gateway уже ответил.

Проверь через 10-15 секунд:

bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps balancer balancer-2
# оба должны быть (healthy)
И метрики:

bash
curl -s http://localhost:9098/metrics | grep balancer_backend_healthy
curl -s http://localhost:8080/metrics 2>/dev/null | grep balancer_backend_healthy || echo "balancer-2 metrics не проброшены на хост (9098 занят balancer-1)"
Заметь: у balancer-2 в compose не проброшен 9098 на хост (только 8080:8079). Значит :9098/metrics — это только balancer-1. Если хочешь видеть метрики обоих — надо пробросить 9099:9098 для balancer-2. P3, не блокер.

🧪 Быстрый smoke — что ещё стоит проверить
bash
# 1. nginx жив
curl -s -o /dev/null -w "nginx /health: %{http_code}\n" http://localhost/health

# 2. балансер-1 через хост
curl -s -o /dev/null -w "balancer-1 /ready: %{http_code}\n" http://localhost:8079/ready

# 3. балансер-2 через хост
curl -s -o /dev/null -w "balancer-2 /ready: %{http_code}\n" http://localhost:8080/ready

# 4. метрики балансера-1
curl -s http://localhost:9098/metrics | grep -E "balancer_(backend_healthy|active_connections|requests_total)"

# 5. ретрай через nginx: 5 быстрых запросов подряд
for i in {1..5}; do curl -s -o /dev/null -w "%{http_code} " http://localhost/health; done; echo

# 6. распределение по бэкендам (X-Balancer-Backend)
for i in {1..10}; do curl -sI http://localhost/health | grep -i "X-Balancer-Backend"; done | sort | uniq -c
Ожидаемо:

1-3 → 200.

4 → три метрики.

5 → 200 200 200 200 200.

6 → распределение между gateway:8080, gateway-2:8080, gateway-3:8080 (least connections, при равной нагрузке — примерно поровну).

Но по делу: после каждого изменения в services/balancer/ нужен:

make build-all SERVICES=balancer — бинарь.

docker build -t eastwesser/balancer:latest -f Dockerfile.balancer.bin . — образ.

docker compose ... up -d --force-recreate balancer balancer-2 — пересоздать контейнеры.

make deploy — только для полного рестарта стека (например, после make down или перезагрузки). Для итераций по одному сервису — точечный up -d --force-recreate <service>.

Есть make rebuild-services SVC=balancer — это bash-скрипт, который, скорее всего, делает 1+2+3 автоматически. Посмотри scripts/rebuild-services.sh — если он делает всё правильно, используй его. Это удобнее, чем три команды вручную.

-------------- TESTING --------------
[denismatveev@c0der event_horizon]$ docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps balancer balancer-2
# оба должны быть (healthy)
NAME                     IMAGE                        COMMAND       SERVICE      CREATED         STATUS                   PORTS
deployments-balancer-1   eastwesser/balancer:latest   "/balancer"   balancer     7 minutes ago   Up 7 minutes (healthy)   0.0.0.0:8079->8079/tcp, [::]:8079->8079/tcp, 0.0.0.0:9098->9098/tcp, [::]:9098->9098/tcp
deployments-balancer-2   eastwesser/balancer:latest   "/balancer"   balancer-2   7 minutes ago   Up 7 minutes (healthy)   9098/tcp, 0.0.0.0:8080->8079/tcp, [::]:8080->8079/tcp
[denismatveev@c0der event_horizon]$ curl -s http://localhost:9098/metrics | grep balancer_backend_healthy
curl -s http://localhost:8080/metrics 2>/dev/null | grep balancer_backend_healthy || echo "balancer-2 metrics не проброшены на хост (9098 занят balancer-1)"
# HELP balancer_backend_healthy 1 if backend /ready returned 2xx on the last check, 0 otherwise
# TYPE balancer_backend_healthy gauge
balancer_backend_healthy{backend="gateway-2:8080"} 1
balancer_backend_healthy{backend="gateway-3:8080"} 1
balancer_backend_healthy{backend="gateway:8080"} 1
balancer-2 metrics не проброшены на хост (9098 занят balancer-1)
[denismatveev@c0der event_horizon]$ 

Надо чтобы были тоже проблошены
Заметь: у balancer-2 в compose не проброшен 9098 на хост (только 8080:8079). Значит :9098/metrics — это только balancer-1. Если хочешь видеть метрики обоих — надо пробросить 9099:9098 для balancer-2. P3, не блокер.

[denismatveev@c0der event_horizon]$ curl -s -o /dev/null -w "nginx /health: %{http_code}\n" http://localhost/health
nginx /health: 200
[denismatveev@c0der event_horizon]$ curl -s -o /dev/null -w "balancer-1 /ready: %{http_code}\n" http://localhost:8079/ready
balancer-1 /ready: 200
[denismatveev@c0der event_horizon]$ curl -s -o /dev/null -w "balancer-2 /ready: %{http_code}\n" http://localhost:8080/ready
balancer-2 /ready: 200
[denismatveev@c0der event_horizon]$ curl -s http://localhost:9098/metrics | grep -E "balancer_(backend_healthy|active_connections|requests_total)"
# HELP balancer_active_connections Total active connections across all backends
# TYPE balancer_active_connections gauge
balancer_active_connections 0
# HELP balancer_backend_healthy 1 if backend /ready returned 2xx on the last check, 0 otherwise
# TYPE balancer_backend_healthy gauge
balancer_backend_healthy{backend="gateway-2:8080"} 1
balancer_backend_healthy{backend="gateway-3:8080"} 1
balancer_backend_healthy{backend="gateway:8080"} 1
# HELP balancer_requests_total Total number of requests processed by balancer
# TYPE balancer_requests_total counter
balancer_requests_total 2
[denismatveev@c0der event_horizon]$ for i in {1..5}; do curl -s -o /dev/null -w "%{http_code} " http://localhost/health; done; echo
200 200 200 200 200 
[denismatveev@c0der event_horizon]$ for i in {1..10}; do curl -sI http://localhost/health | grep -i "X-Balancer-Backend"; done | sort | uniq -c
     10 X-Balancer-Backend: gateway:8080
[denismatveev@c0der event_horizon]$ 
