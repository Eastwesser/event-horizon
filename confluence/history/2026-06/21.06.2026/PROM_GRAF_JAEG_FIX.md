Что сейчас есть:

Prometheus — запущен на 9090
Grafana — запущена на 3000
Jaeger — запущен на 16686
Prometheus targets — прописаны для всех сервисов
Проблемы:

Prometheus не видит метрики — потому что в prometheus.yml используются auth:50051, но метрики отдаются на других портах:
Auth: 9091
Billing: 9093
Game: 9092
Leaderboard: 9094
Gateway: 9095, 9096, 9097 (но они на хосте, а не в Docker-сети)
Gateway метрики — ты используешь host.docker.internal:9095, но gateway запущен в Docker, а не локально.
Исправляем prometheus.yml:

bash
cd /home/denismatveev/event_horizon

cat > deployments/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'auth'
    static_configs:
      - targets: ['auth:9091']
    metrics_path: '/metrics'

  - job_name: 'billing'
    static_configs:
      - targets: ['billing:9093']
    metrics_path: '/metrics'

  - job_name: 'game'
    static_configs:
      - targets: ['game:9092']
    metrics_path: '/metrics'

  - job_name: 'leaderboard'
    static_configs:
      - targets: ['leaderboard:9094']
    metrics_path: '/metrics'

  - job_name: 'gateway'
    static_configs:
      - targets: 
          - 'gateway:9095'
          - 'gateway-2:9096'
          - 'gateway-3:9097'

  - job_name: 'balancer'
    static_configs:
      - targets: ['balancer:8079']
    metrics_path: '/metrics'

  - job_name: 'nats'
    static_configs:
      - targets: ['nats:8222']
    metrics_path: '/metrics'
EOF
Перезапускаем Prometheus:

bash
docker-compose -f deployments/docker-compose.cluster.yml restart prometheus
Проверяем метрики:

bash
# 1. Проверяем, что метрики отдаются
curl http://localhost:9091/metrics | head -5
curl http://localhost:9092/metrics | head -5
curl http://localhost:9093/metrics | head -5
curl http://localhost:9094/metrics | head -5

# 2. Проверяем, что Prometheus видит targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
Настраиваем Grafana:

Открываешь http://localhost:3000
Логин: admin / admin
Добавляешь источник данных:
Type: Prometheus
URL: http://prometheus:9090
Сохраняешь
Импортируешь дашборд или создаёшь свой
Проверяем Jaeger:

Открываешь http://localhost:16686
Выбираешь сервис (gateway, auth, billing, etc.)
Смотришь трассы
Покажи что выдает curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'. Нужно убедиться, что все targets UP.