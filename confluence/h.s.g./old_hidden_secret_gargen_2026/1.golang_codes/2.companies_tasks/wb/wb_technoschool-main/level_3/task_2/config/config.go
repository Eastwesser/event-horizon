package config

import (
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	cfg := config.New()
	cfg.SetDefault("port", "8080")
	cfg.SetDefault("redis_addr", "localhost:6379")
	cfg.SetDefault("redis_db", 0)
	
	cfg.EnableEnv("")
	cfg.LoadEnvFiles(".env")

	port := cfg.GetString("port")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:          port,
		RedisAddr:     cfg.GetString("redis_addr"),
		RedisPassword: cfg.GetString("redis_password"),
		RedisDB:       cfg.GetInt("redis_db"),
	}
}
