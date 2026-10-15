9) payment — webhook + worker.

payment
bash
find services/payment -type f -name '*.go' | sort
echo "===== main ====="
cat services/payment/cmd/main.go
echo "===== app ====="
for f in services/payment/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/payment/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/payment/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/payment/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== webhook ====="
for f in services/payment/internal/webhook/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/payment/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
