#!/bin/bash
# Быстрый перезапуск всех сервисов после перезагрузки системы

set -e

cd ~/event_horizon

echo "🛑 Stopping old processes..."
pkill -f "auth-service" || true
pkill -f "leaderboard-service" || true
pkill -f "game-service" || true
pkill -f "billing-service" || true
pkill -f "gateway" || true

echo "🐳 Starting Docker containers..."
docker-compose -f deployments/docker-compose.cluster.yml up -d

echo "🚀 Starting services..."
cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service &
cd ../leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service &
cd ../game && go build -o game-service ./cmd/main.go && ./game-service &
cd ../billing && go build -o billing-service ./cmd/main.go && ./billing-service &
cd ../gateway && go build -o gateway ./cmd/main.go && ./gateway &

cd ~/event_horizon
echo "✅ All services started!"
echo ""
echo "Check logs:"
echo "  tail -f /tmp/auth.log"
echo "  tail -f /tmp/game.log"
echo "  tail -f /tmp/billing.log"
echo "  tail -f /tmp/leaderboard.log"
echo "  tail -f /tmp/gateway.log"