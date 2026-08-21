package captchaimpl

import (
	"testing"
	"time"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Expiry != 300 {
		t.Errorf("expected 300, got %d", cfg.Expiry)
	}
	if cfg.Height != 80 {
		t.Errorf("expected 80, got %d", cfg.Height)
	}
	if cfg.Width != 240 {
		t.Errorf("expected 240, got %d", cfg.Width)
	}
	if cfg.Length != 6 {
		t.Errorf("expected 6, got %d", cfg.Length)
	}
	if cfg.NoiseLevel != 0.7 {
		t.Errorf("expected 0.7, got %f", cfg.NoiseLevel)
	}
	if cfg.BackgroundCircles != 80 {
		t.Errorf("expected 80, got %d", cfg.BackgroundCircles)
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Expiry: 600, Height: 100, Width: 300, Length: 4, NoiseLevel: 0.5, BackgroundCircles: 50}
	cfg.SetDefault()
	if cfg.Expiry != 600 {
		t.Errorf("expected 600, got %d", cfg.Expiry)
	}
	if cfg.Height != 100 {
		t.Errorf("expected 100, got %d", cfg.Height)
	}
}

func TestConfigExpiry(t *testing.T) {
	cfg := &Config{Expiry: 120}
	if cfg.expiry() != 120*time.Second {
		t.Errorf("expected 120s, got %v", cfg.expiry())
	}
}
