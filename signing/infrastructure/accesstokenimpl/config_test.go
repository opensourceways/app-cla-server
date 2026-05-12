package accesstokenimpl

import (
	"testing"
	"time"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Expire != 60 {
		t.Errorf("expected 60, got %d", cfg.Expire)
	}
	if cfg.expire() != 60*time.Second {
		t.Errorf("expected 60s, got %v", cfg.expire())
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Expire: 120}
	cfg.SetDefault()
	if cfg.Expire != 120 {
		t.Errorf("expected 120, got %d", cfg.Expire)
	}
}
