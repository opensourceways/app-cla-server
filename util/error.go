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
	ErrNotReadyToSign            = "not_ready_to_sign"
	ErrNotSupportedPlatform      = "not_supported_platform"
	ErrNotYoursOrg               = "not_yours_org"
	ErrInvalidAccountOrPw        = "invalid_account_or_pw"
	ErrNoDBRecord                = "no_db_record"
	ErrRecordExists              = "db_record_exists"
	ErrNoPlatformOrOrg           = "no_platform_or_org"
	ErrInvalidEmail              = "invalid_email"
	ErrInvalidManagerID          = "invalid_manager_id"
	ErrNoCorpManager             = "no_corp_manager"
	ErrAuthFailed                = "auth_failed"
	ErrUnauthorized              = "unauthorized"
	ErrSystemError               = "system_error"
	ErrCLAIsUsed                 = "cla_is_used"
)

type IAppError interface {
	Error() string
	ErrCode() string
}

type AppError struct {
	Code string
	Err  error
}

func (this AppError) Error() string {
	return this.Err.Error()
}

func (this AppError) ErrCode() string {
	return this.Code
}

func (this AppError) IsErrorOf(code string) bool {
	return this.Code == code
}
