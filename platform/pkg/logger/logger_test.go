package logger

import (
	"log/slog"
	"testing"
)

func TestNewJSONAndNop(t *testing.T) {
	log := New(staticConfig{level: "debug", format: "json"})
	if log == nil {
		t.Fatal("nil logger")
	}
	log.Info("ok")
	if Nop() == nil {
		t.Fatal("nop nil")
	}
	_ = slog.LevelInfo
}

func TestParseLevel(t *testing.T) {
	if parseLevel("error") != slog.LevelError {
		t.Fatal("error level")
	}
}
