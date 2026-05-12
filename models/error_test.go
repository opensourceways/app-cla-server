package models

import (
	"errors"
	"testing"
)

func TestNewModelError(t *testing.T) {
	err := newModelError(ErrSystemError, errors.New("test error"))
	if err == nil {
		t.Fatal("expected non-nil")
	}
	if !err.IsErrorOf(ErrSystemError) {
		t.Errorf("expected ErrSystemError, got '%s'", err.ErrCode())
	}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
}

func TestNewModelErrorNilErr(t *testing.T) {
	err := newModelError(ErrNoLink, nil)
	if err.Error() != "" {
		t.Errorf("expected empty string, got '%s'", err.Error())
	}
}

func TestModelErrorIsErrorOf(t *testing.T) {
	err := newModelError(ErrUnsigned, errors.New("unsigned"))
	if err.IsErrorOf(ErrNoLink) {
		t.Error("expected false for different code")
	}
	if !err.IsErrorOf(ErrUnsigned) {
		t.Error("expected true for matching code")
	}
}

func TestModelErrorErrCode(t *testing.T) {
	err := newModelError(ErrWrongVerificationCode, nil)
	if err.ErrCode() != ErrWrongVerificationCode {
		t.Errorf("expected '%s', got '%s'", ErrWrongVerificationCode, err.ErrCode())
	}
}

func TestIModelErrorInterface(t *testing.T) {
	var _ IModelError = NewModelError(ErrSystemError, errors.New("e"))
}

func TestCheckEmailFormat(t *testing.T) {
	if err := checkEmailFormat("test@example.com"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
	if err := checkEmailFormat("not-an-email"); err == nil {
		t.Error("expected invalid for bad email")
	}
	if err := checkEmailFormat(""); err == nil {
		t.Error("expected invalid for empty")
	}
}

func TestErrorCodeConstants(t *testing.T) {
	codes := []string{
		ErrSystemError, ErrUnknownDBError, ErrWrongVerificationCode,
		ErrVerificationCodeExpired, ErrUnmatchedUserID, ErrUnmatchedEmail,
		ErrNotAnEmail, ErrNoLink, ErrNoLinkOrResigned, ErrUnsigned,
		ErrSamePassword, ErrNoLinkOrNoManager, ErrCorpManagerExists,
		ErrInvalidManagerID, ErrDuplicateManagerID, ErrEmptyPayload,
		ErrAdminAsManager, ErrNotSameCorp, ErrManyEmployeeManagers,
		ErrOrgEmailNotExists, ErrLinkExists, ErrUnsupportedCLALang,
		ErrNoCLAField, ErrManyCLAField, ErrCLAFieldID, ErrNoOrgSignature,
		ErrMissgingCLA, ErrMissgingEmail, ErrNoLinkOrCLAExists,
		ErrNoLinkOrUnuploaed, ErrUnmatchedEmailDomain, ErrRestrictedEmailSuffix,
		ErrInvalidPWRetrievalKey, ErrInvalidPassword, ErrBadRequestParameter,
		ErrNoCorpEmployeeManager, ErrUnuploaed, ErrGoToSignEmployeeCLA,
		ErrWrongIDOrPassword, ErrPrivacyConsentInvalid, ErrNoRefreshToken,
		ErrInvalidToken, ErrCLAExists, ErrTooManyRequest, ErrUserLoginFrozen,
		ErrUserNotExists, ErrCaptchaInvalid, ErrCLAIsUsed, ErrLinkIsUsed,
		ErrNoPermission,
	}
	for _, c := range codes {
		if c == "" {
			t.Error("empty error code constant")
		}
	}
}
