package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("METRICS_PORT", "")
	t.Setenv("KAFKA_BROKERS", "")
	cfg := Load()
	if cfg.MetricsPort != "9101" {
		t.Fatalf("MetricsPort=%q want 9101", cfg.MetricsPort)
	}
	if cfg.KafkaBrokers != "" {
		t.Fatalf("KafkaBrokers=%q want empty", cfg.KafkaBrokers)
	}
	if cfg.AssembleDelaySec != 10 {
		t.Fatalf("AssembleDelaySec=%d want 10", cfg.AssembleDelaySec)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("METRICS_PORT", "9119")
	t.Setenv("KAFKA_BROKERS", "kafka:9092")
	cfg := Load()
	if cfg.MetricsPort != "9119" {
		t.Fatalf("MetricsPort=%q", cfg.MetricsPort)
	}
	if cfg.KafkaBrokers != "kafka:9092" {
		t.Fatalf("KafkaBrokers=%q", cfg.KafkaBrokers)
	}
}
