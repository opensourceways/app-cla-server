package vcratelimitimpl

import "testing"

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()

	if cfg.Threshold != 10 {
		t.Errorf("Threshold: got %d, want 10", cfg.Threshold)
	}
	if cfg.Window != 60 {
		t.Errorf("Window: got %d, want 60", cfg.Window)
	}
}

func TestConfigSetDefaultPreservesExisting(t *testing.T) {
	cfg := &Config{Threshold: 20, Window: 120}
	cfg.SetDefault()

	if cfg.Threshold != 20 {
		t.Errorf("Threshold should be preserved: got %d, want 20", cfg.Threshold)
	}
	if cfg.Window != 120 {
		t.Errorf("Window should be preserved: got %d, want 120", cfg.Window)
	}
}

func TestConfigWindow(t *testing.T) {
	cfg := &Config{Window: 30}
	if d := cfg.window(); d.Seconds() != 30 {
		t.Errorf("window: got %f, want 30", d.Seconds())
	}
}
