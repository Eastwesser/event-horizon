//go:build !kafka

package kafka

import "log/slog"

func newSaramaProducer(cfg Config, topic string, log *slog.Logger) Producer {
	if log != nil {
		log.Warn("kafka: built without -tags kafka; using noop producer (go get github.com/IBM/sarama && go build -tags kafka)",
			"brokers", cfg.Brokers, "topic", topic)
	}
	return &NoopProducer{Topic: topic, Logger: log}
}

func newSaramaConsumer(cfg Config, group string, topics []string, log *slog.Logger) Consumer {
	_ = group
	if log != nil {
		log.Warn("kafka: built without -tags kafka; using noop consumer",
			"brokers", cfg.Brokers, "topics", topics)
	}
	return &NoopConsumer{Topics: topics, Logger: log}
}
