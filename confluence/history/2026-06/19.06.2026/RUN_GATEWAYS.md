chmod +x scripts/start_gateways.sh

cd /home/denismatveev/event_horizon
./scripts/start_gateways.sh

🧪 Проверяем

bash
# Проверить, что все 3 Gateway работают
curl -s http://localhost:8081/health && echo " ✅ 8081"
curl -s http://localhost:8082/health && echo " ✅ 8082"
curl -s http://localhost:8083/health && echo " ✅ 8083"

# Проверить балансировщик
curl -s http://localhost:8079/health && echo " ✅ Balancer"

# Посмотреть логи балансировщика (если запущен)
ps aux | grep balancer



--

🚀 Запускаем Gateway вручную на трёх портах

bash
cd /home/denismatveev/event_horizon/services/gateway

# Убедимся, что бинарник существует
ls -la gateway

# Запускаем три экземпляра (каждый в своей сессии или с &)
PORT=8081 ./gateway > /tmp/gateway_8081.log 2>&1 &
PORT=8082 ./gateway > /tmp/gateway_8082.log 2>&1 &
PORT=8083 ./gateway > /tmp/gateway_8083.log 2>&1 &

PORT=8081 /home/denismatveev/event_horizon/services/gateway/gateway > /tmp/gateway_8081.log 2>&1 &
PORT=8082 /home/denismatveev/event_horizon/services/gateway/gateway > /tmp/gateway_8082.log 2>&1 &
PORT=8083 /home/denismatveev/event_horizon/services/gateway/gateway > /tmp/gateway_8083.log 2>&1 &

# Проверяем, что процессы запустились
ps aux | grep gateway

# Проверяем, что порты слушают
ss -tlnp | grep -E "8081|8082|8083"

# Проверяем health
curl -s http://localhost:8081/health && echo " ✅ 8081 OK"
curl -s http://localhost:8082/health && echo " ✅ 8082 OK"
curl -s http://localhost:8083/health && echo " ✅ 8083 OK"
🔧 Если Gateway не принимает флаг -port или переменную PORT:

Нужно убедиться, что в services/gateway/cmd/main.go есть поддержка порта:

go
// В main() добавить:
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

srv := &http.Server{
    Addr:    ":" + port,
    Handler: r,
}
Если нет — добавь и пересобери:

bash
cd /home/denismatveev/event_horizon/services/gateway
go build -o gateway ./cmd/main.go
🧪 После запуска Gateway — проверяем балансировщик

bash
for i in {1..10}; do
  curl -s http://localhost:8079/health
  echo ""
done
Должны увидеть {"status":"ok"} 10 раз.


---


cd /home/denismatveev/event_horizon

# Остановить всё
pkill -f gateway || true
pkill -f balancer || true
sleep 1

# Запустить Gateway
./scripts/start_gateways.sh

# Запустить балансировщик
./scripts/start_balancer.sh

# Проверить
curl -s http://localhost:8079/health && echo " ✅ All systems go!"

----

🔧 Останавливаем всё

bash
pkill -f gateway
pkill -f balancer
sleep 1
ss -tlnp | grep -E "8081|8082|8083|8079"  # должно быть пусто
🚀 Запускаем заново (по порядку)

bash
cd /home/denismatveev/event_horizon/services/gateway

# Терминал 1 (или фон)
PORT=8081 ./gateway > /tmp/gateway_8081.log 2>&1 &

# Терминал 2
PORT=8082 ./gateway > /tmp/gateway_8082.log 2>&1 &

# Терминал 3
PORT=8083 ./gateway > /tmp/gateway_8083.log 2>&1 &

# Проверить
sleep 2
curl -s http://127.0.0.1:8081/health && echo " ✅ 8081"
curl -s http://127.0.0.1:8082/health && echo " ✅ 8082"
curl -s http://127.0.0.1:8083/health && echo " ✅ 8083"
🧪 Проверяем балансировщик

bash
cd /home/denismatveev/event_horizon/services/balancer
./balancer > /tmp/balancer.log 2>&1 &

# Проверить
curl -s http://127.0.0.1:8079/health