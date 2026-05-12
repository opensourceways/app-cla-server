package watch

import (
	"testing"
	"time"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Interval != 3600 {
		t.Errorf("expected 3600, got %d", cfg.Interval)
	}
	if cfg.intervalDuration() != 3600*time.Second {
		t.Errorf("expected 1h, got %v", cfg.intervalDuration())
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Interval: 7200}
	cfg.SetDefault()
	if cfg.Interval != 7200 {
		t.Errorf("expected 7200, got %d", cfg.Interval)
	}
}

func TestCLAUpdateConfigSetDefault(t *testing.T) {
	cfg := &CLAUpdateConfig{}
	cfg.SetDefault()
	if cfg.MsgChannelSize != 100 {
		t.Errorf("expected 100, got %d", cfg.MsgChannelSize)
	}
	if cfg.PythonRetryTimes != 3 {
		t.Errorf("expected 3, got %d", cfg.PythonRetryTimes)
	}
	if cfg.GenAllDiffInterval != 600 {
		t.Errorf("expected 600, got %d", cfg.GenAllDiffInterval)
	}
	if cfg.genAllDiffInterval() != 600*time.Second {
		t.Errorf("expected 10m, got %v", cfg.genAllDiffInterval())
	}
}

func TestNotifyAdminConfigSetDefault(t *testing.T) {
	cfg := &NotifyAdminConfig{}
	cfg.SetDefault()
	if cfg.SendEmailInterval != 10 {
		t.Errorf("expected 10, got %d", cfg.SendEmailInterval)
	}
	if cfg.NotifyCorpAdminInterval != 1200 {
		t.Errorf("expected 1200, got %d", cfg.NotifyCorpAdminInterval)
	}
}

func TestNotifyAdminConfigIntervals(t *testing.T) {
	cfg := &NotifyAdminConfig{SendEmailInterval: 20, NotifyCorpAdminInterval: 2400}
	if cfg.genSendEmailInterval() != 20*time.Second {
		t.Errorf("expected 20s, got %v", cfg.genSendEmailInterval())
	}
	if cfg.genNotifyCorpAdminInterval() != 2400*time.Second {
		t.Errorf("expected 2400s, got %v", cfg.genNotifyCorpAdminInterval())
	}
}

func TestConfigConfigItems(t *testing.T) {
	cfg := &Config{}
	items := cfg.ConfigItems()
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}
