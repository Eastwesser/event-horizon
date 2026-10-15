16) nats-hub — хаб.

nats-hub
bash
find services/nats-hub -type f -name '*.go' | sort
echo "===== main ====="
cat services/nats-hub/cmd/main.go
echo "===== app ====="
for f in services/nats-hub/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
