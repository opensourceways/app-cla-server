package controllers

import (
	"testing"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.CookieTimeout != 1800 {
		t.Errorf("expected 1800, got %d", cfg.CookieTimeout)
	}
	if cfg.WebRedirectDirOnSuccessForEmail != "/config-email" {
		t.Errorf("expected '/config-email', got '%s'", cfg.WebRedirectDirOnSuccessForEmail)
	}
	if cfg.MaxSizeOfCorpCLAPDF <= 0 {
		t.Error("expected positive default for max size")
	}
	if cfg.PasswordRetrievalExpiry != 3600 {
		t.Errorf("expected 3600, got %d", cfg.PasswordRetrievalExpiry)
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{
		CookieTimeout:                   3600,
		WebRedirectDirOnSuccessForEmail: "/custom",
		MaxSizeOfCorpCLAPDF:            10 << 20,
		PasswordRetrievalExpiry:         7200,
	}
	cfg.SetDefault()
	if cfg.CookieTimeout != 3600 {
		t.Errorf("expected 3600, got %d", cfg.CookieTimeout)
	}
	if cfg.WebRedirectDirOnSuccessForEmail != "/custom" {
		t.Errorf("expected '/custom', got '%s'", cfg.WebRedirectDirOnSuccessForEmail)
	}
	if cfg.PasswordRetrievalExpiry != 7200 {
		t.Errorf("expected 7200, got %d", cfg.PasswordRetrievalExpiry)
	}
}

func TestConfigSigningURL(t *testing.T) {
	cfg := &Config{CLAPlatformURL: "https://cla.example.com/"}
	u := cfg.signingURL("link-1")
	if u != "https://cla.example.com/link-1" {
		t.Errorf("expected 'https://cla.example.com/link-1', got '%s'", u)
	}
}

func TestConfigSigningURLNoTrailingSlash(t *testing.T) {
	cfg := &Config{CLAPlatformURL: "https://cla.example.com"}
	u := cfg.signingURL("link-1")
	if u != "https://cla.example.comlink-1" {
		t.Errorf("expected 'https://cla.example.comlink-1', got '%s'", u)
	}
}
