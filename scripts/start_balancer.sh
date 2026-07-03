#!/bin/bash
cd /home/denismatveev/event_horizon/services/balancer
pkill -f "balancer" || true
./balancer > /tmp/balancer.log 2>&1 &
echo "✅ Balancer started on :8079"
