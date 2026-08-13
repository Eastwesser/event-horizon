package config_test

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
)

func TestConfigProviderViews(t *testing.T) {
	cfg := config.Load()
	var _ config.ConfigProvider = cfg

	if cfg.GRPCConfig().Port() == "" {
		t.Fatal("empty grpc port")
	}
	if cfg.PostgresConfig().DSN() == "" {
		t.Fatal("empty dsn")
	}
	if cfg.JWTConfig().Secret() == "" {
		t.Fatal("empty jwt")
	}
	if cfg.LoggerConfig().Level() == "" {
		t.Fatal("empty log level")
	}
}
