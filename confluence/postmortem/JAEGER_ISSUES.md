🔧 Проблема

Jaeger в Docker-сети event-horizon-net слушает на порту 4317, но Go-сервисы на хосте не могут к нему подключиться, потому что они не в Docker-сети.

🎯 Решение: запустить Go-сервисы в Docker-сети

Вариант А: Подключить Go-сервисы к сети event-horizon-net

В docker-compose.cluster.yml нужно добавить Go-сервисы (Auth, Game, Billing, Leaderboard, Gateway) как контейнеры.

Но это потребует переписывания Makefile и запуска сервисов через Docker.

Вариант Б: Использовать host.docker.internal с правильной настройкой Docker

На Linux для работы host.docker.internal нужно добавить флаг при запуске Docker:

bash
# Добавить в /etc/docker/daemon.json
{
  "extra_hosts": ["host.docker.internal:host-gateway"]
}
Или в docker-compose:

yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  container_name: event-horizon-jaeger
  ports:
    - "16686:16686"
    - "4318:4318"
  extra_hosts:
    - "host.docker.internal:host-gateway"
  networks:
    - event-horizon-net

Вариант В: Использовать IP-адрес хоста (правильный)

Если host.docker.internal не работает на Linux, нужно использовать реальный IP хоста, но не через sed — через переменную окружения.

Добавь в каждый main.go чтение из переменной:
go
jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
if jaegerEndpoint == "" {
    jaegerEndpoint = "localhost:4317"
}
exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint(jaegerEndpoint),
    otlptracegrpc.WithInsecure(),
)
В Makefile или restart.sh задай переменную:
bash
export JAEGER_ENDPOINT=$(ip route get 1 | awk '{print $NF;exit}'):4317
💡 Самый простой (и правильный) вариант — добавить сеть и запускать всё через Docker

Но это потребует времени. Давай пока сделаем временный костыль, чтобы Jaeger заработал:

bash
# Узнать IP хоста
HOST_IP=$(ip route get 1 | awk '{print $NF;exit}')
echo "HOST_IP=$HOST_IP"

# Заменить во всех main.go на этот IP
find services -name "main.go" -exec sed -i "s/host.docker.internal/$HOST_IP/g" {} \;
Это временное решение, пока мы не переведём все сервисы в Docker.


--

Ошибка: produced zero addresses — Auth не может разрешить host.docker.internal, потому что Go-сервис запущен на хосте, а не в Docker-контейнере.

🤯 Проблема

Jaeger в Docker — знает host.docker.internal
Auth на хосте — не знает host.docker.internal
Go на хосте пытается резолвить host.docker.internal → ничего не находит
🎯 Решение: использовать IP хоста для Go-сервисов

Шаг 1: Узнать IP хоста

bash
ip route get 1 | awk '{print $NF;exit}'
# Например: 192.168.1.100 или 10.0.2.15
Шаг 2: В коде использовать переменную окружения

В каждом main.go замени жёстко зашитый адрес на чтение из os.Getenv:

go
jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
if jaegerEndpoint == "" {
    jaegerEndpoint = "localhost:4317"
}

exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint(jaegerEndpoint),
    otlptracegrpc.WithInsecure(),
)
Шаг 3: В restart.sh или Makefile задать правильный IP

bash
export JAEGER_ENDPOINT=$(ip route get 1 | awk '{print $NF;exit}'):4317
💨 Быстрое решение (прямо сейчас)

bash
# 1. Узнай IP хоста
HOST_IP=$(ip route get 1 | awk '{print $NF;exit}')
echo "HOST_IP=$HOST_IP"

# 2. Заменить host.docker.internal на IP в main.go
find services -name "main.go" -exec sed -i "s/host.docker.internal/$HOST_IP/g" {} \;

# 3. Пересобрать и запустить
make all

# 4. Проверить логи
tail -f /tmp/auth.log

=====================

# 1. Узнай IP хоста
HOST_IP=$(ip route get 1 | awk '{print $NF;exit}')
echo "HOST_IP=$HOST_IP"

# 2. Заменить host.docker.internal на IP в main.go
find services -name "main.go" -exec sed -i "s/host.docker.internal/$HOST_IP/g" {} \;

# 3. Пересобрать и запустить
make all

# 4. Проверить логи
tail -f /tmp/auth.log

=====================

Auth пытается подключиться к 172.17.0.1:4317, но получает connection refused
ss -tlnp | grep 4317 — ничего не показывает (порт не слушается на хосте)
Jaeger в Docker слушает на [::]:4317 внутри контейнера, но порт не проброшен на хост!
🚨 Проблема: порт 4317 не проброшен в docker-compose

В docker-compose.cluster.yml у Jaeger есть только:

yaml
ports:
  - "16686:16686"   # UI
  - "4318:4318"     # OTLP HTTP
А порт 4317 (gRPC OTLP) НЕ ПРОБРОШЕН!

🔧 Решение: добавить порт 4317

Обнови docker-compose.cluster.yml:

yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  container_name: event-horizon-jaeger
  ports:
    - "16686:16686"   # UI
    - "4318:4318"     # OTLP HTTP
    - "4317:4317"     # OTLP gRPC  👈 ДОБАВИТЬ
  environment:
    - COLLECTOR_OTLP_ENABLED=true
  restart: unless-stopped
  extra_hosts:
    - "host.docker.internal:host-gateway"
  networks:
    - event-horizon-net