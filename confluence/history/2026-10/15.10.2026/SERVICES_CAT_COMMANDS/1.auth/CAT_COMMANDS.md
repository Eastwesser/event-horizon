1) auth — простой, без воркеров.

# Структура
find services/auth -type f -name '*.go' | sort

# Main и app
cat services/auth/cmd/main.go
echo "===== app ====="
cat services/auth/internal/app/*.go

# Handlers
echo "===== handler ====="
for f in services/auth/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done

# Service
echo "===== service ====="
for f in services/auth/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done

# Repository
echo "===== repository ====="
for f in services/auth/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done

# Converter
echo "===== converter ====="
for f in services/auth/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done

# JWT
echo "===== jwt ====="
for f in services/auth/internal/jwt/*.go; do echo "--- $f ---"; cat "$f"; done

# Interceptor
echo "===== interceptor ====="
for f in services/auth/internal/interceptor/*.go; do echo "--- $f ---"; cat "$f"; done

# Config
echo "===== config ====="
for f in services/auth/internal/config/*.go; do echo "--- $f ---"; cat "$f"; done
