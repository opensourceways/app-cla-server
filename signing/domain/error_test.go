package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewDomainError(t *testing.T) {
	e := NewDomainError(ErrorCodeUserFrozen)
	if e.Error() != "user frozen" {
		t.Errorf("expected 'user frozen', got '%s'", e.Error())
	}
	if e.ErrorCode() != ErrorCodeUserFrozen {
		t.Errorf("expected '%s', got '%s'", ErrorCodeUserFrozen, e.ErrorCode())
	}
}

func TestNewNotFoundDomainError(t *testing.T) {
	e := NewNotFoundDomainError(ErrorCodeUserNotExists)
	if e.Error() != "user not exists" {
		t.Errorf("expected 'user not exists', got '%s'", e.Error())
	}
	if !e.NotFound() {
		t.Error("expected NotFound() to return true")
	}
	if e.ErrorCode() != ErrorCodeUserNotExists {
		t.Errorf("expected '%s', got '%s'", ErrorCodeUserNotExists, e.ErrorCode())
	}
}

func TestIsErrorOf(t *testing.T) {
	e := NewDomainError(ErrorCodeUserFrozen)
	if !IsErrorOf(e, ErrorCodeUserFrozen) {
		t.Error("expected IsErrorOf to return true")
	}
	if IsErrorOf(e, ErrorCodeUserExists) {
		t.Error("expected IsErrorOf to return false for different code")
	}
	if IsErrorOf(errors.New("plain error"), ErrorCodeUserFrozen) {
		t.Error("expected IsErrorOf to return false for plain error")
	}
}

func TestDomainErrorFormatsUnderscores(t *testing.T) {
	e := NewDomainError("user_wrong_account_or_password")
	if e.Error() != "user wrong account or password" {
		t.Errorf("expected underscores replaced, got '%s'", e.Error())
	}
}

func TestDomainErrorErrorInterface(t *testing.T) {
	var _ error = NewDomainError("test")
	var _ error = NewNotFoundDomainError("test")
}

func TestDomainErrorAllCodes(t *testing.T) {
	codes := []string{
		ErrorCodeUserFrozen,
		ErrorCodeUserExists,
		ErrorCodeUserNotExists,
		ErrorCodeUserSamePassword,
		ErrorPrivacyConsentInvalid,
		ErrorCodeUserInvalidAccount,
		ErrorCodeUserInvalidPassword,
		ErrorCodeUserUnmatchedPassword,
		ErrorCodeUserWrongAccountOrPassword,
		ErrorCodeCorpAdminExists,
		ErrorCodeCorpPDFNotFound,
		ErrorCodeCorpSigningNotFound,
		ErrorCodeCorpSigningReSigning,
		ErrorCodeCorpSigningCanNotDelete,
		ErrorCodeCorpSigningCLAIsLatest,
		ErrorCodeCorpEmailDomainExists,
		ErrorCodeCorpEmailDomainNotMatch,
		ErrorCodeEmployeeManagerExists,
		ErrorCodeEmployeeManagerTooMany,
		ErrorCodeEmployeeManagerNotExists,
		ErrorCodeEmployeeManagerNotSameCorp,
		ErrorCodeEmployeeManagerAdminAsManager,
		ErrorCodeEmployeeNotSameCorp,
		ErrorCodeEmployeeSigningNotFound,
		ErrorCodeEmployeeSigningReSigning,
		ErrorCodeEmployeeSigningNoManager,
		ErrorCodeEmployeeSigningEnableAgain,
		ErrorCodeEmployeeSigningDisableAgain,
		ErrorCodeEmployeeSigningCanNotDelete,
		ErrorCodeIndividualSigningReSigning,
		ErrorCodeIndividualSigningCorpExists,
		ErrorCodeIndividualSigningCLAIsLatest,
		ErrorCodeVerificationCodeBusy,
		ErrorCodeVerificationCodeWrong,
		ErrorCodeEmailCredentialNotFound,
		ErrorCodeGmailNoRefreshToken,
		ErrorCodeAccessTokenInvalid,
		ErrorCodeCaptchaInvalid,
		ErrorCodeCLAExists,
		ErrorCodeCLANotExists,
		ErrorCodeCLACanNotRemove,
		ErrorCodeLinkExists,
		ErrorCodeLinkNotExists,
		ErrorCodeLinkCanNotRemove,
		ErrorCodeNoPermission,
		ErrorCodeLinkCanNotMigrate,
	}
	for _, code := range codes {
		e := NewDomainError(code)
		if e.ErrorCode() != code {
			t.Errorf("ErrorCode mismatch for %s", code)
		}
		_ = fmt.Sprint(e)
	}
}
