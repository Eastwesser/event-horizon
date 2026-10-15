10) billing — outbox + worker.

billing
bash
find services/billing -type f -name '*.go' | sort
echo "===== main ====="
cat services/billing/cmd/main.go
echo "===== app ====="
for f in services/billing/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/billing/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/billing/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/billing/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/billing/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/billing/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
