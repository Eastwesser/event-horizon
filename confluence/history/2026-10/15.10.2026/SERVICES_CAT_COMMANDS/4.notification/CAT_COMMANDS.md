4) notification — простой + воркер.

notification
bash
find services/notification -type f -name '*.go' | sort
echo "===== main ====="
cat services/notification/cmd/main.go
echo "===== app ====="
for f in services/notification/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/notification/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/notification/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/notification/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/notification/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done