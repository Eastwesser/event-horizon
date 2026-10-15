5) analytics — простой + воркер.

analytics
bash
find services/analytics -type f -name '*.go' | sort
echo "===== main ====="
cat services/analytics/cmd/main.go
echo "===== app ====="
for f in services/analytics/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/analytics/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/analytics/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/analytics/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/analytics/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
