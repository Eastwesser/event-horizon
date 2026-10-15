3) history — простой + воркер.

history
bash
find services/history -type f -name '*.go' | sort
echo "===== main ====="
cat services/history/cmd/main.go
echo "===== app ====="
for f in services/history/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/history/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/history/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/history/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/history/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== validate ====="
for f in services/history/internal/validate/*.go; do echo "--- $f ---"; cat "$f"; done