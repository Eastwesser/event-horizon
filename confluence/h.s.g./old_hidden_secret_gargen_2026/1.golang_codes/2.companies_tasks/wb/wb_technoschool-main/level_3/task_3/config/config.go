package config

import (
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Port string
}

func Load() *Config {
	cfg := config.New()
	cfg.SetDefault("port", "8080")
	
	cfg.EnableEnv("")
	cfg.LoadEnvFiles(".env")

	port := cfg.GetString("port")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port: port,
	}
}
