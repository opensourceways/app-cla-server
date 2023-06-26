package domain

import "strings"

const (
	ErrorCodeUserExists = "user_exists"

	ErrorCodeCorpAdminExists = "corp_admin_exists"

	ErrorCodeCorpPDFNotFound = "corp_pdf_not_found"

	ErrorCodeCorpSigningNotFound  = "corp_signing_not_found"
	ErrorCodeCorpSigningReSigning = "corp_signing_resigning"

	ErrorCodeEmployeeSigningReSigning    = "employee_signing_resigning"
	ErrorCodeEmployeeSigningEnableAgain  = "employee_signing_enable_again"
	ErrorCodeEmployeeSigningDisableAgain = "employee_signing_disable_again"
	ErrorCodeEmployeeSigningCanNotDelete = "employee_signing_can_not_delete"
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
