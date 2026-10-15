# Структура
echo "===== tree ====="
find services/balancer -type f -name '*.go' | sort

# Main
echo "===== main ====="
cat services/balancer/cmd/main.go

# App
echo "===== app ====="
for f in services/balancer/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done

# Backend
echo "===== backend ====="
for f in services/balancer/internal/backend/*.go; do echo "--- $f ---"; cat "$f"; done

# Balancer
echo "===== balancer ====="
for f in services/balancer/internal/balancer/*.go; do echo "--- $f ---"; cat "$f"; done

# Metrics
echo "===== metrics ====="
for f in services/balancer/internal/metrics/*.go; do echo "--- $f ---"; cat "$f"; done

# Proxy
echo "===== proxy ====="
for f in services/balancer/internal/proxy/*.go; do echo "--- $f ---"; cat "$f"; done

# Config (если есть)
echo "===== config ====="
for f in services/balancer/internal/config/*.go; do echo "--- $f ---"; cat "$f"; done