#!/bin/bash
echo "Testing PostgreSQL..."
PGPASSWORD=eventhorizon psql -h localhost -p 5433 -U eventhorizon -d eventhorizon -c "SELECT 1" && echo "✅ PostgreSQL OK"

echo "Testing Redis..."
redis-cli -h localhost -p 6379 ping && echo "✅ Redis OK"

echo "Testing NATS..."
nats-server --version && echo "✅ NATS CLI OK"