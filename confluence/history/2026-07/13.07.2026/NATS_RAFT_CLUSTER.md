🎯 Почему нам нужно мониторить ВСЕ 3 ноды

Лидер меняется - NATS кластер использует Raft для выборов лидера. Лидером может стать любая из 3 нод (nats-1, nats-2 или nats-3).

Отказоустойчивость - если одна нода падает, кластер продолжает работать, но мы должны знать какая упала.

Диагностика - если проблемы с производительностью, нужно видеть какая нода перегружена.

🔍 Как NATS показывает лидера

bash
# Проверяем какая нода лидер
curl -s http://localhost:8222/varz | jq '.leader'

# Или через кластерный API
curl -s http://localhost:8222/varz | jq '.cluster.leader'
В ответе будет server_id лидера.

✅ Правильное решение - 3 экспортера с правильными портами

Давай запустим 3 экспортера правильно:

bash
# Удаляем все старые
docker rm -f nats-exporter-1 nats-exporter-2 nats-exporter-3 2>/dev/null

# Экспортер для nats-1 (порт 7777)
docker run -d --name nats-exporter-1 \
  --network deployments_event-horizon-net \
  -p 7777:7777 \
  natsio/prometheus-nats-exporter:latest \
  -listen :7777 \
  -varz http://nats-1:8222

# Экспортер для nats-2 (порт 7778)
docker run -d --name nats-exporter-2 \
  --network deployments_event-horizon-net \
  -p 7778:7778 \
  natsio/prometheus-nats-exporter:latest \
  -listen :7778 \
  -varz http://nats-2:8222

# Экспортер для nats-3 (порт 7779)
docker run -d --name nats-exporter-3 \
  --network deployments_event-horizon-net \
  -p 7779:7779 \
  natsio/prometheus-nats-exporter:latest \
  -listen :7779 \
  -varz http://nats-3:8222

# Проверяем что все работают
curl -s http://localhost:7777/metrics | head -3
curl -s http://localhost:7778/metrics | head -3
curl -s http://localhost:7779/metrics | head -3

📊 В Prometheus будет 3 таргета

yaml
- job_name: 'nats'
  static_configs:
    - targets: 
        - 'nats-exporter-1:7777'
        - 'nats-exporter-2:7778'
        - 'nats-exporter-3:7779'

🔍 Как узнать кто лидер
В Grafana можно сделать дашборд с запросом:

promql
gnatsd_varz_leader{server_id="http://nats-1:8222"}

Или через метрику gnatsd_varz_cluster_leader.

В итоге: 3 экспортера = 3 точки мониторинга = видим кто лидер, кто упал, кто перегружен. 🚀

