package config

import "os"

type Config struct {
	MetricsPort    string
	KafkaBrokers   string
	NATSURL        string
	TelegramToken  string
	TelegramChatID string
	HTTPAddr       string
}

func Load() *Config {
	return &Config{
		MetricsPort:    getenv("METRICS_PORT", "9102"),
		KafkaBrokers:   getenv("KAFKA_BROKERS", ""),
		NATSURL:        getenv("NATS_URL", "nats://localhost:4222"),
		TelegramToken:  getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID: getenv("TELEGRAM_CHAT_ID", ""),
		HTTPAddr:       getenv("HTTP_ADDR", ":8088"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
