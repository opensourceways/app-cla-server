package oauth

import (
	"testing"
)

func TestCodePlatformAuthWebRedirectDir(t *testing.T) {
	auth := &codePlatformAuth{
		webRedirectDir: webRedirectDirConfig{
			WebRedirectDirOnSuccess: "/success",
			WebRedirectDirOnFailure: "/failure",
		},
	}
	if dir := auth.WebRedirectDir(true); dir != "/success" {
		t.Errorf("expected '/success', got '%s'", dir)
	}
	if dir := auth.WebRedirectDir(false); dir != "/failure" {
		t.Errorf("expected '/failure', got '%s'", dir)
	}
}

func TestCodePlatformAuthGetAuthInstance(t *testing.T) {
	auth := &codePlatformAuth{
		clients: map[string]AuthInterface{
			"github": &authClient{},
		},
	}
	c, err := auth.GetAuthInstance("github")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if c == nil {
		t.Error("expected non-nil client")
	}

	_, err = auth.GetAuthInstance("unknown")
	if err == nil {
		t.Error("expected error for unknown platform")
	}
}

func TestAuthConstants(t *testing.T) {
	if AuthApplyToLogin != "login" {
		t.Errorf("expected 'login', got '%s'", AuthApplyToLogin)
	}
}
