13) fulfillment — fulfillment.

fulfillment
bash
find services/fulfillment -type f -name '*.go' | sort
echo "===== main ====="
cat services/fulfillment/cmd/main.go
echo "===== app ====="
for f in services/fulfillment/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/fulfillment/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
