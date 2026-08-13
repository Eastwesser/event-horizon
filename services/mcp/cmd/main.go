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
