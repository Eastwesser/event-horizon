package config

import (
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Port          string
	RabbitMQURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
	TelegramToken string
}

func Load() *Config {
	cfg := config.New()
	cfg.SetDefault("port", "8080")
	cfg.SetDefault("rabbitmq_url", "amqp://guest:guest@localhost:5872/")
	cfg.SetDefault("redis_addr", "localhost:6379")
	cfg.SetDefault("redis_db", 0)
	cfg.SetDefault("smtp_port", 587)
	
	cfg.EnableEnv("")
	cfg.LoadEnvFiles(".env")

	port := cfg.GetString("port")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:          port,
		RabbitMQURL:   cfg.GetString("rabbitmq_url"),
		RedisAddr:     cfg.GetString("redis_addr"),
		RedisPassword: cfg.GetString("redis_password"),
		RedisDB:       cfg.GetInt("redis_db"),
		SMTPHost:      cfg.GetString("smtp_host"),
		SMTPPort:      cfg.GetInt("smtp_port"),
		SMTPUser:      cfg.GetString("smtp_user"),
		SMTPPassword:  cfg.GetString("smtp_password"),
		SMTPFrom:      cfg.GetString("smtp_from"),
		TelegramToken: cfg.GetString("telegram_bot_token"),
	}
}
