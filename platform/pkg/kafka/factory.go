package kafka

import (
	"log/slog"
	"os"
	"strings"
)

// Config from env.
type Config struct {
	Brokers []string
	Enabled bool
}

// LoadConfig reads KAFKA_BROKERS (comma-separated). Empty → disabled (noop).
func LoadConfig() Config {
	raw := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if raw == "" {
		return Config{Enabled: false}
	}
	parts := strings.Split(raw, ",")
	var brokers []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			brokers = append(brokers, p)
		}
	}
	return Config{Brokers: brokers, Enabled: len(brokers) > 0}
}

// NewProducer returns a producer for topic. Without -tags kafka (or no brokers) → NoopProducer.
func NewProducer(cfg Config, topic string, log *slog.Logger) Producer {
	if !cfg.Enabled {
		return &NoopProducer{Topic: topic, Logger: log}
	}
	return newSaramaProducer(cfg, topic, log)
}

// NewConsumer returns a consumer for topics/group. Without -tags kafka → NoopConsumer.
func NewConsumer(cfg Config, group string, topics []string, log *slog.Logger) Consumer {
	if !cfg.Enabled {
		return &NoopConsumer{Topics: topics, Logger: log}
	}
	return newSaramaConsumer(cfg, group, topics, log)
}
