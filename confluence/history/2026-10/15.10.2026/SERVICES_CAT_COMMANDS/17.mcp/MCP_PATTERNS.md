Разбор mcp
Что делает сервис
mcp — это MCP-сервер (Model Context Protocol) для AI-ассистентов:

6 tools (инструментов) для AI:

nats_list_streams — список JetStream streams.

nats_list_consumers — список consumers для stream.

postgres_query — read-only SQL.

redis_get — GET ключа.

redis_keys — SCAN ключей.

search_prydwen — RAG-поиск по knowledge base.

RAG без embedding API — offline TF-IDF.

SQL Sandbox — только SELECT / WITH.

Graceful Degradation — тулы деградируют с понятной ошибкой, если соединение недоступно.

stdio transport — общение через stdin/stdout (MCP-стандарт).

Найденные паттерны
1. MCP Server (главный паттерн)
Файл: cmd/main.go.

Почему: mcp — это MCP-сервер через mark3labs/mcp-go. Каждый tool — это JSON-RPC метод, доступный AI-ассистенту.

Код:

go
s := server.NewMCPServer(
    "event-horizon-mcp",
    "1.0.0",
    server.WithToolCapabilities(true),
)

s.AddTool(mcp.NewTool("nats_list_streams",
    mcp.WithDescription("List JetStream streams (active NATS messaging topology for Event Horizon)"),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ...
})
MCP Server — новый паттерн для AI-интеграции. Уникально для mcp.

2. Tool Registration (6 tools)
Файл: cmd/main.go.

Почему: Регистрация 6 инструментов через s.AddTool. Каждый tool имеет:

Имя (nats_list_streams).

Описание.

Параметры (required/optional).

Handler.

Код:

go
s.AddTool(mcp.NewTool("postgres_query",
    mcp.WithDescription("Run a read-only SELECT/WITH query against Event Horizon Postgres (MCP_POSTGRES_DSN). Mutations forbidden."),
    mcp.WithString("sql", mcp.Required(), mcp.Description("SELECT statement without semicolon")),
    mcp.WithNumber("limit", mcp.Description("Max rows (default 100, max 500)")),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ...
})
Tool Registration — декларативная регистрация инструментов.

3. Graceful Degradation (per-tool)
Файл: cmd/main.go.

Почему: Каждый tool проверяет, доступен ли клиент:

go
if natsClient == nil {
    return mcp.NewToolResultError("NATS not connected — set NATS_URL"), nil
}
if pg == nil {
    return mcp.NewToolResultError("Postgres not connected — set MCP_POSTGRES_DSN"), nil
}
if rdb == nil {
    return mcp.NewToolResultError("Redis not connected — set REDIS_ADDR"), nil
}
if idx == nil || idx.N == 0 {
    return mcp.NewToolResultError("RAG index empty — set PRYDWEN_ROOT to confluence/agents/prydwen_knowledge"), nil
}
Graceful Degradation — сервер стартует всегда, даже если все соединения недоступны. Каждый tool даёт точную ошибку с подсказкой, что настроить.

Уникально для mcp — в других сервисах деградация сервиса целиком, а тут — по инструментам.

4. Optional Dependencies (все клиенты опциональны)
Файл: cmd/main.go.

Почему: NATS, Postgres, Redis — все опциональны. Сервер стартует всегда:

go
var natsClient *natsx.Client
if nc, err := natsx.Connect(cfg.NATSURL); err != nil {
    log.Printf("nats warn: %v", err)
} else {
    natsClient = nc
    defer natsClient.Close()
}
Optional Dependencies — все зависимости soft.

5. SQL Sandbox (ValidateSelect)
Файл: internal/pgxtool/client.go, ValidateSelect.

Почему: postgres_query — read-only. Валидация запрещает:

; (multiple statements).

