11) game — несколько игр + boost.

game
bash
find services/game -type f -name '*.go' | sort
echo "===== main ====="
cat services/game/cmd/main.go
echo "===== app ====="
for f in services/game/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/game/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/game/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/game/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/game/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/game/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
