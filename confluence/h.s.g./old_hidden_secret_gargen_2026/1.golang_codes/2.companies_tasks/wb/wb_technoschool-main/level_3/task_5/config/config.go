package config

import (
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Port              string
	DatabaseURL       string
	DefaultTimeoutMin int
}

func Load() *Config {
	cfg := config.New()

	// Load .env file if exists
	_ = cfg.LoadEnvFiles(".env")

	// Enable environment variables
	cfg.EnableEnv("")

	cfg.SetDefault("port", "8080")
	cfg.SetDefault("database_url", "postgres://postgres:8832@localhost:5555/eventbooker?sslmode=disable")
	cfg.SetDefault("default_timeout_min", "15")

	return &Config{
		Port:              cfg.GetString("port"),
		DatabaseURL:       cfg.GetString("database_url"),
		DefaultTimeoutMin: cfg.GetInt("default_timeout_min"),
	}
}
