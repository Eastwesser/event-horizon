#!/bin/bash
set -e

echo "🔍 Checking Postgres..."
docker exec event-horizon-postgres pg_isready -U eventhorizon

echo "🔍 Checking Redis..."
docker exec event-horizon-redis redis-cli ping

echo "🔍 Checking NATS..."
docker exec event-horizon-nats nats-server --version

echo "🔍 Checking NATS JetStream..."
docker logs event-horizon-nats 2>&1 | grep -q "JetStream is enabled" && echo "✅ JetStream enabled" || echo "⚠️ JetStream not enabled"

echo "🎉 All systems operational!"