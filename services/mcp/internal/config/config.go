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
