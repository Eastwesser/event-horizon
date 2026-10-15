8) shop — сложный, миграции + саги.

shop
bash
find services/shop -type f -name '*.go' | sort
echo "===== main ====="
cat services/shop/cmd/main.go
echo "===== app ====="
for f in services/shop/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/shop/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/shop/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/shop/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/shop/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/shop/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
