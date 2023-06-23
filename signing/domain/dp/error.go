package dp

import "strings"

// domainError
type domainError string

func (e domainError) Error() string {
	return strings.ReplaceAll(string(e), "_", " ")
}

func (e domainError) Code() string {
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
