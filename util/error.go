package util

// All the error codes used by this app
const (
	ErrInvalidParameter          = "invalid_parameter"
	ErrHasSigned                 = "has_signed"
	ErrHasNotSigned              = "has_not_signed"
	ErrMissingToken              = "missing_token"
	ErrUnknownToken              = "unknown_token"
	ErrInvalidToken              = "invalid_token"
	ErrSigningUncompleted        = "uncompleted_signing"
	ErrUnknownEmailPlatform      = "unknown_email_platform"
	ErrSendingEmail              = "failed_to_send_email"
	ErrWrongVerificationCode     = "wrong_verification_code"
	ErrVerificationCodeExpired   = "expired_verification_code"
	ErrPDFHasNotUploaded         = "pdf_has_not_uploaded"
	ErrNumOfCorpManagersExceeded = "num_of_corp_managers_exceeded"
	ErrCorpManagerHasAdded       = "corp_manager_exists"
	ErrNoCLABindingDoc           = "no_cla_binding"
	ErrNotSameCorp               = "not_same_corp"
	ErrNoOrgEmail                = "no_org_email"
	ErrSystemError               = "system_error"
)
