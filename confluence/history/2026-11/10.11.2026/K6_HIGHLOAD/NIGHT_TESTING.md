Про balancer_active_connections — давай посмотрим под нагрузкой
Ты прав — не надо верить на слово. Запусти P=200, и в момент прогона в втором терминале:

bash
# Каждые 2 секунды, 50 раз (100 секунд — покрывает ramp+hold)
for i in $(seq 1 50); do
  echo "$(date +%H:%M:%S) $(docker exec deployments-balancer-1 wget -qO- http://127.0.0.1:9098/metrics 2>/dev/null | grep '^balancer_active_connections ' | awk '{print $2}')"
  sleep 2
done
Это покажет реальное значение balancer_active_connections каждые 2 секунды во время прогона. Скинь — по нему точно скажу, сломана ли метрика или нет.

Плюс — проверь под нагрузкой все три реплики Gateway:

bash
for i in $(seq 1 50); do
  echo "$(date +%H:%M:%S) $(curl -s 'http://localhost:9090/api/v1/query?query=gateway_request_duration_seconds_count{path="/api/v1/shop/items"}' | jq -r '.data.result | map(.value[1]) | add // 0')"
  sleep 2
done
Это покажет растёт ли shop/items counter во время прогона. Если растёт — Prometheus скрейпит Gateway. Если нет — проблема в скрейпе.