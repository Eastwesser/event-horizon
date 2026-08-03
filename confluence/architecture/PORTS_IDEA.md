Самое главное — в docker-compose.cluster.yml для Inventory написано:

yaml
ports:
  - "50059:50059"   # Это gRPC
  - "9096:9096"     # Это метрики
А в config.go Inventory читает переменную INVENTORY_METRICS_PORT, которая равна 9096. ✅ Всё верно!

И да, ты заметил очень важную вещь: метрики Gateway (9095-9097) пересекаются с метриками Shop (9095) и Profile (9099). Это не пересечение, а разные порты:

Gateway-1: 9095, Gateway-2: 9096, Gateway-3: 9097.

Shop: 9095, Profile: 9099, Inventory: 9096.

Всё это разные порты на хосте, просто цифры совпадают, но они у разных сервисов.

🧠 ВЫВОД ПО ПОРТАМ
Кто	Что слушает	На каком порту хоста
Auth	gRPC + metrics	50051, 9091
Billing	gRPC + metrics	50053, 9093
Shop	gRPC + metrics	50055, 9095
Inventory	gRPC + metrics	50059, 9096
Profile	gRPC + metrics	50060, 9099
Gateway-1	HTTP + metrics	8081, 9095
Gateway-2	HTTP + metrics	8082, 9096
Gateway-3	HTTP + metrics	8083, 9097
Balancer	HTTP + metrics	8079, 9098

📝 ЗАПОМНИ (для собеседования)
«У нас все микросервисы имеют свои собственные порты для gRPC и метрик. В Docker Compose мы маппим их на хостовые порты. Inventory использует 50059 для gRPC и 9096 для метрик, чтобы не конфликтовать с Profile и Gateway.»

Чтобы чекнуть все порты:

echo "=== ВСЕ ПОРТЫ EVENT HORIZON ==="

echo -e "\n🔴 PostgreSQL:"
sudo ss -tlnp | grep -E "5460|5461|5462|5463|5464|5465"

echo -e "\n🔴 Redis:"
sudo ss -tlnp | grep -E "6379|6380|6381|6382|6383"

echo -e "\n🔴 gRPC:"
sudo ss -tlnp | grep -E "50051|50052|50053|50054|50055|50059|50060"

echo -e "\n🔴 Metrics:"
sudo ss -tlnp | grep -E "9091|9092|9093|9094|9095|9096|9097|9098|9099"

echo -e "\n🔴 HTTP:"
sudo ss -tlnp | grep -E "8079|8081|8082|8083"

echo -e "\n🔴 NATS:"
sudo ss -tlnp | grep -E "4222|4223|4224|8222|8223|8224"

echo -e "\n🔴 Мониторинг:"
sudo ss -tlnp | grep -E "9090|3000|16686|7777|9187|9121"