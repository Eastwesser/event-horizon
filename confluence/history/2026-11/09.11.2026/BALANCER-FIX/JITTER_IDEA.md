🎯 Jitter — сниппет (из прошлого сообщения, но с math/rand/v2)
Файл: services/balancer/internal/balancer/least_conn.go

Место: функция getLeastConnBackendExcluding — заменить полностью.

Сниппет:

go
// getLeastConnBackendExcluding skips healthy=false and any backend already in `skip`.
// When several backends share the minimum ActiveConns, pick one at random (jitter)
// so sequential traffic is spread across the pool instead of always hitting the first.
func (lb *LeastConnBalancer) getLeastConnBackendExcluding(skip map[*Backend]struct{}) *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.backends) == 0 {
		return nil
	}

	candidates := make([]*Backend, 0, len(lb.backends))
	var minConns int32 = 2147483647 // MaxInt32

	for _, b := range lb.backends {
		if !b.Healthy.Load() {
			continue
		}
		if skip != nil {
			if _, ok := skip[b]; ok {
				continue
			}
		}
		conns := atomic.LoadInt32(&b.ActiveConns)
		switch {
		case conns < minConns:
			minConns = conns
			candidates = candidates[:0]
			candidates = append(candidates, b)
		case conns == minConns:
			candidates = append(candidates, b)
		}
	}

	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	return candidates[rand.IntN(len(candidates))]
}
И импорт math/rand/v2:

Место: блок импортов в least_conn.go.

Сниппет:

go
import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/Eastwesser/event-horizon/services/balancer/internal/metrics"
)
Почему math/rand/v2:

Go 1.22+ (у тебя 1.25).

rand.IntN (заглавная N) — thread-safe без глобального мьютекса (использует runtime.fastrand).

Не нужен rand.Seed (v2 автосеется).

В hot path (каждый запрос) — минимальный overhead.

Почему не math/rand (v1): rand.Intn (строчная n) использует глобальный sync.Mutex → contention на hot path. v2 — per-goroutine state.

✅ После jitter — пересборка + проверка
bash
# 1. Собрать
make build-all SERVICES=balancer

# 2. Образ
DOCKER_BUILDKIT=1 docker build -t eastwesser/balancer:latest -f Dockerfile.balancer.bin .

# 3. Пересоздать оба балансера
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d --force-recreate balancer balancer-2

# 4. Проверить распределение (20 последовательных — теперь jitter должен размазать)
for i in {1..20}; do curl -sI http://localhost/health | grep -i "X-Balancer-Backend"; done | sort | uniq -c

# 5. Параллельные (20)
seq 1 20 | xargs -P20 -I{} curl -sI http://localhost/health | grep -i "X-Balancer-Backend" | sort | uniq -c
Ожидаемо (последовательные): ~7/7/6 между gateway, gateway-2, gateway-3. Не 20/0/0.



[denismatveev@c0der event_horizon]$ make build-all SERVICES=balancer
Building all services (CGO_ENABLED=0, linux/amd64)...
===== balancer =====
^Cmake: *** [Makefile:48: build-all] Прерывание

[denismatveev@c0der event_horizon]$ make build-all
Building all services (CGO_ENABLED=0, linux/amd64)...
===== auth =====
===== billing =====
===== game =====
===== leaderboard =====
===== profile =====
===== shop =====
===== gateway =====
===== balancer =====
===== nats-hub =====
===== inventory =====
===== payment =====
===== authors =====
===== history =====
===== analytics =====
===== fulfillment =====
===== notification =====
✅ Static binaries ready → next: make docker-build-all && make deploy
[denismatveev@c0der event_horizon]$ DOCKER_BUILDKIT=1 docker build -t eastwesser/balancer:latest -f Dockerfile.balancer.bin .
[+] Building 10.7s (8/8) FINISHED                                                                                                                                         docker:default
 => [internal] load build definition from Dockerfile.balancer.bin                                                                                                                   0.8s
 => => transferring dockerfile: 197B                                                                                                                                                0.4s
 => [internal] load metadata for docker.io/library/alpine:3.20                                                                                                                      0.1s
 => [internal] load .dockerignore                                                                                                                                                   0.3s
 => => transferring context: 174B                                                                                                                                                   0.2s
 => [1/3] FROM docker.io/library/alpine:3.20                                                                                                                                        0.0s
 => [internal] load build context                                                                                                                                                   1.6s
 => => transferring context: 15.29MB                                                                                                                                                1.3s
 => CACHED [2/3] RUN apk add --no-cache ca-certificates wget                                                                                                                        0.0s
 => [3/3] COPY services/balancer/balancer-service /balancer                                                                                                                         1.4s
 => exporting to image                                                                                                                                                              5.0s
 => => exporting layers                                                                                                                                                             4.2s
 => => writing image sha256:20ed6d6b901bb3b777a30c8508d6505dc0f8d2e54b6771680fc49c9f90d5236e                                                                                        0.1s
 => => naming to docker.io/eastwesser/balancer:latest                                                                                                                               0.5s
[denismatveev@c0der event_horizon]$ docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d --force-recreate balancer balancer-2
[+] up 8/8
 ✔ Container event-horizon-nats-1    Running                                                                                                                                         0.0s
 ✔ Container event-horizon-nats-2    Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-1   Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-3-1 Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-2-1 Running                                                                                                                                         0.0s
 ✔ Container event-horizon-nats-3    Running                                                                                                                                         0.0s
 ✔ Container deployments-balancer-1  Started                                                                                                                                        24.4s
 ✔ Container deployments-balancer-2  Started                                                                                                                                        21.6s
[denismatveev@c0der event_horizon]$ docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d --force-recreate balancer balancer-2
[+] up 8/8
 ✔ Container event-horizon-nats-1    Running                                                                                                                                         0.0s
 ✔ Container event-horizon-nats-3    Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-2-1 Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-3-1 Running                                                                                                                                         0.0s
 ✔ Container event-horizon-nats-2    Running                                                                                                                                         0.0s
 ✔ Container deployments-gateway-1   Running                                                                                                                                         0.0s
 ✔ Container deployments-balancer-1  Started                                                                                                                                        15.4s
 ✔ Container deployments-balancer-2  Started                                                                                                                                        12.9s
[denismatveev@c0der event_horizon]$ for i in {1..20}; do curl -sI http://localhost/health | grep -i "X-Balancer-Backend"; done | sort | uniq -c
     11 X-Balancer-Backend: gateway-2:8080
      7 X-Balancer-Backend: gateway-3:8080
      2 X-Balancer-Backend: gateway:8080
[denismatveev@c0der event_horizon]$ seq 1 20 | xargs -P20 -I{} curl -sI http://localhost/health | grep -i "X-Balancer-Backend" | sort | uniq -c
      8 X-Balancer-Backend: gateway-2:8080
      4 X-Balancer-Backend: gateway-3:8080
      8 X-Balancer-Backend: gateway:8080
[denismatveev@c0der event_horizon]$ 