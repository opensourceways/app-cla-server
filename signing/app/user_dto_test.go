package app

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCmdToChangePasswordValidate(t *testing.T) {
	old, _ := dp.NewPassword([]byte("old"))
	newPw, _ := dp.NewPassword([]byte("new"))
	cmd := &CmdToChangePassword{Id: "user-1", OldOne: old, NewOne: newPw}
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCmdToChangePasswordValidateSame(t *testing.T) {
	same, _ := dp.NewPassword([]byte("same"))
	cmd := &CmdToChangePassword{OldOne: same, NewOne: same}
	if err := cmd.Validate(); err == nil {
		t.Error("expected error for same password")
	}
}

func TestCmdToLoginClear(t *testing.T) {
	pw, _ := dp.NewPassword([]byte("secret"))
	cmd := &CmdToLogin{Password: pw}
	cmd.clear()
	// After clear, password bytes should be zeroed
	if string(pw.Password()) == "secret" {
		t.Error("expected password to be cleared")
	}
}

func TestCmdToChangePasswordClear(t *testing.T) {
	old, _ := dp.NewPassword([]byte("old"))
	newPw, _ := dp.NewPassword([]byte("new"))
	cmd := &CmdToChangePassword{OldOne: old, NewOne: newPw}
	cmd.clear()
	if string(old.Password()) == "old" || string(newPw.Password()) == "new" {
		t.Error("expected passwords to be cleared")
	}
}

func TestCmdToResetPasswordClear(t *testing.T) {
	newPw, _ := dp.NewPassword([]byte("new"))
	cmd := &CmdToResetPassword{NewOne: newPw}
	cmd.clear()
	if string(newPw.Password()) == "new" {
		t.Error("expected password to be cleared")
	}
}

func TestCmdToGenKeyForPasswordRetrievalPurpose(t *testing.T) {
	e, _ := dp.NewEmailAddr("user@test.com")
	cmd := &CmdToGenKeyForPasswordRetrieval{
		Id:        "link-1",
		EmailAddr: e,
	}
	p, err := cmd.purpose()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p == nil || p.Purpose() == "" {
		t.Error("expected non-empty purpose")
	}
}
