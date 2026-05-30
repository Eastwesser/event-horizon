#!/bin/bash
cd ~/event_horizon
echo "🛑 Stopping services..."
make stop-services
echo "🐳 Restarting Docker..."
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d
echo "🚀 Starting Go services..."
make all
echo "✅ Status:"
make status
echo "🌐 Gateway health:"
curl -s http://localhost:8080/health
