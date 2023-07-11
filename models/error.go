package models

type ModelErrCode = string

var NewModelError = newModelError

const (
	ErrSystemError             ModelErrCode = "system_error"
	ErrUnknownDBError          ModelErrCode = "unknown_db_error"
	ErrWrongVerificationCode   ModelErrCode = "wrong_verification_code"
	ErrVerificationCodeExpired ModelErrCode = "expired_verification_code"
	ErrUnmatchedUserID         ModelErrCode = "unmatched_user_id"
	ErrUnmatchedEmail          ModelErrCode = "unmatched_email"
	ErrNotAnEmail              ModelErrCode = "not_an_email"
	ErrNoLink                  ModelErrCode = "no_link"
	ErrNoLinkOrResigned        ModelErrCode = "no_link_or_resigned"
	ErrNoLinkOrUnsigned        ModelErrCode = "no_link_or_unsigned"
	ErrUnsigned                ModelErrCode = "unsigned"
	ErrSamePassword            ModelErrCode = "same_password"
	ErrNoLinkOrNoManager       ModelErrCode = "no_link_or_no_manager"
	ErrNoLinkOrManagerExists   ModelErrCode = "no_link_or_manager_exists"
	ErrCorpManagerExists       ModelErrCode = "corp_manager_exists"
	ErrInvalidManagerID        ModelErrCode = "invalid_manager_id"
	ErrDuplicateManagerID      ModelErrCode = "duplicate_manager_id"
	ErrEmptyPayload            ModelErrCode = "empty_payload"
	ErrAdminAsManager          ModelErrCode = "admin_as_manager"
	ErrNotSameCorp             ModelErrCode = "not_same_corp"
	ErrManyEmployeeManagers    ModelErrCode = "many_employee_managers"
	ErrOrgEmailNotExists       ModelErrCode = "org_email_not_exists"
	ErrLinkExists              ModelErrCode = "link_exists"
	ErrUnsupportedCLALang      ModelErrCode = "unsupported_cla_lang"
	ErrNoCLAField              ModelErrCode = "no_cla_field"
	ErrManyCLAField            ModelErrCode = "many_cla_field"
	ErrCLAFieldID              ModelErrCode = "invalid_cla_field_id"
	ErrNoOrgSignature          ModelErrCode = "missing_org_signature"
	ErrMissgingCLA             ModelErrCode = "missing_cla"
	ErrMissgingEmail           ModelErrCode = "missing_email"
	ErrNoLinkOrCLAExists       ModelErrCode = "no_link_or_cla_exists"
	ErrNoLinkOrUnuploaed       ModelErrCode = "no_link_or_unuploaded"
	ErrUnmatchedEmailDomain    ModelErrCode = "unmatched_email_domain"
	ErrRestrictedEmailSuffix   ModelErrCode = "restricted_email_suffix"
	ErrInvalidPWRetrievalKey   ModelErrCode = "invalid_pw_retrieval_key"
	ErrInvalidPassword         ModelErrCode = "invalid_password"
	ErrBadRequestParameter     ModelErrCode = "bad_request_parameter"
	ErrNoCorpEmployeeManager   ModelErrCode = "no_employee_manager"
	ErrUnuploaed               ModelErrCode = "unuploaded"
	ErrGoToSignEmployeeCLA     ModelErrCode = "go_to_sign_employee_cla"
	ErrWrongIDOrPassword       ModelErrCode = "wrong_id_or_pw"
	ErrNoRefreshToken          ModelErrCode = "no_refresh_token"
	ErrInvalidToken            ModelErrCode = "invalid_token"
	ErrCLAExists               ModelErrCode = "cla_exists"
	ErrTooManyRequest          ModelErrCode = "too_many_request"
)

type IModelError interface {
	Error() string
	IsErrorOf(ModelErrCode) bool
	ErrCode() ModelErrCode
}

type modelError struct {
	code ModelErrCode
	err  error
}

func (this modelError) Error() string {
	if this.err == nil {
		return ""
	}
	return this.err.Error()
}

func (this modelError) IsErrorOf(code ModelErrCode) bool {
	return this.code == code
}

func (this modelError) ErrCode() ModelErrCode {
	return this.code
}

func newModelError(code ModelErrCode, err error) IModelError {
	return modelError{code: code, err: err}
}
