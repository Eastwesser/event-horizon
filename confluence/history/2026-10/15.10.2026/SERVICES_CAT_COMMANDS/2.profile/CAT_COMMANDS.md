2) profile — простой + миграции.

profile
bash
find services/profile -type f -name '*.go' | sort
echo "===== main ====="
cat services/profile/cmd/main.go
echo "===== app ====="
for f in services/profile/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/profile/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/profile/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/profile/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/profile/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done