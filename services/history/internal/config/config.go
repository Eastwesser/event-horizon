package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GRPCPort      string
	MetricsPort   string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	NATSURL       string
	RetentionDays int
	LogLevel      string
	LogFormat     string
}

func Load() *Config {
	days, _ := strconv.Atoi(getenv("RETENTION_DAYS", "30"))
	if days <= 0 {
		days = 30
	}
	return &Config{
		GRPCPort:      getenv("GRPC_PORT", "50062"),
		MetricsPort:   getenv("METRICS_PORT", "9105"),
		DBHost:        getenv("DB_HOST", "localhost"),
		DBPort:        getenv("DB_PORT", "5469"),
		DBUser:        getenv("DB_USER", "eventhorizon"),
		DBPass:        getenv("DB_PASSWORD", "eventhorizon"),
		DBName:        getenv("DB_NAME", "eventhorizon_history"),
		NATSURL:       getenv("NATS_URL", "nats://localhost:4222"),
		RetentionDays: days,
		LogLevel:      getenv("LOG_LEVEL", "info"),
		LogFormat:     getenv("LOG_FORMAT", "text"),
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
