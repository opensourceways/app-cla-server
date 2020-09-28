package util

// All the error codes used by this app
const (
	ErrorStart = iota
	ErrInvalidParameter
	ErrHasSigned
	ErrHasNotSigned
	ErrMissingToken
	ErrUnknownToken
	ErrInvalidToken
	ErrSigningUncompleted
	ErrUnknownEmailPlatform
	ErrSendingEmail
	ErrWrongVerificationCode
	ErrVerificationCodeExpired
	ErrPDFHasNotUploaded
	ErrNumOfCorpManagersExceeded
	ErrCorpManagerHasAdded
)
