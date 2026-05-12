package adapter

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func TestToModelError(t *testing.T) {
	err := toModelError(domain.NewDomainError(domain.ErrorCodeUserFrozen))
	if err == nil {
		t.Fatal("expected non-nil")
	}
	if !err.IsErrorOf(models.ErrUserLoginFrozen) {
		t.Errorf("expected ErrUserLoginFrozen, got '%s'", err.ErrCode())
	}
}

func TestToModelErrorUnknownCode(t *testing.T) {
	err := toModelError(errors.New("plain error"))
	if err == nil {
		t.Fatal("expected non-nil")
	}
	if !err.IsErrorOf(models.ErrSystemError) {
		t.Errorf("expected ErrSystemError, got '%s'", err.ErrCode())
	}
}

func TestToModelErrorDomainErrorWithMappedCode(t *testing.T) {
	tests := []struct {
		domainCode string
		modelCode  string
	}{
		{domain.ErrorCodeCorpAdminExists, models.ErrNoLinkOrManagerExists},
		{domain.ErrorCodeCorpSigningReSigning, models.ErrNoLinkOrResigned},
		{domain.ErrorCodeCorpSigningNotFound, models.ErrUnsigned},
		{domain.ErrorCodeCorpPDFNotFound, models.ErrUnuploaed},
		{domain.ErrorCodeUserInvalidPassword, models.ErrInvalidPassword},
		{domain.ErrorCodeUserNotExists, models.ErrUserNotExists},
		{domain.ErrorCodeCLAExists, models.ErrCLAExists},
		{domain.ErrorCodeCLACanNotRemove, models.ErrCLAIsUsed},
		{domain.ErrorCodeLinkNotExists, models.ErrNoLink},
		{domain.ErrorCodeLinkExists, models.ErrLinkExists},
		{domain.ErrorCodeNoPermission, models.ErrNoPermission},
		{domain.ErrorCodeCaptchaInvalid, models.ErrCaptchaInvalid},
	}

	for _, tt := range tests {
		derr := domain.NewDomainError(tt.domainCode)
		merr := toModelError(derr)
		if !merr.IsErrorOf(tt.modelCode) {
			t.Errorf("for %s: expected %s, got %s", tt.domainCode, tt.modelCode, merr.ErrCode())
		}
	}
}

func TestErrBadRequestParameter(t *testing.T) {
	derr := domain.NewDomainError(domain.ErrorCodeUserInvalidAccount)
	merr := errBadRequestParameter(derr)
	if merr == nil {
		t.Fatal("expected non-nil")
	}
	if !merr.IsErrorOf(models.ErrInvalidManagerID) {
		t.Errorf("expected ErrInvalidManagerID, got '%s'", merr.ErrCode())
	}
}

func TestErrBadRequestParameterPlainError(t *testing.T) {
	merr := errBadRequestParameter(errors.New("plain"))
	if !merr.IsErrorOf(models.ErrBadRequestParameter) {
		t.Errorf("expected ErrBadRequestParameter, got '%s'", merr.ErrCode())
	}
}

func TestCodeMap(t *testing.T) {
	if v := codeMap(domain.ErrorCodeUserFrozen); v != models.ErrUserLoginFrozen {
		t.Errorf("expected ErrUserLoginFrozen, got '%s'", v)
	}
	if v := codeMap("unknown_code"); v != models.ErrBadRequestParameter {
		t.Errorf("expected ErrBadRequestParameter for unknown, got '%s'", v)
	}
}

func TestAllErrorCodeMap(t *testing.T) {
	// Ensure all mapped domain codes have valid model codes
	for domainCode, modelCode := range errorCodeMap {
		if domainCode == "" || modelCode == "" {
			t.Errorf("empty key/value in errorCodeMap: %s -> %s", domainCode, modelCode)
		}
	}
}
