package controllers

import (
	"testing"
)

func TestAccessControllerGetUserOwner(t *testing.T) {
	ac := &accessController{
		Permission: PermissionOwnerOfOrg,
		Payload:    &acForCorpManagerPayload{UserId: "user-1", LinkID: "link-1"},
	}
	u := ac.getUser()
	if u != "user-1" {
		t.Errorf("expected 'user-1', got '%s'", u)
	}
}

func TestAccessControllerGetUserOther(t *testing.T) {
	ac := &accessController{
		Permission: PermissionCorpAdmin,
		Payload:    &acForCorpManagerPayload{UserId: "user-2", LinkID: "link-2"},
	}
	u := ac.getUser()
	if u != "link-2/user-2" {
		t.Errorf("expected 'link-2/user-2', got '%s'", u)
	}
}

func TestAccessControllerGetUserWrongPayload(t *testing.T) {
	ac := &accessController{
		Payload: "wrong type",
	}
	u := ac.getUser()
	if u != "" {
		t.Errorf("expected '', got '%s'", u)
	}
}

func TestAccessControllerVerify(t *testing.T) {
	ac := &accessController{
		RemoteAddr: "192.168.1.1",
		Permission: PermissionCorpAdmin,
	}
	if err := ac.verify([]string{PermissionCorpAdmin}, "192.168.1.1"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ac.verify([]string{PermissionCorpAdmin}, "10.0.0.1"); err == nil {
		t.Error("expected error for wrong addr")
	}
	if err := ac.verify([]string{PermissionOwnerOfOrg}, "192.168.1.1"); err == nil {
		t.Error("expected error for wrong permission")
	}
}

func TestNewAccessController(t *testing.T) {
	tests := []struct {
		perm string
	}{
		{PermissionOwnerOfOrg},
		{PermissionCorpAdmin},
		{PermissionEmployeeManager},
		{"unknown"},
	}
	for _, tt := range tests {
		ac := (&baseController{}).newAccessController(tt.perm)
		if tt.perm == "unknown" {
			if ac.Payload != nil {
				t.Errorf("expected nil payload for unknown permission")
			}
		} else {
			if ac.Payload == nil {
				t.Errorf("expected non-nil payload for %s", tt.perm)
			}
		}
	}
}

func TestPermissionConstants(t *testing.T) {
	if PermissionCorpAdmin == "" || PermissionOwnerOfOrg == "" || PermissionEmployeeManager == "" {
		t.Error("expected non-empty permission constants")
	}
}

func TestAccessControllerCheckPrivacyConsentNoCheck(t *testing.T) {
	ac := &accessController{
		Payload: "no-check",
	}
	if err := ac.checkPrivacyConsent(); err != nil {
		t.Errorf("expected nil for non-privacy-consent payload, got: %v", err)
	}
}
