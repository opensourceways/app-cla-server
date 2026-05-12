package mongodb

import (
	"testing"
	"time"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Timeout != 10 {
		t.Errorf("expected 10, got %d", cfg.Timeout)
	}
	if cfg.timeout() != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.timeout())
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Timeout: 30}
	cfg.SetDefault()
	if cfg.Timeout != 30 {
		t.Errorf("expected 30, got %d", cfg.Timeout)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{Conn: "mongodb://host?ssl=true"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	cfg2 := &Config{Conn: "mongodb://host"}
	if err := cfg2.Validate(); err == nil {
		t.Error("expected error for non-ssl conn")
	}
}
