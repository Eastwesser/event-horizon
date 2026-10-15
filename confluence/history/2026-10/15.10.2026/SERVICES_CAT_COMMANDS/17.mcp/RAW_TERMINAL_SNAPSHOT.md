[denismatveev@c0der event_horizon]$ find services/mcp -type f -name '*.go' | sort
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
services/mcp/cmd/main.go
services/mcp/internal/config/config.go
services/mcp/internal/natsx/client.go
services/mcp/internal/pgxtool/client.go
services/mcp/internal/pgxtool/client_test.go
services/mcp/internal/rag/index.go
services/mcp/internal/rag/index_test.go
services/mcp/internal/redisx/client.go
===== main =====
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/Eastwesser/event-horizon/services/mcp/internal/config"
	"github.com/Eastwesser/event-horizon/services/mcp/internal/natsx"
	"github.com/Eastwesser/event-horizon/services/mcp/internal/pgxtool"
	"github.com/Eastwesser/event-horizon/services/mcp/internal/rag"
	"github.com/Eastwesser/event-horizon/services/mcp/internal/redisx"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	prydwen := resolvePrydwen(cfg.PrydwenRoot)
	idx, err := rag.EnsureIndex(prydwen, cfg.RAGIndexPath)
	if err != nil {
		log.Printf("rag index warn: %v (search_prydwen will fail until fixed)", err)
	} else {
		log.Printf("rag index ready: %d chunks from %s", idx.N, prydwen)
	}

	var natsClient *natsx.Client
	if nc, err := natsx.Connect(cfg.NATSURL); err != nil {
		log.Printf("nats warn: %v", err)
	} else {
		natsClient = nc
		defer natsClient.Close()
	}

	var pg *pgxtool.Client
	if cfg.PostgresDSN != "" {
		if c, err := pgxtool.Connect(ctx, cfg.PostgresDSN); err != nil {
			log.Printf("postgres warn: %v", err)
		} else {
			pg = c
			defer pg.Close()
		}
	}

	var rdb *redisx.Client
	if c, err := redisx.Connect(cfg.RedisAddr); err != nil {
		log.Printf("redis warn: %v", err)
	} else {
		rdb = c
		defer rdb.Close()
	}

	s := server.NewMCPServer(
		"event-horizon-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("nats_list_streams",
		mcp.WithDescription("List JetStream streams (active NATS messaging topology for Event Horizon)"),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if natsClient == nil {
			return mcp.NewToolResultError("NATS not connected — set NATS_URL"), nil
		}
		streams, err := natsClient.ListStreams(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(natsx.JSON(streams)), nil
	})

	s.AddTool(mcp.NewTool("nats_list_consumers",
		mcp.WithDescription("List JetStream consumers (durable subscriptions) for a stream"),
		mcp.WithString("stream", mcp.Required(), mcp.Description("JetStream stream name")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if natsClient == nil {
			return mcp.NewToolResultError("NATS not connected — set NATS_URL"), nil
		}
		stream, err := req.RequireString("stream")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		cons, err := natsClient.ListConsumers(ctx, stream)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(natsx.JSON(cons)), nil
	})

	s.AddTool(mcp.NewTool("postgres_query",
		mcp.WithDescription("Run a read-only SELECT/WITH query against Event Horizon Postgres (MCP_POSTGRES_DSN). Mutations forbidden."),
		mcp.WithString("sql", mcp.Required(), mcp.Description("SELECT statement without semicolon")),
		mcp.WithNumber("limit", mcp.Description("Max rows (default 100, max 500)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if pg == nil {
			return mcp.NewToolResultError("Postgres not connected — set MCP_POSTGRES_DSN"), nil
		}
		sql, err := req.RequireString("sql")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := 100
		if args := req.GetArguments(); args != nil {
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}
		}
		out, err := pg.Query(ctx, sql, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	})

	s.AddTool(mcp.NewTool("redis_get",
		mcp.WithDescription("GET a Redis key (cache/session inspection)"),
		mcp.WithString("key", mcp.Required(), mcp.Description("Redis key")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if rdb == nil {
			return mcp.NewToolResultError("Redis not connected — set REDIS_ADDR"), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		val, err := rdb.Get(ctx, key)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(val), nil
	})

	s.AddTool(mcp.NewTool("redis_keys",
		mcp.WithDescription("SCAN Redis keys by pattern (capped)"),
		mcp.WithString("pattern", mcp.Description("Glob pattern, default *")),
		mcp.WithNumber("limit", mcp.Description("Max keys, default 50")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if rdb == nil {
			return mcp.NewToolResultError("Redis not connected — set REDIS_ADDR"), nil
		}
		pattern := "*"
		limit := int64(50)
		if args := req.GetArguments(); args != nil {
			if v, ok := args["pattern"].(string); ok && v != "" {
				pattern = v
			}
			if v, ok := args["limit"].(float64); ok {
				limit = int64(v)
			}
		}
		keys, err := rdb.Keys(ctx, pattern, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.MarshalIndent(keys, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	s.AddTool(mcp.NewTool("search_prydwen",
		mcp.WithDescription("RAG search over Event Horizon Prydwen knowledge base (Outbox, NATS, ports, ClickHouse, etc.)"),
		mcp.WithString("query", mcp.Required(), mcp.Description(`Natural language question, e.g. "how does Outbox work in Shop"`)),
		mcp.WithNumber("k", mcp.Description("Top-k chunks, default 5")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if idx == nil || idx.N == 0 {
			return mcp.NewToolResultError("RAG index empty — set PRYDWEN_ROOT to confluence/agents/prydwen_knowledge"), nil
		}
		q, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		k := 5
		if args := req.GetArguments(); args != nil {
			if v, ok := args["k"].(float64); ok && v > 0 {
				k = int(v)
			}
		}
		hits := idx.Search(q, k)
		b, _ := json.MarshalIndent(hits, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

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
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return configured
}
===== config =====
--- services/mcp/internal/config/config.go ---
package config

import "os"

// Config for the Event Horizon MCP server (stdio). Connections are optional —
// tools degrade with a clear error if the matching env is empty / unreachable.
type Config struct {
	NATSURL       string
	PostgresDSN   string // prefer a read-only role
	RedisAddr     string
	PrydwenRoot   string // path to confluence/agents/prydwen_knowledge
	RAGIndexPath  string // cached TF-IDF index JSON (optional)
}

func Load() Config {
	root := getenv("PRYDWEN_ROOT", "")
	if root == "" {
		// default: walk up from CWD is handled by caller; keep empty here
		root = "confluence/agents/prydwen_knowledge"
	}
	return Config{
		NATSURL:      getenv("NATS_URL", "nats://localhost:4222"),
		PostgresDSN:  getenv("MCP_POSTGRES_DSN", getenv("DATABASE_URL", "")),
		RedisAddr:    getenv("REDIS_ADDR", "localhost:6379"),
		PrydwenRoot:  root,
		RAGIndexPath: getenv("RAG_INDEX_PATH", ""),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
===== natsx =====
--- services/mcp/internal/natsx/client.go ---
package natsx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func Connect(url string) (*Client, error) {
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second), nats.Name("event-horizon-mcp"))
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Client{nc: nc, js: js}, nil
}

func (c *Client) Close() {
	if c != nil && c.nc != nil {
		_ = c.nc.Drain()
	}
}

type StreamInfo struct {
	Name     string   `json:"name"`
	Subjects []string `json:"subjects"`
	Messages uint64   `json:"messages"`
	Bytes    uint64   `json:"bytes"`
	Consumer int      `json:"consumer_count"`
}

func (c *Client) ListStreams(_ context.Context) ([]StreamInfo, error) {
	if c == nil || c.js == nil {
		return nil, fmt.Errorf("nats not connected")
	}
	var out []StreamInfo
	for name := range c.js.StreamNames() {
		si, err := c.js.StreamInfo(name)
		if err != nil {
			continue
		}
		out = append(out, StreamInfo{
			Name:     si.Config.Name,
			Subjects: si.Config.Subjects,
			Messages: si.State.Msgs,
			Bytes:    si.State.Bytes,
			Consumer: si.State.Consumers,
		})
	}
	return out, nil
}

type ConsumerInfo struct {
	Stream   string `json:"stream"`
	Name     string `json:"name"`
	Durable  string `json:"durable"`
	Filter   string `json:"filter_subject"`
	Pending  int    `json:"num_pending"`
	AckFloor uint64 `json:"ack_floor_seq"`
}

func (c *Client) ListConsumers(_ context.Context, stream string) ([]ConsumerInfo, error) {
	if c == nil || c.js == nil {
		return nil, fmt.Errorf("nats not connected")
	}
	var out []ConsumerInfo
	for name := range c.js.ConsumerNames(stream) {
		ci, err := c.js.ConsumerInfo(stream, name)
		if err != nil {
			continue
		}
		out = append(out, ConsumerInfo{
			Stream:   stream,
			Name:     ci.Name,
			Durable:  ci.Config.Durable,
			Filter:   ci.Config.FilterSubject,
			Pending:  int(ci.NumPending),
			AckFloor: ci.AckFloor.Stream,
		})
	}
	return out, nil
}

func JSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
===== pgxtool =====
--- services/mcp/internal/pgxtool/client.go ---
package pgxtool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("MCP_POSTGRES_DSN / DATABASE_URL not set")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Client{pool: pool}, nil
}

func (c *Client) Close() {
	if c != nil && c.pool != nil {
		c.pool.Close()
	}
}

// ValidateSelect rejects anything that is not a single SELECT/WITH query.
func ValidateSelect(sql string) error {
	s := strings.TrimSpace(sql)
	if s == "" {
		return fmt.Errorf("empty query")
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("multiple statements / trailing semicolon not allowed")
	}
	lower := strings.ToLower(s)
	// strip leading comments
	for strings.HasPrefix(lower, "--") || strings.HasPrefix(lower, "/*") {
		return fmt.Errorf("comments not allowed in MCP queries")
	}
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

func (c *Client) Query(ctx context.Context, sql string, limit int) (string, error) {
	if c == nil || c.pool == nil {
		return "", fmt.Errorf("postgres not connected")
	}
	if err := ValidateSelect(sql); err != nil {
		return "", err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	// limit is an int this process clamped. The inner text already passed ValidateSelect.
	wrapped := fmt.Sprintf("SELECT * FROM (%s) AS _mcp_q LIMIT %d", sql, limit)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := c.pool.Query(ctx, wrapped)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = string(f.Name)
	}

	var out strings.Builder
	out.WriteString(strings.Join(cols, "\t"))
	out.WriteByte('\n')
	n := 0
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return "", err
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = fmt.Sprint(v)
		}
		out.WriteString(strings.Join(parts, "\t"))
		out.WriteByte('\n')
		n++
	}
	out.WriteString(fmt.Sprintf("\n(%d rows)\n", n))
	return out.String(), rows.Err()
}
--- services/mcp/internal/pgxtool/client_test.go ---
package pgxtool

import "testing"

func TestValidateSelect_AllowsSelect(t *testing.T) {
	if err := ValidateSelect("SELECT id FROM users LIMIT 10"); err != nil {
		t.Fatal(err)
	}
}

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
===== rag =====
--- services/mcp/internal/rag/index.go ---
package rag

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Offline TF-IDF RAG over Prydwen markdown — no embedding API required.
// Good enough for "how does Outbox work in Shop?" style retrieval.

type Chunk struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type Index struct {
	Chunks []Chunk             `json:"chunks"`
	DF     map[string]int      `json:"df"`
	TF     []map[string]float64 `json:"tf"` // per-chunk term frequencies
	N      int                 `json:"n"`
}

type Hit struct {
	Score float64 `json:"score"`
	Path  string  `json:"path"`
	Title string  `json:"title"`
	Text  string  `json:"text"`
}

var wordRe = regexp.MustCompile(`[[:alnum:]_]+`)

func tokenize(s string) []string {
	s = strings.ToLower(s)
	raw := wordRe.FindAllString(s, -1)
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		if len(w) < 2 {
			continue
		}
		out = append(out, w)
	}
	return out
}

func BuildFromDir(root string) (*Index, error) {
	var chunks []Chunk
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		parts := splitChunks(string(b), 1200)
		for i, p := range parts {
			title := firstHeading(p)
			if title == "" {
				title = rel
			}
			chunks = append(chunks, Chunk{
				ID:    rel + "#" + itoa(i),
				Path:  rel,
				Title: title,
				Text:  strings.TrimSpace(p),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	idx := &Index{Chunks: chunks, DF: map[string]int{}, N: len(chunks)}
	idx.TF = make([]map[string]float64, len(chunks))
	for i, c := range chunks {
		tf := map[string]float64{}
		toks := tokenize(c.Title + " " + c.Text)
		for _, t := range toks {
			tf[t]++
		}
		seen := map[string]struct{}{}
		for t, n := range tf {
			tf[t] = n / float64(len(toks)+1)
			if _, ok := seen[t]; !ok {
				idx.DF[t]++
				seen[t] = struct{}{}
			}
		}
		idx.TF[i] = tf
	}
	return idx, nil
}

func (idx *Index) Search(query string, k int) []Hit {
	if idx == nil || idx.N == 0 {
		return nil
	}
	if k <= 0 {
		k = 5
	}
	qtf := map[string]float64{}
	qtoks := tokenize(query)
	for _, t := range qtoks {
		qtf[t]++
	}
	for t := range qtf {
		qtf[t] = qtf[t] / float64(len(qtoks)+1)
	}

	type scored struct {
		i int
		s float64
	}
	var scores []scored
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
		// also accumulate doc norm leftover lightly
		for t, dv := range idx.TF[i] {
			if _, ok := qtf[t]; ok {
				continue
			}
			idf := math.Log(1 + float64(idx.N)/(1+float64(idx.DF[t])))
			dw := dv * idf
			nd += dw * dw
		}
		if nq == 0 || nd == 0 {
			continue
		}
		cos := dot / (math.Sqrt(nq) * math.Sqrt(nd))
		if cos > 0 {
			scores = append(scores, scored{i: i, s: cos})
		}
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].s > scores[b].s })
	if len(scores) > k {
		scores = scores[:k]
	}
	hits := make([]Hit, 0, len(scores))
	for _, sc := range scores {
		c := idx.Chunks[sc.i]
		text := c.Text
		if len(text) > 800 {
			text = text[:800] + "…"
		}
		hits = append(hits, Hit{Score: sc.s, Path: c.Path, Title: c.Title, Text: text})
	}
	return hits
}

func (idx *Index) Save(path string) error {
	b, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Load(path string) (*Index, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func splitChunks(s string, maxRunes int) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	paras := strings.Split(s, "\n\n")
	var (
		out  []string
		cur  strings.Builder
		size int
	)
	flush := func() {
		t := strings.TrimSpace(cur.String())
		if t != "" {
			out = append(out, t)
		}
		cur.Reset()
		size = 0
	}
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		r := utf8Len(p)
		if size > 0 && size+r > maxRunes {
			flush()
		}
		if size > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(p)
		size += r
		if size >= maxRunes {
			flush()
		}
	}
	flush()
	if len(out) == 0 && strings.TrimSpace(s) != "" {
		return []string{strings.TrimSpace(s)}
	}
	return out
}

func firstHeading(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	// first non-empty line
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return truncate(line, 80)
		}
	}
	return ""
}

func utf8Len(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

// EnsureIndex loads from cache or builds from root.
func EnsureIndex(root, cachePath string) (*Index, error) {
	if cachePath != "" {
		if idx, err := Load(cachePath); err == nil && idx.N > 0 {
			return idx, nil
		}
	}
	idx, err := BuildFromDir(root)
	if err != nil {
		return nil, err
	}
	if cachePath != "" {
		_ = idx.Save(cachePath)
	}
	return idx, nil
}
--- services/mcp/internal/rag/index_test.go ---
package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchFindsOutbox(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "4.architecture_patterns")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `# Integration patterns — Event Horizon

## Outbox (transactional messaging)

Shop writes purchase and outbox in one Postgres transaction, then a worker publishes shop.purchased to NATS.
`
	if err := os.WriteFile(filepath.Join(sub, "03_ARCH_INTEGRATION_PATTERNS.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	idx, err := BuildFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	hits := idx.Search("how does Outbox work in Shop", 3)
	if len(hits) == 0 {
		t.Fatal("expected at least one hit")
	}
	if hits[0].Path == "" || hits[0].Score <= 0 {
		t.Fatalf("bad hit: %+v", hits[0])
	}
}
===== redisx =====
--- services/mcp/internal/redisx/client.go ---
package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func Connect(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("REDIS_ADDR not set")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DialTimeout: 3 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() {
	if c != nil && c.rdb != nil {
		_ = c.rdb.Close()
	}
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", fmt.Errorf("redis not connected")
	}
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Keys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	if c == nil || c.rdb == nil {
		return nil, fmt.Errorf("redis not connected")
	}
	if pattern == "" {
		pattern = "*"
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var (
		cursor uint64
		out    []string
	)
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, limit).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		cursor = next
		if cursor == 0 || int64(len(out)) >= limit {
			break
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}
[denismatveev@c0der event_horizon]$