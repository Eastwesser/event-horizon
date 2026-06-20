# scripts/start_gateways.sh
#!/bin/bash
cd /home/denismatveev/event_horizon/services/gateway

pkill -f "gateway" || true
sleep 1

PORT=8081 ./gateway > /tmp/gateway_8081.log 2>&1 &
PORT=8082 ./gateway > /tmp/gateway_8082.log 2>&1 &
PORT=8083 ./gateway > /tmp/gateway_8083.log 2>&1 &

sleep 2
curl -s http://127.0.0.1:8081/health && echo " ✅ 8081"
curl -s http://127.0.0.1:8082/health && echo " ✅ 8082"
curl -s http://127.0.0.1:8083/health && echo " ✅ 8083"