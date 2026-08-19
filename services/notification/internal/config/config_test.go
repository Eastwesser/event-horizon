package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("METRICS_PORT", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("HTTP_ADDR", "")
	cfg := Load()
	if cfg.MetricsPort != "9102" {
		t.Fatalf("MetricsPort=%q want 9102", cfg.MetricsPort)
	}
	if cfg.HTTPAddr != ":8088" {
		t.Fatalf("HTTPAddr=%q want :8088", cfg.HTTPAddr)
	}
	if cfg.TelegramToken != "" || cfg.TelegramChatID != "" {
		t.Fatalf("telegram defaults should be empty")
	}
}
