#!/bin/bash

echo "🔥 Запуск нагрузки k6 (50 VUs, 60 секунд)..."
cd ~/event_horizon/deployments/k6
k6 run --vus 50 --duration 60s e2e-test.js > /tmp/k6_load.log 2>&1 &

K6_PID=$!

# Ждём 5 секунд, чтобы нагрузка разогрелась
sleep 5

echo "📊 Снятие CPU профиля (30 секунд)..."
curl -s "http://localhost:9092/debug/pprof/profile?seconds=30" > /tmp/cpu_under_load.prof &
PROFILE_PID=$!

# Ждём завершения профиля
wait $PROFILE_PID

# Ждём завершения k6
wait $K6_PID

echo "✅ Нагрузка и профиль завершены!"

echo ""
echo "📊 Анализ CPU профиля:"
go tool pprof -top /tmp/cpu_under_load.prof | head -30

echo ""
echo "📊 Анализ Heap профиля:"
curl -s "http://localhost:9092/debug/pprof/heap" > /tmp/heap_under_load.prof
go tool pprof -top /tmp/heap_under_load.prof | head -30
