package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("METRICS_PORT", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("DB_HOST", "")
	cfg := Load()
	if cfg.MetricsPort != "9102" {
		t.Fatalf("MetricsPort=%q want 9102", cfg.MetricsPort)
	}
	if cfg.GRPCPort != "50063" {
		t.Fatalf("GRPCPort=%q want 50063", cfg.GRPCPort)
	}
	if cfg.TelegramToken != "" || cfg.TelegramChatID != "" {
		t.Fatalf("telegram defaults should be empty")
	}
}