Комментарии (--, /*).

Всё, что не начинается с SELECT / WITH.

Ключевые слова: INSERT, UPDATE, DELETE, DROP, ALTER, CREATE, TRUNCATE, GRANT, REVOKE, COPY, CALL, DO, INTO, UNION, EXCEPT, INTERSECT.

Код:

go
func ValidateSelect(sql string) error {
    s := strings.TrimSpace(sql)
    if s == "" {
        return fmt.Errorf("empty query")
    }
    if strings.Contains(s, ";") {
        return fmt.Errorf("multiple statements / trailing semicolon not allowed")
    }
    lower := strings.ToLower(s)
    if !(strings.HasPrefix(lower, "select") || strings.HasPrefix(lower, "with")) {
        return fmt.Errorf("only SELECT / WITH (read-only) queries are allowed")
    }
    forbidden := []string{
        " insert ", " update ", " delete ", " drop ", " alter ", " create ",
        " truncate ", " grant ", " revoke ", " copy ", " call ", " do ",
        " into ", " union ", " except ", " intersect ",
    }
    padded := " " + lower + " "
    for _, f := range forbidden {
        if strings.Contains(padded, f) {
            return fmt.Errorf("forbidden keyword in query: %s", strings.TrimSpace(f))
        }
    }
    return nil
}
SQL Sandbox — защита от инъекций через AI-запросы. Уникально для mcp.

Плюс: SQL оборачивается в подзапрос с LIMIT, чтобы гарантировать ограничение:

go
wrapped := fmt.Sprintf("SELECT * FROM (%s) AS _mcp_q LIMIT %d", sql, limit)
6. Offline RAG (TF-IDF без embedding API)
Файл: internal/rag/index.go.

Почему: RAG реализован без embedding API через TF-IDF + cosine similarity. Работает offline, не требует OpenAI/Anthropic API.

Код:

go
// Offline TF-IDF RAG over Prydwen markdown — no embedding API required.
// Good enough for "how does Outbox work in Shop?" style retrieval.

type Index struct {
    Chunks []Chunk              `json:"chunks"`
    DF     map[string]int       `json:"df"`     // document frequency
    TF     []map[string]float64 `json:"tf"`     // term frequency per chunk
    N      int                  `json:"n"`
}
Offline RAG — RAG без зависимости от внешнего API. Уникально для mcp.

7. TF-IDF Scoring (cosine similarity)
Файл: internal/rag/index.go, Search.

Код:

go
for i := range idx.Chunks {
    var dot, nq, nd float64
    for t, qv := range qtf {
        idf := math.Log(1 + float64(idx.N)/(1+float64(idx.DF[t])))
        qw := qv * idf
        dw := idx.TF[i][t] * idf
        dot += qw * dw
        nq += qw * qw
        nd += dw * dw
    }
    // ...
    cos := dot / (math.Sqrt(nq) * math.ND(nd))
    if cos > 0 {
        scores = append(scores, scored{i: i, s: cos})
    }
}
TF-IDF + Cosine Similarity — классический Information Retrieval.

8. Chunking Strategy (splitChunks)
Файл: internal/rag/index.go, splitChunks.

Почему: Markdown разбивается на чанки по 1200 рун, по границам параграфов:

go
func splitChunks(s string, maxRunes int) []string {
    s = strings.ReplaceAll(s, "\r\n", "\n")
    paras := strings.Split(s, "\n\n")
    // ...
    for _, p := range paras {
        // ...
        if size > 0 && size+r > maxRunes {
            flush()
        }
        // ...
    }
}
Chunking Strategy — разбиение на чанки по параграфам с ограничением размера.

9. Index Caching (Load/Save/EnsureIndex)
Файл: internal/rag/index.go, EnsureIndex.

Почему: Индекс кешируется на диск в JSON. При старте EnsureIndex сначала пробует загрузить кеш, потом строит:

go
func EnsureIndex(root, cachePath string) (*Index, error) {
    if cachePath != "" {
        if idx, err := Load(cachePath); err == nil && idx.N > 0 {
            return idx, nil
        }
    }
    idx, err := BuildFromDir(root)
    // ...
    if cachePath != "" {
        _ = idx.Save(cachePath)
    }
    return idx, nil
}
Index Caching — долгоживущий кеш индекса.

10. Path Resolution (resolvePrydwen)
Файл: cmd/main.go, resolvePrydwen.

Почему: Путь к knowledge base ищется по нескольким кандидатам (начиная с CWD):

go
func resolvePrydwen(configured string) string {
    candidates := []string{configured}
    if wd, err := os.Getwd(); err == nil {
        candidates = append(candidates,
            filepath.Join(wd, configured),
            filepath.Join(wd, "confluence/agents/prydwen_knowledge"),
            filepath.Join(wd, "../confluence/agents/prydwen_knowledge"),
            filepath.Join(wd, "../../confluence/agents/prydwen_knowledge"),
        )
    }
    // ...
}
Path Resolution — гибкий поиск пути.

11. Adapter (natsx, pgxtool, redisx)
Файлы: internal/natsx/client.go, internal/pgxtool/client.go, internal/redisx/client.go.

Почему: Каждый пакет — адаптер над внешней библиотекой:

natsx — над nats.go.

pgxtool — над pgxpool.

redisx — над redis/v9.

Adapter Pattern — три адаптера в одном сервисе.

12. stdio Transport
Файл: cmd/main.go.

Код:

go
if err := server.ServeStdio(s); err != nil {
    fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
    os.Exit(1)
}
stdio Transport — стандарт для MCP. Сервер общается через stdin/stdout.

13. Config with Optional DSNs
Файл: internal/config/config.go.

Код:

go
type Config struct {
    NATSURL       string
    PostgresDSN   string // prefer a read-only role
    RedisAddr     string
    PrydwenRoot   string
    RAGIndexPath  string
}
Config with Optional DSNs — все соединения опциональны.

Плюс: Postgres DSN fallback на DATABASE_URL:

go
PostgresDSN: getenv("MCP_POSTGRES_DSN", getenv("DATABASE_URL", "")),
14. Testable SQL Validator
Файл: internal/pgxtool/client_test.go.

Код:

go
func TestValidateSelect_RejectsMutation(t *testing.T) {
    cases := []string{
        "DELETE FROM users",
        "SELECT * FROM users; DROP TABLE users",
        "UPDATE users SET role='admin'",
        "INSERT INTO users VALUES (1)",
    }
    for _, c := range cases {
        if err := ValidateSelect(c); err == nil {
            t.Fatalf("expected reject for %q", c)
        }
    }
}
Testable SQL Validator — тесты на SQL injection.

15. Testable RAG
Файл: internal/rag/index_test.go.

Код:

go
func TestSearchFindsOutbox(t *testing.T) {
    dir := t.TempDir()
    // ... пишем markdown про Outbox ...
    idx, err := BuildFromDir(dir)
    hits := idx.Search("how does Outbox work in Shop", 3)
    if len(hits) == 0 {
        t.Fatal("expected at least one hit")
    }
}
Testable RAG — тест проверяет, что RAG находит Outbox.

Итог по mcp
Паттерн	Где
MCP Server	cmd/main.go, server.NewMCPServer
Tool Registration (6 tools)	cmd/main.go, s.AddTool
Graceful Degradation (per-tool)	cmd/main.go
Optional Dependencies	cmd/main.go
SQL Sandbox	pgxtool/client.go, ValidateSelect
Offline RAG (TF-IDF)	rag/index.go
TF-IDF Scoring (cosine similarity)	rag/index.go, Search
Chunking Strategy	rag/index.go, splitChunks
Index Caching	rag/index.go, EnsureIndex
Path Resolution	cmd/main.go, resolvePrydwen
Adapter (natsx, pgxtool, redisx)	internal/natsx/, internal/pgxtool/, internal/redisx/
stdio Transport	cmd/main.go, server.ServeStdio
Config with Optional DSNs	internal/config/config.go
Testable SQL Validator	pgxtool/client_test.go
Testable RAG	rag/index_test.go
Итого 15 паттернов и приёмов.

Уникальные фичи mcp:

MCP Server — единственный сервис для AI-интеграции.

Tool Registration — декларативные инструменты.

Graceful Degradation per-tool — сервер работает, даже если всё упало.

SQL Sandbox — защита от injection.

Offline RAG (TF-IDF) — RAG без embedding API.

stdio Transport — MCP-стандарт.

Готовый markdown для MCP_PATTERNS.md
md
# mcp — Programming Patterns

> Сервис: MCP (Model Context Protocol) сервер для AI-ассистентов.
> 6 tools: NATS, Postgres, Redis, RAG. stdio transport.

## Что делает сервис

MCP-сервер для Claude/Cursor: 6 инструментов (`nats_list_streams`, `nats_list_consumers`, `postgres_query`, `redis_get`, `redis_keys`, `search_prydwen`). RAG по Prydwen knowledge base через offline TF-IDF. Все соединения опциональны.

## Паттерны

### 1. MCP Server
- **Файл:** `cmd/main.go`
- **Код:** `server.NewMCPServer("event-horizon-mcp", "1.0.0", server.WithToolCapabilities(true))`

### 2. Tool Registration (6 tools)
- **Файл:** `cmd/main.go`
- **Код:** `s.AddTool(mcp.NewTool("nats_list_streams", mcp.WithDescription(...)), handler)`

### 3. Graceful Degradation (per-tool)
- **Файл:** `cmd/main.go`
- **Код:**
  ```go
  if natsClient == nil {
      return mcp.NewToolResultError("NATS not connected — set NATS_URL"), nil
  }
4. Optional Dependencies
Файл: cmd/main.go

Суть: все клиенты soft — сервер стартует, даже если всё упало.

5. SQL Sandbox
Файл: internal/pgxtool/client.go, ValidateSelect

Код:

go
forbidden := []string{" insert ", " update ", " delete ", " drop ", /* ... */ }
6. Offline RAG (TF-IDF без embedding API)
Файл: internal/rag/index.go

Комментарий: Offline TF-IDF RAG over Prydwen markdown — no embedding API required.

7. TF-IDF Scoring (cosine similarity)
Файл: internal/rag/index.go, Search

8. Chunking Strategy
Файл: internal/rag/index.go, splitChunks

Суть: markdown по параграфам, макс 1200 рун.

9. Index Caching
Файл: internal/rag/index.go, EnsureIndex

10. Path Resolution
Файл: cmd/main.go, resolvePrydwen

11. Adapter (natsx, pgxtool, redisx)
Файлы: internal/natsx/client.go, internal/pgxtool/client.go, internal/redisx/client.go

12. stdio Transport
Файл: cmd/main.go

Код: server.ServeStdio(s)

13. Config with Optional DSNs
Файл: internal/config/config.go

14. Testable SQL Validator
Файл: internal/pgxtool/client_test.go

15. Testable RAG
Файл: internal/rag/index_test.go
