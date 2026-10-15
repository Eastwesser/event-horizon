6) leaderboard — простой.

leaderboard
bash
find services/leaderboard -type f -name '*.go' | sort
echo "===== main ====="
cat services/leaderboard/cmd/main.go
echo "===== app ====="
for f in services/leaderboard/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/leaderboard/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/leaderboard/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/leaderboard/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done


