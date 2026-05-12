package redisdb

import (
	"testing"
	"time"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Timeout != 3 {
		t.Errorf("expected 3, got %d", cfg.Timeout)
	}
	if cfg.timeout() != 3*time.Second {
		t.Errorf("expected 3s, got %v", cfg.timeout())
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Timeout: 10}
	cfg.SetDefault()
	if cfg.Timeout != 10 {
		t.Errorf("expected 10, got %d", cfg.Timeout)
	}
}
