#!/bin/bash
cd ~/event_horizon

echo "Stopping old processes..."
pkill -f "auth-service" || true
pkill -f "leaderboard-service" || true
pkill -f "game-service" || true
pkill -f "billing-service" || true
pkill -f "gateway" || true

echo "Starting Docker containers..."
docker-compose -f deployments/docker-compose.cluster.yml up -d

echo "Building and starting services..."
cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service > /tmp/auth.log 2>&1 &
cd ~/event_horizon/services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
cd ~/event_horizon/services/game && go build -o game-service ./cmd/main.go && ./game-service > /tmp/game.log 2>&1 &
cd ~/event_horizon/services/billing && go build -o billing-service ./cmd/main.go && ./billing-service > /tmp/billing.log 2>&1 &
cd ~/event_horizon/services/gateway && go build -o gateway ./cmd/main.go && ./gateway > /tmp/gateway.log 2>&1 &

sleep 3
echo "✅ Done!"
echo "Check logs: tail -f /tmp/auth.log"