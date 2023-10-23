package domain

import "strings"

const (
	ErrorCodeUserFrozen                 = "user_frozen"
	ErrorCodeUserExists                 = "user_exists"
	ErrorCodeUserNotExists              = "user_not_exists"
	ErrorCodeUserSamePassword           = "user_same_password"
	ErrorCodeUserInvalidAccount         = "user_invalid_account"
	ErrorCodeUserInvalidPassword        = "user_invalid_password"
	ErrorCodeUserUnmatchedPassword      = "user_unmatched_password"
	ErrorCodeUserWrongAccountOrPassword = "user_wrong_account_or_password"

	ErrorCodeCorpAdminExists = "corp_admin_exists"

	ErrorCodeCorpPDFNotFound = "corp_pdf_not_found"

	ErrorCodeCorpSigningNotFound     = "corp_signing_not_found"
	ErrorCodeCorpSigningReSigning    = "corp_signing_resigning"
	ErrorCodeCorpSigningCanNotDelete = "corp_signing_can_not_delete"

	ErrorCodeCorpEmailDomainExists   = "corp_email_domain_exists"
	ErrorCodeCorpEmailDomainNotMatch = "corp_email_domain_not_match"

	ErrorCodeEmployeeManagerExists         = "employee_manager_exists"
	ErrorCodeEmployeeManagerTooMany        = "employee_manager_too_many"
	ErrorCodeEmployeeManagerNotExists      = "employee_manager_not_exists"
	ErrorCodeEmployeeManagerNotSameCorp    = "employee_manager_not_same_corp"
	ErrorCodeEmployeeManagerAdminAsManager = "employee_manager_admin_as_manager"

	ErrorCodeEmployeeSigningNotFound     = "employee_signing_not_found"
	ErrorCodeEmployeeSigningReSigning    = "employee_signing_resigning"
	ErrorCodeEmployeeSigningNoManager    = "employee_signing_no_manager"
	ErrorCodeEmployeeSigningEnableAgain  = "employee_signing_enable_again"
	ErrorCodeEmployeeSigningDisableAgain = "employee_signing_disable_again"
	ErrorCodeEmployeeSigningCanNotDelete = "employee_signing_can_not_delete"

	ErrorCodeIndividualSigningReSigning  = "individual_signing_resigning"
	ErrorCodeIndividualSigningCorpExists = "individual_signing_corp_exists"

	ErrorCodeVerificationCodeBusy  = "verification_code_busy"
	ErrorCodeVerificationCodeWrong = "verification_code_wrong"

	ErrorCodeEmailCredentialNotFound = "email_credential_not_found"

	ErrorCodeGmailNoRefreshToken = "gmail_no_refresh_token"

	ErrorCodeAccessTokenInvalid = "access_token_invalid"

	ErrorCodeCLAExists       = "cla_exists"
	ErrorCodeCLANotExists    = "cla_not_exists"
	ErrorCodeCLACanNotRemove = "cla_can_not_remove"

	ErrorCodeLinkExists       = "link_exists"
	ErrorCodeLinkNotExists    = "link_not_exists"
	ErrorCodeLinkCanNotRemove = "link_can_not_remove"
)

// domainError
type domainError string

func (e domainError) Error() string {
	return strings.ReplaceAll(string(e), "_", " ")
}

func (e domainError) ErrorCode() string {
	return string(e)
}

// notfoudError
type notfoudError struct {
	domainError
}

func (e notfoudError) NotFound() {}

// NewDomainError
func NewDomainError(v string) domainError {
	return domainError(v)
}

// NewNotFoundDomainError
func NewNotFoundDomainError(v string) notfoudError {
	return notfoudError{domainError(v)}
}
