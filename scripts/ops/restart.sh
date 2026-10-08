#!/bin/bash
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

# cd handled by REPO_ROOT
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
