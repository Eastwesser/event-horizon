12) authors — applications + portfolio.

authors
bash
find services/authors -type f -name '*.go' | sort
echo "===== main ====="
cat services/authors/cmd/main.go
echo "===== app ====="
for f in services/authors/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/authors/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/authors/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/authors/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== validate ====="
for f in services/authors/internal/validate/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/authors/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done