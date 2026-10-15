17) mcp — MCP + RAG.

mcp
bash
find services/mcp -type f -name '*.go' | sort
echo "===== main ====="
cat services/mcp/cmd/main.go
echo "===== config ====="
for f in services/mcp/internal/config/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== natsx ====="
for f in services/mcp/internal/natsx/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== pgxtool ====="
for f in services/mcp/internal/pgxtool/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== rag ====="
for f in services/mcp/internal/rag/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== redisx ====="
for f in services/mcp/internal/redisx/*.go; do echo "--- $f ---"; cat "$f"; done
