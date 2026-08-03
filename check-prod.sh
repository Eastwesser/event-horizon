#!/bin/bash
echo "=== 1. Optimistic Locking (version) ==="
grep -rn "Version" --include="*.go" services/*/internal/ 2>/dev/null | wc -l

echo "=== 2. Транзакции (BeginTx) ==="
grep -rn "BeginTx" --include="*.go" services/ 2>/dev/null | wc -l

echo "=== 3. Индексы (CREATE INDEX) ==="
grep -rn "CREATE INDEX" --include="*.sql" services/*/migrations/ 2>/dev/null | wc -l

echo "=== 4. Redis (SetBalance) ==="
grep -rn "SetBalance" --include="*.go" services/billing/ 2>/dev/null | wc -l

echo "=== 5. Миграции Down ==="
grep -rn "+goose Down" --include="*.sql" services/*/migrations/ 2>/dev/null | wc -l

echo "=== 6. Health check ==="
grep -rn "/health" --include="*.go" services/ 2>/dev/null | wc -l