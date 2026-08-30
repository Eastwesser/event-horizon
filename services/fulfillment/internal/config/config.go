package config

import "os"

type Config struct {
	MetricsPort      string
	KafkaBrokers     string
	NATSURL          string
	AssembleDelaySec int
}

func Load() *Config {
	return &Config{
		MetricsPort:      getenv("METRICS_PORT", "9101"),
		KafkaBrokers:     getenv("KAFKA_BROKERS", ""),
		NATSURL:          getenv("NATS_URL", "nats://localhost:4222"),
		AssembleDelaySec: 10,
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
