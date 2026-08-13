package config

import "os"

type Config struct {
	GRPCPort      string
	MetricsPort   string
	ClickHouseURL string // HTTP base, e.g. http://clickhouse:8123
	ClickHouseDB  string
	NATSURL       string
	LogLevel      string
	LogFormat     string
}

func Load() *Config {
	return &Config{
		GRPCPort:      getenv("GRPC_PORT", "50057"),
		MetricsPort:   getenv("METRICS_PORT", "9106"),
		ClickHouseURL: getenv("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDB:  getenv("CLICKHOUSE_DB", "eventhorizon"),
		NATSURL:       getenv("NATS_URL", "nats://localhost:4222"),
		LogLevel:      getenv("LOG_LEVEL", "info"),
		LogFormat:     getenv("LOG_FORMAT", "text"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
