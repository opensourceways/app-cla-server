package domain

import "strings"

const (
	ErrorCodeCorpSigningReSigning = "corp_signing_resigning"
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
