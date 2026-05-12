package adapter

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func init() {
	dp.Init(&dp.Config{
		MaxLengthOfEmail: 100,
	})
}

func TestCreateCodeForSigning(t *testing.T) {
	called := false
	f := func(cmd *app.CmdToCreateVerificationCode) (string, error) {
		called = true
		if cmd.Id != "test-link" {
			t.Errorf("expected 'test-link', got '%s'", cmd.Id)
		}
		if cmd.EmailAddr.EmailAddr() != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got '%s'", cmd.EmailAddr.EmailAddr())
		}
		return "123456", nil
	}

	code, err := createCodeForSigning("test-link", "test@example.com", f)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected f to be called")
	}
	if code != "123456" {
		t.Errorf("expected '123456', got '%s'", code)
	}
}

func TestCreateCodeForSigningInvalidEmail(t *testing.T) {
	_, err := createCodeForSigning("link", "not-an-email", nil)
	if err == nil {
		t.Error("expected error for invalid email")
	}
}

func TestCreateCodeForSigningWithError(t *testing.T) {
	f := func(cmd *app.CmdToCreateVerificationCode) (string, error) {
		return "", errors.New("service error")
	}
	_, err := createCodeForSigning("link", "test@example.com", f)
	if err == nil {
		t.Error("expected error")
	}
}
