✅ Проверь, что Stream EVENTS работает в кластере
bash
# Проверить, что Stream создан
docker exec event-horizon-nats-1 nats-server --help 2>/dev/null || echo "nats cli not available"

# Альтернатива — проверить через curl на каждой ноде
curl -s http://localhost:8222/varz | jq '.server_id'
curl -s http://localhost:4223/varz | jq '.server_id'
curl -s http://localhost:4224/varz | jq '.server_id'

# Проверить, что Stream EVENTS существует (через nats-hub)
docker logs event-horizon-nats-hub --tail=5 | grep "Stream"

🧠 Если хочешь убедиться, что кластер работает
bash
# Посмотреть, что NATS видит другие ноды
curl -s http://localhost:8222/varz | jq '.cluster.connections'
Если там есть записи о nats-2 и nats-3 — кластер работает.