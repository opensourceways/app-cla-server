package app

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func init() {
	dp.Init(&dp.Config{MaxLengthOfEmail: 100})
}

func TestCmdToCreateVerificationCodeGenPurpose(t *testing.T) {
	e, _ := dp.NewEmailAddr("test@example.com")
	cmd := &CmdToCreateVerificationCode{
		Id:        "link-1",
		EmailAddr: e,
	}
	p, err := cmd.genPurpose("corp")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p == nil || p.Purpose() == "" {
		t.Error("expected non-empty purpose")
	}
}

func TestCmdToCreateCodeForCorpSigningPurpose(t *testing.T) {
	e, _ := dp.NewEmailAddr("corp@example.com")
	cmd := &cmdToCreateCodeForCorpSigning{
		Id:        "cs-1",
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

func TestCmdToCreateCodeForEmployeeSigningPurpose(t *testing.T) {
	e, _ := dp.NewEmailAddr("employee@example.com")
	cmd := &cmdToCreateCodeForEmployeeSigning{
		Id:        "emp-1",
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

func TestCmdToCreateCodeForIndividualSigningPurpose(t *testing.T) {
	e, _ := dp.NewEmailAddr("individual@example.com")
	cmd := &cmdToCreateCodeForIndividualSigning{
		Id:        "ind-1",
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
