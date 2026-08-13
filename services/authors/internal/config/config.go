package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort    string
	MetricsPort string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	RedisAddr   string
	NATSURL     string
	LogLevel    string
	LogFormat   string
}

func Load() *Config {
	return &Config{
		GRPCPort:    getenv("GRPC_PORT", "50061"),
		MetricsPort: getenv("METRICS_PORT", "9104"),
		DBHost:      getenv("DB_HOST", "localhost"),
		DBPort:      getenv("DB_PORT", "5468"),
		DBUser:      getenv("DB_USER", "eventhorizon"),
		DBPass:      getenv("DB_PASSWORD", "eventhorizon"),
		DBName:      getenv("DB_NAME", "eventhorizon_authors"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6387"),
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
		LogFormat:   getenv("LOG_FORMAT", "text"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
