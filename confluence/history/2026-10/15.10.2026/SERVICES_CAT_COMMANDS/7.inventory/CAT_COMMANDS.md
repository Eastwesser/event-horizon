7) inventory — есть seed, конвертер.

inventory
bash
find services/inventory -type f -name '*.go' | sort
echo "===== main ====="
cat services/inventory/cmd/main.go
echo "===== app ====="
for f in services/inventory/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/inventory/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/inventory/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/inventory/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/inventory/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== seed ====="
for f in services/inventory/internal/seed/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/inventory/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
