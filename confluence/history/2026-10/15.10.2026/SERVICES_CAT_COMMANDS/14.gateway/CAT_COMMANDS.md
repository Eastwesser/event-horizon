14) gateway — HTTP + gRPC + WebSocket + ratelimit + circuit.

gateway
bash
find services/gateway -type f -name '*.go' | sort
echo "===== main ====="
cat services/gateway/cmd/main.go
echo "===== app ====="
for f in services/gateway/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== middleware ====="
for f in services/gateway/internal/middleware/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== ratelimit ====="
for f in services/gateway/internal/ratelimit/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== circuit ====="
for f in services/gateway/internal/circuit/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== cache ====="
for f in services/gateway/internal/cache/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== client ====="
for f in services/gateway/internal/client/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== dto ====="
for f in services/gateway/internal/dto/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== websocket ====="
for f in services/gateway/internal/websocket/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== webhook ====="
for f in services/gateway/internal/webhook/*.go; do echo "--- $f ---"; cat "$f"; done
